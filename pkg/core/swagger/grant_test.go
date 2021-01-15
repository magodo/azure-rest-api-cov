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
		error  bool
	}{
		{
			path: filepath.Join(grantBaseDir, "normal"),
			expect: map[string]SchemaGrant{
				SchemaMetaInfo{
					Name:           "schema1",
					SwaggerRelPath: "a/b/c/grant.json",
				}.Ref(): {
					Comment: "granted",
				},
				SchemaMetaInfo{
					Name:           "schema2",
					SwaggerRelPath: "a/b/c/grant.json",
				}.Ref(): {
					Properties: map[string]string{
						"prop1":         "granted",
						"prop2/prop2_1": "granted",
					},
				},
			},
		},
		{
			path:  filepath.Join(grantBaseDir, "abnormal"),
			error: true,
		},
	}

	for idx, c := range cases {
		grants, err := NewSchemaGrantsFromFiles(c.path)
		if c.error {
			require.Error(t, err, idx)
			continue
		}
		require.NoError(t, err, idx)
		require.Equal(t, c.expect, grants, idx)
	}
}
