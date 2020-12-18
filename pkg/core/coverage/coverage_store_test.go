package coverage_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/magodo/azure-rest-api-cov/pkg/core/coverage"
	"github.com/magodo/azure-rest-api-cov/pkg/core/workspace"

	"github.com/magodo/azure-rest-api-cov/pkg/core/swagger"

	"github.com/stretchr/testify/require"
)

func TestNewEmptyRepoCoverageStore(t *testing.T) {
	pwd, err := os.Getwd()
	require.NoError(t, err, "pwd")
	specRootDir := filepath.Join(filepath.Dir(pwd), "testdata", "spec")

	expect := coverage.RepoCoverageStore{
		Commit: "xxx",
		ResourceProviders: coverage.ResourceProviders{
			"dummy": coverage.Operations{
				coverage.Operation{
					APIPath:       "/{host}/{foo}/def_base",
					OperationKind: "PUT",
				}: coverage.APIVersions{
					"2020-12-01": nil,
				},
				coverage.Operation{
					APIPath:       "/{host}/{foo}/def_bar",
					OperationKind: "POST",
				}: coverage.APIVersions{
					"2020-12-01": nil,
				},
				coverage.Operation{
					APIPath:       "/{host}/{foo}/def_regular",
					OperationKind: "GET",
				}: coverage.APIVersions{
					"2020-12-01": nil,
				},
				coverage.Operation{
					APIPath:       "/{host}/{foo}/def_regular",
					OperationKind: "PUT",
				}: coverage.APIVersions{
					"2020-12-01": nil,
				},
			},
		},
	}

	idx, err := workspace.NewAzureRepoIndex(specRootDir, "xxx")
	require.NoError(t, err, "generate index")

	cov := coverage.NewEmptyRepoCoverageStore(*idx)

	require.Equal(t, expect, cov)

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
					"rp1": coverage.Operations{
						coverage.Operation{
							APIPath:       "path1",
							OperationKind: "PUT",
						}: coverage.APIVersions{
							"version1": &coverage.SchemaCoverageStore{
								SchemaMetaInfo: swagger.SchemaMetaInfo{
									Name:           "schema1",
									Version:        "2077-01-01",
									SwaggerRelPath: "foo/bar.json",
								},
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
			expect: `
{
  "commit": "xxx",
  "resource_providers": {
    "rp1": {
      "path1:PUT": {
        "version1": {
          "name": "schema1",
          "version": "2077-01-01",
          "swagger_rel_path": "foo/bar.json",
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
      "path1:PUT": {
        "version1": {
          "name": "schema1",
          "version": "2077-01-01",
          "swagger_rel_path": "foo/bar.json",
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
`,
			expect: coverage.RepoCoverageStore{
				Commit: "xxx",
				ResourceProviders: coverage.ResourceProviders{
					"rp1": coverage.Operations{
						coverage.Operation{
							APIPath:       "path1",
							OperationKind: "PUT",
						}: coverage.APIVersions{
							"version1": &coverage.SchemaCoverageStore{
								SchemaMetaInfo: swagger.SchemaMetaInfo{
									Name:           "schema1",
									Version:        "2077-01-01",
									SwaggerRelPath: "foo/bar.json",
								},
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
	}

	for i, c := range cases {
		cov := coverage.RepoCoverageStore{}
		err := json.Unmarshal([]byte(c.input), &cov)
		require.NoError(t, err, i)
		require.Equal(t, c.expect, cov, i)
	}
}
