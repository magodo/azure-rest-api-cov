package feeddb

import (
	"errors"
	"fmt"
	"github.com/magodo/azure-rest-api-cov/analyzer/coverage"
	"github.com/magodo/azure-rest-api-cov/analyzer/schemainput"
	"github.com/magodo/azure-rest-api-cov/analyzer/swagger"
	"github.com/magodo/azure-rest-api-cov/analyzer/workspace"
	log "github.com/sirupsen/logrus"
	"path/filepath"
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

	rp := pathInfo.ResourceProvider
	paths, ok := cov.ResourceProviders[rp]
	if !ok {
		return fmt.Errorf("no resource provider %q found", rp)
	}

	path, ok := paths.Paths[locator.ApiPath]
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

	cov.ResourceProviders[rp].Paths[locator.ApiPath].Operations[string(locator.Operation)].Versions[locator.ApiVersion] = APIVersion{
		Version:   version.Version,
		IsActive:  s.Active,
		IsPreview: pathInfo.IsPreview,
		Model:     model,
	}
	return nil
}

func (cov *Coverage) IncludeUncoveredOperationModels(idx *workspace.AzureRepoIndex) error {
	log.Info("Gathering uncovered operation models...")
	uncoveredSchemas, err := cov.gatherUncoveredOperationModels(idx)
	if err != nil {
		return err
	}

	for k, schema := range uncoveredSchemas {
		log.Infof("Processing (uncovered): %s", k)
		if err := cov.AddSchema(schema, k); err != nil {
			return err
		}
	}
	return nil
}

// gatherUncoveredOperationModels walk through the coverage data and gather the schemas for those uncovered operations.
// This means for operations that none of the available versions is covered, then all of those versions of this operation will be converted into schema inputs.
// This can be some API path has covered some operations but not the others. This can also happen when some API paths are not covered at all.
func (cov *Coverage) gatherUncoveredOperationModels(idx *workspace.AzureRepoIndex) (coverage.FlatCoverage, error) {
	out := coverage.FlatCoverage{}
	for rpName, rp := range cov.ResourceProviders {
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
				for version := range op.Versions {
					log.Infof("gathering for %s - %s - %s - %s", rpName, pathName, opName, version)
					loc := workspace.SchemaLocator{
						SpecLocator: workspace.SpecLocator{
							ApiVersion: version,
							ApiPath:    pathName,
						},
						Operation: workspace.OperationKind(opName),
					}
					schemaInfo, err := idx.LocateSchema(loc)
					if err != nil {
						if errors.Is(err, workspace.ErrFailedToLocateSwagger) || errors.Is(err, workspace.ErrFailedToLocateSchema) {
							log.Warn(err)
							continue
						}
						return nil, err
					}

					// Skip schemaless operations
					if schemaInfo.Name == "" {
						continue
					}

					// Convert the schema into schema input.
					// In most case we can simpy initialize a schema input here, however, we should take care about the discriminator schemas (i.e., the schema itself is a discriminator)
					schemaRef := swagger.SchemaNameToRef(filepath.Join(idx.SpecRootDir, string(schemaInfo.SpecRelPath)), schemaInfo.Name)
					variantRefSet, err := swagger.GetVariantRefSet(schemaRef)
					if err != nil {
						return nil, fmt.Errorf("getting variant ref set: %v", err)
					}
					var schema *swagger.Schema
					if len(variantRefSet) == 0 {
						// The schema is not a discriminator, we can simply convert it to schema input
						si := schemainput.SchemaInput{
							SchemaLocator: loc,
							Properties:    []string{},
						}
						schema, err = si.ToSchema(idx, true, false)
						if err != nil {
							return nil, fmt.Errorf("converting schema input to schema for %q: %v", loc, err)
						}
					} else {
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
					}

					out[loc] = *schema
				}
			}
		}
	}
	return out, nil
}
