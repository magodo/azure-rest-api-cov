package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/magodo/azure-rest-api-cov/pkg/core/swagger"
)

const usage = `azure-rest-api-cov <option>`

func main() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "%s\n", usage)
		flag.PrintDefaults()
	}

	swaggerSpecPath := flag.String("spec-base-dir", "", "The path to the swagger spec directory (.../azure-rest-api-specs/specification)")
	inputPath := flag.String("schema-cov-dir", "", "The path to the directory contains schema input files")

	flag.Parse()

	if *swaggerSpecPath == "" ||
		*inputPath == "" {
		flag.Usage()
		os.Exit(1)
	}

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
		var schemaInput swagger.SchemaInput
		if err := json.Unmarshal(b, &schemaInput); err != nil {
			return err
		}
		schema, err := schemaInput.ToSchema(*swaggerSpecPath)
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
