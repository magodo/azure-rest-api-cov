package workspace_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/magodo/azure-rest-api-cov/pkg/core/workspace"

	"github.com/stretchr/testify/require"
)

func TestNewAzureRepoIndex(t *testing.T) {
	pwd, err := os.Getwd()
	require.NoError(t, err, "pwd")
	specRootDir := filepath.Join(filepath.Dir(pwd), "testdata", "spec")
	expect := fmt.Sprintf(`
{
  "spec_root_dir": "%s",
  "commit": "",
  "resource_providers": {
    "dummy": {
      "2020-12-01-preview": {
        "dummy/resource-manager/Microsoft.Dummy/preview/2020-12-01-preview/bar.json": {
          "/{host}/{bar}": {}
        }
      },
      "2020-12-01": {
        "dummy/resource-manager/Microsoft.Dummy/stable/2020-12-01/foo.json": {
          "/{host}/{foo}/def_foo": {
            "PUT": {
              "name": "def_foo",
              "spec_rel_path": "dummy/resource-manager/Microsoft.Dummy/stable/2020-12-01/foo.json"
            },
            "GET": {
              "name": "def_foo",
              "spec_rel_path": "dummy/resource-manager/Microsoft.Dummy/stable/2020-12-01/foo.json"
            }
          },
          "/{host}/{foo}/def_bar": {
            "POST": {
              "name": "def_bar",
              "spec_rel_path": "dummy/resource-manager/Microsoft.Dummy/preview/2020-12-01-preview/bar.json"
            }
          },
          "/{host}/{foo}/def_base": {
            "PUT": {
              "name": "def_base",
              "spec_rel_path": "dummy/resource-manager/Microsoft.Dummy/stable/2020-12-01/foo.json"
            }
          },
          "/{host}/{foo}/def_regular": {
            "GET": {
              "name": "def_regular",
              "spec_rel_path": "dummy/resource-manager/Microsoft.Dummy/stable/2020-12-01/foo.json"
            },
            "PUT": {
              "name": "def_regular",
              "spec_rel_path": "dummy/resource-manager/Microsoft.Dummy/stable/2020-12-01/foo.json"
            }
          }
        }
      },
      "2021-01-01": {
        "dummy/resource-manager/Microsoft.Dummy/stable/2021-01-01/foo.json": {
          "/{host}/{foo}/def_foo": {
            "GET": {
              "name": "def_foo",
              "spec_rel_path": "dummy/resource-manager/Microsoft.Dummy/stable/2021-01-01/foo.json"
            }
          },
          "/{host}/{foo}/def_base": {
            "PUT": {
              "name": "def_base",
              "spec_rel_path": "dummy/resource-manager/Microsoft.Dummy/stable/2021-01-01/foo.json"
            }
          },
          "/{host}/{foo}/def_regular": {
            "GET": {
              "name": "def_regular",
              "spec_rel_path": "dummy/resource-manager/Microsoft.Dummy/stable/2021-01-01/foo.json"
            },
            "PUT": {
              "name": "def_regular",
              "spec_rel_path": "dummy/resource-manager/Microsoft.Dummy/stable/2021-01-01/foo.json"
            }
          }
        }
      }
    }
  }
}`, specRootDir)
	idx, err := workspace.NewAzureRepoIndex(specRootDir, "")
	require.NoError(t, err, "generate index")
	b, err := json.MarshalIndent(idx, "", "  ")
	require.NoError(t, err, "marshal index")
	require.JSONEq(t, expect, string(b))
}

//func TestAzureRepoIndexRefresh(t *testing.T) {
//	pwd, err := os.Getwd()
//	require.NoError(t, err, "pwd")
//	specRootDir := filepath.Join(filepath.Dir(pwd), "testdata", "spec")
//	specRefreshRootDir := filepath.Join(filepath.Dir(pwd), "testdata", "specRefresh")
//	expect := fmt.Sprintf(`{
//  "resource_providers": {
//    "dummy": {
//      "2020-12-10-preview:preview": {
//        "dummy/resource-manager/Microsoft.Dummy/preview/2020-12-10-preview/bar.json": {
//          "rel_path": "dummy/resource-manager/Microsoft.Dummy/preview/2020-12-10-preview/bar.json",
//          "api_paths": [
//            "/{host}/{bar}"
//          ],
//          "api_put_schemas": []
//        }
//      },
//      "2020-12-01:stable": {
//        "dummy/resource-manager/Microsoft.Dummy/stable/2020-12-01/foo.json": {
//          "rel_path": "dummy/resource-manager/Microsoft.Dummy/stable/2020-12-01/foo.json",
//          "api_paths": [
//            "/{host}/{foo}"
//          ],
//          "api_put_schemas": []
//        }
//      }
//    }
//  },
//  "spec_root_dir": "%s"
//}
//`, specRefreshRootDir)
//	idx, err := workspace.NewAzureRepoIndex(specRootDir)
//	require.NoError(t, err, "generate index")
//
//	idx.SpecRootDir = specRefreshRootDir
//	require.NoError(t, idx.Refresh(), "refresh")
//
//	b, err := json.MarshalIndent(idx, "", "  ")
//	require.NoError(t, err, "marshal index")
//	require.JSONEq(t, expect, string(b))
//}

func TestAzureRepoIndexExportImport(t *testing.T) {
	pwd, err := os.Getwd()
	require.NoError(t, err, "pwd")
	specRootDir := filepath.Join(filepath.Dir(pwd), "testdata", "spec")

	idx, err := workspace.NewAzureRepoIndex(specRootDir, "")
	require.NoError(t, err, "generate index")

	tmpDir := t.TempDir()
	indexPath := filepath.Join(tmpDir, "index.json")
	require.NoError(t, idx.Export(indexPath), "export")

	newidx, err := workspace.ImportAzureRepoIndex(indexPath)
	require.NoError(t, err, "import")

	require.Equal(t, idx, newidx)
}

func TestAzureRepoIndexLocateSchema(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	specBaseDir := filepath.Join(filepath.Dir(pwd), "testdata", "spec")

	type result struct {
		schemaName  string
		specRelPath string
	}

	cases := []struct {
		locator workspace.SchemaLocator
		expect  result
	}{
		{
			locator: workspace.SchemaLocator{
				SpecLocator: workspace.SpecLocator{
					ApiVersion: "2020-12-01",
					ApiPath:    "/{host}/{foo}/def_bar",
				},
				Operation: "POST",
			},
			expect: result{
				schemaName:  "def_bar",
				specRelPath: filepath.Join("dummy", "resource-manager", "Microsoft.Dummy", "preview", "2020-12-01-preview", "bar.json"),
			},
		},
	}

	index, err := workspace.NewAzureRepoIndex(specBaseDir, "")
	require.NoError(t, err)

	for idx, c := range cases {
		name, p, err := index.LocateSchema(c.locator)
		require.NoError(t, err, idx)
		require.Equal(t, c.expect, result{
			schemaName:  name,
			specRelPath: string(p),
		}, idx)
	}
}
