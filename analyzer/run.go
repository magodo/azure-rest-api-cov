package analyzer

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/magodo/azure-rest-api-cov/analyzer/grant"
	"io"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"

	"github.com/magodo/azure-rest-api-cov/analyzer/structuralize"

	"github.com/magodo/azure-rest-api-cov/analyzer/swaggerloader"

	"github.com/magodo/azure-rest-api-cov/analyzer/schemainput"
	log "github.com/sirupsen/logrus"

	"github.com/magodo/azure-rest-api-cov/analyzer/swagger"
	"github.com/magodo/azure-rest-api-cov/analyzer/workspace"
)

type SchemaKey struct {
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

func (schemas Schemas) AddSchema(locator workspace.SchemaLocator, s swagger.Schema) error {
	schemaKey := SchemaKey{
		SchemaLocator: locator,
	}
	existing, ok := schemas[schemaKey]
	if !ok {
		schemas[schemaKey] = s
		return nil
	}
	return existing.Merge(s)
}

func (schemas Schemas) predict(schemaInputs []schemainput.SchemaInput, wsp *workspace.Workspace) error {
	activeSchemas := map[workspace.SchemaLocator]bool{}
	for _, schemaInput := range schemaInputs {
		activeSchemas[schemaInput.SchemaLocator] = true
	}

predictLoop:
	for _, schemaInput := range schemaInputs {
		specPath, err := wsp.AzureRepoIndex.LocateSpec(schemaInput.SpecLocator)
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

		swg, err := swaggerloader.LoadSwagger(filepath.Join(wsp.AzureRepoIndex.SpecRootDir, string(specPath)))
		if err != nil {
			return err
		}

		rp, currentVersion := pathInfo.ResourceProvider, swg.Info.Version

		var newApiVersions []workspace.APIVersion
		for apiVersion := range wsp.AzureRepoIndex.ResourceProviders[workspace.ResourceProvider(rp)] {
			if string(apiVersion) > currentVersion {
				// Modify the API version to shift to the newer ones
				schemaInput.SchemaLocator.SpecLocator.ApiVersion = string(apiVersion)

				// If at least one newer version has already been explicitly specified via schemaInput files, then skip prediction based on current
				// schemaInput. We should use the latest existing schemaInput to do the prediction on newer versions.
				if activeSchemas[schemaInput.SchemaLocator] {
					continue predictLoop
				}
				newApiVersions = append(newApiVersions, apiVersion)
			}
		}

		for _, apiVersion := range newApiVersions {
			// Modify the API version to shift to the newer ones
			schemaInput.SchemaLocator.SpecLocator.ApiVersion = string(apiVersion)

			logCurrentProcessingSchemaInput(schemaInput, true)
			schema, err := schemaInput.ToSchema(wsp.AzureRepoIndex, false, false)
			if err != nil {
				if errors.Is(err, workspace.ErrFailedToLocateSwagger) || errors.Is(err, workspace.ErrFailedToLocateSchema) {
					log.Warn(err)
					continue
				}
				return err
			}
			if err := schemas.AddSchema(schemaInput.SchemaLocator, *schema); err != nil {
				return err
			}
		}
	}
	return nil
}

type RunContext struct {
	WorkspaceOpt workspace.InitOption

	// Whether predict property coverage for newer api version models.
	Predict bool

	// Whether to complement the coverage report with the not covered RP/resources.
	All bool

	// Additional message to add into the coverage report.
	AdditionalMessage string

	// Whether to generate migrate grant hint file.
	MigrateGrantHint bool

	// Whether to try to use the existing grant file for models with different API versions.
	TryMigrate bool

	// A list of schema locator pattern in the format of: <API Path>:<Operation>:<API Version>, to skip the coverage calculation.
	// This is useful to ignore some "infinite" model like data factory pipeline.
	IgnoreSchema []string
}

func Run(schemaInputsIn io.Reader, ctx RunContext) (string, error) {
	// Initialize the workspace, which will return the repo and repo index.
	wsp, err := workspace.Init(ctx.WorkspaceOpt)
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

	// Ignore some specified schemaInput
	if ctx.IgnoreSchema != nil {
		var ignorePatterns []workspace.SchemaLocator
		for _, p := range ctx.IgnoreSchema {
			loc, err := workspace.NewSchemaLocatorFromPattern(p)
			if err != nil {
				return "", err
			}
			ignorePatterns = append(ignorePatterns, loc)
		}

		tmpSchemaInputs := make([]schemainput.SchemaInput, 0, len(schemaInputs))
	ignoreLoop:
		for _, schemaInput := range schemaInputs {
			for _, loc := range ignorePatterns {
				if loc.Match(schemaInput.SchemaLocator) {
					log.Infof("schema %q is ignored for property coverage calculation", schemaInput.SchemaLocator)
					continue ignoreLoop
				}
			}
			tmpSchemaInputs = append(tmpSchemaInputs, schemaInput)
		}
		schemaInputs = tmpSchemaInputs
	}

	// Consume the schema inputs to get the property coverage

	// schemas is used to store all the schemas with every property state (cov/un-cov/granted/ignored) set
	schemas := Schemas{}
	for _, schemaInput := range schemaInputs {
		logCurrentProcessingSchemaInput(schemaInput, false)

		schema, err := schemaInput.ToSchema(wsp.AzureRepoIndex, true, true)
		if err != nil {
			if errors.Is(err, workspace.ErrFailedToLocateSwagger) {
				log.Warn(err)
				continue
			}
			return "", err
		}

		if err := schemas.AddSchema(schemaInput.SchemaLocator, *schema); err != nil {
			return "", err
		}
	}

	// (Optionally) Predict the coverage of the Specs of newer API versions based on the schema inputs
	if ctx.Predict {
		if err := schemas.predict(schemaInputs, wsp); err != nil {
			return "", err
		}
	}

	// Grant
	if ctx.WorkspaceOpt.GrantRepoURL != "" {
		hintMessages, err := schemas.grant(wsp, ctx.TryMigrate)
		if err != nil {
			return "", err
		}

		if ctx.MigrateGrantHint {
			if err := ioutil.WriteFile("possible_grant_migration.txt", []byte(strings.Join(hintMessages, "\n")), 0644); err != nil {
				return "", err
			}
		}
	}

	azureRepoURLPrefix := fmt.Sprintf("https://github.com/Azure/azure-rest-api-specs/blob/%s/specification/", wsp.AzureRepoIndex.Commit)

	// Generate coverage result (only the covered models)
	if !ctx.All {
		var keys SchemaKeys
		for k := range schemas {
			keys = append(keys, k)
		}
		sort.Sort(keys)

		var covs []*structuralize.SchemaCoverage
		for _, k := range keys {
			schema := schemas[k]
			cov, err := structuralize.NewSchemaCoverage(azureRepoURLPrefix, schema)
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
	cov := structuralize.NewEmptyCoverage(*(wsp.AzureRepoIndex), ctx.AdditionalMessage)
	for k, schema := range schemas {
		if err := cov.AddSchema(azureRepoURLPrefix, schema, k.SchemaLocator); err != nil {
			return "", err
		}
	}

	b, err = json.MarshalIndent(cov, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// grant leverage the grant repo to grant the schemas (both active and predicted schemas).
// In case there is any grant exist but is for some other API version, a hint message will be returned.
// If tryMigrate is specified, these API mismatched grant will take effect.
func (schemas Schemas) grant(wsp *workspace.Workspace, tryMigrate bool) ([]string, error) {
	grants, err := grant.NewSchemaGrantsFromFiles(wsp.GrantRepo.RootDir, wsp.AzureRepoIndex.SpecRootDir, nil)
	if err != nil {
		return nil, err
	}

	var hintMessages []string
	for k := range schemas {
		schema := schemas[k]
		grant, matchRef, err := grants.GetGrantForSchema(schema.Ref())
		if err != nil {
			return nil, err
		}
		if grant != nil {
			// The grant file is exactly for this schema.
			if matchRef == schema.Ref() {
				if err := grant.Grant(&schema, true); err != nil {
					return nil, fmt.Errorf("granting for schema %v: %v", k.SchemaLocator, err)
				}
			} else {
				hintMessages = append(hintMessages, fmt.Sprintf("there is a grant with different API version %q, that is able to be migrated for %q", matchRef, schema.Ref()))
				if tryMigrate {
					if err := grant.Grant(&schema, false); err != nil {
						return nil, fmt.Errorf("(trying) granting for schema %v: %v", k.SchemaLocator, err)
					}
				}
			}
		}
	}
	return hintMessages, nil
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
