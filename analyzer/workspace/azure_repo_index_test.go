package workspace_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/magodo/azure-rest-api-cov/analyzer/specpath"
	"github.com/magodo/azure-rest-api-cov/analyzer/workspace"

	"github.com/stretchr/testify/require"
)

func TestNewAzureRepoIndex(t *testing.T) {
	pwd, err := os.Getwd()
	require.NoError(t, err, "pwd")
	specRootDir := filepath.Join(filepath.Dir(pwd), "testdata", "spec")
	idx, err := workspace.NewAzureRepoIndex(specRootDir)
	require.NoError(t, err, "generate index")
	require.Equal(t, 1, len(idx.Index.ResourceProviders))
	require.Equal(t, 2, len(idx.Index.ResourceProviders["MICROSOFT.FOO"]))
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
					ApiVersion: "2021-01-01",
					ApiPath:    "/SUBSCRIPTIONS/{}/PROVIDERS/MICROSOFT.FOO/BARS/{}",
				},
				Operation: "PUT",
			},
			expect: &workspace.OperationInfo{
				Model: &workspace.ModelInfo{
					Name:        "bar",
					SpecRelPath: filepath.Join("dummy", "resource-manager", "Microsoft.Dummy", "stable", "2021-01-01", "foo.json"),
				},
				SpecRelPath: specpath.Path(filepath.Join("dummy", "resource-manager", "Microsoft.Dummy", "stable", "2021-01-01", "foo.json")),
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
