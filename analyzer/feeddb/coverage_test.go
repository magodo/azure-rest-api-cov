package feeddb_test

import (
	"encoding/json"
	"github.com/magodo/azure-rest-api-cov/analyzer/feeddb"
	"github.com/magodo/azure-rest-api-cov/analyzer/schemainput"
	"github.com/magodo/azure-rest-api-cov/analyzer/workspace"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestNewEmptyCoverage(t *testing.T) {
	pwd, err := os.Getwd()
	require.NoError(t, err, "pwd")
	specRootDir := filepath.Join(filepath.Dir(pwd), "testdata", "spec")

	expect := feeddb.Coverage{
		Ref:     "xxx",
		Comment: "comment",
		ResourceProviders: feeddb.ResourceProviders{
			"dummy": feeddb.ResourceProvider{
				Name: "dummy",
				Paths: feeddb.APIPaths{
					"/{}/{}": feeddb.APIPath{
						Name:       "/{}/{}",
						Operations: feeddb.Operations{},
					},
					"/{}/{}/DEF_FOO": feeddb.APIPath{
						Name: "/{}/{}/DEF_FOO",
						Operations: feeddb.Operations{
							"PUT": feeddb.Operation{
								Name: "PUT",
								Versions: feeddb.APIVersions{
									"2020-12-01": feeddb.APIVersion{
										Version: "2020-12-01",
									},
								},
							},
							"GET": feeddb.Operation{
								Name: "GET",
								Versions: feeddb.APIVersions{
									"2020-12-01": feeddb.APIVersion{
										Version: "2020-12-01",
									},
									"2021-01-01": feeddb.APIVersion{
										Version: "2021-01-01",
									},
								},
							},
							"DELETE": feeddb.Operation{
								Name: "DELETE",
								Versions: feeddb.APIVersions{
									"2020-12-01": feeddb.APIVersion{
										Version: "2020-12-01",
									},
									"2021-01-01": feeddb.APIVersion{
										Version: "2021-01-01",
									},
								},
							},
						},
					},
					"/{}/{}/DEF_BAR": feeddb.APIPath{
						Name: "/{}/{}/DEF_BAR",
						Operations: feeddb.Operations{
							"POST": {
								Name: "POST",
								Versions: feeddb.APIVersions{
									"2020-12-01": {
										Version: "2020-12-01",
									},
								},
							},
						},
					},
					"/{}/{}/DEF_BASE": feeddb.APIPath{
						Name: "/{}/{}/DEF_BASE",
						Operations: feeddb.Operations{
							"PUT": {
								Name: "PUT",
								Versions: feeddb.APIVersions{
									"2020-12-01": {
										Version: "2020-12-01",
									},
									"2021-01-01": {
										Version: "2021-01-01",
									},
								},
							},
						},
					},
					"/{}/{}/DEF_REGULAR": feeddb.APIPath{
						Name: "/{}/{}/DEF_REGULAR",
						Operations: feeddb.Operations{
							"GET": {
								Name: "GET",
								Versions: feeddb.APIVersions{
									"2020-12-01": {
										Version: "2020-12-01",
									},
									"2021-01-01": {
										Version: "2021-01-01",
									},
								},
							},
							"PUT": {
								Name: "PUT",
								Versions: feeddb.APIVersions{
									"2020-12-01": {
										Version: "2020-12-01",
									},
									"2021-01-01": {
										Version: "2021-01-01",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	idx, err := workspace.NewAzureRepoIndex(specRootDir, "xxx")
	require.NoError(t, err, "generate index")
	cov := feeddb.NewEmptyCoverage(idx, "comment")
	require.Equal(t, expect, cov)
}

func TestCoverageAddSchema(t *testing.T) {
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
  "ref": "",
  "comment": "comment",
  "resourceProviders": {
    "dummy": {
      "name": "dummy",
      "paths": {
        "/{}/{}": {
          "name": "/{}/{}",
          "operations": {}
        },
        "/{}/{}/DEF_BAR": {
          "name": "/{}/{}/DEF_BAR",
          "operations": {
            "POST": {
              "name": "POST",
              "versions": {
                "2020-12-01": {
                  "version": "2020-12-01",
                  "isActive": false,
                  "isPreview": false
                }
              }
            }
          }
        },
        "/{}/{}/DEF_BASE": {
          "name": "/{}/{}/DEF_BASE",
          "operations": {
            "PUT": {
              "name": "PUT",
              "versions": {
                "2020-12-01": {
                  "version": "2020-12-01",
                  "isActive": true,
                  "isPreview": false,
                  "model": {
                    "name": "def_base",
                    "properties": {
                      "{def_variant1}": {
                        "properties": {
                          "type": {
                            "isCovered": true
                          }
                        }
                      },
                      "{def_variant2}": {
                        "properties": {
                          "type": {
                            "isCovered": false
                          }
                        }
                      }
                    }
                  }
                },
                "2021-01-01": {
                  "version": "2021-01-01",
                  "isActive": false,
                  "isPreview": false,
                  "model": {
                    "name": "def_base",
                    "properties": {
                      "{def_variant1}": {
                        "properties": {
                          "new_prop": {
                            "isCovered": false
                          },
                          "type": {
                            "isCovered": true
                          }
                        }
                      },
                      "{def_variant2}": {
                        "properties": {
                          "new_prop": {
                            "isCovered": false
                          },
                          "type": {
                            "isCovered": false
                          }
                        }
                      }
                    }
                  }
                }
              }
            }
          }
        },
        "/{}/{}/DEF_FOO": {
          "name": "/{}/{}/DEF_FOO",
          "operations": {
            "DELETE": {
              "name": "DELETE",
              "versions": {
                "2020-12-01": {
                  "version": "2020-12-01",
                  "isActive": false,
                  "isPreview": false
                },
                "2021-01-01": {
                  "version": "2021-01-01",
                  "isActive": false,
                  "isPreview": false
                }
              }
            },
            "GET": {
              "name": "GET",
              "versions": {
                "2020-12-01": {
                  "version": "2020-12-01",
                  "isActive": false,
                  "isPreview": false
                },
                "2021-01-01": {
                  "version": "2021-01-01",
                  "isActive": false,
                  "isPreview": false
                }
              }
            },
            "PUT": {
              "name": "PUT",
              "versions": {
                "2020-12-01": {
                  "version": "2020-12-01",
                  "isActive": false,
                  "isPreview": false
                }
              }
            }
          }
        },
        "/{}/{}/DEF_REGULAR": {
          "name": "/{}/{}/DEF_REGULAR",
          "operations": {
            "GET": {
              "name": "GET",
              "versions": {
                "2020-12-01": {
                  "version": "2020-12-01",
                  "isActive": false,
                  "isPreview": false
                },
                "2021-01-01": {
                  "version": "2021-01-01",
                  "isActive": false,
                  "isPreview": false
                }
              }
            },
            "PUT": {
              "name": "PUT",
              "versions": {
                "2020-12-01": {
                  "version": "2020-12-01",
                  "isActive": false,
                  "isPreview": false
                },
                "2021-01-01": {
                  "version": "2021-01-01",
                  "isActive": false,
                  "isPreview": false
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
		b, err := ioutil.ReadFile(c.schemaInput)
		require.NoError(t, err, idx)

		cov := feeddb.NewEmptyCoverage(index, "comment")
		var schemaInput schemainput.SchemaInput

		// Add active schemaInput
		err = json.Unmarshal(b, &schemaInput)
		require.NoError(t, err, idx)

		schema, err := schemaInput.ToSchema(index, true, true)
		require.NoError(t, err, idx)
		cov.AddSchema(*schema, schemaInput.SchemaLocator)

		// Add in-active schemaInput
		schemaInput.ApiVersion = "2021-01-01"
		schema, err = schemaInput.ToSchema(index, true, false)
		require.NoError(t, err, idx)
		cov.AddSchema(*schema, schemaInput.SchemaLocator)

		b, err = json.MarshalIndent(cov, "", "  ")
		require.NoError(t, err, idx)

		require.JSONEq(t, c.expect, string(b), idx)
	}
}

func TestCoverageIncludeUncoveredOperationModels(t *testing.T) {
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
  "ref": "",
  "comment": "comment",
  "resourceProviders": {
    "dummy": {
      "name": "dummy",
      "paths": {
        "/{}/{}": {
          "name": "/{}/{}",
          "operations": {}
        },
        "/{}/{}/DEF_BAR": {
          "name": "/{}/{}/DEF_BAR",
          "operations": {
            "POST": {
              "name": "POST",
              "versions": {
                "2020-12-01": {
                  "version": "2020-12-01",
                  "isActive": false,
                  "isPreview": true,
                  "model": {
                    "name": "def_bar",
                    "properties": {
                      "bar": {
                        "isCovered": false
                      }
                    }
                  }
                }
              }
            }
          }
        },
        "/{}/{}/DEF_BASE": {
          "name": "/{}/{}/DEF_BASE",
          "operations": {
            "PUT": {
              "name": "PUT",
              "versions": {
                "2020-12-01": {
                  "version": "2020-12-01",
                  "isActive": true,
                  "isPreview": false,
                  "model": {
                    "name": "def_base",
                    "properties": {
                      "{def_variant1}": {
                        "properties": {
                          "type": {
                            "isCovered": true
                          }
                        }
                      },
                      "{def_variant2}": {
                        "properties": {
                          "type": {
                            "isCovered": false
                          }
                        }
                      }
                    }
                  }
                },
                "2021-01-01": {
                  "version": "2021-01-01",
                  "isActive": false,
                  "isPreview": false,
                  "model": {
                    "name": "def_base",
                    "properties": {
                      "{def_variant1}": {
                        "properties": {
                          "new_prop": {
                            "isCovered": false
                          },
                          "type": {
                            "isCovered": true
                          }
                        }
                      },
                      "{def_variant2}": {
                        "properties": {
                          "new_prop": {
                            "isCovered": false
                          },
                          "type": {
                            "isCovered": false
                          }
                        }
                      }
                    }
                  }
                }
              }
            }
          }
        },
        "/{}/{}/DEF_FOO": {
          "name": "/{}/{}/DEF_FOO",
          "operations": {
            "DELETE": {
              "name": "DELETE",
              "versions": {
                "2020-12-01": {
                  "version": "2020-12-01",
                  "isActive": false,
                  "isPreview": false
                },
                "2021-01-01": {
                  "version": "2021-01-01",
                  "isActive": false,
                  "isPreview": false
                }
              }
            },
            "GET": {
              "name": "GET",
              "versions": {
                "2020-12-01": {
                  "version": "2020-12-01",
                  "isActive": false,
                  "isPreview": false,
                  "model": {
                    "name": "def_foo",
                    "properties": {
                      "foo": {
                        "isCovered": false
                      }
                    }
                  }
                },
                "2021-01-01": {
                  "version": "2021-01-01",
                  "isActive": false,
                  "isPreview": false,
                  "model": {
                    "name": "def_foo",
                    "properties": {
                      "foo": {
                        "isCovered": false
                      }
                    }
                  }
                }
              }
            },
            "PUT": {
              "name": "PUT",
              "versions": {
                "2020-12-01": {
                  "version": "2020-12-01",
                  "isActive": false,
                  "isPreview": false,
                  "model": {
                    "name": "def_foo",
                    "properties": {
                      "foo": {
                        "isCovered": false
                      }
                    }
                  }
                }
              }
            }
          }
        },
        "/{}/{}/DEF_REGULAR": {
          "name": "/{}/{}/DEF_REGULAR",
          "operations": {
            "GET": {
              "name": "GET",
              "versions": {
                "2020-12-01": {
                  "version": "2020-12-01",
                  "isActive": false,
                  "isPreview": false,
                  "model": {
                    "name": "def_regular",
                    "properties": {
                      "prop_array_of_object": {
                        "properties": {
                          "prop_nested1": {
                            "isCovered": false
                          },
                          "prop_nested2": {
                            "isCovered": false
                          }
                        }
                      },
                      "prop_array_of_primitive": {
                        "isCovered": false
                      },
                      "prop_object": {
                        "properties": {
                          "prop_nested": {
                            "isCovered": false
                          }
                        }
                      },
                      "prop_primitive": {
                        "isCovered": false
                      }
                    }
                  }
                },
                "2021-01-01": {
                  "version": "2021-01-01",
                  "isActive": false,
                  "isPreview": false,
                  "model": {
                    "name": "def_regular",
                    "properties": {
                      "prop_array_of_object": {
                        "properties": {
                          "prop_nested1": {
                            "isCovered": false
                          },
                          "prop_nested2": {
                            "isCovered": false
                          }
                        }
                      },
                      "prop_object": {
                        "properties": {
                          "prop_nested": {
                            "isCovered": false
                          }
                        }
                      },
                      "prop_primitive": {
                        "isCovered": false
                      }
                    }
                  }
                }
              }
            },
            "PUT": {
              "name": "PUT",
              "versions": {
                "2020-12-01": {
                  "version": "2020-12-01",
                  "isActive": false,
                  "isPreview": false,
                  "model": {
                    "name": "def_regular",
                    "properties": {
                      "prop_array_of_object": {
                        "properties": {
                          "prop_nested1": {
                            "isCovered": false
                          },
                          "prop_nested2": {
                            "isCovered": false
                          }
                        }
                      },
                      "prop_array_of_primitive": {
                        "isCovered": false
                      },
                      "prop_object": {
                        "properties": {
                          "prop_nested": {
                            "isCovered": false
                          }
                        }
                      },
                      "prop_primitive": {
                        "isCovered": false
                      }
                    }
                  }
                },
                "2021-01-01": {
                  "version": "2021-01-01",
                  "isActive": false,
                  "isPreview": false,
                  "model": {
                    "name": "def_regular",
                    "properties": {
                      "prop_array_of_object": {
                        "properties": {
                          "prop_nested1": {
                            "isCovered": false
                          },
                          "prop_nested2": {
                            "isCovered": false
                          }
                        }
                      },
                      "prop_object": {
                        "properties": {
                          "prop_nested": {
                            "isCovered": false
                          }
                        }
                      },
                      "prop_primitive": {
                        "isCovered": false
                      }
                    }
                  }
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
		b, err := ioutil.ReadFile(c.schemaInput)
		require.NoError(t, err, idx)

		cov := feeddb.NewEmptyCoverage(index, "comment")
		var schemaInput schemainput.SchemaInput

		// Add active schemaInput
		err = json.Unmarshal(b, &schemaInput)
		require.NoError(t, err, idx)

		schema, err := schemaInput.ToSchema(index, true, true)
		require.NoError(t, err, idx)
		cov.AddSchema(*schema, schemaInput.SchemaLocator)

		// Add in-active schemaInput
		schemaInput.ApiVersion = "2021-01-01"
		schema, err = schemaInput.ToSchema(index, true, false)
		require.NoError(t, err, idx)
		cov.AddSchema(*schema, schemaInput.SchemaLocator)

		require.NoError(t, cov.IncludeUncoveredOperationModels(index), idx)

		b, err = json.MarshalIndent(cov, "", "  ")
		require.NoError(t, err, idx)

		require.JSONEq(t, c.expect, string(b), idx)
	}
}
