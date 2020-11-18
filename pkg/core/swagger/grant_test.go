package swagger

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewSchemaGrantsFromFiles(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	grantBaseDir := filepath.Join(filepath.Dir(pwd), "testdata", "grant")

	cases := []struct {
		path   string
		expect SchemaGrants
	}{
		{
			path: grantBaseDir,
			expect: map[SchemaAddr]SchemaGrant{
				{SwaggerRelPath: "a/b/c/grant.json", Name: "schema1"}: {
					Comment: "granted",
				},
				{SwaggerRelPath: "a/b/c/grant.json", Name: "schema2"}: {
					Properties: map[string]string{
						"prop1":         "granted",
						"prop2.prop2_1": "granted",
					},
				},
			},
		},
	}

	for idx, c := range cases {
		grants, err := NewSchemaGrantsFromFiles(c.path)
		require.NoError(t, err, idx)
		require.Equal(t, c.expect, grants, idx)
	}
}
