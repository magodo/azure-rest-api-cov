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
	grantNormalBaseDir := filepath.Join(filepath.Dir(pwd), "testdata", "grant_adhoc", "normal")
	grantAbnormalPropBaseDir := filepath.Join(filepath.Dir(pwd), "testdata", "grant_adhoc", "abnormal_prop")
	grantAbnormalNonExistBaseDir := filepath.Join(filepath.Dir(pwd), "testdata", "grant_adhoc", "abnormal_non_exist")

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
		grants, err := NewSchemaGrantsFromFiles(c.path, specBasePath, nil)
		if c.error {
			require.Error(t, err, idx)
			continue
		}
		require.NoError(t, err, idx)
		require.Equal(t, c.expect, grants, idx)
	}
}

func TestSchemaGrantFor(t *testing.T) {
	ref20200501 := "dummy/resource-manager/Microsoft.Dummy/stable/2020-05-01/dummy.json#/definitions/Model"
	ref20200601 := "dummy/resource-manager/Microsoft.Dummy/stable/2020-06-01/dummy.json#/definitions/Model"

	grant20200501 := SchemaGrant{}
	grant20200601 := SchemaGrant{}

	schemaGrants := SchemaGrants{
		ref20200501: grant20200501,
		ref20200601: grant20200601,
	}

	type expectResult struct {
		grant *SchemaGrant
		exact bool
	}

	cases := []struct {
		ref    string
		expect expectResult
	}{
		// 0
		{
			ref: ref20200601,
			expect: expectResult{
				grant: &grant20200601,
				exact: true,
			},
		},
		// 1
		{
			ref: "dummy/resource-manager/Microsoft.Dummy/stable/2020-07-01/dummy.json#/definitions/Model",
			expect: expectResult{
				grant: &grant20200601,
				exact: false,
			},
		},
		// 2
		{
			ref: "dummy/resource-manager/Microsoft.Dummy/preview/2020-07-01-preview/dummy.json#/definitions/Model",
			expect: expectResult{
				grant: &grant20200601,
				exact: false,
			},
		},
		// 3
		{
			ref: "foo/resource-manager/Microsoft.Foo/stable/2020-06-01/foo.json#/definitions/Model",
			expect: expectResult{
				grant: nil,
				exact: false,
			},
		},
	}

	for idx, c := range cases {
		grant, matchRef, err := schemaGrants.SchemaGrantFor(c.ref)
		require.NoError(t, err, idx)
		require.Equal(t, c.expect, expectResult{
			grant: grant,
			exact: matchRef == c.ref,
		}, idx)
	}
}
