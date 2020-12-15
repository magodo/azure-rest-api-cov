package workspace_test

import (
	"os"
	"testing"

	"github.com/magodo/azure-rest-api-cov/pkg/core/workspace"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	if os.Getenv(envTestAll) == "" {
		t.Skipf(`Skipping %s since "%s" environment variable is not set...`, t.Name(), envTestAll)
	}
	tmpDir := t.TempDir()
	t.Log("start to init (1)")
	_, _, err := workspace.Init(tmpDir, workspace.InitOption{
		Force:    false,
		RepoHash: "",
	})
	require.NoError(t, err, "first init")

	t.Log("start to init (2)")
	_, _, err = workspace.Init(tmpDir, workspace.InitOption{
		Force:    false,
		RepoHash: "",
	})
	require.NoError(t, err, "second init")

	t.Log("start to init (3)")
	_, _, err = workspace.Init(tmpDir, workspace.InitOption{
		Force:    false,
		RepoHash: "ee92c0739efe6acd5b947a3f7829a349800b2426",
	})
	require.NoError(t, err, "third init")
}
