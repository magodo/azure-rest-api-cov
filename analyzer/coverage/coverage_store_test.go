package coverage_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/magodo/azure-rest-api-cov/analyzer/schemainput"

	"github.com/magodo/azure-rest-api-cov/analyzer/coverage"
	"github.com/magodo/azure-rest-api-cov/analyzer/workspace"

	"github.com/stretchr/testify/require"
)

func TestNewEmptyRepoCoverageStore(t *testing.T) {
	pwd, err := os.Getwd()
	require.NoError(t, err, "pwd")
	specRootDir := filepath.Join(filepath.Dir(pwd), "testdata", "spec")

	expect := coverage.RepoCoverageStore{
		Commit: "xxx",
		ResourceProviders: coverage.ResourceProviders{
			"dummy": coverage.APIPaths{
				"/{host}/{bar}": coverage.Operations{},
				"/{host}/{foo}/def_foo": coverage.Operations{
					"PUT": coverage.APIVersions{
						"2020-12-01": nil,
					},
					"GET": coverage.APIVersions{
						"2020-12-01": nil,
						"2021-01-01": nil,
					},
				},
				"/{host}/{foo}/def_bar": coverage.Operations{
					"POST": coverage.APIVersions{
						"2020-12-01": nil,
					},
				},
				"/{host}/{foo}/def_base": coverage.Operations{
					"PUT": coverage.APIVersions{
						"2020-12-01": nil,
						"2021-01-01": nil,
					},
				},
				"/{host}/{foo}/def_regular": coverage.Operations{
					"GET": coverage.APIVersions{
						"2020-12-01": nil,
						"2021-01-01": nil,
					},
					"PUT": coverage.APIVersions{
						"2020-12-01": nil,
						"2021-01-01": nil,
					},
				},
			},
		},
	}

	idx, err := workspace.NewAzureRepoIndex(specRootDir, "xxx")
	require.NoError(t, err, "generate index")
	cov := coverage.NewEmptyRepoCoverageStore(*idx)
	require.Equal(t, expect, cov)
}

func TestRepoCoverageStoreAddSchema(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	specBaseDir := filepath.Join(filepath.Dir(pwd), "testdata", "spec")
	schemaBasePath := filepath.Join(filepath.Dir(pwd), "testdata", "schema_input")
	cases := []struct {
		schemaInput string
		expect      string
	}{
		{
			schemaInput: filepath.Join(schemaBasePath, "def_base", "def_variant1.json"),
			expect: `
{
  "commit": "",
  "resource_providers": {
    "dummy": {
      "/{host}/{bar}": {},
      "/{host}/{foo}/def_foo": {
	    "GET": {
          "2020-12-01": null,
          "2021-01-01": null
        },
	    "PUT": {
          "2020-12-01": null
        }
      },
      "/{host}/{foo}/def_bar": {
	    "POST": {
          "2020-12-01": null
        }
      },
      "/{host}/{foo}/def_base": {
       "PUT": {
          "2020-12-01": {
			"active": true,
            "name": "def_base",
            "url": "prefix/dummy/resource-manager/Microsoft.Dummy/stable/2020-12-01/foo.json",
            "node": {
              "total_amount": 2,
              "covered_amount": 1,
              "children": {
                "{def_variant1}": {
                  "total_amount": 1,
                  "covered_amount": 1,
                  "children": {
                    "type": {
                      "total_amount": 1,
                      "covered_amount": 1
                    }
                  }
                },
                "{def_variant2}": {
                  "total_amount": 1,
                  "covered_amount": 0
                }
              }
            }
          },
          "2021-01-01": {
			"active": false,
            "name": "def_base",
            "url": "prefix/dummy/resource-manager/Microsoft.Dummy/stable/2021-01-01/foo.json",
            "node": {
              "total_amount": 3,
              "covered_amount": 1,
              "children": {
                "{def_variant1}": {
                  "total_amount": 2,
                  "covered_amount": 1,
                  "children": {
					"new_prop": {
                      "total_amount": 1,
                      "covered_amount": 0
					},
                    "type": {
                      "total_amount": 1,
                      "covered_amount": 1
                    }
                  }
                },
                "{def_variant2}": {
                  "total_amount": 1,
                  "covered_amount": 0
                }
              }
            }
          }
        }
      },
      "/{host}/{foo}/def_regular": {
        "GET": {
          "2020-12-01": null,
          "2021-01-01": null
        },
        "PUT": {
          "2020-12-01": null,
          "2021-01-01": null
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
		b, err := ioutil.ReadFile(c.schemaInput)
		require.NoError(t, err, idx)

		cov := coverage.NewEmptyRepoCoverageStore(*index)
		var schemaInput schemainput.SchemaInput

		// Add active schemaInput
		err = json.Unmarshal(b, &schemaInput)
		require.NoError(t, err, idx)

		schema, err := schemaInput.ToSchema(index, true)
		require.NoError(t, err, idx)
		require.NoError(t, cov.AddSchema("prefix/", *schema, schemaInput.SchemaLocator, true), idx)

		// Add in-active schemaInput
		schemaInput.ApiVersion = "2021-01-01"
		schema, err = schemaInput.ToSchema(index, true)
		require.NoError(t, err, idx)
		require.NoError(t, cov.AddSchema("prefix/", *schema, schemaInput.SchemaLocator, false), idx)

		b, err = json.MarshalIndent(cov, "", "  ")
		require.NoError(t, err, idx)

		require.JSONEq(t, c.expect, string(b), idx)
	}
}

func TestRepoCoverageStoreMarshal(t *testing.T) {
	cases := []struct {
		input  coverage.RepoCoverageStore
		expect string
	}{
		{
			input: coverage.RepoCoverageStore{
				Commit: "xxx",
				ResourceProviders: coverage.ResourceProviders{
					"rp1": coverage.APIPaths{
						"path1": coverage.Operations{
							"PUT": coverage.APIVersions{
								"version1": &coverage.SchemaCoverageStore{
									Active: true,
									Name:   "schema1",
									URL:    "foo/bar.json",
									Node: coverage.PropertyCoverageNode{
										TotalAmount:   1,
										CoveredAmount: 1,
									},
								},
								"version2": nil,
							},
						},
					},
				},
			},
			expect: `
{
  "commit": "xxx",
  "resource_providers": {
    "rp1": {
      "path1": {
		"PUT": {
          "version1": {
			"active": true,
            "name": "schema1",
            "url": "foo/bar.json",
            "node": {
              "total_amount": 1,
              "covered_amount": 1
            }
          },
          "version2": null
		}
      }
    }
  }
}
`,
		},
	}

	for i, c := range cases {
		b, err := json.MarshalIndent(c.input, "", "  ")
		require.NoError(t, err, i)
		require.JSONEq(t, c.expect, string(b), i)
	}
}

func TestRepoCoverageStoreUnmarshal(t *testing.T) {
	cases := []struct {
		input  string
		expect coverage.RepoCoverageStore
	}{
		{
			input: `
{
  "commit": "xxx",
  "resource_providers": {
    "rp1": {
      "path1": {
		"PUT": {
          "version1": {
			"active": true,
            "name": "schema1",
            "url": "foo/bar.json",
            "node": {
              "total_amount": 1,
              "covered_amount": 1
            }
          },
          "version2": null
		}
      }
    }
  }
}
`,
			expect: coverage.RepoCoverageStore{
				Commit: "xxx",
				ResourceProviders: coverage.ResourceProviders{
					"rp1": coverage.APIPaths{
						"path1": coverage.Operations{
							"PUT": coverage.APIVersions{
								"version1": &coverage.SchemaCoverageStore{
									Active: true,
									Name:   "schema1",
									URL:    "foo/bar.json",
									Node: coverage.PropertyCoverageNode{
										TotalAmount:   1,
										CoveredAmount: 1,
									},
								},
								"version2": nil,
							},
						},
					},
				},
			},
		},
	}

	for i, c := range cases {
		cov := coverage.RepoCoverageStore{}
		err := json.Unmarshal([]byte(c.input), &cov)
		require.NoError(t, err, i)
		require.Equal(t, c.expect, cov, i)
	}
}
