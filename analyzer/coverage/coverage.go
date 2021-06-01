package coverage

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/magodo/azure-rest-api-cov/analyzer/addr"
	"github.com/magodo/azure-rest-api-cov/analyzer/schemainput"
	"github.com/magodo/azure-rest-api-cov/analyzer/swagger"
	"github.com/magodo/azure-rest-api-cov/analyzer/swaggerloader"
	"github.com/magodo/azure-rest-api-cov/analyzer/workspace"
	log "github.com/sirupsen/logrus"
)

type Coverage ResourceProviders

type ResourceProviders map[string]*APIPaths

type APIPaths map[string]*APIPath

type APIPath struct {
	GrantComment string     `json:"grant_comment,omitempty"`
	Operations   Operations `json:"operations,omitempty"`
}

type Operations map[string]APIVersions

type APIVersions map[string]swagger.Schema

func NewEmptyCoverage() Coverage {
	return Coverage{}
}

func (cov *Coverage) AddSchema(schemaLocator workspace.SchemaLocator, schema swagger.Schema) error {
	p := schema.SwaggerRelPath
	pathInfo, err := p.PathInfo()
	if err != nil {
		return err
	}
	rp := pathInfo.ResourceProvider
	apiPath, op, apiVersion := schemaLocator.ApiPath, string(schemaLocator.Operation), schemaLocator.ApiVersion

	apiPaths, ok := (*cov)[rp]
	if !ok {
		apiPaths = &APIPaths{}
		(*cov)[rp] = apiPaths
	}

	path, ok := (*apiPaths)[apiPath]
	if !ok {
		path = &APIPath{
			Operations: Operations{},
		}
		(*apiPaths)[apiPath] = path
	}

	apiVersions, ok := path.Operations[op]
	if !ok {
		apiVersions = APIVersions{}
		path.Operations[op] = apiVersions
	}

	existSchema, ok := apiVersions[apiVersion]
	if !ok {
		apiVersions[apiVersion] = schema
		return nil
	}
	return existSchema.Merge(schema)
}

// ToSchemaInputs converts the coverage back to an array of SchemaInputs.
// Note that the order in the SchemaInputs are not guaranteed. Also, the granted api path are not converted.
func (cov Coverage) ToSchemaInputs() []schemainput.SchemaInput {
	var schemaInputs []schemainput.SchemaInput

	for _, apiPaths := range cov {
		for apiPath, path := range *apiPaths {
			for apiOperation, apiVersions := range path.Operations {
				for apiVersion, schema := range apiVersions {
					schemaLocator := workspace.SchemaLocator{
						SpecLocator: workspace.SpecLocator{
							ApiVersion: apiVersion,
							ApiPath:    apiPath,
						},
						Operation: workspace.OperationKind(apiOperation),
					}

					if len(schema.Properties) == 0 {
						schemaInputs = append(schemaInputs, schemainput.SchemaInput{
							SchemaLocator: schemaLocator,
						})
						continue
					}

					// If the schema has top level variant, then it will be transforms to numerous schemaInputs, each corresponds to a variant.
					// We peek any property to see whether it is a top level variant, since when a schema contains top level properties, all the property addresses
					// start with a variant.
					var isTopLevelDiscriminator bool
					for paddr := range schema.Properties {
						prop := addr.MustParseProperty(paddr)
						if prop[0].DiscriminatorValue != "" && prop[0].Name == "" {
							isTopLevelDiscriminator = true
						}
						break
					}

					if !isTopLevelDiscriminator {
						schemaInput := schemainput.SchemaInput{
							SchemaLocator: schemaLocator,
						}
						for paddr, prop := range schema.Properties {
							if prop.IsCovered() {
								schemaInput.Properties = append(schemaInput.Properties, paddr)
							}
						}
						schemaInputs = append(schemaInputs, schemaInput)
						continue
					}

					variantProperties := map[string][]string{}
					for paddr := range schema.Properties {
						prop := addr.MustParseProperty(paddr)
						if len(prop) > 1 {
							variantProps := variantProperties[prop[0].DiscriminatorValue]
							if variantProps == nil {
								variantProps = []string{}
							}
							variantProps = append(variantProps, prop[1:].String())
							variantProperties[prop[0].DiscriminatorValue] = variantProps
						}
					}

					for variant, variantProps := range variantProperties {
						schemaInputs = append(schemaInputs, schemainput.SchemaInput{
							SchemaLocator: schemaLocator,
							Variant:       variant,
							Properties:    variantProps,
						})
					}
				}
			}
		}
	}

	return schemaInputs
}

func (cov *Coverage) Predict(idx *workspace.AzureRepoIndex) error {
	activeSchemaInputs := cov.ToSchemaInputs()

	activeSchemaInputMap := map[workspace.SchemaLocator]bool{}
	for _, schemaInput := range activeSchemaInputs {
		activeSchemaInputMap[schemaInput.SchemaLocator] = true
	}

predictLoop:
	for _, schemaInput := range activeSchemaInputs {
		specPath, err := idx.LocateSpec(schemaInput.SpecLocator)
		if err != nil {
			if errors.Is(err, workspace.ErrFailedToLocateSwagger) {
				log.Warn(err)
				continue
			}
			return err
		}

		pathInfo, err := specPath.PathInfo()
		if err != nil {
			return err
		}

		swg, err := swaggerloader.LoadSwagger(filepath.Join(idx.SpecRootDir, string(specPath)))
		if err != nil {
			return err
		}

		rp, currentVersion := pathInfo.ResourceProvider, swg.Info.Version

		var newApiVersions []workspace.APIVersion
		for apiVersion := range idx.ResourceProviders[workspace.ResourceProvider(rp)] {
			if string(apiVersion) > currentVersion {
				// Modify the API version to shift to the newer ones
				schemaInput.SchemaLocator.SpecLocator.ApiVersion = string(apiVersion)

				// If at least one newer version has already been explicitly specified via schemaInput files, then skip prediction based on current
				// schemaInput. We should use the latest existing schemaInput to do the prediction on newer versions.
				if activeSchemaInputMap[schemaInput.SchemaLocator] {
					continue predictLoop
				}
				newApiVersions = append(newApiVersions, apiVersion)
			}
		}

		for _, apiVersion := range newApiVersions {
			// Modify the API version to shift to the newer ones
			schemaInput.SchemaLocator.SpecLocator.ApiVersion = string(apiVersion)

			// Log out the current schema under prediction
			msg := fmt.Sprintf("Processing (predict): %s (%s - %s)", schemaInput.ApiPath, schemaInput.Operation, schemaInput.ApiVersion)
			if schemaInput.Variant != "" {
				msg = fmt.Sprintf("%s [variant: %s]", msg, schemaInput.Variant)
			}
			log.Infoln(msg)

			schema, err := schemaInput.ToSchema(idx, false, false)
			if err != nil {
				if errors.Is(err, workspace.ErrFailedToLocateSwagger) || errors.Is(err, workspace.ErrFailedToLocateSchema) {
					log.Warn(err)
					continue
				}
				return err
			}

			if err := cov.AddSchema(schemaInput.SchemaLocator, *schema); err != nil {
				return err
			}
		}
	}
	return nil
}

type FlatCoverage map[workspace.SchemaLocator]swagger.Schema

// Flatten return the flattened coverage which is basically a map whose key is the schema locator, and the value is the schema.
// Note: the granted api path are not converted.
func (cov Coverage) Flatten() FlatCoverage {
	flatCov := FlatCoverage{}
	for _, apiPaths := range cov {
		for apiPath, path := range *apiPaths {
			for apiOperation, apiVersions := range path.Operations {
				for apiVersion, schema := range apiVersions {
					schemaLocator := workspace.SchemaLocator{
						SpecLocator: workspace.SpecLocator{
							ApiVersion: apiVersion,
							ApiPath:    apiPath,
						},
						Operation: workspace.OperationKind(apiOperation),
					}

					flatCov[schemaLocator] = schema
				}
			}
		}
	}
	return flatCov
}

func (flatCov FlatCoverage) Structuralize() (Coverage, error) {
	cov := NewEmptyCoverage()
	for locator, schema := range flatCov {
		if err := cov.AddSchema(locator, schema); err != nil {
			return Coverage{}, err
		}
	}
	return cov, nil
}
