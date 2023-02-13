package analyzer

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"runtime"

	"github.com/magodo/azure-rest-api-cov/analyzer/coverage"
	"github.com/magodo/azure-rest-api-cov/analyzer/operation"
	"github.com/magodo/azure-rest-api-cov/analyzer/propmatch"
	"github.com/magodo/azure-rest-api-cov/analyzer/schemainput"
	"github.com/magodo/azure-rest-api-cov/analyzer/utils"
	"github.com/magodo/azure-rest-api-cov/analyzer/workspace"
	log "github.com/sirupsen/logrus"
)

type CoverageContext struct {
	WorkspaceOpt workspace.InitOption

	// Whether predict property coverage for newer api version models.
	Predict bool

	// A list of schema locator pattern in the format of: <API Path>:<Operation>:<API Version>, to skip the coverage calculation.
	// This is useful to ignore some "infinite" model like data factory pipeline.
	IgnoreSchema []string

	// A list of property locator pattern, to allow a mismatch from its API spec definition.
	// This is mainly used for workaround the artificially added attributes like in: https://github.com/hashicorp/pandora/commit/13e177637e196c42d0b0bdc82df1fa703d2c0da7
	AllowMismatchPropMatchers []propmatch.PropMatcher
}

func Coverage(in io.Reader, ctx CoverageContext) (string, error) {
	log.Info("CMD: Coverage")

	wsp, err := workspace.Init(ctx.WorkspaceOpt)
	if err != nil {
		return "", err
	}

	// Parse the schema inputs
	b, err := ioutil.ReadAll(in)
	if err != nil {
		return "", err
	}
	var schemaInputs []schemainput.SchemaInput
	if err := json.Unmarshal(b, &schemaInputs); err != nil {
		return "", err
	}

	// Ignore some specified schemaInput
	if ctx.IgnoreSchema != nil {
		var ignorePatterns []workspace.OperationLocator
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
				if loc.Match(schemaInput.OperationLocator) {
					log.Infof("schema %q is ignored for property coverage calculation", schemaInput.OperationLocator)
					continue ignoreLoop
				}
			}
			tmpSchemaInputs = append(tmpSchemaInputs, schemaInput)
		}
		schemaInputs = tmpSchemaInputs
	}

	// Consume the schema inputs to get the property coverage
	cov := coverage.NewEmptyCoverage()

	// Set up the work pool
	wp := utils.NewWorkPool(runtime.NumCPU())
	wp.Run(func(i interface{}) error {
		return cov.AddOperation(i.(*operation.Info))
	})

	for _, schemaInput := range schemaInputs {
		schemaInput := schemaInput
		wp.AddTask(func() (interface{}, error) {
			// Log output
			msg := fmt.Sprintf("Processing: %s (%s - %s)", schemaInput.ApiPath, schemaInput.Operation, schemaInput.ApiVersion)
			if schemaInput.Variant != "" {
				msg = fmt.Sprintf("%s [variant: %s]", msg, schemaInput.Variant)
			}
			log.Infoln(msg)

			opInfo, err := schemaInput.ToOperationInfo(wsp.AzureRepoIndex, true, true, ctx.AllowMismatchPropMatchers)
			if err != nil {
				if errors.Is(err, workspace.ErrFailedToLocateSwagger) {
					log.Warn(err)
					return opInfo, utils.ErrSkipTask
				}
			}
			return opInfo, err
		})
	}

	if err := wp.Done(); err != nil {
		return "", err
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

func logCurrentProcessingSchemaInput(schemaInput schemainput.SchemaInput) {
	msg := fmt.Sprintf("Processing: %s (%s - %s)", schemaInput.ApiPath, schemaInput.Operation, schemaInput.ApiVersion)
	if schemaInput.Variant != "" {
		msg = fmt.Sprintf("%s [variant: %s]", msg, schemaInput.Variant)
	}
	log.Infoln(msg)
}
