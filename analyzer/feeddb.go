package analyzer

import (
	"encoding/json"
	"github.com/magodo/azure-rest-api-cov/analyzer/coverage"
	"github.com/magodo/azure-rest-api-cov/analyzer/feeddb"
	"github.com/magodo/azure-rest-api-cov/analyzer/workspace"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
)

type FeedDBContext struct {
	WorkspaceOpt workspace.InitOption

	// Build comment to add into the coverage report.
	BuildComment string
}

func FeedDB(in io.Reader, ctx FeedDBContext) (string, error) {
	log.Info("CMD: FeedDB")

	wsp, err := workspace.Init(ctx.WorkspaceOpt)
	if err != nil {
		return "", err
	}

	// Parse the coverage
	b, err := ioutil.ReadAll(in)
	if err != nil {
		return "", err
	}
	var cov coverage.Coverage
	if err := json.Unmarshal(b, &cov); err != nil {
		return "", err
	}

	// Adding the schema coverage to report
	flatCov := cov.Flatten()
	covReport := feeddb.NewEmptyCoverage(wsp.AzureRepoIndex, ctx.BuildComment)
	for k, schema := range flatCov {
		log.Infof("Processing: %s", k)
		if err := covReport.AddSchema(schema, k); err != nil {
			return "", err
		}
	}

	if err := covReport.IncludeUncoveredOperationModels(wsp.AzureRepoIndex); err != nil {
		return "", err
	}

	b, err = json.MarshalIndent(covReport, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}
