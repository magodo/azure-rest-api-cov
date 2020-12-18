package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/magodo/azure-rest-api-cov/pkg/core/coverage"

	"github.com/magodo/azure-rest-api-cov/pkg/core/schemainput"

	"github.com/magodo/azure-rest-api-cov/pkg/core/workspace"
)

const usage = `azure-rest-api-cov <option>`

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "%s\n", usage)
		flag.PrintDefaults()
	}

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		log.Fatal(err)
	}
	workspaceDir := flag.String("w", filepath.Join(cacheDir, "azure-rest-api-cov"),
		"The path to the workspace for caching the azure-rest-api repository and other stuffs")
	inputPath := flag.String("i", "",
		"The path to the schema input file, which is expected to hold a JSON array of input schemas")
	force := flag.Bool("f", false, "whether to force create the workspace")
	repoHash := flag.String("c", "", "The azure-rest-api-spec repo commit hash to checkout")

	flag.Parse()

	if *inputPath == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Initialize the workspace, which will return the repo and repo index.
	_, index, err := workspace.Init(*workspaceDir, workspace.InitOption{
		Force:    *force,
		RepoHash: *repoHash,
	})
	if err != nil {
		log.Fatal(err)
	}

	b, err := ioutil.ReadFile(*inputPath)
	if err != nil {
		log.Fatal(err)
	}
	var schemaInputs []schemainput.SchemaInput
	if err := json.Unmarshal(b, &schemaInputs); err != nil {
		log.Fatal(err)
	}

	for _, schemaInput := range schemaInputs {
		msg := fmt.Sprintf("Processing: %s (%s - %s)", schemaInput.ApiPath, schemaInput.Operation, schemaInput.ApiVersion)
		if schemaInput.Variant != "" {
			msg = fmt.Sprintf("%s [variant: %s]", msg, schemaInput.Variant)
		}
		log.Println(msg)

		schema, err := schemaInput.ToSchema(index)
		if err != nil {
			if errors.Is(err, workspace.ErrFailedToLocateSwagger) {
				continue
			}
			log.Fatal(err)
		}
		cov, err := coverage.NewSchemaCoverageStore(*schema)
		b, err := json.MarshalIndent(cov, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(b))
		//covered, total := cov.SchemaCoverage()
		//fmt.Printf("%s: (%d/%d)\n", schema.SchemaAddr.String(), covered, total)
	}
}
