package workspace

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

const envTestAll = "COV_TEST_ALL"

func TestNewAzureRepo(t *testing.T) {
	if os.Getenv(envTestAll) == "" {
		t.Skipf(`Skipping %s since "COV_TEST_ALL" environment variable is not set...`, t.Name())
	}
	tmpDir := t.TempDir()
	repo, err := NewAzureRepo(tmpDir)
	require.NoError(t, err, "first new")
	repo, err = NewAzureRepo(tmpDir)
	require.NoError(t, err, "second new")
	require.True(t, repo.Repo != nil, "repo.Repo is not nil")
	require.Equal(t, filepath.Join(tmpDir, "specification"), repo.SpecRootDir)
}

func TestAzureRepoGenerateIndex(t *testing.T) {
	pwd, err := os.Getwd()
	require.NoError(t, err, "pwd")
	specRootDir := filepath.Join(filepath.Dir(pwd), "testdata", "spec")
	repo := AzureRepo{
		SpecRootDir: specRootDir,
	}
	expect := fmt.Sprintf(`{
  "resource_providers": {
    "dummy": {
      "2020-12-01-preview:preview": {
        "dummy/resource-manager/Microsoft.Dummy/preview/2020-12-01-preview/bar.json": {
          "path": "dummy/resource-manager/Microsoft.Dummy/preview/2020-12-01-preview/bar.json",
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
`, specRootDir)
	idx, err := repo.GenerateIndex()
	require.NoError(t, err, "generate index")
	b, err := json.MarshalIndent(idx, "", "  ")
	require.NoError(t, err, "marshal index")
	require.JSONEq(t, expect, string(b))
}
