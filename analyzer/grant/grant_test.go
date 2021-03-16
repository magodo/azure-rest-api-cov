package grant

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/magodo/azure-rest-api-cov/analyzer/addr"
	"github.com/magodo/azure-rest-api-cov/analyzer/swagger"

	"github.com/stretchr/testify/require"
)

func TestNewSchemaGrantsFromFiles(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	specBasePath := filepath.Join(filepath.Dir(pwd), "testdata", "swagger")
	grantBaseDir := filepath.Join(filepath.Dir(pwd), "testdata", "grant_adhoc")

	cases := []struct {
		dir    string
		expect SchemaGrants
		error  bool
	}{
		// 0. schema granted
		{
			dir: "schema",
			expect: map[string]SchemaGrant{
				"foo.json#/definitions/def_foo": {
					Comment:           "granted",
					grantedProperties: map[string]string{},
				},
			},
		},

		// 1. schema of variant granted
		{
			dir: "schema_variant",
			expect: map[string]SchemaGrant{
				"foo.json#/definitions/def_base": {
					PropertyGrantRules: PropertyGrantRules{
						PropertyGrantRule{
							Prop:    addr.MustParseProperty("").AsVariant("def_variant1"),
							Comment: "granted",
						},
					},
					grantedProperties: map[string]string{
						"{def_variant1}/type": "granted",
					},
				},
			},
		},

		// 2. schema property granted
		{
			dir: "property_simple",
			expect: map[string]SchemaGrant{
				"foo.json#/definitions/def_regular": {
					PropertyGrantRules: PropertyGrantRules{
						PropertyGrantRule{
							Prop:    addr.MustParseProperty("prop_primitive"),
							Comment: "granted",
						},
						PropertyGrantRule{
							Prop:    addr.MustParseProperty("prop_array_of_object/prop_nested1"),
							Comment: "granted",
						},
					},
					grantedProperties: map[string]string{
						"prop_primitive":                    "granted",
						"prop_array_of_object/prop_nested1": "granted",
					},
				},
			},
		},

		// 3. schema property grant rule
		{
			dir: "property_rule",
			expect: map[string]SchemaGrant{
				"foo.json#/definitions/def_regular": {
					PropertyGrantRules: PropertyGrantRules{
						PropertyGrantRule{
							Prop:    addr.MustParseProperty("prop_array_of_object"),
							Comment: "granted",
						},
						PropertyGrantRule{
							Prop:   addr.MustParseProperty("prop_array_of_object/prop_nested2"),
							Negate: true,
						},
					},
					grantedProperties: map[string]string{
						"prop_array_of_object/prop_nested1": "granted",
					},
				},
			},
		},

		// 4. schema variant property grant rule
		{
			dir: "property_variant_rule",
			expect: map[string]SchemaGrant{
				"foo.json#/definitions/def_to_def_base": {
					PropertyGrantRules: PropertyGrantRules{
						PropertyGrantRule{
							Prop:    addr.MustParseProperty("prop"),
							Comment: "granted",
						},
						PropertyGrantRule{
							Prop:   addr.MustParseProperty("prop{def_variant1}"),
							Negate: true,
						},
					},
					grantedProperties: map[string]string{
						"prop{def_variant2}/type": "granted",
					},
				},
			},
		},
	}

	for idx, c := range cases {
		grants, err := NewSchemaGrantsFromFiles(filepath.Join(grantBaseDir, c.dir), specBasePath, nil)
		if c.error {
			require.Error(t, err, idx)
			continue
		}
		require.NoError(t, err, idx)
		require.Equal(t, c.expect, grants, idx)
	}
}

func TestSchemaGrants_GetGrantForSchema(t *testing.T) {
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
		grant, matchRef, err := schemaGrants.GetGrantForSchema(c.ref)
		require.NoError(t, err, idx)
		require.Equal(t, c.expect, expectResult{
			grant: grant,
			exact: matchRef == c.ref,
		}, idx)
	}
}

func TestSchemaGrant_Grant(t *testing.T) {
	var cases = []struct {
		grant        SchemaGrant
		schema       swagger.Schema
		expectSchema swagger.Schema
		hasError     bool
	}{
		// 0: schema is granted
		{
			grant: SchemaGrant{
				Comment: "schema is granted",
			},
			schema: swagger.Schema{},
			expectSchema: swagger.Schema{
				IsGranted:    true,
				GrantComment: "schema is granted",
			},
		},
		// 1: property granted
		{
			grant: SchemaGrant{
				grantedProperties: map[string]string{
					"prop1": "granted",
				},
			},
			schema: swagger.Schema{
				Properties: map[string]*swagger.Property{
					"prop1": {
						PropertyState: swagger.PropertyStateUncovered,
					},
				},
			},
			expectSchema: swagger.Schema{
				Properties: map[string]*swagger.Property{
					"prop1": {
						PropertyState: swagger.NewPropertyState(swagger.PropStateGranted, "granted"),
					},
				},
			},
		},
		// 2: grant unknown property cause error
		{
			grant: SchemaGrant{
				grantedProperties: map[string]string{
					"prop1": "granted",
				},
			},
			schema: swagger.Schema{
				Properties: map[string]*swagger.Property{},
			},
			hasError: true,
		},
		// 3: grant property of non-uncovered state cause error
		{
			grant: SchemaGrant{
				grantedProperties: map[string]string{
					"prop1": "granted",
				},
			},
			schema: swagger.Schema{
				Properties: map[string]*swagger.Property{
					"prop1": {
						PropertyState: swagger.PropertyStateCovered,
					},
				},
			},
			hasError: true,
		},
	}

	for idx, c := range cases {
		err := c.grant.Grant(&c.schema, true)
		if c.hasError {
			require.Error(t, err, idx)
			continue
		}
		require.NoError(t, err, idx)
		require.Equal(t, c.expectSchema, c.schema, idx)
	}
}
