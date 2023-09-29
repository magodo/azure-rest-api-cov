package coverage

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/magodo/azure-rest-api-cov/analyzer/operation"
	"github.com/magodo/azure-rest-api-cov/analyzer/utils"

	"github.com/magodo/azure-rest-api-cov/analyzer/addr"
	"github.com/magodo/azure-rest-api-cov/analyzer/schemainput"
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

type Operations map[string]*APIVersions

type APIVersions map[string]*operation.Info

func NewEmptyCoverage() Coverage {
	return Coverage{}
}

func (cov *Coverage) AddOperation(opInfo *operation.Info) error {
	rp, apiPath, op, apiVersion := opInfo.RP, opInfo.APIPath, string(opInfo.Operation), opInfo.APIVersion

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
		apiVersions = &APIVersions{}
		path.Operations[op] = apiVersions
	}

	existOp, ok := (*apiVersions)[apiVersion]
	if !ok {
		(*apiVersions)[apiVersion] = opInfo
		return nil
	}

	if existOp.Model == nil || opInfo.Model == nil {
		return nil
	}

	return existOp.Model.Merge(*opInfo.Model)
}

// ToSchemaInputs converts the coverage back to an array of SchemaInputs.
// Note that the order in the SchemaInputs are not guaranteed. Also, the granted api path are not converted.
func (cov Coverage) ToSchemaInputs() []schemainput.SchemaInput {
	var schemaInputs []schemainput.SchemaInput

	for _, apiPaths := range cov {
		for apiPath, path := range *apiPaths {
			for apiOperation, apiVersions := range path.Operations {
				for apiVersion, opInfo := range *apiVersions {
					schemaLocator := workspace.OperationLocator{
						SpecLocator: workspace.SpecLocator{
							ApiVersion: apiVersion,
							ApiPath:    apiPath,
						},
						Operation: workspace.OperationKind(apiOperation),
					}

					if opInfo.Model == nil {
						schemaInputs = append(schemaInputs, schemainput.SchemaInput{
							OperationLocator: schemaLocator,
						})
						continue
					}

					// If the schema has top level variant, then it will be transformed to numerous schemaInputs, each corresponds to a variant.
					// We peek any property to see whether it is a top level variant, since when a schema contains top level properties, all the property addresses
					// start with a variant.
					var isTopLevelDiscriminator bool
					for paddr := range opInfo.Model.Properties {
						prop := addr.MustParseProperty(paddr)
						if prop[0].DiscriminatorValue != "" && prop[0].Name == "" {
							isTopLevelDiscriminator = true
						}
						break
					}

					if !isTopLevelDiscriminator {
						schemaInput := schemainput.SchemaInput{
							OperationLocator: schemaLocator,
						}
						for paddr, prop := range opInfo.Model.Properties {
							if prop.IsCovered() {
								schemaInput.Properties = append(schemaInput.Properties, paddr)
							}
						}
						schemaInputs = append(schemaInputs, schemaInput)
						continue
					}

					variantProperties := map[string][]string{}
					for paddr, pinfo := range opInfo.Model.Properties {
						if !pinfo.IsCovered() {
							continue
						}
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
							OperationLocator: schemaLocator,
							Variant:          variant,
							Properties:       variantProps,
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

	activeSchemaInputMap := map[workspace.OperationLocator]bool{}
	for _, schemaInput := range activeSchemaInputs {
		activeSchemaInputMap[schemaInput.OperationLocator] = true
	}

	// Set up the work pool
	wp := utils.NewWorkPool(runtime.NumCPU())
	wp.Run(func(i interface{}) error {
		return cov.AddOperation(i.(*operation.Info))
	})

predictLoop:
	for _, schemaInput := range activeSchemaInputs {
		opInfo, err := idx.LocateOperation(schemaInput.OperationLocator)
		if err != nil {
			if errors.Is(err, workspace.ErrFailedToLocateSwagger) {
				log.Warn(err)
				continue
			}
			return err
		}

		specPath := opInfo.SpecRelPath
		pathInfo, err := specPath.PathInfo()
		if err != nil {
			return err
		}

		swg, err := swaggerloader.LoadSwagger(filepath.Join(idx.SpecRootDir, string(specPath)))
		if err != nil {
			return err
		}

		rp, currentVersion := pathInfo.ResourceProvider, swg.Info.Version

		var newApiVersions []string
		for apiVersion := range idx.Index.ResourceProviders[rp] {
			if string(apiVersion) > currentVersion {
				// Modify the API version to shift to the newer ones
				schemaInput.OperationLocator.SpecLocator.ApiVersion = string(apiVersion)

				// If at least one newer version has already been explicitly specified via schemaInput files, then skip prediction based on current
				// schemaInput. We should use the latest existing schemaInput to do the prediction on newer versions.
				if activeSchemaInputMap[schemaInput.OperationLocator] {
					continue predictLoop
				}
				newApiVersions = append(newApiVersions, apiVersion)
			}
		}

		for _, apiVersion := range newApiVersions {
			// Modify the API version to shift to the newer ones
			schemaInput.OperationLocator.SpecLocator.ApiVersion = string(apiVersion)

			// Log out the current schema under prediction
			msg := fmt.Sprintf("Processing (predict): %s (%s - %s)", schemaInput.ApiPath, schemaInput.Operation, schemaInput.ApiVersion)
			if schemaInput.Variant != "" {
				msg = fmt.Sprintf("%s [variant: %s]", msg, schemaInput.Variant)
			}
			log.Infoln(msg)

			schemaInput := schemaInput
			wp.AddTask(func() (interface{}, error) {
				opInfo, err := schemaInput.ToOperationInfo(idx, false, false, nil)
				if err != nil {
					if errors.Is(err, workspace.ErrFailedToLocateSwagger) || errors.Is(err, workspace.ErrFailedToLocateSchema) {
						log.Warn(err)
						return nil, utils.ErrSkipTask
					}
				}
				return opInfo, err
			},
			)
		}
	}
	return wp.Done()
}

type FlatCoverage map[workspace.OperationLocator]*operation.Info

// Flatten return the flattened coverage which is basically a map whose key is the operation locator, and the value is the operation info.
// Note: the granted api path are not converted.
func (cov Coverage) Flatten() FlatCoverage {
	flatCov := FlatCoverage{}
	for _, apiPaths := range cov {
		for apiPath, path := range *apiPaths {
			for apiOperation, apiVersions := range path.Operations {
				for apiVersion, opInfo := range *apiVersions {
					opLoc := workspace.OperationLocator{
						SpecLocator: workspace.SpecLocator{
							ApiVersion: apiVersion,
							ApiPath:    apiPath,
						},
						Operation: workspace.OperationKind(apiOperation),
					}

					flatCov[opLoc] = opInfo
				}
			}
		}
	}
	return flatCov
}

func (flatCov FlatCoverage) Structuralize() (Coverage, error) {
	cov := NewEmptyCoverage()
	for _, opInfo := range flatCov {
		if err := cov.AddOperation(opInfo); err != nil {
			return Coverage{}, err
		}
	}
	return cov, nil
}

func (cov *Coverage) UnmarshalJSON(b []byte) error {
	rps := map[string]*APIPaths{}
	if err := json.Unmarshal(b, &rps); err != nil {
		return err
	}

	for rpName, paths := range rps {
		for pathName, path := range *paths {
			for opName, versions := range path.Operations {
				for verName, ver := range *versions {
					ver.RP = rpName
					ver.APIPath = pathName
					ver.Operation = workspace.OperationKind(opName)
					ver.APIVersion = verName
				}
			}
		}
	}

	*cov = rps
	return nil
}
