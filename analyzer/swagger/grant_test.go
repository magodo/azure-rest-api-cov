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
	specBasePath = filepath.Join(filepath.Dir(pwd), "testdata", "swagger")
	grantNormalBaseDir := filepath.Join(filepath.Dir(pwd), "testdata", "grant", "normal")
	grantAbnormalPropBaseDir := filepath.Join(filepath.Dir(pwd), "testdata", "grant", "abnormal_prop")
	grantAbnormalNonExistBaseDir := filepath.Join(filepath.Dir(pwd), "testdata", "grant", "abnormal_non_exist")

	cases := []struct {
		path   string
		expect SchemaGrants
		error  bool
	}{
		{
			path: grantNormalBaseDir,
			expect: map[string]SchemaGrant{
				SchemaMetaInfo{
					Name:           "def_foo",
					SwaggerRelPath: "foo.json",
				}.Ref(): {
					Comment: "schema",
				},
				SchemaMetaInfo{
					Name:           "def_base",
					SwaggerRelPath: "foo.json",
				}.Ref(): {
					Variant: "def_variant1",
					Comment: "schema variant",
				},
				SchemaMetaInfo{
					Name:           "def_regular",
					SwaggerRelPath: "foo.json",
				}.Ref(): {
					Properties: map[string]string{
						"prop_primitive":                   "property",
						"prop_array_of_object/prop_nested": "property",
					},
				},
			},
		},
		{
			path:  grantAbnormalPropBaseDir,
			error: true,
		},
		{
			path:  grantAbnormalNonExistBaseDir,
			error: true,
		},
	}

	for idx, c := range cases {
		grants, err := NewSchemaGrantsFromFiles(c.path, specBasePath)
		if c.error {
			require.Error(t, err, idx)
			continue
		}
		require.NoError(t, err, idx)
		require.Equal(t, c.expect, grants, idx)
	}
}
