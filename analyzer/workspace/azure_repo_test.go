package workspace_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/magodo/azure-rest-api-cov/analyzer/workspace"

	"github.com/stretchr/testify/require"
)

func TestNewAzureRepo(t *testing.T) {
	if os.Getenv(envTestAll) == "" {
		t.Skipf(`Skipping %s since "%s" environment variable is not set...`, t.Name(), envTestAll)
	}
	tmpDir := t.TempDir()
	t.Log("start to new azure repo (1)")
	repo, err := workspace.NewAzureRepo(tmpDir)
	require.NoError(t, err, "first new")
	t.Log("start to new azure repo (2)")
	repo, err = workspace.NewAzureRepo(tmpDir)
	require.NoError(t, err, "second new")
	require.True(t, repo.Repo != nil, "repo.Repo is not nil")
	require.Equal(t, filepath.Join(tmpDir, "specification"), repo.SpecRootDir)
}
