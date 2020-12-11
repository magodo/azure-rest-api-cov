package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

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
	workspaceDir := flag.String("workspace", filepath.Join(cacheDir, "azure-rest-api-cov"),
		"The path to the workspace for caching the azure-rest-api repository and other stuffs")
	inputPath := flag.String("schema-cov-dir", "",
		"The path to the directory contains schema schema input files")

	flag.Parse()

	if *workspaceDir == "" ||
		*inputPath == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Initialize the workspace, which will return the repo and repo index.
	repo, index, err := workspace.Init(*workspaceDir, false)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Walk the schema schema input files...")
	if err := filepath.Walk(*inputPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() ||
			!strings.HasSuffix(info.Name(), ".json") {
			return nil
		}

		b, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		var schemaInput schemainput.SchemaInput
		if err := json.Unmarshal(b, &schemaInput); err != nil {
			return err
		}

		swaggerRelPath, ok := index.SpecLookupView()[schemaInput.SpecLocator]
		if !ok {
			log.Printf("[WARNING] Can't locate swagger spec by (%q, %q), this might because this tuple ends up with multiple swagger specs.",
				schemaInput.SpecLocator.ApiVersion, schemaInput.SpecLocator.ApiPath)
		}
		schema, err := schemaInput.ToSchema(repo.SpecRootDir, swaggerRelPath)
		if err != nil {
			return err
		}
		cov, err := schema.CalcCoverage()
		if err != nil {
			return err
		}
		covered, total := cov.SchemaCoverage()
		fmt.Printf("%s: (%d/%d)\n", schema.SchemaAddr.String(), covered, total)
		return nil
	}); err != nil {
		fmt.Fprint(os.Stderr, err)
		os.Exit(1)
	}
}
