package grant

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewRPGrantsFromFiles(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	grantBaseDir := filepath.Join(filepath.Dir(pwd), "testdata", "schema_grants")

	cases := []struct {
		dir    string
		expect RPGrants
	}{
		{
			dir: filepath.Join(grantBaseDir, "schema"),
			expect: RPGrants{
				"dummy": {
					"resource.com/foo/bar": "The grant reason.",
				},
			},
		},
	}

	for idx, c := range cases {
		gs, err := NewRPGrantsFromFiles(c.dir)
		require.NoError(t, err, idx)
		require.Equal(t, c.expect, gs, idx)
	}
}
