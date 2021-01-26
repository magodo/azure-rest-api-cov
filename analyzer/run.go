package analyzer

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"sort"

	"github.com/magodo/azure-rest-api-cov/analyzer/coverage"

	"github.com/magodo/azure-rest-api-cov/analyzer/swaggerloader"

	"github.com/magodo/azure-rest-api-cov/analyzer/schemainput"
	log "github.com/sirupsen/logrus"

	"github.com/magodo/azure-rest-api-cov/analyzer/swagger"
	"github.com/magodo/azure-rest-api-cov/analyzer/workspace"
)

type SchemaKey struct {
	Active        bool
	SchemaLocator workspace.SchemaLocator
}

type SchemaKeys []SchemaKey

func (keys SchemaKeys) Len() int {
	return len(keys)
}

func (keys SchemaKeys) Less(i, j int) bool {
	if keys[i].SchemaLocator.ApiPath < keys[j].SchemaLocator.ApiPath {
		return true
	}
	if keys[i].SchemaLocator.ApiPath > keys[j].SchemaLocator.ApiPath {
		return false
	}
	if keys[i].SchemaLocator.ApiVersion < keys[j].SchemaLocator.ApiVersion {
		return true
	}
	if keys[i].SchemaLocator.ApiVersion > keys[j].SchemaLocator.ApiVersion {
		return false
	}
	return keys[i].SchemaLocator.Operation < keys[j].SchemaLocator.Operation
}

func (keys SchemaKeys) Swap(i, j int) {
	keys[i], keys[j] = keys[j], keys[i]
}

type Schemas map[SchemaKey]swagger.Schema

func (schemas Schemas) AddSchema(locator workspace.SchemaLocator, s swagger.Schema, active bool) error {
	schemaKey := SchemaKey{
		Active:        active,
		SchemaLocator: locator,
	}
	existing, ok := schemas[schemaKey]
	if !ok {
		schemas[schemaKey] = s
		return nil
	}
	return existing.Merge(s)
}

func Run(schemaInputsIn io.Reader, workspaceOpt workspace.InitOption, predict bool, all bool, additionalMessage string) (string, error) {

	// Initialize the workspace, which will return the repo and repo index.
	wsp, err := workspace.Init(workspaceOpt)
	if err != nil {
		return "", err
	}

	// Parse the schema inputs
	b, err := ioutil.ReadAll(schemaInputsIn)
	if err != nil {
		return "", err
	}
	var schemaInputs []schemainput.SchemaInput
	if err := json.Unmarshal(b, &schemaInputs); err != nil {
		return "", err
	}

	// Consume the schema inputs
	schemas := Schemas{}
	for _, schemaInput := range schemaInputs {
		logCurrentProcessingSchemaInput(schemaInput, false)
		schema, err := schemaInput.ToSchema(wsp.AzureRepoIndex, true)
		if err != nil {
			if errors.Is(err, workspace.ErrFailedToLocateSwagger) {
				log.Warn(err)
				continue
			}
			return "", err
		}

		if err := schemas.AddSchema(schemaInput.SchemaLocator, *schema, true); err != nil {
			return "", err
		}
	}

	// (Optionally) Predict the coverage of the Specs of newer API versions based on the schema inputs
	if predict {
		explicitSchemas := map[workspace.SchemaLocator]bool{}
		for _, schemaInput := range schemaInputs {
			explicitSchemas[schemaInput.SchemaLocator] = true
		}

	loop:
		for _, schemaInput := range schemaInputs {
			specPath, err := wsp.AzureRepoIndex.LocateSpec(schemaInput.SpecLocator)
			if err != nil {
				if errors.Is(err, workspace.ErrFailedToLocateSwagger) {
					log.Warn(err)
					continue
				}
				return "", err
			}

			pathInfo, err := specPath.PathInfo()
			if err != nil {
				return "", err
			}

			swg, err := swaggerloader.LoadSwagger(filepath.Join(wsp.AzureRepoIndex.SpecRootDir, string(specPath)))
			if err != nil {
				return "", err
			}

			rp, currentVersion := pathInfo.ResourceProvider, swg.Info.Version

			var newApiVersions []workspace.APIVersion
			for apiVersion := range wsp.AzureRepoIndex.ResourceProviders[workspace.ResourceProvider(rp)] {
				if string(apiVersion) > currentVersion {
					// Modify the API version to shift to the newer ones
					schemaInput.SchemaLocator.SpecLocator.ApiVersion = string(apiVersion)

					// If at least one newer version has already been explicitly specified via schemaInput files, then skip prediction based on current
					// schemaInput. We should use the latest existing schemaInput to do the prediction on newer versions.
					if explicitSchemas[schemaInput.SchemaLocator] {
						continue loop
					}
					newApiVersions = append(newApiVersions, apiVersion)
				}
			}

			for _, apiVersion := range newApiVersions {
				// Modify the API version to shift to the newer ones
				schemaInput.SchemaLocator.SpecLocator.ApiVersion = string(apiVersion)

				logCurrentProcessingSchemaInput(schemaInput, true)
				schema, err := schemaInput.ToSchema(wsp.AzureRepoIndex, false)
				if err != nil {
					if errors.Is(err, workspace.ErrFailedToLocateSwagger) || errors.Is(err, workspace.ErrFailedToLocateSchema) {
						log.Warn(err)
						continue
					}
					return "", err
				}
				if err := schemas.AddSchema(schemaInput.SchemaLocator, *schema, false); err != nil {
					return "", err
				}
			}
		}
	}

	// Grant
	if workspaceOpt.GrantRepoURL != "" {
		grants, err := swagger.NewSchemaGrantsFromFiles(wsp.GrantRepo.RootDir, wsp.AzureRepoIndex.SpecRootDir)
		if err != nil {
			return "", err
		}

		for k, schema := range schemas {
			grant, ok := grants[schema.Ref()]
			if !ok {
				continue
			}
			if err := schema.Grant(grant, k.Active); err != nil {
				return "", fmt.Errorf("granting for schema %v: %v", k.SchemaLocator, err)
			}
		}
	}

	// Generate coverage result
	azureRepoURLPrefix := fmt.Sprintf("https://github.com/Azure/azure-rest-api-specs/blob/%s/specification/", wsp.AzureRepoIndex.Commit)
	if !all {
		var keys SchemaKeys
		for k := range schemas {
			keys = append(keys, k)
		}
		sort.Sort(keys)

		var covs []*coverage.SchemaCoverageStore
		for _, k := range keys {
			schema := schemas[k]
			cov, err := coverage.NewSchemaCoverageStore(azureRepoURLPrefix, schema, k.Active)
			if err != nil {
				return "", err
			}
			covs = append(covs, cov)
		}

		b, err = json.MarshalIndent(covs, "", "  ")
		if err != nil {
			return "", err
		}
		return string(b), nil
	}

	// Complement the coverage report with all uncovered models.
	cov := coverage.NewEmptyRepoCoverageStore(*(wsp.AzureRepoIndex), additionalMessage)
	for k, schema := range schemas {
		if err := cov.AddSchema(azureRepoURLPrefix, schema, k.SchemaLocator, k.Active); err != nil {
			return "", err
		}
	}

	b, err = json.MarshalIndent(cov, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func logCurrentProcessingSchemaInput(schemaInput schemainput.SchemaInput, predict bool) {
	decoration := ""
	if predict {
		decoration = " (predict)"
	}
	msg := fmt.Sprintf("Processing%s: %s (%s - %s)", decoration, schemaInput.ApiPath, schemaInput.Operation, schemaInput.ApiVersion)
	if schemaInput.Variant != "" {
		msg = fmt.Sprintf("%s [variant: %s]", msg, schemaInput.Variant)
	}
	log.Infoln(msg)
}
