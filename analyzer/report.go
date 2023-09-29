package analyzer

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/magodo/azure-rest-api-cov/analyzer/coverage"
	"github.com/magodo/azure-rest-api-cov/analyzer/report"
	"github.com/magodo/azure-rest-api-cov/analyzer/workspace"
	log "github.com/sirupsen/logrus"
)

type ReportContext struct {
	WorkspaceOpt workspace.InitOption

	// Additional message to add into the coverage report.
	AdditionalMessage string
}

func Report(in io.Reader, ctx ReportContext) (string, error) {
	log.Info("CMD: Report")

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
	azureRepoURLPrefix := fmt.Sprintf("https://github.com/Azure/azure-rest-api-specs/blob/%s/specification/", wsp.AzureRepoIndex.Index.Commit)

	covReport := report.NewEmptyCoverage(wsp.AzureRepoIndex, ctx.AdditionalMessage)
	for _, opInfo := range flatCov {
		if err := covReport.AddOperation(azureRepoURLPrefix, opInfo); err != nil {
			return "", err
		}
	}

	// Adding any granted api path to report
	for rpName, apiPaths := range cov {
		for pathName, apiPath := range *apiPaths {
			if apiPath.GrantComment != "" {
				(*covReport.ResourceProviders[report.ResourceProvider(rpName)])[pathName].GrantComment = apiPath.GrantComment
			}
		}
	}

	b, err = json.MarshalIndent(covReport, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}
