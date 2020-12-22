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

	"github.com/magodo/azure-rest-api-cov/pkg/core/swagger"

	"github.com/magodo/azure-rest-api-cov/pkg/core/coverage"

	"github.com/magodo/azure-rest-api-cov/pkg/core/schemainput"

	"github.com/magodo/azure-rest-api-cov/pkg/core/workspace"
)

const usage = `azure-rest-api-cov <option>`

type Schemas map[workspace.SchemaLocator]swagger.Schema

func (schemas Schemas) AddSchema(locator workspace.SchemaLocator, s swagger.Schema) error {
	existing, ok := schemas[locator]
	if !ok {
		schemas[locator] = s
		return nil
	}
	return existing.Merge(s)
}

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
	regenIndex := flag.Bool("r", false, "whether to force regenerate the repository index in the workspace")
	repoHash := flag.String("c", "", "The azure-rest-api-spec repo commit hash to checkout")
	all := flag.Bool("a", false, "whether to export the RP level coverage")

	flag.Parse()

	if *inputPath == "" {
		flag.Usage()
		os.Exit(1)
	}

	// Initialize the workspace, which will return the repo and repo index.
	_, index, err := workspace.Init(*workspaceDir, workspace.InitOption{
		Force:      *force,
		RegenIndex: *regenIndex,
		RepoHash:   *repoHash,
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

	schemas := Schemas{}
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

		if err := schemas.AddSchema(schemaInput.SchemaLocator, *schema); err != nil {
			log.Fatal(err)
		}
	}

	if !*all {
		for _, schema := range schemas {
			cov, err := coverage.NewSchemaCoverageStore(schema)
			if err != nil {
				log.Fatal(err)
			}
			b, err = json.MarshalIndent(cov, "", "  ")
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(string(b))
		}
		return
	}

	// All
	cov := coverage.NewEmptyRepoCoverageStore(*index)
	for locator, schema := range schemas {
		if err := cov.AddSchema(schema, locator); err != nil {
			log.Fatal(err)
		}
	}

	b, err = json.MarshalIndent(cov, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))
}
