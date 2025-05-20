package workspace_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/magodo/azure-rest-api-cov/analyzer/workspace"

	"github.com/stretchr/testify/require"
)

func TestNewAzureRepoIndex(t *testing.T) {
	pwd, err := os.Getwd()
	require.NoError(t, err, "pwd")
	specRootDir := filepath.Join(filepath.Dir(pwd), "testdata", "spec")
	expect := workspace.ResourceProviders{
		"dummy": workspace.APIVersions{
			"2020-12-01": workspace.Specs{
				"dummy/resource-manager/Microsoft.Dummy/stable/2020-12-01/foo.json": workspace.APIPaths{
					"/PROVIDERS/MICROSOFT.DUMMY/DEF_BAR": {
						"POST": true,
					},
					"/PROVIDERS/MICROSOFT.DUMMY/DEF_BASE": {
						"PUT": true,
					},
					"/PROVIDERS/MICROSOFT.DUMMY/DEF_FOO": {
						"PUT":    true,
						"GET":    true,
						"DELETE": true,
					},
					"/PROVIDERS/MICROSOFT.DUMMY/DEF_REGULAR": {
						"GET": true,
						"PUT": true,
					},
				},
			},
			"2020-12-01-preview": workspace.Specs{
				"dummy/resource-manager/Microsoft.Dummy/preview/2020-12-01-preview/bar.json": workspace.APIPaths{
					"/PROVIDERS/MICROSOFT.DUMMY/DEF_BAR": {
						"GET": true,
					},
				},
			},
			"2021-01-01": workspace.Specs{
				"dummy/resource-manager/Microsoft.Dummy/stable/2021-01-01/foo.json": workspace.APIPaths{
					"/PROVIDERS/MICROSOFT.DUMMY/DEF_BASE": {
						"PUT": true,
					},
					"/PROVIDERS/MICROSOFT.DUMMY/DEF_FOO": {
						"DELETE": true,
						"GET":    true,
					},
					"/PROVIDERS/MICROSOFT.DUMMY/DEF_REGULAR": {
						"GET": true,
						"PUT": true,
					},
				},
			},
		},
	}
	idx, err := workspace.NewAzureRepoIndex(specRootDir)
	require.NoError(t, err, "generate index")
	require.Equal(t, expect, idx.ResourceProviders)
}

func TestAzureRepoIndexExportImport(t *testing.T) {
	pwd, err := os.Getwd()
	require.NoError(t, err, "pwd")
	specRootDir := filepath.Join(filepath.Dir(pwd), "testdata", "spec")

	idx, err := workspace.NewAzureRepoIndex(specRootDir)
	require.NoError(t, err, "generate index")

	tmpDir := t.TempDir()
	indexPath := filepath.Join(tmpDir, "index.json")
	require.NoError(t, idx.Export(indexPath), "export")

	newidx, err := workspace.ImportAzureRepoIndex(specRootDir, indexPath)
	require.NoError(t, err, "import")

	require.Equal(t, idx, newidx)
}

func TestAzureRepoIndexLocateOperation(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	specBaseDir := filepath.Join(filepath.Dir(pwd), "testdata", "spec")

	cases := []struct {
		locator workspace.OperationLocator
		expect  *workspace.OperationInfo
	}{
		{
			locator: workspace.OperationLocator{
				SpecLocator: workspace.SpecLocator{
					ApiVersion: "2020-12-01",
					ApiPath:    "/PROVIDERS/MICROSOFT.DUMMY/DEF_BAR",
				},
				Operation: "POST",
			},
			expect: &workspace.OperationInfo{
				Model: &workspace.ModelInfo{
					Name:        "def_bar",
					SpecRelPath: filepath.Join("dummy", "resource-manager", "Microsoft.Dummy", "preview", "2020-12-01-preview", "bar.json"),
				},
				APIPath:     "/PROVIDERS/MICROSOFT.DUMMY/DEF_BAR",
				SpecRelPath: filepath.Join("dummy", "resource-manager", "Microsoft.Dummy", "stable", "2020-12-01", "foo.json"),
			},
		},
	}

	index, err := workspace.NewAzureRepoIndex(specBaseDir)
	require.NoError(t, err)

	for idx, c := range cases {
		info, err := index.LocateOperation(c.locator)
		require.NoError(t, err, idx)
		require.Equal(t, c.expect, info, idx)
	}
}
