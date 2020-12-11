package core

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/magodo/azure-rest-api-cov/pkg/core/swagger"

	"github.com/magodo/azure-rest-api-cov/pkg/core/schemainput"
	"github.com/magodo/azure-rest-api-cov/pkg/core/workspace"
	"github.com/stretchr/testify/require"
)

func TestSchemaMerge(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	specBaseDir := filepath.Join(filepath.Dir(pwd), "testdata", "spec")
	schemaBasePath := filepath.Join(filepath.Dir(pwd), "testdata", "schema_input")
	cases := []struct {
		schemaInputDir   string
		specRelPath      string
		expectProperties string
	}{
		{
			schemaInputDir: filepath.Join(schemaBasePath, "def_base"),
			expectProperties: `{
	"{def_variant1}.type": {
		"state": "Covered"
	},
	"{def_variant2}.type":{
		"state": "Covered"
	}
}`,
		},
	}

	index, err := workspace.NewAzureRepoIndex(specBaseDir)
	require.NoError(t, err)

	var mergedSchema *swagger.Schema
	for idx, c := range cases {
		err := filepath.Walk(c.schemaInputDir, func(path string, info os.FileInfo, err error) error {
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
			schema, err := schemaInput.ToSchema(specBaseDir, index.SpecLookupView()[schemaInput.SpecLocator])
			if err != nil {
				return err
			}
			if mergedSchema == nil {
				mergedSchema = schema
			} else {
				if err := mergedSchema.Merge(*schema); err != nil {
					return err
				}
			}
			return nil
		})
		require.NoError(t, err, idx)

		b, err := mergedSchema.Properties.MarshalJSON()
		require.NoError(t, err, idx)
		require.JSONEq(t, c.expectProperties, string(b), idx)
	}
}
