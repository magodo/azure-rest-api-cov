package analyzer

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/magodo/azure-rest-api-cov/analyzer/coverage"
	"github.com/magodo/azure-rest-api-cov/analyzer/feeddb"
	"github.com/magodo/azure-rest-api-cov/analyzer/workspace"
	log "github.com/sirupsen/logrus"
)

type FeedDBContext struct {
	WorkspaceOpt workspace.InitOption

	// Build comment to add into the coverage report.
	BuildComment string

	// A list of schema locator pattern in the format of: <API Path>:<Operation>:<API Version>, to skip the from reporting.
	// This is useful to ignore some "infinite" model like data factory pipeline.
	IgnoreSchema []string
}

func FeedDB(in io.Reader, ctx FeedDBContext) error {
	log.Info("CMD: FeedDB")

	tempDir, err := os.MkdirTemp("", "feeddb-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tempDir)

	wsp, err := workspace.Init(ctx.WorkspaceOpt)
	if err != nil {
		return err
	}

	// Parse the coverage
	b, err := ioutil.ReadAll(in)
	if err != nil {
		return err
	}
	var cov coverage.Coverage
	if err := json.Unmarshal(b, &cov); err != nil {
		return err
	}

	// Adding the schema coverage to report
	flatCov := cov.Flatten()
	covReport := feeddb.NewEmptyCoverage(wsp.AzureRepoIndex, ctx.BuildComment)
	for k, v := range flatCov {
		log.Infof("Processing: %s", k)
		if err := covReport.AddOperation(v); err != nil {
			return err
		}
	}

	out, err := covReport.GenerateWithUncoveredOperationModels(wsp.AzureRepoIndex, ctx.IgnoreSchema)
	if err != nil {
		return err
	}
	fmt.Println(out)
	return err
}
