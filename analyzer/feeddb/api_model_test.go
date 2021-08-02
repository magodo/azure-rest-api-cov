package feeddb_test

import (
	"github.com/magodo/azure-rest-api-cov/analyzer/addr"
	"github.com/magodo/azure-rest-api-cov/analyzer/feeddb"
	"github.com/magodo/azure-rest-api-cov/analyzer/swagger"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestProperty_Add(t *testing.T) {
	type swgproperty struct {
		addr  addr.Property
		state swagger.PropertyState
	}
	cases := []struct {
		propertiesToAdd []swgproperty
		expect          feeddb.Property
	}{
		{
			propertiesToAdd: []swgproperty{
				{
					addr:  addr.MustParseProperty("prop1/not_covered"),
					state: swagger.PropertyStateUncovered,
				},
				{
					addr:  addr.MustParseProperty("prop1/covered"),
					state: swagger.PropertyStateCovered,
				},
				{
					addr:  addr.MustParseProperty("prop2/covered"),
					state: swagger.PropertyStateCovered,
				},
				{
					addr:  addr.MustParseProperty("prop_ignored"),
					state: swagger.PropertyStateIgnoredForReqROProp,
				},
			},
			expect: feeddb.Property{
				Properties: feeddb.Properties{
					addr.MustParsePropertySegment("prop1"): {
						Properties: feeddb.Properties{
							addr.MustParsePropertySegment("not_covered"): {
								IsCovered: boolPtr(false),
							},
							addr.MustParsePropertySegment("covered"): {
								IsCovered: boolPtr(true),
							},
						},
					},
					addr.MustParsePropertySegment("prop2"): {
						Properties: feeddb.Properties{
							addr.MustParsePropertySegment("covered"): {
								IsCovered: boolPtr(true),
							},
						},
					},
				},
			},
		},
	}

	for idx, c := range cases {
		var property feeddb.Property
		for _, prop := range c.propertiesToAdd {
			property.Add(prop.addr, prop.state)
		}
		require.Equal(t, c.expect, property, idx)
	}
}

func TestNewAPIModel(t *testing.T) {
	cases := []struct {
		schema swagger.Schema
		expect *feeddb.APIModel
	}{
		{
			schema: swagger.Schema{
				SchemaMetaInfo: swagger.SchemaMetaInfo{
					Name:           "foo",
					SwaggerRelPath: "foo/resource-manager/Microsoft.Foo/preview/2021-02-01/foo.json",
					Active:         true,
				},
				Properties: swagger.Properties{
					"prop1/covered": &swagger.Property{
						PropertyState: swagger.PropertyStateCovered,
					},
					"prop1/not_covered": {
						PropertyState: swagger.PropertyStateUncovered,
					},
					"prop2/covered": {
						PropertyState: swagger.PropertyStateCovered,
					},
					"prop2/ignored": {
						PropertyState: swagger.PropertyStateIgnoredForReqROProp,
					},
				},
			},
			expect: &feeddb.APIModel{
				Name:      "foo",
				Properties: feeddb.Properties{
					addr.MustParsePropertySegment("prop1"): {
						Properties: feeddb.Properties{
							addr.MustParsePropertySegment("not_covered"): {
								IsCovered: boolPtr(false),
							},
							addr.MustParsePropertySegment("covered"): {
								IsCovered: boolPtr(true),
							},
						},
					},
					addr.MustParsePropertySegment("prop2"): {
						Properties: feeddb.Properties{
							addr.MustParsePropertySegment("covered"): {
								IsCovered: boolPtr(true),
							},
						},
					},
				},
			},
		},
	}

	for idx, c := range cases {
		store, err := feeddb.NewAPIModel(c.schema)
		require.NoError(t, err, idx)
		require.Equal(t, c.expect, store, idx)
	}
}

func boolPtr(b bool) *bool {
	return &b
}
