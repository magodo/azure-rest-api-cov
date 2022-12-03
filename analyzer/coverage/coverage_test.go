package coverage

import (
	"encoding/json"
	"github.com/magodo/azure-rest-api-cov/analyzer/swagger"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/magodo/azure-rest-api-cov/analyzer/schemainput"
	"github.com/magodo/azure-rest-api-cov/analyzer/workspace"
	"github.com/stretchr/testify/require"
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
			schemaInputs: []string{filepath.Join(schemaBasePath, "def_base", "def_variant1.json"),
				filepath.Join(schemaBasePath, "def_base", "def_variant2.json"),
				filepath.Join(schemaBasePath, "def_foo", "def_foo_delete.json")},
			expect: `
{
  "dummy": {
    "/{}/{}/DEF_BASE": {
      "operations": {
        "PUT": {
          "2020-12-01": {
            "is_active": true,
            "is_preview": false,
            "swagger_rel_path": "dummy/resource-manager/Microsoft.Dummy/stable/2020-12-01/foo.json",
            "model": {
              "name": "def_base",
              "swagger_rel_path": "dummy/resource-manager/Microsoft.Dummy/stable/2020-12-01/foo.json",
              "properties": {
                "{def_base}/type": {
                  "state": "Uncovered"
                },
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
    },
    "/{}/{}/DEF_FOO": {
      "operations": {
        "DELETE": {
          "2020-12-01": {
            "is_active": true,
            "is_preview": false,
            "swagger_rel_path": "dummy/resource-manager/Microsoft.Dummy/stable/2020-12-01/foo.json"
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

			opInfo, err := schemaInput.ToOperationInfo(index, true, true)
			require.NoError(t, err, idx)
			require.NoError(t, cov.AddOperation(opInfo), idx)
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
			schemaInputs: []string{
				filepath.Join(schemaBasePath, "def_base", "def_variant1.json"),
				filepath.Join(schemaBasePath, "def_base", "def_variant2.json"),
				filepath.Join(schemaBasePath, "def_foo", "def_foo_delete.json"),
			},
		},
		{
			schemaInputs: []string{
				filepath.Join(schemaBasePath, "def_base", "def_variant1_new_ver.json"),
				filepath.Join(schemaBasePath, "def_base", "def_variant2_new_ver.json"),
			},
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
			opInfo, err := schemaInput.ToOperationInfo(index, true, false)
			require.NoError(t, err, idx)
			require.NoError(t, cov.AddOperation(opInfo), idx)
		}

		// Since the re-constructed schemaInputs doesn't guarantee the order, so we instead check whether it is able to convert from the re-constructed schemaInputs
		// back to the Coverage.
		genSchemaInputs := cov.ToSchemaInputs()

		newCov := NewEmptyCoverage()
		for _, schemaInput := range genSchemaInputs {
			opInfo, err := schemaInput.ToOperationInfo(index, true, false)
			require.NoError(t, err, idx)
			require.NoError(t, newCov.AddOperation(opInfo), idx)
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
			schemaInputs: []string{filepath.Join(schemaBasePath, "def_base", "def_variant1.json"),
				filepath.Join(schemaBasePath, "def_base", "def_variant2.json"),
				filepath.Join(schemaBasePath, "def_foo", "def_foo_delete.json")},
			expect: `
{
  "dummy": {
    "/{}/{}/DEF_FOO": {
      "operations": {
        "DELETE": {
          "2020-12-01": {
            "is_active": true,
            "is_preview": false,
            "swagger_rel_path": "dummy/resource-manager/Microsoft.Dummy/stable/2020-12-01/foo.json"
          },
          "2021-01-01": {
            "is_active": false,
            "is_preview": false,
            "swagger_rel_path": "dummy/resource-manager/Microsoft.Dummy/stable/2021-01-01/foo.json"
          }
        }
      }
    },
    "/{}/{}/DEF_BASE": {
      "operations": {
        "PUT": {
          "2020-12-01": {
            "is_active": true,
            "is_preview": false,
            "swagger_rel_path": "dummy/resource-manager/Microsoft.Dummy/stable/2020-12-01/foo.json",
            "model": {
              "name": "def_base",
              "swagger_rel_path": "dummy/resource-manager/Microsoft.Dummy/stable/2020-12-01/foo.json",
              "properties": {
                "{def_base}/type": {
                  "state": "Uncovered"
                },
                "{def_variant1}/type": {
                  "state": "Covered"
                },
                "{def_variant2}/type": {
                  "state": "Covered"
                }
              }
            }
          },
          "2021-01-01": {
            "is_active": false,
            "is_preview": false,
            "swagger_rel_path": "dummy/resource-manager/Microsoft.Dummy/stable/2021-01-01/foo.json",
            "model": {
              "name": "def_base",
              "swagger_rel_path": "dummy/resource-manager/Microsoft.Dummy/stable/2021-01-01/foo.json",
              "properties": {
                "{def_base}/type": {
                  "state": "Uncovered"
                },
                "{def_variant1}/new_prop": {
                  "state": "Uncovered"
                },
                "{def_variant1}/type": {
                  "state": "Covered"
                },
                "{def_variant2}/new_prop": {
                  "state": "Uncovered"
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
			opInfo, err := schemaInput.ToOperationInfo(index, true, true)
			require.NoError(t, err, idx)
			require.NoError(t, cov.AddOperation(opInfo), idx)
		}

		require.NoError(t, cov.Predict(index), idx)

		b, err := json.MarshalIndent(cov, "", "  ")
		require.NoError(t, err, idx)

		require.JSONEq(t, c.expect, string(b), idx)
	}
}

func TestCoverage_Unmarshal(t *testing.T) {
	cases := []struct {
		input  string
		expect Coverage
	}{
		{
			input: `
{
  "dummy": {
    "/{}/{}/DEF_BASE": {
      "operations": {
        "PUT": {
          "2020-12-01": {
            "is_active": true,
            "is_preview": false,
            "swagger_rel_path": "dummy/resource-manager/Microsoft.Dummy/stable/2020-12-01/foo.json",
            "model": {
              "name": "def_base",
              "swagger_rel_path": "dummy/resource-manager/Microsoft.Dummy/stable/2020-12-01/foo.json",
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
    },
    "/{}/{}/DEF_FOO": {
      "operations": {
        "DELETE": {
          "2020-12-01": {
            "is_active": true,
            "is_preview": false,
            "swagger_rel_path": "dummy/resource-manager/Microsoft.Dummy/stable/2020-12-01/foo.json"
          }
        }
      }
    }
  }
}
`,
			expect: Coverage{
				"dummy": &APIPaths{
					"/{}/{}/DEF_BASE": &APIPath{
						Operations: Operations{
							"PUT": &APIVersions{
								"2020-12-01": {
									RP:             "dummy",
									APIPath:        "/{}/{}/DEF_BASE",
									APIVersion:     "2020-12-01",
									Operation:      workspace.OperationKindPut,
									IsActive:       true,
									IsPreview:      false,
									SwaggerRelPath: "dummy/resource-manager/Microsoft.Dummy/stable/2020-12-01/foo.json",
									Model: &swagger.Schema{
										SchemaMetaInfo: swagger.SchemaMetaInfo{
											Name:           "def_base",
											SwaggerRelPath: "dummy/resource-manager/Microsoft.Dummy/stable/2020-12-01/foo.json",
										},
										Properties: map[string]*swagger.Property{
											"{def_variant1}/type": {
												PropertyState: swagger.PropertyStateCovered,
											},
											"{def_variant2}/type": {
												PropertyState: swagger.PropertyStateCovered,
											},
										},
									},
								},
							},
						},
					},
					"/{}/{}/DEF_FOO": &APIPath{
						Operations: Operations{
							"DELETE": &APIVersions{
								"2020-12-01": {
									RP:             "dummy",
									APIPath:        "/{}/{}/DEF_FOO",
									APIVersion:     "2020-12-01",
									Operation:      workspace.OperationKindDelete,
									IsActive:       true,
									IsPreview:      false,
									SwaggerRelPath: "dummy/resource-manager/Microsoft.Dummy/stable/2020-12-01/foo.json",
								},
							},
						},
					},
				},
			},
		},
	}

	for idx, c := range cases {
		var actual Coverage
		require.NoError(t, json.Unmarshal([]byte(c.input), &actual), idx)
		require.Equal(t, c.expect, actual)
	}
}
