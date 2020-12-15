package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

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
	repo, index, err := workspace.Init(*workspaceDir, workspace.InitOption{
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
		name := schemaInput.Name
		if schemaInput.Variant != "" {
			name = fmt.Sprintf("%s[%s]", name, schemaInput.Variant)
		}
		log.Printf("Processing: %s(%s): %s", schemaInput.ApiPath, schemaInput.ApiVersion, name)
		swaggerRelPath, ok := index.SpecLookupView()[schemaInput.SpecLocator]
		if !ok {
			log.Printf("[WARNING] Can't locate swagger spec by (%q, %q), this might because this tuple ends up with multiple swagger specs.",
				schemaInput.SpecLocator.ApiVersion, schemaInput.SpecLocator.ApiPath)
			continue
		}

		schema, err := schemaInput.ToSchema(repo.SpecRootDir, swaggerRelPath)
		if err != nil {
			log.Fatal(err)
		}
		cov, err := schema.CalcCoverage()
		if err != nil {
			log.Fatal(err)
		}
		covered, total := cov.SchemaCoverage()
		fmt.Printf("%s: (%d/%d)\n", schema.SchemaAddr.String(), covered, total)
	}
}
