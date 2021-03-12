package analyzer

import (
	"encoding/json"
	"fmt"
	"github.com/magodo/azure-rest-api-cov/analyzer/coverage"
	"github.com/magodo/azure-rest-api-cov/analyzer/report"
	"github.com/magodo/azure-rest-api-cov/analyzer/workspace"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"sort"
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

	flatCov := cov.Flatten()
	azureRepoURLPrefix := fmt.Sprintf("https://github.com/Azure/azure-rest-api-specs/blob/%s/specification/", wsp.AzureRepoIndex.Commit)

	var locators workspace.SchemaLocators
	for k := range flatCov {
		locators = append(locators, k)
	}
	sort.Sort(locators)

	var covReport []*report.SchemaCoverage
	for _, k := range locators {
		schema := flatCov[k]
		cov, err := report.NewSchemaCoverage(azureRepoURLPrefix, schema)
		if err != nil {
			return "", err
		}
		covReport = append(covReport, cov)
	}

	b, err = json.MarshalIndent(covReport, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}
