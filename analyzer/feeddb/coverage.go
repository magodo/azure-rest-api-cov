package feeddb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-multierror"
	"github.com/magodo/azure-rest-api-cov/analyzer/coverage"
	"github.com/magodo/azure-rest-api-cov/analyzer/schemainput"
	"github.com/magodo/azure-rest-api-cov/analyzer/swagger"
	"github.com/magodo/azure-rest-api-cov/analyzer/utils"
	"github.com/magodo/azure-rest-api-cov/analyzer/workspace"
	log "github.com/sirupsen/logrus"
)

type Coverage struct {
	Ref               string            `json:"ref"`
	Comment           string            `json:"comment"`
	ResourceProviders ResourceProviders `json:"resourceProviders"`
}

type ResourceProviders map[string]ResourceProvider

type ResourceProvider struct {
	Name  string   `json:"name"`
	Paths APIPaths `json:"paths"`
}

type APIPaths map[string]APIPath

type APIPath struct {
	Name       string     `json:"name"`
	Operations Operations `json:"operations"`
}

type Operations map[string]Operation

type Operation struct {
	Name     string      `json:"name"`
	Versions APIVersions `json:"versions"`
}

type APIVersions map[string]APIVersion

type APIVersion struct {
	Version   string    `json:"version"`
	IsActive  bool      `json:"isActive"`
	IsPreview bool      `json:"isPreview"`
	Model     *APIModel `json:"model,omitempty"`
}

func NewEmptyCoverage(index *workspace.AzureRepoIndex, comment string) Coverage {
	coverage := Coverage{
		Ref:               index.Commit,
		Comment:           comment,
		ResourceProviders: ResourceProviders{},
	}

	for rp, versions := range index.ResourceProviders {
		rpName := string(rp)
		paths := APIPaths{}
		coverage.ResourceProviders[rpName] = ResourceProvider{
			Name:  string(rp),
			Paths: paths,
		}
		for version, specs := range versions {
			versionName := string(version)
			for _, apiPaths := range specs {
				for pathName, ops := range apiPaths {
					if _, ok := paths[pathName]; !ok {
						paths[pathName] = APIPath{
							Name:       pathName,
							Operations: Operations{},
						}
					}

					for op := range ops {
						opName := string(op)
						if _, ok := paths[pathName].Operations[opName]; !ok {
							paths[pathName].Operations[opName] = Operation{
								Name:     opName,
								Versions: APIVersions{},
							}
						}

						paths[pathName].Operations[opName].Versions[versionName] = APIVersion{
							Version: versionName,
						}
					}
				}
			}
		}
	}

	return coverage
}

func (cov *Coverage) AddSchema(s swagger.Schema, locator workspace.SchemaLocator) error {
	p := s.SwaggerRelPath
	pathInfo, err := p.PathInfo()
	if err != nil {
		return err
	}

	rpName := pathInfo.ResourceProvider
	rp, ok := cov.ResourceProviders[rpName]
	if !ok {
		return fmt.Errorf("no resource provider %q found", rp)
	}
	return rp.AddSchema(s, locator)
}

func (rp *ResourceProvider) AddSchema(s swagger.Schema, locator workspace.SchemaLocator) error {
	p := s.SwaggerRelPath
	pathInfo, err := p.PathInfo()
	if err != nil {
		return err
	}

	path, ok := rp.Paths[locator.ApiPath]
	if !ok {
		return fmt.Errorf("no API path %q found for resource provider %q", locator.ApiPath, rp)
	}

	op, ok := path.Operations[string(locator.Operation)]
	if !ok {
		return fmt.Errorf("no operation %q found for API path %q, resource provider %q", locator.Operation, locator.ApiPath, rp)
	}

	version, ok := op.Versions[locator.ApiVersion]
	if !ok {
		return fmt.Errorf("no version %q found for operation %q, API path %q, resource provider %q", locator.ApiVersion, locator.Operation, locator.ApiPath, rp)
	}

	model, err := NewAPIModel(s)
	if err != nil {
		return err
	}

	rp.Paths[locator.ApiPath].Operations[string(locator.Operation)].Versions[locator.ApiVersion] = APIVersion{
		Version:   version.Version,
		IsActive:  s.Active,
		IsPreview: pathInfo.IsPreview,
		Model:     model,
	}
	return nil
}

func (rp ResourceProvider) Copy() ResourceProvider {
	return ResourceProvider{
		Name:  rp.Name,
		Paths: rp.Paths,
	}
}

// GenerateWithUncoveredOperationModels walks through the coverage data and gather the schemas for those uncovered operations. For each RP, merge both covered and uncovered schemas and generate in json format.
// This means for operations that none of the available versions is covered, then all of those versions of this operation will be converted into schema inputs.
// This can be some API path has covered some operations but not the others. This can also happen when some API paths are not covered at all.
func (cov *Coverage) GenerateWithUncoveredOperationModels(idx *workspace.AzureRepoIndex, ignoreSchema []string, outdir string) error {
	log.Info("Generating with uncovered operation models...")
	type input struct {
		loc workspace.SchemaLocator
		idx *workspace.AzureRepoIndex
	}
	type output struct {
		loc    workspace.SchemaLocator
		schema *swagger.Schema
	}

	var ignorePatterns []workspace.SchemaLocator
	for _, p := range ignoreSchema {
		loc, err := workspace.NewSchemaLocatorFromPattern(p)
		if err != nil {
			return err
		}
		ignorePatterns = append(ignorePatterns, loc)
	}

	for rpName, rp := range cov.ResourceProviders {
		log.Debugf("Generating for %q...", rpName)

		// Make a shallow copy of the RP cov so that any mutation only applies to this copy, rather than the original one inside the "cov".
		// This is for saving memory usage.
		rp := rp.Copy()

		// Setup the work pool
		size := 0
		for _, path := range rp.Paths {
			for _, op := range path.Operations {
				size += len(op.Versions)
			}
		}
		wp := utils.NewWorkPool(size)
		wp.Run()

		for pathName, path := range rp.Paths {
			for opName, op := range path.Operations {
				// Iterate all the versions to see whether there is an active one, if so it means this operation is covered already, we don't need to expand the uncovered ones.
				// Since in this case we only need to expand the newer API versions, which has been done in the "coverage" stage (with "-predict").
				var hasActive bool
				for _, version := range op.Versions {
					if version.IsActive {
						hasActive = true
						break
					}
				}
				if hasActive {
					continue
				}

				// There is no version covered for this operation, we will need to expand the model of this operation.
			VersionLoop:
				for version := range op.Versions {
					loc := workspace.SchemaLocator{
						SpecLocator: workspace.SpecLocator{
							ApiVersion: version,
							ApiPath:    pathName,
						},
						Operation: workspace.OperationKind(opName),
					}

					for _, igloc := range ignorePatterns {
						if igloc.Match(loc) {
							log.Infof("schema %q is ignored", loc)
							continue VersionLoop
						}
					}

					// Construct the worker task and send it into the task queue.
					task := utils.NewTask(
						context.TODO(),
						input{loc, idx},
						func(_ context.Context, i interface{}) (interface{}, error) {
							in := i.(input)
							log.Debugf("Expanding uncovered schema for %q", in.loc)
							schema, err := expandUncoveredSchema(in.loc, in.idx)
							return output{loc: in.loc, schema: schema}, err
						},
					)
					wp.AddTask(task)
				}
			}
		}

		var result error
		uncoveredSchemas := coverage.FlatCoverage{}

		// Iterate the task results and gather all the expanded uncoverd schemas
		for res := range wp.Done() {
			if res.Error != nil {
				result = multierror.Append(result, res.Error)
				continue
			}
			o := res.Value.(output)
			if o.schema == nil {
				continue
			}
			uncoveredSchemas[o.loc] = *o.schema
		}

		// Apply the uncovered schemas to the rp cov, which contained the covered schemas' coverage data.
		for k, schema := range uncoveredSchemas {
			if err := rp.AddSchema(schema, k); err != nil {
				return err
			}
		}

		// Encode and write out the RP coverage into its own file
		outputFile := filepath.Join(outdir, rpName+".json")
		rpCov, err := json.Marshal(rp)
		if err != nil {
			return fmt.Errorf("json encoding for %q: %v", rpName, err)
		}

		log.Debugf("Writing coverage data for %q: %q", rpName, outputFile)
		if err := os.WriteFile(outputFile, rpCov, 0644); err != nil {
			return fmt.Errorf("writing to %q: %v", outputFile, err)
		}
	}

	return nil
}

// expandUncoveredSchema expands uncovered schema to its end. In case the schema specified is a schemaless one, it returns nil.
func expandUncoveredSchema(loc workspace.SchemaLocator, idx *workspace.AzureRepoIndex) (*swagger.Schema, error) {
	log.Infof("Processing (uncovered) %q", loc)
	schemaInfo, err := idx.LocateSchema(loc)
	if err != nil {
		if errors.Is(err, workspace.ErrFailedToLocateSwagger) || errors.Is(err, workspace.ErrFailedToLocateSchema) {
			log.Warn(err)
			return nil, nil
		}
		return nil, err
	}

	// Skip schemaless operations
	if schemaInfo.Name == "" {
		return nil, nil
	}

	// Convert the schema into schema input.
	// In most case we can simpy initialize a schema input here, however, we should take care about the discriminator schemas (i.e., the schema itself is a discriminator)
	schemaRef := swagger.SchemaNameToRef(filepath.Join(idx.SpecRootDir, string(schemaInfo.SpecRelPath)), schemaInfo.Name)
	variantRefSet, err := swagger.GetVariantRefSet(schemaRef)
	if err != nil {
		return nil, fmt.Errorf("getting variant ref set: %v", err)
	}

	// The schema is not a discriminator, we can simply convert it to schema input
	if len(variantRefSet) == 0 {
		si := schemainput.SchemaInput{
			SchemaLocator: loc,
			Properties:    []string{},
		}
		schema, err := si.ToSchema(idx, true, false)
		if err != nil {
			return nil, fmt.Errorf("converting schema input to schema for %q: %v", loc, err)
		}
		return schema, nil
	}

	// The schema is a discriminator, we will need to expand all its variants and merge them into one schema.
	var schema *swagger.Schema
	for ref := range variantRefSet {
		si := schemainput.SchemaInput{
			SchemaLocator: loc,
			Variant:       ref,
			Properties:    []string{},
		}
		oschema, err := si.ToSchema(idx, true, false)
		if err != nil {
			return nil, fmt.Errorf("converting schema input to schema for %q: %v", loc, err)
		}
		if schema == nil {
			schema = oschema
		} else {
			if err := schema.Merge(*oschema); err != nil {
				return nil, fmt.Errorf("mergeing schema for %q: %v", loc, err)
			}
		}
	}
	return schema, nil
}
