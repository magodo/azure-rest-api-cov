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

func TestAPIsMarshal(t *testing.T) {
	cases := []struct {
		apis   workspace.APIs
		expect string
	}{
		{
			apis: workspace.APIs{
				workspace.API{
					Version:   "2020-12-12",
					IsPreview: true,
				}: workspace.APISpecs{
					"path": workspace.APISpec{
						RelPath:        "path",
						APIPaths:       []string{"p1", "p2"},
						APIRootSchemas: []string{"schema1", "schema2"},
					},
				},
			},
			expect: `
{
  "2020-12-12:preview": {
  "path": {
    "rel_path": "path",
    "api_paths": [
      "p1",
      "p2"
     ],
     "api_put_schemas": [
       "schema1",
       "schema2"
     ]
    }
  }
}
`,
		},
	}

	for idx, c := range cases {
		b, err := json.MarshalIndent(c.apis, "", "  ")
		require.NoError(t, err, idx)
		require.JSONEq(t, c.expect, string(b), idx)
	}
}

func TestAPIsUnmarshal(t *testing.T) {
	cases := []struct {
		input  string
		expect workspace.APIs
	}{
		{
			input: `
{
  "2020-12-12:preview": {
  "path": {
    "rel_path": "path",
    "api_paths": [
      "p1",
      "p2"
     ],
     "api_put_schemas": [
       "schema1",
       "schema2"
     ]
    }
  }
}
`,
			expect: workspace.APIs{
				workspace.API{
					Version:   "2020-12-12",
					IsPreview: true,
				}: workspace.APISpecs{
					"path": workspace.APISpec{
						RelPath:        "path",
						APIPaths:       []string{"p1", "p2"},
						APIRootSchemas: []string{"schema1", "schema2"},
					},
				},
			},
		},
	}

	for idx, c := range cases {
		var apis workspace.APIs
		err := json.Unmarshal([]byte(c.input), &apis)
		require.NoError(t, err, idx)
		require.Equal(t, c.expect, apis, idx)
	}
}

func TestResourceProvidersMarshal(t *testing.T) {
	cases := []struct {
		rps    workspace.ResourceProviders
		expect string
	}{
		{
			rps: workspace.ResourceProviders{
				workspace.ResourceProvider{
					Name: "rp",
				}: workspace.APIs{
					workspace.API{
						Version:   "2020-12-12",
						IsPreview: true,
					}: workspace.APISpecs{
						"path": workspace.APISpec{
							RelPath:        "path",
							APIPaths:       []string{"p1", "p2"},
							APIRootSchemas: []string{"schema1", "schema2"},
						},
					},
				},
			},
			expect: `
{
  "rp": {
	  "2020-12-12:preview": {
		"path": {
		  "rel_path": "path",
		  "api_paths": [
			"p1",
			"p2"
		  ],
		  "api_put_schemas": [
			"schema1",
			"schema2"
		  ]
		}
	  }
  }
}
`,
		},
	}

	for idx, c := range cases {
		b, err := json.MarshalIndent(c.rps, "", "  ")
		require.NoError(t, err, idx)
		require.JSONEq(t, c.expect, string(b), idx)
	}
}

func TestResourceProvidersUnmarshal(t *testing.T) {
	cases := []struct {
		input  string
		expect workspace.ResourceProviders
	}{
		{
			input: `
{
  "rp": {
	  "2020-12-12:preview": {
		"path": {
		  "rel_path": "path",
		  "api_paths": [
			"p1",
			"p2"
		  ],
		  "api_put_schemas": [
			"schema1",
			"schema2"
		  ]
		}
	  }
  }
}
`,
			expect: workspace.ResourceProviders{
				workspace.ResourceProvider{
					Name: "rp",
				}: workspace.APIs{
					workspace.API{
						Version:   "2020-12-12",
						IsPreview: true,
					}: workspace.APISpecs{
						"path": workspace.APISpec{
							RelPath:        "path",
							APIPaths:       []string{"p1", "p2"},
							APIRootSchemas: []string{"schema1", "schema2"},
						},
					},
				},
			},
		},
	}

	for idx, c := range cases {
		var rps workspace.ResourceProviders
		err := json.Unmarshal([]byte(c.input), &rps)
		require.NoError(t, err, idx)
		require.Equal(t, c.expect, rps, idx)
	}
}

func TestNewAzureRepoIndex(t *testing.T) {
	pwd, err := os.Getwd()
	require.NoError(t, err, "pwd")
	specRootDir := filepath.Join(filepath.Dir(pwd), "testdata", "spec")
	expect := fmt.Sprintf(`{
  "resource_providers": {
    "dummy": {
      "2020-12-01-preview:preview": {
        "dummy/resource-manager/Microsoft.Dummy/preview/2020-12-01-preview/bar.json": {
          "rel_path": "dummy/resource-manager/Microsoft.Dummy/preview/2020-12-01-preview/bar.json",
          "api_paths": [
            "/{host}/{bar}"
          ],
          "api_put_schemas": []
        }
      },
      "2020-12-01:stable": {
        "dummy/resource-manager/Microsoft.Dummy/stable/2020-12-01/foo.json": {
          "rel_path": "dummy/resource-manager/Microsoft.Dummy/stable/2020-12-01/foo.json",
          "api_paths": [
            "/{host}/{foo}"
          ],
          "api_put_schemas": []
        }
      }
    }
  },
  "spec_root_dir": "%s"
}
`, specRootDir)
	idx, err := workspace.NewAzureRepoIndex(specRootDir)
	require.NoError(t, err, "generate index")
	b, err := json.MarshalIndent(idx, "", "  ")
	require.NoError(t, err, "marshal index")
	require.JSONEq(t, expect, string(b))
}

func TestAzureRepoIndexRefresh(t *testing.T) {
	pwd, err := os.Getwd()
	require.NoError(t, err, "pwd")
	specRootDir := filepath.Join(filepath.Dir(pwd), "testdata", "spec")
	specRefreshRootDir := filepath.Join(filepath.Dir(pwd), "testdata", "specRefresh")
	expect := fmt.Sprintf(`{
  "resource_providers": {
    "dummy": {
      "2020-12-10-preview:preview": {
        "dummy/resource-manager/Microsoft.Dummy/preview/2020-12-10-preview/bar.json": {
          "rel_path": "dummy/resource-manager/Microsoft.Dummy/preview/2020-12-10-preview/bar.json",
          "api_paths": [
            "/{host}/{bar}"
          ],
          "api_put_schemas": []
        }
      },
      "2020-12-01:stable": {
        "dummy/resource-manager/Microsoft.Dummy/stable/2020-12-01/foo.json": {
          "rel_path": "dummy/resource-manager/Microsoft.Dummy/stable/2020-12-01/foo.json",
          "api_paths": [
            "/{host}/{foo}"
          ],
          "api_put_schemas": []
        }
      }
    }
  },
  "spec_root_dir": "%s"
}
`, specRefreshRootDir)
	idx, err := workspace.NewAzureRepoIndex(specRootDir)
	require.NoError(t, err, "generate index")

	idx.SpecRootDir = specRefreshRootDir
	require.NoError(t, idx.Refresh(), "refresh")

	b, err := json.MarshalIndent(idx, "", "  ")
	require.NoError(t, err, "marshal index")
	require.JSONEq(t, expect, string(b))
}

func TestAzureRepoIndexExportImport(t *testing.T) {
	pwd, err := os.Getwd()
	require.NoError(t, err, "pwd")
	specRootDir := filepath.Join(filepath.Dir(pwd), "testdata", "spec")
	expect := fmt.Sprintf(`{
  "resource_providers": {
    "dummy": {
      "2020-12-01-preview:preview": {
        "dummy/resource-manager/Microsoft.Dummy/preview/2020-12-01-preview/bar.json": {
          "rel_path": "dummy/resource-manager/Microsoft.Dummy/preview/2020-12-01-preview/bar.json",
          "api_paths": [
            "/{host}/{bar}"
          ],
          "api_put_schemas": []
        }
      },
      "2020-12-01:stable": {
        "dummy/resource-manager/Microsoft.Dummy/stable/2020-12-01/foo.json": {
          "rel_path": "dummy/resource-manager/Microsoft.Dummy/stable/2020-12-01/foo.json",
          "api_paths": [
            "/{host}/{foo}"
          ],
          "api_put_schemas": []
        }
      }
    }
  },
  "spec_root_dir": "%s"
}
`, specRootDir)
	idx, err := workspace.NewAzureRepoIndex(specRootDir)
	require.NoError(t, err, "generate index")

	tmpDir := t.TempDir()
	indexPath := filepath.Join(tmpDir, "index.json")
	require.NoError(t, idx.Export(indexPath), "export")

	idx, err = workspace.ImportAzureRepoIndex(indexPath)
	require.NoError(t, err, "import")

	b, err := json.MarshalIndent(idx, "", "  ")
	require.NoError(t, err, "marshal index")
	require.JSONEq(t, expect, string(b))
}
