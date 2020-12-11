package workspace_test

import (
	"os"
	"testing"

	"github.com/magodo/azure-rest-api-cov/pkg/core/workspace"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
	if os.Getenv(envTestAll) == "" {
		t.Skipf(`Skipping %s since "COV_TEST_ALL" environment variable is not set...`, t.Name())
	}
	tmpDir := t.TempDir()
	t.Log("start to init (1)")
	_, _, err := workspace.Init(tmpDir, false)
	require.NoError(t, err, "first init")
	t.Log("start to init (2)")
	_, _, err = workspace.Init(tmpDir, false)
	require.NoError(t, err, "second init")
}
