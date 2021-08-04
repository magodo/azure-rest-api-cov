package feeddb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/magodo/azure-rest-api-cov/analyzer/operation"
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

func (cov *Coverage) AddOperation(opInfo *operation.Info) error {
	rpName := opInfo.RP
	rp, ok := cov.ResourceProviders[rpName]
	if !ok {
		return fmt.Errorf("no resource provider %q found", rpName)
	}
	return rp.AddOperation(opInfo)
}

func (rp *ResourceProvider) AddOperation(opInfo *operation.Info) error {
	path, ok := rp.Paths[opInfo.APIPath]
	if !ok {
		return fmt.Errorf("no API path %q found for resource provider %q", opInfo.APIPath, rp.Name)
	}

	op, ok := path.Operations[string(opInfo.Operation)]
	if !ok {
		return fmt.Errorf("no operation %q found for API path %q, resource provider %q", opInfo.Operation, opInfo.APIPath, rp.Name)
	}

	version, ok := op.Versions[opInfo.APIVersion]
	if !ok {
		return fmt.Errorf("no version %q found for operation %q, API path %q, resource provider %q", opInfo.APIVersion, opInfo.Operation, opInfo.APIPath, rp.Name)
	}

	v := APIVersion{
		Version:   version.Version,
		IsActive:  opInfo.IsActive,
		IsPreview: opInfo.IsPreview,
	}

	if opInfo.Model != nil {
		model, err := NewAPIModel(*opInfo.Model)
		if err != nil {
			return err
		}
		v.Model = model
	}
	rp.Paths[opInfo.APIPath].Operations[string(opInfo.Operation)].Versions[opInfo.APIVersion] = v

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
		loc workspace.OperationLocator
		idx *workspace.AzureRepoIndex
	}
	type output struct {
		loc    workspace.OperationLocator
		opInfo *operation.Info
	}

	var ignorePatterns []workspace.OperationLocator
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
					loc := workspace.OperationLocator{
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
							opInfo, err := getUncoveredOperation(in.loc, in.idx)
							return output{loc, opInfo}, err
						},
					)
					wp.AddTask(task)
				}
			}
		}

		var result error
		uncoveredOps := coverage.FlatCoverage{}

		// Iterate the task results and gather all the expanded uncoverd schemas
		for res := range wp.Done() {
			if res.Error != nil {
				result = multierror.Append(result, res.Error)
				continue
			}
			o := res.Value.(output)
			uncoveredOps[o.loc] = o.opInfo
		}

		// Apply the uncovered schemas to the rp cov, which contained the covered schemas' coverage data.
		for _, v := range uncoveredOps {
			if err := rp.AddOperation(v); err != nil {
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

// getUncoveredOperation search the index for an uncovered operation. If that operation has a model, further expands it to its end.
func getUncoveredOperation(loc workspace.OperationLocator, idx *workspace.AzureRepoIndex) (*operation.Info, error) {
	log.Infof("Processing (uncovered) %q", loc)
	opInfo, err := idx.LocateOperation(loc)
	if err != nil {
		if errors.Is(err, workspace.ErrFailedToLocateSwagger) || errors.Is(err, workspace.ErrFailedToLocateSchema) {
			log.Warn(err)
			return nil, nil
		}
		return nil, err
	}

	// Schemaless operations
	if opInfo.Model == nil {
		si := schemainput.SchemaInput{
			OperationLocator: loc,
			Properties:       []string{},
		}
		out, err := si.ToOperationInfo(idx, true, false)
		if err != nil {
			return nil, fmt.Errorf("converting schema input to opInfo for %q: %v", loc, err)
		}
		return out, nil
	}

	// Operations with schemas
	// In most case we can simpy initialize a schema input here, however, we should take care about the discriminator schemas (i.e., the schema itself is a discriminator)
	schemaRef := swagger.SchemaNameToRef(filepath.Join(idx.SpecRootDir, opInfo.Model.SpecRelPath), opInfo.Model.Name)
	variantRefSet, err := swagger.GetVariantRefSet(schemaRef)
	if err != nil {
		return nil, fmt.Errorf("getting variant ref set: %v", err)
	}

	// The schema is not a discriminator, we can simply convert it to schema input and process on
	if len(variantRefSet) == 0 {
		si := schemainput.SchemaInput{
			OperationLocator: loc,
			Properties:       []string{},
		}
		out, err := si.ToOperationInfo(idx, true, false)
		if err != nil {
			return nil, fmt.Errorf("converting schema input to opInfo for %q: %v", loc, err)
		}
		return out, nil
	}

	// The schema is a discriminator, we will need to expand all its variants and merge them into one schema.
	var out *operation.Info
	for ref := range variantRefSet {
		si := schemainput.SchemaInput{
			OperationLocator: loc,
			Variant:          ref,
			Properties:       []string{},
		}
		info, err := si.ToOperationInfo(idx, true, false)
		if err != nil {
			return nil, fmt.Errorf("converting schema input to opInfo for %q: %v", loc, err)
		}
		if out == nil {
			out = info
		} else {
			if err := out.Model.Merge(*info.Model); err != nil {
				return nil, fmt.Errorf("merging schema for %q: %v", loc, err)
			}
		}
	}
	return out, nil
}
