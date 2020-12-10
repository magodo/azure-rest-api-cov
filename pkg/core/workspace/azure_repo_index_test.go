package workspace

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAPIsMarshal(t *testing.T) {
	cases := []struct {
		apis   APIs
		expect string
	}{
		{
			apis: APIs{
				API{
					Version:   "2020-12-12",
					IsPreview: true,
				}: APISpecs{
					"path": APISpec{
						Path:           "path",
						APIPaths:       []string{"p1", "p2"},
						APIRootSchemas: []string{"schema1", "schema2"},
					},
				},
			},
			expect: `
{
  "2020-12-12:preview": {
  "path": {
    "path": "path",
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
		expect APIs
	}{
		{
			input: `
{
  "2020-12-12:preview": {
  "path": {
    "path": "path",
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
			expect: APIs{
				API{
					Version:   "2020-12-12",
					IsPreview: true,
				}: APISpecs{
					"path": APISpec{
						Path:           "path",
						APIPaths:       []string{"p1", "p2"},
						APIRootSchemas: []string{"schema1", "schema2"},
					},
				},
			},
		},
	}

	for idx, c := range cases {
		var apis APIs
		err := json.Unmarshal([]byte(c.input), &apis)
		require.NoError(t, err, idx)
		require.Equal(t, c.expect, apis, idx)
	}
}

func TestResourceProvidersMarshal(t *testing.T) {
	cases := []struct {
		rps    ResourceProviders
		expect string
	}{
		{
			rps: ResourceProviders{
				ResourceProvider{
					Name: "rp",
				}: APIs{
					API{
						Version:   "2020-12-12",
						IsPreview: true,
					}: APISpecs{
						"path": APISpec{
							Path:           "path",
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
		  "path": "path",
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
		expect ResourceProviders
	}{
		{
			input: `
{
  "rp": {
	  "2020-12-12:preview": {
		"path": {
		  "path": "path",
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
			expect: ResourceProviders{
				ResourceProvider{
					Name: "rp",
				}: APIs{
					API{
						Version:   "2020-12-12",
						IsPreview: true,
					}: APISpecs{
						"path": APISpec{
							Path:           "path",
							APIPaths:       []string{"p1", "p2"},
							APIRootSchemas: []string{"schema1", "schema2"},
						},
					},
				},
			},
		},
	}

	for idx, c := range cases {
		var rps ResourceProviders
		err := json.Unmarshal([]byte(c.input), &rps)
		require.NoError(t, err, idx)
		require.Equal(t, c.expect, rps, idx)
	}
}

func TestAzureRepoIndexRefresh(t *testing.T) {
	pwd, err := os.Getwd()
	require.NoError(t, err, "pwd")
	specRootDir := filepath.Join(filepath.Dir(pwd), "testdata", "spec")
	specRefreshRootDir := filepath.Join(filepath.Dir(pwd), "testdata", "specRefresh")
	repo := AzureRepo{
		SpecRootDir: specRootDir,
	}
	expect := fmt.Sprintf(`{
  "resource_providers": {
    "dummy": {
      "2020-12-10-preview:preview": {
        "dummy/resource-manager/Microsoft.Dummy/preview/2020-12-10-preview/bar.json": {
          "path": "dummy/resource-manager/Microsoft.Dummy/preview/2020-12-10-preview/bar.json",
          "api_paths": [
            "/{host}/{bar}"
          ],
          "api_put_schemas": []
        }
      },
      "2020-12-01:stable": {
        "dummy/resource-manager/Microsoft.Dummy/stable/2020-12-01/foo.json": {
          "path": "dummy/resource-manager/Microsoft.Dummy/stable/2020-12-01/foo.json",
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
	idx, err := repo.GenerateIndex()
	require.NoError(t, err, "generate index")

	idx.SpecRootDir = specRefreshRootDir
	require.NoError(t, idx.Refresh(), "refresh")

	b, err := json.MarshalIndent(idx, "", "  ")
	require.NoError(t, err, "marshal index")
	require.JSONEq(t, expect, string(b))
}
