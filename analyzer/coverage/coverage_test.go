package coverage

import (
	"encoding/json"
	"github.com/magodo/azure-rest-api-cov/analyzer/schemainput"
	"github.com/magodo/azure-rest-api-cov/analyzer/workspace"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestCoverage_AddSchema(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	specBaseDir := filepath.Join(filepath.Dir(pwd), "testdata", "spec")
	schemaBasePath := filepath.Join(filepath.Dir(pwd), "testdata", "schema_input")
	cases := []struct {
		schemaInputs []string
		expect       string
	}{
		{
			schemaInputs: []string{filepath.Join(schemaBasePath, "def_base", "def_variant1.json"), filepath.Join(schemaBasePath, "def_base", "def_variant2.json")},
			expect: `
{
  "dummy": {
    "/{host}/{foo}/def_base": {
      "PUT": {
        "2020-12-01": {
          "name": "def_base",
          "swagger_rel_path": "dummy/resource-manager/Microsoft.Dummy/stable/2020-12-01/foo.json",
          "active": false,
          "properties": {
            "{def_variant1}/type": {
              "state": "Covered"
            },
            "{def_variant2}/type": {
              "state": "Covered"
            }
          }
        }
      }
    }
  }
}
`,
		},
	}

	index, err := workspace.NewAzureRepoIndex(specBaseDir, "")
	require.NoError(t, err)
	for idx, c := range cases {
		cov := NewEmptyCoverage()
		for _, schemaInput := range c.schemaInputs {
			b, err := ioutil.ReadFile(schemaInput)
			require.NoError(t, err, idx)

			var schemaInput schemainput.SchemaInput

			// Add active schemaInput
			err = json.Unmarshal(b, &schemaInput)
			require.NoError(t, err, idx)

			schema, err := schemaInput.ToSchema(index, true, false)
			require.NoError(t, err, idx)
			require.NoError(t, cov.AddSchema(schemaInput.SchemaLocator, *schema), idx)
		}

		b, err := json.MarshalIndent(cov, "", "  ")
		require.NoError(t, err, idx)

		require.JSONEq(t, c.expect, string(b), idx)
	}
}

func TestCoverage_ToSchemaInputs(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	specBaseDir := filepath.Join(filepath.Dir(pwd), "testdata", "spec")
	schemaBasePath := filepath.Join(filepath.Dir(pwd), "testdata", "schema_input")
	cases := []struct {
		schemaInputs []string
	}{
		{
			schemaInputs: []string{filepath.Join(schemaBasePath, "def_base", "def_variant1.json"), filepath.Join(schemaBasePath, "def_base", "def_variant2.json")},
		},
	}

	index, err := workspace.NewAzureRepoIndex(specBaseDir, "")
	require.NoError(t, err)
	for idx, c := range cases {
		var activeSchemaInputs []schemainput.SchemaInput
		for _, schemaInput := range c.schemaInputs {
			b, err := ioutil.ReadFile(schemaInput)
			require.NoError(t, err, idx)

			var schemaInput schemainput.SchemaInput

			// Add active schemaInput
			err = json.Unmarshal(b, &schemaInput)
			require.NoError(t, err, idx)

			activeSchemaInputs = append(activeSchemaInputs, schemaInput)
		}

		cov := NewEmptyCoverage()
		for _, schemaInput := range activeSchemaInputs {
			schema, err := schemaInput.ToSchema(index, true, false)
			require.NoError(t, err, idx)
			require.NoError(t, cov.AddSchema(schemaInput.SchemaLocator, *schema), idx)
		}

		// Since the re-constructed schemaInputs doesn't guarantee the order, so we instead check whether it is able to convert from the re-constructed schemaInputs
		// back to the Coverage.
		genSchemaInputs := cov.ToSchemaInputs()

		newCov := NewEmptyCoverage()
		for _, schemaInput := range genSchemaInputs {
			schema, err := schemaInput.ToSchema(index, true, false)
			require.NoError(t, err, idx)
			require.NoError(t, newCov.AddSchema(schemaInput.SchemaLocator, *schema), idx)
		}

		b, err := json.MarshalIndent(cov, "", "  ")
		require.NoError(t, err, idx)

		newb, err := json.MarshalIndent(newCov, "", "  ")
		require.NoError(t, err, idx)

		require.Equal(t, string(b), string(newb), idx)
	}
}

func TestCoverage_Predict(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	specBaseDir := filepath.Join(filepath.Dir(pwd), "testdata", "spec")
	schemaBasePath := filepath.Join(filepath.Dir(pwd), "testdata", "schema_input")
	cases := []struct {
		schemaInputs []string
		expect       string
	}{
		{
			schemaInputs: []string{filepath.Join(schemaBasePath, "def_base", "def_variant1.json"), filepath.Join(schemaBasePath, "def_base", "def_variant2.json")},
			expect: `
{
  "dummy": {
    "/{host}/{foo}/def_base": {
      "PUT": {
        "2020-12-01": {
          "name": "def_base",
          "swagger_rel_path": "dummy/resource-manager/Microsoft.Dummy/stable/2020-12-01/foo.json",
          "active": false,
          "properties": {
            "{def_variant1}/type": {
              "state": "Covered"
            },
            "{def_variant2}/type": {
              "state": "Covered"
            }
          }
        },
        "2021-01-01": {
          "name": "def_base",
          "swagger_rel_path": "dummy/resource-manager/Microsoft.Dummy/stable/2021-01-01/foo.json",
          "active": false,
          "properties": {
            "{def_variant1}/type": {
              "state": "Covered"
            },
            "{def_variant1}/new_prop": {
              "state": "Uncovered"
            },
            "{def_variant2}/type": {
              "state": "Covered"
            },
            "{def_variant2}/new_prop": {
              "state": "Uncovered"
            }
          }
        }
      }
    }
  }
}
`,
		},
	}

	index, err := workspace.NewAzureRepoIndex(specBaseDir, "")
	require.NoError(t, err)
	for idx, c := range cases {
		cov := NewEmptyCoverage()
		var activeSchemaInputs []schemainput.SchemaInput
		for _, schemaInput := range c.schemaInputs {
			b, err := ioutil.ReadFile(schemaInput)
			require.NoError(t, err, idx)

			var schemaInput schemainput.SchemaInput

			// Add active schemaInput
			err = json.Unmarshal(b, &schemaInput)
			require.NoError(t, err, idx)

			activeSchemaInputs = append(activeSchemaInputs, schemaInput)
		}

		for _, schemaInput := range activeSchemaInputs {
			schema, err := schemaInput.ToSchema(index, true, false)
			require.NoError(t, err, idx)
			require.NoError(t, cov.AddSchema(schemaInput.SchemaLocator, *schema), idx)
		}

		require.NoError(t, cov.Predict(index), idx)

		b, err := json.MarshalIndent(cov, "", "  ")
		require.NoError(t, err, idx)

		require.JSONEq(t, c.expect, string(b), idx)
	}
}
