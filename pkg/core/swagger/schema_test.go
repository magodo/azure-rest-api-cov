package swagger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/magodo/azure-rest-api-cov/pkg/core/swaggerloader"

	"github.com/magodo/azure-rest-api-cov/pkg/core/addr"
	"github.com/stretchr/testify/require"
)

var (
	propStateUncovered = NewPropertyState(PropStateUncovered, "")
	propStateCovered   = NewPropertyState(PropStateCovered, "")
)

func TestSchema(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	specBasePathLocal := filepath.Join(filepath.Dir(pwd), "testdata", "swagger")
	specFooPathLocal := filepath.Join(specBasePathLocal, "foo.json")
	specBarPathLocal := filepath.Join(specBasePathLocal, "bar.json")
	specBazPathLocal := filepath.Join(specBasePathLocal, "baz", "baz.json")
	specRemotePathLocal := filepath.Join(specBasePathLocal, "remote.json")
	specFoo, err := swaggerloader.LoadSwagger(specFooPathLocal)
	require.NoError(t, err)
	specBar, err := swaggerloader.LoadSwagger(specBarPathLocal)
	require.NoError(t, err)
	specBaz, err := swaggerloader.LoadSwagger(specBazPathLocal)
	require.NoError(t, err)
	specRemote, err := swaggerloader.LoadSwagger(specRemotePathLocal)
	require.NoError(t, err)

	specRemoteBaseURL := "https://gist.githubusercontent.com/magodo/f054bb1c2e7a1c74fd78f65eb42a17bb/raw/8308d9295287bc590be9b431ea0353b626da7ada"
	specRemoteFileURL := specRemoteBaseURL + "/remote.json"

	_ = specFoo
	_ = specBar
	_ = specBaz
	_ = specRemote
	_ = specRemoteFileURL

	cases := []struct {
		specBaseURL string
		specRelPath string
		schemaName  string
		err         error
		expect      Schema
	}{
		{
			specBaseURL: specBasePathLocal,
			specRelPath: "foo.json",
			schemaName:  "def_foo",
			err:         nil,
			expect: Schema{
				SchemaAddr: SchemaAddr{
					SwaggerRelPath: "foo.json",
					Name:           "def_foo",
				},
				Properties: map[string]*Property{
					"foo": {
						PropertyState: propStateUncovered,
						schema:        specFoo.Definitions["def_foo"].Properties["foo"],
						resolvedRefs: map[string]interface{}{
							specFooPathLocal + "#/definitions/def_foo": struct{}{},
						},
						swaggerURI: specFooPathLocal,
					},
				},
				swaggerURI: specFooPathLocal,
				swagger:    specFoo,
			},
		},
		{
			specBaseURL: specBasePathLocal,
			specRelPath: "foo.json",
			schemaName:  "def_regular",
			err:         nil,
			expect: Schema{
				SchemaAddr: SchemaAddr{
					SwaggerRelPath: "foo.json",
					Name:           "def_regular",
				},
				Properties: map[string]*Property{
					"prop_primitive": {
						PropertyState: propStateUncovered,
						schema:        specFoo.Definitions["def_regular"].Properties["prop_primitive"],
						resolvedRefs: map[string]interface{}{
							specFooPathLocal + "#/definitions/def_regular": struct{}{},
						},
						swaggerURI: specFooPathLocal,
					},
					"prop_array_of_primitive": {
						PropertyState: propStateUncovered,
						schema:        specFoo.Definitions["def_regular"].Properties["prop_array_of_primitive"],
						resolvedRefs: map[string]interface{}{
							specFooPathLocal + "#/definitions/def_regular": struct{}{},
						},
						swaggerURI: specFooPathLocal,
					},
					"prop_array_of_object": {
						PropertyState: propStateUncovered,
						schema:        specFoo.Definitions["def_regular"].Properties["prop_array_of_object"],
						resolvedRefs: map[string]interface{}{
							specFooPathLocal + "#/definitions/def_regular": struct{}{},
						},
						swaggerURI: specFooPathLocal,
					},
					"prop_object": {
						PropertyState: propStateUncovered,
						schema:        specFoo.Definitions["def_regular"].Properties["prop_object"],
						resolvedRefs: map[string]interface{}{
							specFooPathLocal + "#/definitions/def_regular": struct{}{},
						},
						swaggerURI: specFooPathLocal,
					},
				},
				swaggerURI: specFooPathLocal,
				swagger:    specFoo,
			},
		},
		{
			specBaseURL: specBasePathLocal,
			specRelPath: "foo.json",
			schemaName:  "def_prop_in_file_ref",
			err:         nil,
			expect: Schema{
				SchemaAddr: SchemaAddr{
					SwaggerRelPath: "foo.json",
					Name:           "def_prop_in_file_ref",
				},
				Properties: map[string]*Property{
					"prop": {
						PropertyState: propStateUncovered,
						schema:        specFoo.Definitions["def_prop_in_file_ref"].Properties["prop"],
						resolvedRefs: map[string]interface{}{
							specFooPathLocal + "#/definitions/def_prop_in_file_ref": struct{}{},
						},
						swaggerURI: specFooPathLocal,
					},
				},
				swaggerURI: specFooPathLocal,
				swagger:    specFoo,
			},
		},
		{
			specBaseURL: specBasePathLocal,
			specRelPath: "foo.json",
			schemaName:  "def_prop_cross_file_ref",
			err:         nil,
			expect: Schema{
				SchemaAddr: SchemaAddr{
					SwaggerRelPath: "foo.json",
					Name:           "def_prop_cross_file_ref",
				},
				Properties: map[string]*Property{
					"prop": {
						PropertyState: propStateUncovered,
						schema:        specFoo.Definitions["def_prop_cross_file_ref"].Properties["prop"],
						resolvedRefs: map[string]interface{}{
							specFooPathLocal + "#/definitions/def_prop_cross_file_ref": struct{}{},
						},
						swaggerURI: specFooPathLocal,
					},
				},
				swaggerURI: specFooPathLocal,
				swagger:    specFoo,
			},
		},
		{
			specBaseURL: specBasePathLocal,
			specRelPath: "foo.json",
			schemaName:  "def_prop_self_ref",
			err:         nil,
			expect: Schema{
				SchemaAddr: SchemaAddr{
					SwaggerRelPath: "foo.json",
					Name:           "def_prop_self_ref",
				},
				Properties: map[string]*Property{
					"prop": {
						PropertyState: propStateUncovered,
						schema:        specFoo.Definitions["def_prop_self_ref"].Properties["prop"],
						resolvedRefs: map[string]interface{}{
							specFooPathLocal + "#/definitions/def_prop_self_ref": struct{}{},
						},
						swaggerURI: specFooPathLocal,
					},
				},
				swaggerURI: specFooPathLocal,
				swagger:    specFoo,
			},
		},
		{
			specBaseURL: specBasePathLocal,
			specRelPath: "foo.json",
			schemaName:  "def_array_ref",
			err:         nil,
			expect: Schema{
				SchemaAddr: SchemaAddr{
					SwaggerRelPath: "foo.json",
					Name:           "def_array_ref",
				},
				Properties: map[string]*Property{
					"foo": {
						PropertyState: propStateUncovered,
						schema:        specFoo.Definitions["def_foo"].Properties["foo"],
						resolvedRefs: map[string]interface{}{
							specFooPathLocal + "#/definitions/def_foo":       struct{}{},
							specFooPathLocal + "#/definitions/def_array_ref": struct{}{},
						},
						swaggerURI: specFooPathLocal,
					},
				},
				swaggerURI: specFooPathLocal,
				swagger:    specFoo,
			},
		},
		{
			specBaseURL: specBasePathLocal,
			specRelPath: "foo.json",
			schemaName:  "def_in_file_ref",
			err:         nil,
			expect: Schema{
				SchemaAddr: SchemaAddr{
					SwaggerRelPath: "foo.json",
					Name:           "def_in_file_ref",
				},
				Properties: map[string]*Property{
					"foo": {
						PropertyState: propStateUncovered,
						schema:        specFoo.Definitions["def_foo"].Properties["prop"],
						resolvedRefs: map[string]interface{}{
							specFooPathLocal + "#/definitions/def_in_file_ref": struct{}{},
							specFooPathLocal + "#/definitions/def_foo":         struct{}{},
						},
						swaggerURI: specFooPathLocal,
					},
				},
				swaggerURI: specFooPathLocal,
				swagger:    specFoo,
			},
		},
		{
			specBaseURL: specBasePathLocal,
			specRelPath: "foo.json",
			schemaName:  "def_cross_file_ref",
			err:         nil,
			expect: Schema{
				SchemaAddr: SchemaAddr{
					SwaggerRelPath: "foo.json",
					Name:           "def_cross_file_ref",
				},
				Properties: map[string]*Property{
					"bar": {
						PropertyState: propStateUncovered,
						schema:        specBar.Definitions["def_bar"].Properties["bar"],
						resolvedRefs: map[string]interface{}{
							specFooPathLocal + "#/definitions/def_cross_file_ref": struct{}{},
							specBarPathLocal + "#/definitions/def_bar":            struct{}{},
						},
						swaggerURI: specBarPathLocal,
					},
				},
				swaggerURI: specFooPathLocal,
				swagger:    specFoo,
			},
		},
		{
			specBaseURL: specBasePathLocal,
			specRelPath: "foo.json",
			schemaName:  "def_self_ref",
			err:         nil,
			expect: Schema{
				SchemaAddr: SchemaAddr{
					SwaggerRelPath: "foo.json",
					Name:           "def_self_ref",
				},
				Properties: map[string]*Property{
					"": {
						PropertyState: propStateUncovered,
						schema:        specFoo.Definitions["def_self_ref"],
						resolvedRefs: map[string]interface{}{
							specFooPathLocal + "#/definitions/def_self_ref": struct{}{},
						},
						swaggerURI: specFooPathLocal,
					},
				},
				swaggerURI: specFooPathLocal,
				swagger:    specFoo,
			},
		},
		{
			specBaseURL: specBasePathLocal,
			specRelPath: "foo.json",
			schemaName:  "def_all_of",
			expect: Schema{
				SchemaAddr: SchemaAddr{
					SwaggerRelPath: "foo.json",
					Name:           "def_all_of",
				},
				Properties: map[string]*Property{
					"prop_nested": {
						PropertyState: propStateUncovered,
						schema:        specFoo.Definitions["def_all_of"].AllOf[0].Properties["prop_nested"],
						resolvedRefs: map[string]interface{}{
							specFooPathLocal + "#/definitions/def_all_of": struct{}{},
						},
						swaggerURI: specFooPathLocal,
					},
					"foo": {
						PropertyState: propStateUncovered,
						schema:        specFoo.Definitions["def_foo"].Properties["foo"],
						resolvedRefs: map[string]interface{}{
							specFooPathLocal + "#/definitions/def_all_of":  struct{}{},
							specFooPathLocal + "#/definitions/def_all_of2": struct{}{},
							specFooPathLocal + "#/definitions/def_foo":     struct{}{},
						},
						swaggerURI: specFooPathLocal,
					},
					"bar": {
						PropertyState: propStateUncovered,
						schema:        specBar.Definitions["def_bar"].Properties["bar"],
						resolvedRefs: map[string]interface{}{
							specFooPathLocal + "#/definitions/def_all_of": struct{}{},
							specBarPathLocal + "#/definitions/def_bar":    struct{}{},
						},
						swaggerURI: specBarPathLocal,
					},
					"baz": {
						PropertyState: propStateUncovered,
						schema:        specBaz.Definitions["def_baz"].Properties["baz"],
						resolvedRefs: map[string]interface{}{
							specFooPathLocal + "#/definitions/def_all_of": struct{}{},
							specBazPathLocal + "#/definitions/def_all_of": struct{}{},
							specBazPathLocal + "#/definitions/def_baz":    struct{}{},
						},
						swaggerURI: specBazPathLocal,
					},
					"p1": {
						PropertyState: propStateUncovered,
						schema:        specFoo.Definitions["def_all_of"].Properties["p1"],
						resolvedRefs: map[string]interface{}{
							specFooPathLocal + "#/definitions/def_all_of": struct{}{},
						},
						swaggerURI: specFooPathLocal,
					},
				},
				swaggerURI: specFooPathLocal,
				swagger:    specFoo,
			},
		},
		{
			specBaseURL: specRemoteBaseURL,
			specRelPath: "remote.json",
			schemaName:  "def_foo",
			expect: Schema{
				SchemaAddr: SchemaAddr{
					SwaggerRelPath: "remote.json",
					Name:           "def_foo",
				},
				Properties: map[string]*Property{
					"prop_primitive": {
						PropertyState: propStateUncovered,
						schema:        specRemote.Definitions["def_foo"].Properties["prop_primitive"],
						resolvedRefs: map[string]interface{}{
							specRemoteFileURL + "#/definitions/def_foo": struct{}{},
						},
						swaggerURI: specRemoteFileURL,
					},
				},
				swaggerURI: specRemoteFileURL,
				swagger:    specRemote,
			},
		},
		{
			specBaseURL: specBasePathLocal,
			specRelPath: "foo.json",
			schemaName:  "def_base",
			expect: Schema{
				SchemaAddr: SchemaAddr{
					SwaggerRelPath: "foo.json",
					Name:           "def_base",
				},
				Properties: map[string]*Property{
					"{def_variant1}": {
						PropertyState: propStateUncovered,
						schema:        specFoo.Definitions["def_variant1"],
						resolvedRefs: map[string]interface{}{
							specFooPathLocal + "#/definitions/def_variant1": struct{}{},
						},
						swaggerURI: specFooPathLocal,
					},
					"{def_variant2}": {
						PropertyState: propStateUncovered,
						schema:        specFoo.Definitions["variant2_def"],
						resolvedRefs: map[string]interface{}{
							specFooPathLocal + "#/definitions/variant2_def": struct{}{},
						},
						swaggerURI: specFooPathLocal,
					},
				},
				swaggerURI: specFooPathLocal,
				swagger:    specFoo,
			},
		},
		{
			specBaseURL: specBasePathLocal,
			specRelPath: "foo.json",
			schemaName:  "def_to_def_base",
			expect: Schema{
				SchemaAddr: SchemaAddr{
					SwaggerRelPath: "foo.json",
					Name:           "def_to_def_base",
				},
				Properties: map[string]*Property{
					"prop": {
						PropertyState: propStateUncovered,
						schema:        specFoo.Definitions["def_to_def_base"].Properties["prop"],
						resolvedRefs: map[string]interface{}{
							specFooPathLocal + "#/definitions/def_to_def_base": struct{}{},
						},
						swaggerURI: specFooPathLocal,
					},
				},
				swaggerURI: specFooPathLocal,
				swagger:    specFoo,
			},
		},
		{
			specBaseURL: specBasePathLocal,
			specRelPath: "foo.json",
			schemaName:  "non_existent_schema",
			err:         fmt.Errorf("schema %q is not found in swagger spec %q", "non_existent_schema", specBasePathLocal+"/foo.json"),
		},
	}

	for _, c := range cases {
		msg := fmt.Sprintf("%s: %s", filepath.Join(c.specBaseURL, c.specRelPath), c.schemaName)
		actual, err := NewEmptySchema(c.specBaseURL, c.specRelPath, c.schemaName)
		require.Equal(t, c.err, err, msg)
		if err == nil {
			require.Equal(t, c.expect, *actual, msg)
		}
	}
}

func TestSchema_ExpandPropertyOneLevelDeep(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	specBasePath := filepath.Join(filepath.Dir(pwd), "testdata", "swagger")
	specFooPath := filepath.Join(specBasePath, "foo.json")
	specBarPath := filepath.Join(specBasePath, "bar.json")
	specFoo, err := swaggerloader.LoadSwagger(specFooPath)
	specBar, err := swaggerloader.LoadSwagger(specBarPath)
	_ = specBar
	_ = specFoo
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name           string
		swaggerRelPath string
		schemaName     string
		expandAddrs    []addr.Property
		err            error
		expect         Schema
	}{
		{
			name:           `Expand "def_foo"`,
			swaggerRelPath: "foo.json",
			schemaName:     "def_foo",
			expandAddrs: []addr.Property{
				addr.MustParseProperty("foo"),
			},
			expect: Schema{
				SchemaAddr: SchemaAddr{
					SwaggerRelPath: "foo.json",
					Name:           "def_foo",
				},
				Properties: map[string]*Property{
					"foo": {
						PropertyState: propStateUncovered,
						schema:        specFoo.Definitions["def_foo"].Properties["foo"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_foo": struct{}{},
						},
						swaggerURI: specFooPath,
					},
				},
				swaggerURI: specFooPath,
				swagger:    specFoo,
			},
		},
		{
			name:           `Expand "def_foo" multiple times should be fine (re-entry)`,
			swaggerRelPath: "foo.json",
			schemaName:     "def_foo",
			expandAddrs: []addr.Property{
				addr.MustParseProperty("foo"),
				addr.MustParseProperty("foo"),
			},
			err: nil,
			expect: Schema{
				SchemaAddr: SchemaAddr{
					SwaggerRelPath: "foo.json",
					Name:           "def_foo",
				},
				Properties: map[string]*Property{
					"foo": {
						PropertyState: propStateUncovered,
						schema:        specFoo.Definitions["def_foo"].Properties["foo"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_foo": struct{}{},
						},
						swaggerURI: specFooPath,
					},
				},
				swaggerURI: specFooPath,
				swagger:    specFoo,
			},
		},
		{
			name:           `Expand "def_regular"`,
			swaggerRelPath: "foo.json",
			schemaName:     "def_regular",
			expandAddrs: []addr.Property{
				addr.MustParseProperty("prop_array_of_object"),
				addr.MustParseProperty("prop_object"),
			},
			err: nil,
			expect: Schema{
				SchemaAddr: SchemaAddr{
					SwaggerRelPath: "foo.json",
					Name:           "def_regular",
				},
				Properties: map[string]*Property{
					"prop_primitive": {
						PropertyState: propStateUncovered,
						schema:        specFoo.Definitions["def_regular"].Properties["prop_primitive"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_regular": struct{}{},
						},
						swaggerURI: specFooPath,
					},
					"prop_array_of_primitive": {
						PropertyState: propStateUncovered,
						schema:        specFoo.Definitions["def_regular"].Properties["prop_array_of_primitive"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_regular": struct{}{},
						},
						swaggerURI: specFooPath,
					},
					"prop_array_of_object.prop_nested": {
						PropertyState: propStateUncovered,
						schema:        specFoo.Definitions["def_regular"].Properties["prop_array_of_object"].Items.Schema.Properties["prop_nested"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_regular": struct{}{},
						},
						swaggerURI: specFooPath,
					},
					"prop_object.prop_nested": {
						PropertyState: propStateUncovered,
						schema:        specFoo.Definitions["def_regular"].Properties["prop_object"].Properties["prop_nested"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_regular": struct{}{},
						},
						swaggerURI: specFooPath,
					},
				},
				swaggerURI: specFooPath,
				swagger:    specFoo,
			},
		},
		{
			name:           `Expand "def_prop_in_file_ref"`,
			swaggerRelPath: "foo.json",
			schemaName:     "def_prop_in_file_ref",
			expandAddrs: []addr.Property{
				addr.MustParseProperty("prop"),
			},
			expect: Schema{
				SchemaAddr: SchemaAddr{
					SwaggerRelPath: "foo.json",
					Name:           "def_prop_in_file_ref",
				},
				Properties: map[string]*Property{
					"prop.foo": {
						PropertyState: propStateUncovered,
						schema:        specFoo.Definitions["def_foo"].Properties["foo"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_prop_in_file_ref": struct{}{},
							specFooPath + "#/definitions/def_foo":              struct{}{},
						},
						swaggerURI: specFooPath,
					},
				},
				swaggerURI: specFooPath,
				swagger:    specFoo,
			},
		},
		{
			name:           `Expand "def_prop_self_ref"`,
			swaggerRelPath: "foo.json",
			schemaName:     "def_prop_self_ref",
			expandAddrs: []addr.Property{
				addr.MustParseProperty("prop"),
			},
			expect: Schema{
				SchemaAddr: SchemaAddr{
					SwaggerRelPath: "foo.json",
					Name:           "def_prop_self_ref",
				},
				Properties: map[string]*Property{
					"prop": {
						PropertyState: propStateUncovered,
						schema:        specFoo.Definitions["def_prop_self_ref"].Properties["prop"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_prop_self_ref": struct{}{},
						},
						swaggerURI: specFooPath,
					},
				},
				swaggerURI: specFooPath,
				swagger:    specFoo,
			},
		},
		{
			name:           `Expand "def_prop_cross_file_ref"`,
			swaggerRelPath: "foo.json",
			schemaName:     "def_prop_cross_file_ref",
			expandAddrs: []addr.Property{
				addr.MustParseProperty("prop"),
			},
			expect: Schema{
				SchemaAddr: SchemaAddr{
					SwaggerRelPath: "foo.json",
					Name:           "def_prop_cross_file_ref",
				},
				Properties: map[string]*Property{
					"prop.bar": {
						PropertyState: propStateUncovered,
						schema:        specBar.Definitions["def_bar"].Properties["bar"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_prop_cross_file_ref": struct{}{},
							specBarPath + "#/definitions/def_bar":                 struct{}{},
						},
						swaggerURI: specBarPath,
					},
				},
				swaggerURI: specFooPath,
				swagger:    specFoo,
			},
		},
		{
			name:           `Expand "def_in_file_ref"`,
			swaggerRelPath: "foo.json",
			schemaName:     "def_in_file_ref",
			expandAddrs: []addr.Property{
				addr.MustParseProperty("foo"),
			},
			expect: Schema{
				SchemaAddr: SchemaAddr{
					SwaggerRelPath: "foo.json",
					Name:           "def_in_file_ref",
				},
				Properties: map[string]*Property{
					"foo": {
						PropertyState: propStateUncovered,
						schema:        specFoo.Definitions["def_foo"].Properties["foo"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_in_file_ref": struct{}{},
							specFooPath + "#/definitions/def_foo":         struct{}{},
						},
						swaggerURI: specFooPath,
					},
				},
				swaggerURI: specFooPath,
				swagger:    specFoo,
			},
		},
		{
			name:           `Expand "def_cross_file_ref"`,
			swaggerRelPath: "foo.json",
			schemaName:     "def_cross_file_ref",
			expandAddrs: []addr.Property{
				addr.MustParseProperty("bar"),
			},
			expect: Schema{
				SchemaAddr: SchemaAddr{
					SwaggerRelPath: "foo.json",
					Name:           "def_cross_file_ref",
				},
				Properties: map[string]*Property{
					"bar": {
						PropertyState: propStateUncovered,
						schema:        specBar.Definitions["def_bar"].Properties["bar"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_cross_file_ref": struct{}{},
							specBarPath + "#/definitions/def_bar":            struct{}{},
						},
						swaggerURI: specBarPath,
					},
				},
				swaggerURI: specFooPath,
				swagger:    specFoo,
			},
		},
		{
			name:           `Expand "def_self_ref"`,
			swaggerRelPath: "foo.json",
			schemaName:     "def_self_ref",
			expandAddrs: []addr.Property{
				addr.MustParseProperty(""),
			},
			expect: Schema{
				SchemaAddr: SchemaAddr{
					SwaggerRelPath: "foo.json",
					Name:           "def_self_ref",
				},
				Properties: map[string]*Property{
					"": {
						PropertyState: propStateUncovered,
						schema:        specFoo.Definitions["def_self_ref"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_self_ref": struct{}{},
						},
						swaggerURI: specFooPath,
					},
				},
				swaggerURI: specFooPath,
				swagger:    specFoo,
			},
		},
		{
			name:           `Expand "def_base" (schema itself is a discriminator)`,
			swaggerRelPath: "foo.json",
			schemaName:     "def_base",
			expandAddrs: []addr.Property{
				addr.MustParseProperty("{def_variant1}"),
			},
			expect: Schema{
				SchemaAddr: SchemaAddr{
					SwaggerRelPath: "foo.json",
					Name:           "def_base",
				},
				Properties: map[string]*Property{
					"{def_variant1}.type": {
						PropertyState: propStateUncovered,
						schema:        specFoo.Definitions["def_base"].Properties["type"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_variant1": struct{}{},
						},
						swaggerURI: specFooPath,
					},
					"{def_variant2}": {
						PropertyState: propStateUncovered,
						schema:        specFoo.Definitions["variant2_def"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/variant2_def": struct{}{},
						},
						swaggerURI: specFooPath,
					},
				},
				swaggerURI: specFooPath,
				swagger:    specFoo,
			},
		},
		{
			name:           `Expand "def_to_def_base" (property is a discriminator)`,
			swaggerRelPath: "foo.json",
			schemaName:     "def_to_def_base",
			expandAddrs: []addr.Property{
				addr.MustParseProperty("prop"),
			},
			expect: Schema{
				SchemaAddr: SchemaAddr{
					SwaggerRelPath: "foo.json",
					Name:           "def_to_def_base",
				},
				Properties: map[string]*Property{
					"prop{def_variant1}": {
						PropertyState: propStateUncovered,
						schema:        specFoo.Definitions["def_variant1"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_to_def_base": struct{}{},
							specFooPath + "#/definitions/def_variant1":    struct{}{},
						},
						swaggerURI: specFooPath,
					},
					"prop{def_variant2}": {
						PropertyState: propStateUncovered,
						schema:        specFoo.Definitions["variant2_def"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_to_def_base": struct{}{},
							specFooPath + "#/definitions/variant2_def":    struct{}{},
						},
						swaggerURI: specFooPath,
					},
				},
				swaggerURI: specFooPath,
				swagger:    specFoo,
			},
		},
	}

	for _, c := range cases {
		swgschema, err := NewEmptySchema(specBasePath, c.swaggerRelPath, c.schemaName)
		require.Equal(t, c.err, err, c.name)
		if err == nil {
			for _, addr := range c.expandAddrs {
				require.NoError(t, swgschema.expandPropertyOneLevelDeep(addr), c.name)
			}
			require.Equal(t, c.expect, *swgschema, c.name)
		}
	}
}

func TestSchema_AddCovProperty(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	specBasePath := filepath.Join(filepath.Dir(pwd), "testdata", "swagger")
	specFooPath := filepath.Join(specBasePath, "foo.json")
	specBarPath := filepath.Join(specBasePath, "bar.json")
	specFoo, err := swaggerloader.LoadSwagger(specFooPath)
	specBar, err := swaggerloader.LoadSwagger(specBarPath)
	_ = specBar
	if err != nil {
		t.Fatal(err)
	}

	type step struct {
		swgPropAddr addr.Property
		err         bool
		expect      Schema
	}

	cases := []struct {
		name           string
		swaggerRelPath string
		schemaName     string
		steps          []step
	}{
		// 0
		{
			name:           "add same property twice should take no diff (re-entry)",
			swaggerRelPath: "foo.json",
			schemaName:     "def_cov",
			steps: []step{
				{
					swgPropAddr: addr.MustParseProperty("p1"),
					expect: Schema{
						SchemaAddr: SchemaAddr{
							SwaggerRelPath: "foo.json",
							Name:           "def_cov",
						},
						Properties: map[string]*Property{
							"p1": {
								PropertyState: propStateCovered,
								schema:        specFoo.Definitions["def_cov"].Properties["p1"],
								resolvedRefs: map[string]interface{}{
									specFooPath + "#/definitions/def_cov": struct{}{},
								},
								swaggerURI: specFooPath,
							},
							"p2": {
								PropertyState: propStateUncovered,
								schema:        specFoo.Definitions["def_cov"].Properties["p2"],
								resolvedRefs: map[string]interface{}{
									specFooPath + "#/definitions/def_cov": struct{}{},
								},
								swaggerURI: specFooPath,
							},
						},
						swaggerURI: specFooPath,
						swagger:    specFoo,
					},
				},
				// add a second tf link to the same swg property should have no change
				{
					swgPropAddr: addr.MustParseProperty("p1"),
					expect: Schema{
						SchemaAddr: SchemaAddr{
							SwaggerRelPath: "foo.json",
							Name:           "def_cov",
						},
						Properties: map[string]*Property{
							"p1": {
								PropertyState: propStateCovered,
								schema:        specFoo.Definitions["def_cov"].Properties["p1"],
								resolvedRefs: map[string]interface{}{
									specFooPath + "#/definitions/def_cov": struct{}{},
								},
								swaggerURI: specFooPath,
							},
							"p2": {
								PropertyState: propStateUncovered,
								schema:        specFoo.Definitions["def_cov"].Properties["p2"],
								resolvedRefs: map[string]interface{}{
									specFooPath + "#/definitions/def_cov": struct{}{},
								},
								swaggerURI: specFooPath,
							},
						},
						swaggerURI: specFooPath,
						swagger:    specFoo,
					},
				},
			},
		},
		// 1
		{
			name:           "add tf link to swagger child property first, then add another tf link to swagger parent property",
			swaggerRelPath: "foo.json",
			schemaName:     "def_cov",
			steps: []step{
				{
					swgPropAddr: addr.MustParseProperty("p2.p2_1"),
					expect: Schema{
						SchemaAddr: SchemaAddr{
							SwaggerRelPath: "foo.json",
							Name:           "def_cov",
						},
						Properties: map[string]*Property{
							"p1": {
								PropertyState: propStateUncovered,
								schema:        specFoo.Definitions["def_cov"].Properties["p1"],
								resolvedRefs: map[string]interface{}{
									specFooPath + "#/definitions/def_cov": struct{}{},
								},
								swaggerURI: specFooPath,
							},
							"p2.p2_1": {
								PropertyState: propStateCovered,
								schema:        specFoo.Definitions["def_cov"].Properties["p2"].Properties["p2_1"],
								resolvedRefs: map[string]interface{}{
									specFooPath + "#/definitions/def_cov": struct{}{},
								},
								swaggerURI: specFooPath,
							},
							"p2.p2_2": {
								PropertyState: propStateUncovered,
								schema:        specFoo.Definitions["def_cov"].Properties["p2"].Properties["p2_2"],
								resolvedRefs: map[string]interface{}{
									specFooPath + "#/definitions/def_cov": struct{}{},
								},
								swaggerURI: specFooPath,
							},
						},
						swaggerURI: specFooPath,
						swagger:    specFoo,
					},
				},
				{
					swgPropAddr: addr.MustParseProperty("p2"),
					expect: Schema{
						SchemaAddr: SchemaAddr{
							SwaggerRelPath: "foo.json",
							Name:           "def_cov",
						},
						Properties: map[string]*Property{
							"p1": {
								PropertyState: propStateUncovered,
								schema:        specFoo.Definitions["def_cov"].Properties["p1"],
								resolvedRefs: map[string]interface{}{
									specFooPath + "#/definitions/def_cov": struct{}{},
								},
								swaggerURI: specFooPath,
							},
							"p2.p2_1": {
								PropertyState: propStateCovered,
								schema:        specFoo.Definitions["def_cov"].Properties["p2"].Properties["p2_1"],
								resolvedRefs: map[string]interface{}{
									specFooPath + "#/definitions/def_cov": struct{}{},
								},
								swaggerURI: specFooPath,
							},
							"p2.p2_2": {
								PropertyState: propStateCovered,
								schema:        specFoo.Definitions["def_cov"].Properties["p2"].Properties["p2_2"],
								resolvedRefs: map[string]interface{}{
									specFooPath + "#/definitions/def_cov": struct{}{},
								},
								swaggerURI: specFooPath,
							},
						},
						swaggerURI: specFooPath,
						swagger:    specFoo,
					},
				},
			},
		},
		// 2
		{
			name:           "add tf link to swagger parent property first, then add another tf link to swagger child property",
			swaggerRelPath: "foo.json",
			schemaName:     "def_cov",
			steps: []step{
				{
					swgPropAddr: addr.MustParseProperty("p2"),
					expect: Schema{
						SchemaAddr: SchemaAddr{
							SwaggerRelPath: "foo.json",
							Name:           "def_cov",
						},
						Properties: map[string]*Property{
							"p1": {
								PropertyState: propStateUncovered,
								schema:        specFoo.Definitions["def_cov"].Properties["p1"],
								resolvedRefs: map[string]interface{}{
									specFooPath + "#/definitions/def_cov": struct{}{},
								},
								swaggerURI: specFooPath,
							},
							"p2": {
								PropertyState: propStateCovered,
								schema:        specFoo.Definitions["def_cov"].Properties["p2"],
								resolvedRefs: map[string]interface{}{
									specFooPath + "#/definitions/def_cov": struct{}{},
								},
								swaggerURI: specFooPath,
							},
						},
						swaggerURI: specFooPath,
						swagger:    specFoo,
					},
				},
				{
					swgPropAddr: addr.MustParseProperty("p2.p2_1"),
					expect: Schema{
						SchemaAddr: SchemaAddr{
							SwaggerRelPath: "foo.json",
							Name:           "def_cov",
						},
						Properties: map[string]*Property{
							"p1": {
								PropertyState: propStateUncovered,
								schema:        specFoo.Definitions["def_cov"].Properties["p1"],
								resolvedRefs: map[string]interface{}{
									specFooPath + "#/definitions/def_cov": struct{}{},
								},
								swaggerURI: specFooPath,
							},
							"p2.p2_1": {
								PropertyState: propStateCovered,
								schema:        specFoo.Definitions["def_cov"].Properties["p2"].Properties["p2_1"],
								resolvedRefs: map[string]interface{}{
									specFooPath + "#/definitions/def_cov": struct{}{},
								},
								swaggerURI: specFooPath,
							},
							"p2.p2_2": {
								PropertyState: propStateCovered,
								schema:        specFoo.Definitions["def_cov"].Properties["p2"].Properties["p2_2"],
								resolvedRefs: map[string]interface{}{
									specFooPath + "#/definitions/def_cov": struct{}{},
								},
								swaggerURI: specFooPath,
							},
						},
						swaggerURI: specFooPath,
						swagger:    specFoo,
					},
				},
			},
		},
		// 3
		{
			name:           "add discriminator property",
			swaggerRelPath: "foo.json",
			schemaName:     "def_base",
			steps: []step{
				{
					swgPropAddr: addr.MustParseProperty("{def_variant1}.type"),
					expect: Schema{
						SchemaAddr: SchemaAddr{
							SwaggerRelPath: "foo.json",
							Name:           "def_base",
						},
						Properties: map[string]*Property{
							"{def_variant1}.type": {
								PropertyState: propStateCovered,
								schema:        specFoo.Definitions["def_base"].Properties["type"],
								resolvedRefs: map[string]interface{}{
									specFooPath + "#/definitions/def_variant1": struct{}{},
								},
								swaggerURI: specFooPath,
							},
							"{def_variant2}": {
								PropertyState: propStateUncovered,
								schema:        specFoo.Definitions["variant2_def"],
								resolvedRefs: map[string]interface{}{
									specFooPath + "#/definitions/variant2_def": struct{}{},
								},
								swaggerURI: specFooPath,
							},
						},
						swaggerURI: specFooPath,
						swagger:    specFoo,
					},
				},
			},
		},
	}

	for idx, c := range cases {
		schema, err := NewEmptySchema(specBasePath, c.swaggerRelPath, c.schemaName)
		require.NoError(t, err, idx)
		if err == nil {
			for iidx, s := range c.steps {
				if s.err {
					require.Error(t, schema.AddCovProperty(s.swgPropAddr), fmt.Sprintf("%s: %d", c.name, iidx))
					continue
				}
				require.NoError(t, schema.AddCovProperty(s.swgPropAddr), fmt.Sprintf("%s: %d", c.name, iidx))
				require.Equal(t, s.expect, *schema, fmt.Sprintf("%d.%d", idx, iidx))
			}
		}
	}
}

func TestSchema_Marshal(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	specBasePath := filepath.Join(filepath.Dir(pwd), "testdata", "swagger")

	type process func(schema *Schema)

	cases := []struct {
		swaggerRelPath string
		schemaName     string
		process        process
		expect         string
	}{
		{
			swaggerRelPath: "foo.json",
			schemaName:     "def_regular",
			expect: fmt.Sprintf(`{
    "SwaggerRelPath": "foo.json",
    "Name": "def_regular",
    "Properties": {
        "prop_array_of_primitive": {
		  "state": "Uncovered"
		},
        "prop_array_of_object": {
		  "state": "Uncovered"
		},
        "prop_object": {
		  "state": "Uncovered"
		},
        "prop_primitive": {
		  "state": "Uncovered"
		}
    }
}`),
		},
		{
			swaggerRelPath: "foo.json",
			schemaName:     "def_prop_in_file_ref",
			expect: fmt.Sprintf(`{
    "SwaggerRelPath": "foo.json",
    "Name": "def_prop_in_file_ref",
    "Properties": {
        "prop": {
		  "state": "Uncovered"
		}
    }
}`),
		},
		{
			swaggerRelPath: "foo.json",
			schemaName:     "def_all_of",
			expect: fmt.Sprintf(`{
    "SwaggerRelPath": "foo.json",
    "Name": "def_all_of",
    "Properties": {
		"prop_nested": {
		  "state": "Uncovered"
		},
		"foo": {
		  "state": "Uncovered"
		},
		"bar": {
		  "state": "Uncovered"
		},
		"baz": {
		  "state": "Uncovered"
		},
		"p1": {
		  "state": "Uncovered"
		}
    }
}`),
		},
		{
			swaggerRelPath: "foo.json",
			schemaName:     "def_cov",
			process: func(schema *Schema) {
				swgPropAddr := addr.MustParseProperty("p2.p2_1")
				require.NoError(t, schema.AddCovProperty(swgPropAddr))
			},
			expect: fmt.Sprintf(`{
    "SwaggerRelPath": "foo.json",
    "Name": "def_cov",
    "Properties": {
		"p1": {
		  "state": "Uncovered"
		},
		"p2.p2_1": {
		  "state": "Covered"
		},
		"p2.p2_2": {
		  "state": "Uncovered"
		}
    }
}`),
		},
	}

	for idx, c := range cases {
		schema, err := NewEmptySchema(specBasePath, c.swaggerRelPath, c.schemaName)
		require.NoError(t, err, idx)
		if c.process != nil {
			c.process(schema)
		}
		b, err := json.Marshal(schema)
		require.NoError(t, err, idx)
		require.JSONEq(t, c.expect, string(b), idx)
	}
}

func TestSchema_Unmarshal(t *testing.T) {
	cases := []struct {
		input  string
		expect Schema
	}{
		{
			input: `{
   "SwaggerRelPath": "foo.json",
   "Name": "def_cov",
   "Properties": {
		"p1": {
		  "state": "Uncovered"
		},
		"p2.p2_1": {
		  "state": "Covered"
		},
		"p2.p2_2": {
		  "state": "Uncovered"
		}
    }
}`,
			expect: Schema{
				SchemaAddr: SchemaAddr{
					SwaggerRelPath: "foo.json",
					Name:           "def_cov",
				},
				Properties: Properties{
					"p1": &Property{
						PropertyState: propStateUncovered,
					},
					"p2.p2_1": &Property{
						PropertyState: propStateCovered,
					},
					"p2.p2_2": &Property{
						PropertyState: propStateUncovered,
					},
				},
			},
		},
	}

	for idx, c := range cases {
		var actual Schema
		require.NoError(t, json.Unmarshal([]byte(c.input), &actual), idx)
		require.Equal(t, c.expect, actual, idx)
	}
}

func TestSchema_CalcCoverage(t *testing.T) {
	cases := []struct {
		schema    Schema
		expectCov *PropertyCoverageStore
	}{
		{
			schema: Schema{
				Properties: Properties{
					"prop1.covered": &Property{
						PropertyState: NewPropertyState(PropStateCovered, ""),
					},
					"prop1.not_covered": {
						PropertyState: NewPropertyState(PropStateUncovered, ""),
					},
					"prop2.covered": {
						PropertyState: NewPropertyState(PropStateCovered, ""),
					},
					"prop2.granted": {
						PropertyState: NewPropertyState(PropStateGranted, "granted"),
					},
					"prop_granted": {
						PropertyState: NewPropertyState(PropStateGranted, "granted"),
					},
				},
			},
			expectCov: &PropertyCoverageStore{
				node: propertyCoverageNode{
					TotalAmount:   3,
					CoveredAmount: 2,
					Children: map[addr.PropertySegment]*propertyCoverageNode{
						addr.MustParsePropertySegment("prop1"): {
							TotalAmount:   2,
							CoveredAmount: 1,
							Children: map[addr.PropertySegment]*propertyCoverageNode{
								addr.MustParsePropertySegment("not_covered"): {
									TotalAmount:   1,
									CoveredAmount: 0,
									Children:      map[addr.PropertySegment]*propertyCoverageNode{},
								},
								addr.MustParsePropertySegment("covered"): {
									TotalAmount:   1,
									CoveredAmount: 1,
									Children:      map[addr.PropertySegment]*propertyCoverageNode{},
								},
							},
						},
						addr.MustParsePropertySegment("prop2"): {
							TotalAmount:   1,
							CoveredAmount: 1,
							Children: map[addr.PropertySegment]*propertyCoverageNode{
								addr.MustParsePropertySegment("covered"): {
									TotalAmount:   1,
									CoveredAmount: 1,
									Children:      map[addr.PropertySegment]*propertyCoverageNode{},
								},
							},
						},
					},
				},
			},
		},
	}

	for idx, c := range cases {
		cov, err := c.schema.CalcCoverage()
		require.NoError(t, err, idx)
		require.Equal(t, c.expectCov, cov, idx)
	}
}

func TestSchema_Merge(t *testing.T) {
	cases := []struct {
		schemas []Schema
		expect  Schema
	}{
		// 0
		{
			schemas: []Schema{
				{
					SchemaAddr: SchemaAddr{
						SwaggerRelPath: "foo.json",
						Name:           "def_base",
					},
					Properties: map[string]*Property{
						"{def_variant1}.type": {
							PropertyState: propStateCovered,
						},
						"{def_variant2}.type": {
							PropertyState: propStateUncovered,
						},
					},
				},
				{
					SchemaAddr: SchemaAddr{
						SwaggerRelPath: "foo.json",
						Name:           "def_base",
					},
					Properties: map[string]*Property{
						"{def_variant1}.type": {
							PropertyState: propStateUncovered,
						},
						"{def_variant2}.type": {
							PropertyState: propStateCovered,
						},
					},
				},
			},
			expect: Schema{
				SchemaAddr: SchemaAddr{
					SwaggerRelPath: "foo.json",
					Name:           "def_base",
				},
				Properties: map[string]*Property{
					"{def_variant1}.type": {
						PropertyState: propStateCovered,
					},
					"{def_variant2}.type": {
						PropertyState: propStateCovered,
					},
				},
			},
		},
	}

	for idx, c := range cases {
		var mergedSchema Schema
		for i, schema := range c.schemas {
			if i == 0 {
				mergedSchema = schema
				continue
			}
			require.NoError(t, mergedSchema.Merge(schema), i, idx)
		}
		require.Equal(t, c.expect, mergedSchema, idx)
	}
}

func TestSchema_Grant(t *testing.T) {
	var cases = []struct {
		schema       Schema
		grants       SchemaGrants
		expectSchema Schema
	}{
		{
			schema: Schema{
				SchemaAddr: SchemaAddr{"path", "schema1"},
				Properties: map[string]*Property{
					"prop1": {},
				},
			},
			grants: map[SchemaAddr]SchemaGrant{
				{SwaggerRelPath: "path", Name: "schema1"}: {
					Comment: "granted",
				},
			},
			expectSchema: Schema{
				SchemaAddr: SchemaAddr{"path", "schema1"},
				Properties: map[string]*Property{
					"prop1": {},
				},
				IsGranted:    true,
				GrantComment: "granted",
			},
		},
		{
			schema: Schema{
				SchemaAddr: SchemaAddr{"path", "schema1"},
				Properties: map[string]*Property{
					"prop1":   {},
					"prop2":   {},
					"prop2_2": {},
					"prop3":   {},
				},
			},
			grants: map[SchemaAddr]SchemaGrant{
				{SwaggerRelPath: "path", Name: "schema1"}: {
					Properties: map[string]string{
						"prop2_2": "granted",
						"prop3":   "granted",
					},
				},
			},
			expectSchema: Schema{
				SchemaAddr: SchemaAddr{"path", "schema1"},
				Properties: map[string]*Property{
					"prop1": {},
					"prop2": {},
					"prop2_2": {
						PropertyState: NewPropertyState(PropStateGranted, "granted"),
					},
					"prop3": {
						PropertyState: NewPropertyState(PropStateGranted, "granted"),
					},
				},
			},
		},
	}

	for idx, c := range cases {
		require.NoError(t, c.schema.Grant(c.grants), idx)
		require.Equal(t, c.expectSchema, c.schema, idx)
	}
}
