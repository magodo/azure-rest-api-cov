package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"github.com/magodo/azure-rest-api-cov/pkg/core/swaggerloader"

	log "github.com/sirupsen/logrus"

	"github.com/magodo/azure-rest-api-cov/pkg/core/swagger"

	"github.com/magodo/azure-rest-api-cov/pkg/core/coverage"

	"github.com/magodo/azure-rest-api-cov/pkg/core/schemainput"

	"github.com/magodo/azure-rest-api-cov/pkg/core/workspace"
)

const usage = `azure-rest-api-cov -i <schema input> <option>`

type SchemaKey struct {
	Active        bool
	SchemaLocator workspace.SchemaLocator
}

type SchemaKeys []SchemaKey

func (keys SchemaKeys) Len() int {
	return len(keys)
}

func (keys SchemaKeys) Less(i, j int) bool {
	if keys[i].SchemaLocator.ApiPath < keys[j].SchemaLocator.ApiPath {
		return true
	}
	if keys[i].SchemaLocator.ApiPath > keys[j].SchemaLocator.ApiPath {
		return false
	}
	if keys[i].SchemaLocator.ApiVersion < keys[j].SchemaLocator.ApiVersion {
		return true
	}
	if keys[i].SchemaLocator.ApiVersion > keys[j].SchemaLocator.ApiVersion {
		return false
	}
	return keys[i].SchemaLocator.Operation < keys[j].SchemaLocator.Operation
}

func (keys SchemaKeys) Swap(i, j int) {
	keys[i], keys[j] = keys[j], keys[i]
}

type Schemas map[SchemaKey]swagger.Schema

func (schemas Schemas) AddSchema(locator workspace.SchemaLocator, s swagger.Schema, active bool) error {
	schemaKey := SchemaKey{
		Active:        active,
		SchemaLocator: locator,
	}
	existing, ok := schemas[schemaKey]
	if !ok {
		schemas[schemaKey] = s
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
	// Required options
	inputPath := flag.String("i", "",
		"The path to the schema input file, which is expected to hold a JSON array of input schemas")

	// Workspace related options
	workspaceDir := flag.String("w", filepath.Join(cacheDir, "azure-rest-api-cov"),
		"The path to the workspace for caching the azure-rest-api repository and other stuffs")
	force := flag.Bool("f", false, "whether to force create the workspace")
	regenIndex := flag.Bool("r", false, "whether to force regenerate the repository index in the workspace")
	repoHash := flag.String("c", "", "The azure-rest-api-spec repo commit hash to checkout")

	// Behavior options
	all := flag.Bool("a", false, "whether to export the RP level coverage")
	predictVersion := flag.Bool("p", false, "whether to predict coverage for newer API versions")

	// Other options
	debug := flag.Bool("d", false, "Whether to show debug log")

	// Parse flag
	flag.Parse()

	// Initialize logger
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})
	if *debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

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
		logCurrentProcessingSchemaInput(schemaInput, false)
		schema, err := schemaInput.ToSchema(index, true)
		if err != nil {
			if errors.Is(err, workspace.ErrFailedToLocateSwagger) {
				log.Warn(err)
				continue
			}
			log.Fatal(err)
		}

		if err := schemas.AddSchema(schemaInput.SchemaLocator, *schema, true); err != nil {
			log.Fatal(err)
		}
	}

	if *predictVersion {
		explicitSchemas := map[workspace.SchemaLocator]bool{}
		for _, schemaInput := range schemaInputs {
			explicitSchemas[schemaInput.SchemaLocator] = true
		}

	loop:
		for _, schemaInput := range schemaInputs {
			specPath, err := index.LocateSpec(schemaInput.SpecLocator)
			if err != nil {
				if errors.Is(err, workspace.ErrFailedToLocateSwagger) {
					log.Warn(err)
					continue
				}
				log.Fatal(err)
			}

			pathInfo, err := specPath.PathInfo()
			if err != nil {
				log.Fatal(err)
			}

			swg, err := swaggerloader.LoadSwagger(filepath.Join(index.SpecRootDir, string(specPath)))
			if err != nil {
				log.Fatal(err)
			}

			rp, currentVersion := pathInfo.ResourceProvider, swg.Info.Version

			var newApiVersions []workspace.APIVersion
			for apiVersion := range index.ResourceProviders[workspace.ResourceProvider(rp)] {
				if string(apiVersion) > currentVersion {
					// Modify the API version to shift to the newer ones
					schemaInput.SchemaLocator.SpecLocator.ApiVersion = string(apiVersion)

					// If at least one newer version has already been explicitly specified via schemaInput files, then skip prediction based on current
					// schemaInput. We should use the latest existing schemaInput to do the prediction on newer versions.
					if explicitSchemas[schemaInput.SchemaLocator] {
						continue loop
					}
					newApiVersions = append(newApiVersions, apiVersion)
				}
			}

			for _, apiVersion := range newApiVersions {
				// Modify the API version to shift to the newer ones
				schemaInput.SchemaLocator.SpecLocator.ApiVersion = string(apiVersion)

				logCurrentProcessingSchemaInput(schemaInput, true)
				schema, err := schemaInput.ToSchema(index, false)
				if err != nil {
					if errors.Is(err, workspace.ErrFailedToLocateSwagger) || errors.Is(err, workspace.ErrFailedToLocateSchema) {
						log.Warn(err)
						continue
					}
					log.Fatal(err)
				}
				if err := schemas.AddSchema(schemaInput.SchemaLocator, *schema, false); err != nil {
					log.Fatal(err)
				}
			}
		}
	}

	if !*all {
		var keys SchemaKeys
		for k := range schemas {
			keys = append(keys, k)
		}
		sort.Sort(keys)

		var covs []*coverage.SchemaCoverageStore
		for _, k := range keys {
			schema := schemas[k]
			cov, err := coverage.NewSchemaCoverageStore(schema, k.Active)
			if err != nil {
				log.Fatal(err)
			}
			covs = append(covs, cov)
		}

		b, err = json.MarshalIndent(covs, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(string(b))
		return
	}

	// All
	cov := coverage.NewEmptyRepoCoverageStore(*index)
	for k, schema := range schemas {
		if err := cov.AddSchema(schema, k.SchemaLocator, k.Active); err != nil {
			log.Fatal(err)
		}
	}

	b, err = json.MarshalIndent(cov, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(b))
}

func logCurrentProcessingSchemaInput(schemaInput schemainput.SchemaInput, predict bool) {
	decoration := ""
	if predict {
		decoration = " (predict)"
	}
	msg := fmt.Sprintf("Processing%s: %s (%s - %s)", decoration, schemaInput.ApiPath, schemaInput.Operation, schemaInput.ApiVersion)
	if schemaInput.Variant != "" {
		msg = fmt.Sprintf("%s [variant: %s]", msg, schemaInput.Variant)
	}
	log.Infoln(msg)
}
