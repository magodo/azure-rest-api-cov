package analyzer

import (
	"encoding/json"
	"errors"
	"github.com/magodo/azure-rest-api-cov/analyzer/coverage"
	"github.com/magodo/azure-rest-api-cov/analyzer/schemainput"
	"github.com/magodo/azure-rest-api-cov/analyzer/workspace"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
)

type CoverageContext struct {
	WorkspaceOpt workspace.InitOption

	// Whether predict property coverage for newer api version models.
	Predict bool

	// A list of schema locator pattern in the format of: <API Path>:<Operation>:<API Version>, to skip the coverage calculation.
	// This is useful to ignore some "infinite" model like data factory pipeline.
	IgnoreSchema []string
}

func Coverage(schemaInputsIn io.Reader, ctx CoverageContext) (string, error) {
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
	cov := coverage.NewEmptyCoverage()
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

		if err := cov.AddSchema(schemaInput.SchemaLocator, *schema); err != nil {
			return "", err
		}
	}

	// (Optionally) Predict the coverage of the Specs of newer API versions based on the schema inputs
	if ctx.Predict {
		if err := cov.Predict(wsp.AzureRepoIndex); err != nil {
			return "", err
		}
	}

	b, err = json.MarshalIndent(cov, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}
