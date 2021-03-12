package report_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/magodo/azure-rest-api-cov/analyzer/schemainput"

	"github.com/magodo/azure-rest-api-cov/analyzer/report"
	"github.com/magodo/azure-rest-api-cov/analyzer/workspace"

	"github.com/stretchr/testify/require"
)

func TestNewEmptyCoverageStore(t *testing.T) {
	pwd, err := os.Getwd()
	require.NoError(t, err, "pwd")
	specRootDir := filepath.Join(filepath.Dir(pwd), "testdata", "spec")

	expect := report.Coverage{
		Commit: "xxx",
		ResourceProviders: report.ResourceProviders{
			"dummy": report.APIPaths{
				"/{host}/{bar}": report.Operations{},
				"/{host}/{foo}/def_foo": report.Operations{
					"PUT": report.APIVersions{
						"2020-12-01": nil,
					},
					"GET": report.APIVersions{
						"2020-12-01": nil,
						"2021-01-01": nil,
					},
				},
				"/{host}/{foo}/def_bar": report.Operations{
					"POST": report.APIVersions{
						"2020-12-01": nil,
					},
				},
				"/{host}/{foo}/def_base": report.Operations{
					"PUT": report.APIVersions{
						"2020-12-01": nil,
						"2021-01-01": nil,
					},
				},
				"/{host}/{foo}/def_regular": report.Operations{
					"GET": report.APIVersions{
						"2020-12-01": nil,
						"2021-01-01": nil,
					},
					"PUT": report.APIVersions{
						"2020-12-01": nil,
						"2021-01-01": nil,
					},
				},
			},
		},
	}

	idx, err := workspace.NewAzureRepoIndex(specRootDir, "xxx")
	require.NoError(t, err, "generate index")
	cov := report.NewEmptyCoverage(*idx, "")
	require.Equal(t, expect, cov)
}

func TestCoverageStoreAddSchema(t *testing.T) {
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
      "/{host}/{foo}/def_bar": {
        "POST": {
          "2020-12-01": null
        }
      },
      "/{host}/{foo}/def_base": {
        "PUT": {
          "2020-12-01": {
            "name": "def_base",
            "url": "prefix/dummy/resource-manager/Microsoft.Dummy/stable/2020-12-01/foo.json",
            "active": false,
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
                  "covered_amount": 0,
                  "children": {
                    "type": {
                      "total_amount": 1,
                      "covered_amount": 0
                    }
                  }
                }
              }
            }
          },
          "2021-01-01": {
            "name": "def_base",
            "url": "prefix/dummy/resource-manager/Microsoft.Dummy/stable/2021-01-01/foo.json",
            "active": false,
            "node": {
              "total_amount": 4,
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
                  "total_amount": 2,
                  "covered_amount": 0,
                  "children": {
                    "new_prop": {
                      "total_amount": 1,
                      "covered_amount": 0
                    },
                    "type": {
                      "total_amount": 1,
                      "covered_amount": 0
                    }
                  }
                }
              }
            }
          }
        }
      },
      "/{host}/{foo}/def_foo": {
        "GET": {
          "2020-12-01": null,
          "2021-01-01": null
        },
        "PUT": {
          "2020-12-01": null
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

		cov := report.NewEmptyCoverage(*index, "")
		var schemaInput schemainput.SchemaInput

		// Add active schemaInput
		err = json.Unmarshal(b, &schemaInput)
		require.NoError(t, err, idx)

		schema, err := schemaInput.ToSchema(index, true, false)
		require.NoError(t, err, idx)
		require.NoError(t, cov.AddSchema("prefix/", *schema, schemaInput.SchemaLocator), idx)

		// Add in-active schemaInput
		schemaInput.ApiVersion = "2021-01-01"
		schema, err = schemaInput.ToSchema(index, true, false)
		require.NoError(t, err, idx)
		require.NoError(t, cov.AddSchema("prefix/", *schema, schemaInput.SchemaLocator), idx)

		b, err = json.MarshalIndent(cov, "", "  ")
		require.NoError(t, err, idx)

		require.JSONEq(t, c.expect, string(b), idx)
	}
}

func TestCoverageStoreMarshal(t *testing.T) {
	cases := []struct {
		input  report.Coverage
		expect string
	}{
		{
			input: report.Coverage{
				AdditionalMessage: "source: xxx",
				Commit:            "xxx",
				ResourceProviders: report.ResourceProviders{
					"rp1": report.APIPaths{
						"path1": report.Operations{
							"PUT": report.APIVersions{
								"version1": &report.SchemaCoverage{
									Active: true,
									Name:   "schema1",
									URL:    "foo/bar.json",
									Node: report.PropertyCoverageNode{
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
  "additional_message": "source: xxx",
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

func TestCoverageStoreUnmarshal(t *testing.T) {
	cases := []struct {
		input  string
		expect report.Coverage
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
			expect: report.Coverage{
				Commit: "xxx",
				ResourceProviders: report.ResourceProviders{
					"rp1": report.APIPaths{
						"path1": report.Operations{
							"PUT": report.APIVersions{
								"version1": &report.SchemaCoverage{
									Active: true,
									Name:   "schema1",
									URL:    "foo/bar.json",
									Node: report.PropertyCoverageNode{
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
		cov := report.Coverage{}
		err := json.Unmarshal([]byte(c.input), &cov)
		require.NoError(t, err, i)
		require.Equal(t, c.expect, cov, i)
	}
}
