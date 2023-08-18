package feeddb_test

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/magodo/azure-rest-api-cov/analyzer/feeddb"
	"github.com/magodo/azure-rest-api-cov/analyzer/schemainput"
	"github.com/magodo/azure-rest-api-cov/analyzer/workspace"
	"github.com/stretchr/testify/require"
)

func TestNewEmptyCoverage(t *testing.T) {
	pwd, err := os.Getwd()
	require.NoError(t, err, "pwd")
	specRootDir := filepath.Join(filepath.Dir(pwd), "testdata", "spec")

	expect := feeddb.Coverage{
		Ref:     "",
		Comment: "comment",
		ResourceProviders: feeddb.ResourceProviders{
			"dummy": feeddb.ResourceProvider{
				Name: "dummy",
				Paths: feeddb.APIPaths{
					"/PROVIDERS/MICROSOFT.DUMMY/DEF_FOO": feeddb.APIPath{
						Name: "/PROVIDERS/MICROSOFT.DUMMY/DEF_FOO",
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
					"/PROVIDERS/MICROSOFT.DUMMY/DEF_BAR": feeddb.APIPath{
						Name: "/PROVIDERS/MICROSOFT.DUMMY/DEF_BAR",
						Operations: feeddb.Operations{
							"POST": {
								Name: "POST",
								Versions: feeddb.APIVersions{
									"2020-12-01": {
										Version: "2020-12-01",
									},
								},
							},
							"GET": feeddb.Operation{
								Name: "GET",
								Versions: feeddb.APIVersions{
									"2020-12-01-preview": feeddb.APIVersion{
										Version:   "2020-12-01-preview",
										IsActive:  false,
										IsPreview: true,
									},
								},
							},
						},
					},
					"/PROVIDERS/MICROSOFT.DUMMY/DEF_BASE": feeddb.APIPath{
						Name: "/PROVIDERS/MICROSOFT.DUMMY/DEF_BASE",
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
					"/PROVIDERS/MICROSOFT.DUMMY/DEF_REGULAR": feeddb.APIPath{
						Name: "/PROVIDERS/MICROSOFT.DUMMY/DEF_REGULAR",
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

	idx, err := workspace.NewAzureRepoIndex(specRootDir)
	require.NoError(t, err, "generate index")
	cov := feeddb.NewEmptyCoverage(idx, "comment", []byte{})
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
        "/PROVIDERS/MICROSOFT.DUMMY/DEF_BAR": {
          "name": "/PROVIDERS/MICROSOFT.DUMMY/DEF_BAR",
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
            },
            "GET": {
              "name": "GET",
              "versions": {
                "2020-12-01-preview": {
                  "version": "2020-12-01-preview",
                  "isActive": false,
                  "isPreview": true
                }
              }
            }
          }
        },
        "/PROVIDERS/MICROSOFT.DUMMY/DEF_BASE": {
          "name": "/PROVIDERS/MICROSOFT.DUMMY/DEF_BASE",
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
                      "{def_base}": {
                        "properties": {
                          "type": {
                            "isCovered": false
                          }
                        }
                      },
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
                      "{def_base}": {
                        "properties": {
                          "type": {
                            "isCovered": false
                          }
                        }
                      },
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
        "/PROVIDERS/MICROSOFT.DUMMY/DEF_FOO": {
          "name": "/PROVIDERS/MICROSOFT.DUMMY/DEF_FOO",
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
        "/PROVIDERS/MICROSOFT.DUMMY/DEF_REGULAR": {
          "name": "/PROVIDERS/MICROSOFT.DUMMY/DEF_REGULAR",
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

	index, err := workspace.NewAzureRepoIndex(specBaseDir)
	require.NoError(t, err)

	for idx, c := range cases {
		b, err := ioutil.ReadFile(c.schemaInput)
		require.NoError(t, err, idx)

		cov := feeddb.NewEmptyCoverage(index, "comment", []byte{})
		var schemaInput schemainput.SchemaInput

		// Add active schemaInput
		err = json.Unmarshal(b, &schemaInput)
		require.NoError(t, err, idx)

		opInfo, err := schemaInput.ToOperationInfo(index, true, true, nil)
		require.NoError(t, err, idx)
		cov.AddOperation(opInfo)

		// Add in-active schemaInput
		schemaInput.ApiVersion = "2021-01-01"
		opInfo, err = schemaInput.ToOperationInfo(index, true, false, nil)
		require.NoError(t, err, idx)
		cov.AddOperation(opInfo)

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
  "comment": "comment",
  "ref": "",
  "resourceProviders": {
    "dummy": {
      "name": "dummy",
      "paths": {
        "/PROVIDERS/MICROSOFT.DUMMY/DEF_BAR": {
          "name": "/PROVIDERS/MICROSOFT.DUMMY/DEF_BAR",
          "operations": {
            "POST": {
              "name": "POST",
              "versions": {
                "2020-12-01": {
                  "version": "2020-12-01",
                  "isActive": false,
                  "isPreview": false,
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
            },
            "GET": {
              "name": "GET",
              "versions": {
                "2020-12-01-preview": {
                  "version": "2020-12-01-preview",
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
        "/PROVIDERS/MICROSOFT.DUMMY/DEF_BASE": {
          "name": "/PROVIDERS/MICROSOFT.DUMMY/DEF_BASE",
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
                      "{def_base}": {
                        "properties": {
                          "type": {
                            "isCovered": false
                          }
                        }
                      },
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
                      "{def_base}": {
                        "properties": {
                          "type": {
                            "isCovered": false
                          }
                        }
                      },
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
        "/PROVIDERS/MICROSOFT.DUMMY/DEF_FOO": {
          "name": "/PROVIDERS/MICROSOFT.DUMMY/DEF_FOO",
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
        "/PROVIDERS/MICROSOFT.DUMMY/DEF_REGULAR": {
          "name": "/PROVIDERS/MICROSOFT.DUMMY/DEF_REGULAR",
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

	index, err := workspace.NewAzureRepoIndex(specBaseDir)
	require.NoError(t, err)

	for idx, c := range cases {
		b, err := ioutil.ReadFile(c.schemaInput)
		require.NoError(t, err, idx)

		cov := feeddb.NewEmptyCoverage(index, "comment", []byte{})
		var schemaInput schemainput.SchemaInput

		// Add active schemaInput
		err = json.Unmarshal(b, &schemaInput)
		require.NoError(t, err, idx)

		opInfo, err := schemaInput.ToOperationInfo(index, true, true, nil)
		require.NoError(t, err, idx)
		require.NoError(t, cov.AddOperation(opInfo), idx)

		// Add in-active schemaInput
		schemaInput.ApiVersion = "2021-01-01"
		opInfo, err = schemaInput.ToOperationInfo(index, true, false, nil)
		require.NoError(t, err, idx)
		require.NoError(t, cov.AddOperation(opInfo), idx)

		actual, err := cov.GenerateWithUncoveredOperationModels(index, nil, nil)
		require.NoError(t, err, idx)
		require.JSONEq(t, c.expect, actual, idx)
	}
}

func TestCoverageParseRTByPath(t *testing.T) {
	cases := []struct {
		expect       string
		path         string
		ignoreSegCnt bool
	}{
		{
			path:         "/providers/Foo.Bar",
			expect:       "FOO.BAR",
			ignoreSegCnt: false,
		},
		{
			path:         "/providers/Foo.Bar/foos/{fooName}",
			expect:       "FOO.BAR/FOOS",
			ignoreSegCnt: false,
		},
		{
			path:         "/providers/Foo.Bar/foos/{fooName}/bars/{barName}",
			expect:       "FOO.BAR/FOOS/BARS",
			ignoreSegCnt: false,
		},
		{
			path:         "/providers/ABABAB/ababas/{ababaName}/providers/Foo.Bar/foos/{fooName}",
			expect:       "FOO.BAR/FOOS",
			ignoreSegCnt: false,
		},
		{
			// short rt
			path:         "/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Compute/virtualMachines/{vmName}",
			expect:       "Microsoft.Compute/virtualMachines",
			ignoreSegCnt: false,
		},
		{
			// long rt
			path:         "/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Compute/virtualMachines/{vmName}/instanceView/{viewName}",
			expect:       "Microsoft.Compute/virtualMachines/instanceView",
			ignoreSegCnt: false,
		},
		{
			// long rt with ignoreSegCnt
			path:         "/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Compute/vmPool/{poolName}/virtualMachines/{vmName}/shutdown",
			expect:       "Microsoft.Compute/vmPool/virtualMachines",
			ignoreSegCnt: true,
		},
		{
			// multi providers in one path
			path:         "/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Compute/virtualMachines/{vmName}/instanceView/{viewName}/providers/Microsoft.Network/networkInterfaces/{networkInterfaceName}",
			expect:       "MICROSOFT.NETWORK/NETWORKINTERFACES",
			ignoreSegCnt: false,
		},
	}

	for _, c := range cases {
		actual := feeddb.ParseRTbyPath(strings.ToUpper(c.path), c.ignoreSegCnt)
		require.Equal(t, strings.ToUpper(c.expect), actual)
	}
}
