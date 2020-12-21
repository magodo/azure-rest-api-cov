package swagger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-openapi/jsonreference"
	openapispec "github.com/go-openapi/spec"

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
	specNestedDiscriminatorPathLocal := filepath.Join(specBasePathLocal, "nested_discriminator.json")

	specFoo, err := swaggerloader.LoadSwagger(specFooPathLocal)
	require.NoError(t, err)
	specBar, err := swaggerloader.LoadSwagger(specBarPathLocal)
	require.NoError(t, err)
	specBaz, err := swaggerloader.LoadSwagger(specBazPathLocal)
	require.NoError(t, err)
	specRemote, err := swaggerloader.LoadSwagger(specRemotePathLocal)
	require.NoError(t, err)
	specNestedDiscriminator, err := swaggerloader.LoadSwagger(specNestedDiscriminatorPathLocal)
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
		repeat      int
		err         error
		expect      Schema
	}{
		// 0
		{
			specBaseURL: specBasePathLocal,
			specRelPath: "foo.json",
			schemaName:  "def_foo",
			err:         nil,
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
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
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPathLocal, "def_foo")),
						},
						swaggerURI: specFooPathLocal,
					},
				},
				swaggerURI: specFooPathLocal,
				swagger:    specFoo,
			},
		},
		// 1
		{
			specBaseURL: specBasePathLocal,
			specRelPath: "foo.json",
			schemaName:  "def_regular",
			err:         nil,
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
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
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPathLocal, "def_regular")),
						},
						swaggerURI: specFooPathLocal,
					},
					"prop_array_of_primitive": {
						PropertyState: propStateUncovered,
						schema:        specFoo.Definitions["def_regular"].Properties["prop_array_of_primitive"],
						resolvedRefs: map[string]interface{}{
							specFooPathLocal + "#/definitions/def_regular": struct{}{},
						},
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPathLocal, "def_regular")),
						},
						swaggerURI: specFooPathLocal,
					},
					"prop_array_of_object": {
						PropertyState: propStateUncovered,
						schema:        specFoo.Definitions["def_regular"].Properties["prop_array_of_object"],
						resolvedRefs: map[string]interface{}{
							specFooPathLocal + "#/definitions/def_regular": struct{}{},
						},
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPathLocal, "def_regular")),
						},
						swaggerURI: specFooPathLocal,
					},
					"prop_object": {
						PropertyState: propStateUncovered,
						schema:        specFoo.Definitions["def_regular"].Properties["prop_object"],
						resolvedRefs: map[string]interface{}{
							specFooPathLocal + "#/definitions/def_regular": struct{}{},
						},
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPathLocal, "def_regular")),
						},
						swaggerURI: specFooPathLocal,
					},
				},
				swaggerURI: specFooPathLocal,
				swagger:    specFoo,
			},
		},
		// 2
		{
			specBaseURL: specBasePathLocal,
			specRelPath: "foo.json",
			schemaName:  "def_prop_in_file_ref",
			err:         nil,
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
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
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPathLocal, "def_prop_in_file_ref")),
						},
						swaggerURI: specFooPathLocal,
					},
				},
				swaggerURI: specFooPathLocal,
				swagger:    specFoo,
			},
		},
		// 3
		{
			specBaseURL: specBasePathLocal,
			specRelPath: "foo.json",
			schemaName:  "def_prop_cross_file_ref",
			err:         nil,
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
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
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPathLocal, "def_prop_cross_file_ref")),
						},
						swaggerURI: specFooPathLocal,
					},
				},
				swaggerURI: specFooPathLocal,
				swagger:    specFoo,
			},
		},
		// 4
		{
			specBaseURL: specBasePathLocal,
			specRelPath: "foo.json",
			schemaName:  "def_prop_self_ref",
			err:         nil,
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
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
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPathLocal, "def_prop_self_ref")),
						},
						swaggerURI: specFooPathLocal,
					},
				},
				swaggerURI: specFooPathLocal,
				swagger:    specFoo,
			},
		},
		// 5
		{
			specBaseURL: specBasePathLocal,
			specRelPath: "foo.json",
			schemaName:  "def_array_ref",
			err:         nil,
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
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
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPathLocal, "def_foo")),
						},
					},
				},
				swaggerURI: specFooPathLocal,
				swagger:    specFoo,
			},
		},
		// 6
		{
			specBaseURL: specBasePathLocal,
			specRelPath: "foo.json",
			schemaName:  "def_in_file_ref",
			err:         nil,
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
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
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPathLocal, "def_foo")),
						},
						swaggerURI: specFooPathLocal,
					},
				},
				swaggerURI: specFooPathLocal,
				swagger:    specFoo,
			},
		},
		// 7
		{
			specBaseURL: specBasePathLocal,
			specRelPath: "foo.json",
			schemaName:  "def_cross_file_ref",
			err:         nil,
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
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
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specBarPathLocal, "def_bar")),
						},
						swaggerURI: specBarPathLocal,
					},
				},
				swaggerURI: specFooPathLocal,
				swagger:    specFoo,
			},
		},
		// 8
		{
			specBaseURL: specBasePathLocal,
			specRelPath: "foo.json",
			schemaName:  "def_self_ref",
			err:         nil,
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
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
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPathLocal, "def_self_ref")),
						},
						swaggerURI: specFooPathLocal,
					},
				},
				swaggerURI: specFooPathLocal,
				swagger:    specFoo,
			},
		},
		// 9
		{
			specBaseURL: specBasePathLocal,
			specRelPath: "foo.json",
			schemaName:  "def_all_of",
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
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
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPathLocal, "def_all_of")),
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
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPathLocal, "def_foo")),
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
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specBarPathLocal, "def_bar")),
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
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specBazPathLocal, "def_baz")),
						},
						swaggerURI: specBazPathLocal,
					},
					"p1": {
						PropertyState: propStateUncovered,
						schema:        specFoo.Definitions["def_all_of"].Properties["p1"],
						resolvedRefs: map[string]interface{}{
							specFooPathLocal + "#/definitions/def_all_of": struct{}{},
						},
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPathLocal, "def_all_of")),
						},
						swaggerURI: specFooPathLocal,
					},
				},
				swaggerURI: specFooPathLocal,
				swagger:    specFoo,
			},
		},
		// 10
		{
			specBaseURL: specBasePathLocal,
			specRelPath: "foo.json",
			schemaName:  "def_base",
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
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
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPathLocal, "def_variant1")),
						},
						swaggerURI:      specFooPathLocal,
						isDiscriminator: true,
					},
					"{def_variant2}": {
						PropertyState: propStateUncovered,
						schema:        specFoo.Definitions["variant2_def"],
						resolvedRefs: map[string]interface{}{
							specFooPathLocal + "#/definitions/variant2_def": struct{}{},
						},
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPathLocal, "variant2_def")),
						},
						swaggerURI:      specFooPathLocal,
						isDiscriminator: true,
					},
				},
				swaggerURI: specFooPathLocal,
				swagger:    specFoo,
			},
		},
		// 11
		{
			specBaseURL: specBasePathLocal,
			specRelPath: "foo.json",
			schemaName:  "def_to_def_base",
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
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
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPathLocal, "def_to_def_base")),
						},
						swaggerURI: specFooPathLocal,
					},
				},
				swaggerURI: specFooPathLocal,
				swagger:    specFoo,
			},
		},
		// 12
		{
			specBaseURL: specBasePathLocal,
			specRelPath: "foo.json",
			schemaName:  "non_existent_schema",
			err:         fmt.Errorf("schema %q is not found in swagger spec %q", "non_existent_schema", specBasePathLocal+"/foo.json"),
		},
		// 13
		// The "def_base_template" uses the same discriminator values as "def_base",
		// the difference is that the variant schemas are inheriting from different base schemas (via allOf).
		{
			specBaseURL: specBasePathLocal,
			specRelPath: "foo.json",
			schemaName:  "def_base_template",
			repeat:      10, // repeat multiple times to ensure the result is consistent
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "foo.json",
					Name:           "def_base_template",
				},
				Properties: map[string]*Property{
					"{def_variant1}": {
						PropertyState: propStateUncovered,
						schema:        specFoo.Definitions["def_variant1_template"],
						resolvedRefs: map[string]interface{}{
							specFooPathLocal + "#/definitions/def_variant1_template": struct{}{},
						},
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPathLocal, "def_variant1_template")),
						},
						swaggerURI:      specFooPathLocal,
						isDiscriminator: true,
					},
					"{def_variant2}": {
						PropertyState: propStateUncovered,
						schema:        specFoo.Definitions["def_variant2_template"],
						resolvedRefs: map[string]interface{}{
							specFooPathLocal + "#/definitions/def_variant2_template": struct{}{},
						},
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPathLocal, "def_variant2_template")),
						},
						swaggerURI:      specFooPathLocal,
						isDiscriminator: true,
					},
				},
				swaggerURI: specFooPathLocal,
				swagger:    specFoo,
			},
		},
		// 14
		{
			specBaseURL: specBasePathLocal,
			specRelPath: "foo.json",
			schemaName:  "def_parent",
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "foo.json",
					Name:           "def_parent",
				},
				Properties: map[string]*Property{
					"{def_child1}": {
						PropertyState: propStateUncovered,
						schema:        specFoo.Definitions["def_child1"],
						resolvedRefs: map[string]interface{}{
							specFooPathLocal + "#/definitions/def_child1": struct{}{},
						},
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPathLocal, "def_child1")),
						},
						swaggerURI:      specFooPathLocal,
						isDiscriminator: true,
					},
					"{def_child2}": {
						PropertyState: propStateUncovered,
						schema:        specFoo.Definitions["def_child2"],
						resolvedRefs: map[string]interface{}{
							specFooPathLocal + "#/definitions/def_child2": struct{}{},
						},
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPathLocal, "def_child2")),
						},
						swaggerURI:      specFooPathLocal,
						isDiscriminator: true,
					},
				},
				swaggerURI: specFooPathLocal,
				swagger:    specFoo,
			},
		},
		// 15
		{
			specBaseURL: specBasePathLocal,
			specRelPath: "nested_discriminator.json",
			schemaName:  "def_a",
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "nested_discriminator.json",
					Name:           "def_a",
				},
				Properties: map[string]*Property{
					"{def_b}": {
						PropertyState: propStateUncovered,
						schema:        specNestedDiscriminator.Definitions["def_b"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPathLocal + "#/definitions/def_b": struct{}{},
						},
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPathLocal, "def_b")),
						},
						swaggerURI:      specNestedDiscriminatorPathLocal,
						isDiscriminator: true,
						childRefSet: map[string]struct{}{
							fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPathLocal, "def_c"): {},
						},
					},
					"{def_c}": {
						PropertyState: propStateUncovered,
						schema:        specNestedDiscriminator.Definitions["def_c"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPathLocal + "#/definitions/def_c": struct{}{},
						},
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPathLocal, "def_c")),
						},
						swaggerURI:      specNestedDiscriminatorPathLocal,
						isDiscriminator: true,
					},
				},
				swaggerURI: specNestedDiscriminatorPathLocal,
				swagger:    specNestedDiscriminator,
			},
		},
		// 16
		{
			specBaseURL: specBasePathLocal,
			specRelPath: "nested_discriminator.json",
			schemaName:  "def_b",
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "nested_discriminator.json",
					Name:           "def_b",
				},
				Properties: map[string]*Property{
					"type": {
						PropertyState: propStateUncovered,
						schema:        specNestedDiscriminator.Definitions["def_a"].Properties["type"],
						resolvedRefs:  map[string]interface{}{},
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPathLocal, "def_a")),
						},
						swaggerURI: specNestedDiscriminatorPathLocal,
					},
					"{def_c}": {
						PropertyState: propStateUncovered,
						schema:        specNestedDiscriminator.Definitions["def_c"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPathLocal + "#/definitions/def_c": struct{}{},
						},
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPathLocal, "def_c")),
						},
						swaggerURI:      specNestedDiscriminatorPathLocal,
						isDiscriminator: true,
					},
				},
				swaggerURI: specNestedDiscriminatorPathLocal,
				swagger:    specNestedDiscriminator,
			},
		},
		// 17
		{
			specBaseURL: specBasePathLocal,
			specRelPath: "nested_discriminator.json",
			schemaName:  "def_wrap_a",
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "nested_discriminator.json",
					Name:           "def_wrap_a",
				},
				Properties: map[string]*Property{
					"a": {
						PropertyState: propStateUncovered,
						schema:        specNestedDiscriminator.Definitions["def_wrap_a"].Properties["a"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPathLocal + "#/definitions/def_wrap_a": struct{}{},
						},
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPathLocal, "def_wrap_a")),
						},
						swaggerURI: specNestedDiscriminatorPathLocal,
					},
				},
				swaggerURI: specNestedDiscriminatorPathLocal,
				swagger:    specNestedDiscriminator,
			},
		},
		// 18
		{
			specBaseURL: specBasePathLocal,
			specRelPath: "nested_discriminator.json",
			schemaName:  "def_wrap_b",
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "nested_discriminator.json",
					Name:           "def_wrap_b",
				},
				Properties: map[string]*Property{
					"b": {
						PropertyState: propStateUncovered,
						schema:        specNestedDiscriminator.Definitions["def_wrap_b"].Properties["b"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPathLocal + "#/definitions/def_wrap_b": struct{}{},
						},
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPathLocal, "def_wrap_b")),
						},
						swaggerURI: specNestedDiscriminatorPathLocal,
					},
				},
				swaggerURI: specNestedDiscriminatorPathLocal,
				swagger:    specNestedDiscriminator,
			},
		},
		//
		//{
		//	specBaseURL: specRemoteBaseURL,
		//	specRelPath: "remote.json",
		//	schemaName:  "def_foo",
		//	expect: Schema{
		//		SchemaAddr: SchemaAddr{
		//			SwaggerRelPath: "remote.json",
		//			Name:           "def_foo",
		//		},
		//		Properties: map[string]*Property{
		//			"prop_primitive": {
		//				PropertyState: propStateUncovered,
		//				schema:        specRemote.Definitions["def_foo"].Properties["prop_primitive"],
		//				resolvedRefs: map[string]interface{}{
		//					specRemoteFileURL + "#/definitions/def_foo": struct{}{},
		//				},
		//				fromRef: &openapispec.Ref{
		//					Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPathLocal, "def_all_of")),
		//				},
		//				swaggerURI: specRemoteFileURL,
		//			},
		//		},
		//		swaggerURI: specRemoteFileURL,
		//		swagger:    specRemote,
		//	},
		//},
	}

	for idx, c := range cases {
		msg := fmt.Sprintf("(%d) %s: %s", idx, filepath.Join(c.specBaseURL, c.specRelPath), c.schemaName)
		for i := 0; i <= c.repeat; i++ {
			actual, err := NewEmptySchema(c.specBaseURL, c.specRelPath, c.schemaName)
			require.Equal(t, c.err, err, msg)
			if err == nil {
				require.Equal(t, c.expect, *actual, msg)
			}
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
	specNestedDiscriminatorPathLocal := filepath.Join(specBasePath, "nested_discriminator.json")
	specFoo, err := swaggerloader.LoadSwagger(specFooPath)
	specBar, err := swaggerloader.LoadSwagger(specBarPath)
	specNestedDiscriminator, err := swaggerloader.LoadSwagger(specNestedDiscriminatorPathLocal)
	_ = specBar
	_ = specFoo
	_ = specNestedDiscriminator
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		name           string
		swaggerRelPath string
		schemaName     string
		expandAddrs    []addr.Property
		repeat         int
		err            error
		expect         Schema
	}{
		// 0
		{
			name:           `Expand "def_foo"`,
			swaggerRelPath: "foo.json",
			schemaName:     "def_foo",
			expandAddrs: []addr.Property{
				addr.MustParseProperty("foo"),
			},
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
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
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_foo")),
						},
						swaggerURI: specFooPath,
					},
				},
				swaggerURI: specFooPath,
				swagger:    specFoo,
			},
		},
		// 1
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
				SchemaMetaInfo: SchemaMetaInfo{
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
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_foo")),
						},
						swaggerURI: specFooPath,
					},
				},
				swaggerURI: specFooPath,
				swagger:    specFoo,
			},
		},
		// 2
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
				SchemaMetaInfo: SchemaMetaInfo{
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
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_regular")),
						},
						swaggerURI: specFooPath,
					},
					"prop_array_of_primitive": {
						PropertyState: propStateUncovered,
						schema:        specFoo.Definitions["def_regular"].Properties["prop_array_of_primitive"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_regular": struct{}{},
						},
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_regular")),
						},
						swaggerURI: specFooPath,
					},
					"prop_array_of_object/prop_nested": {
						PropertyState: propStateUncovered,
						schema:        specFoo.Definitions["def_regular"].Properties["prop_array_of_object"].Items.Schema.Properties["prop_nested"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_regular": struct{}{},
						},
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_regular")),
						},
						swaggerURI: specFooPath,
					},
					"prop_object/prop_nested": {
						PropertyState: propStateUncovered,
						schema:        specFoo.Definitions["def_regular"].Properties["prop_object"].Properties["prop_nested"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_regular": struct{}{},
						},
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_regular")),
						},
						swaggerURI: specFooPath,
					},
				},
				swaggerURI: specFooPath,
				swagger:    specFoo,
			},
		},
		// 3
		{
			name:           `Expand "def_prop_in_file_ref"`,
			swaggerRelPath: "foo.json",
			schemaName:     "def_prop_in_file_ref",
			expandAddrs: []addr.Property{
				addr.MustParseProperty("prop"),
			},
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "foo.json",
					Name:           "def_prop_in_file_ref",
				},
				Properties: map[string]*Property{
					"prop/foo": {
						PropertyState: propStateUncovered,
						schema:        specFoo.Definitions["def_foo"].Properties["foo"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_prop_in_file_ref": struct{}{},
							specFooPath + "#/definitions/def_foo":              struct{}{},
						},
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_foo")),
						},
						swaggerURI: specFooPath,
					},
				},
				swaggerURI: specFooPath,
				swagger:    specFoo,
			},
		},
		// 4
		{
			name:           `Expand "def_prop_self_ref"`,
			swaggerRelPath: "foo.json",
			schemaName:     "def_prop_self_ref",
			expandAddrs: []addr.Property{
				addr.MustParseProperty("prop"),
			},
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
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
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_prop_self_ref")),
						},
						swaggerURI: specFooPath,
					},
				},
				swaggerURI: specFooPath,
				swagger:    specFoo,
			},
		},
		// 5
		{
			name:           `Expand "def_prop_cross_file_ref"`,
			swaggerRelPath: "foo.json",
			schemaName:     "def_prop_cross_file_ref",
			expandAddrs: []addr.Property{
				addr.MustParseProperty("prop"),
			},
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "foo.json",
					Name:           "def_prop_cross_file_ref",
				},
				Properties: map[string]*Property{
					"prop/bar": {
						PropertyState: propStateUncovered,
						schema:        specBar.Definitions["def_bar"].Properties["bar"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_prop_cross_file_ref": struct{}{},
							specBarPath + "#/definitions/def_bar":                 struct{}{},
						},
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specBarPath, "def_bar")),
						},
						swaggerURI: specBarPath,
					},
				},
				swaggerURI: specFooPath,
				swagger:    specFoo,
			},
		},
		// 6
		{
			name:           `Expand "def_in_file_ref"`,
			swaggerRelPath: "foo.json",
			schemaName:     "def_in_file_ref",
			expandAddrs: []addr.Property{
				addr.MustParseProperty("foo"),
			},
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
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
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_foo")),
						},
						swaggerURI: specFooPath,
					},
				},
				swaggerURI: specFooPath,
				swagger:    specFoo,
			},
		},
		// 7
		{
			name:           `Expand "def_cross_file_ref"`,
			swaggerRelPath: "foo.json",
			schemaName:     "def_cross_file_ref",
			expandAddrs: []addr.Property{
				addr.MustParseProperty("bar"),
			},
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
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
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specBarPath, "def_bar")),
						},
						swaggerURI: specBarPath,
					},
				},
				swaggerURI: specFooPath,
				swagger:    specFoo,
			},
		},
		// 8
		{
			name:           `Expand "def_self_ref"`,
			swaggerRelPath: "foo.json",
			schemaName:     "def_self_ref",
			expandAddrs: []addr.Property{
				addr.MustParseProperty(""),
			},
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
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
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_self_ref")),
						},
						swaggerURI: specFooPath,
					},
				},
				swaggerURI: specFooPath,
				swagger:    specFoo,
			},
		},
		// 9
		{
			name:           `Expand "def_base" (schema itself is a discriminator)`,
			swaggerRelPath: "foo.json",
			schemaName:     "def_base",
			expandAddrs: []addr.Property{
				addr.MustParseProperty("{def_variant1}"),
			},
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "foo.json",
					Name:           "def_base",
				},
				Properties: map[string]*Property{
					"{def_variant1}/type": {
						PropertyState: propStateUncovered,
						schema:        specFoo.Definitions["def_base"].Properties["type"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_variant1": struct{}{},
						},
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_base")),
						},
						swaggerURI: specFooPath,
					},
					"{def_variant2}": {
						PropertyState: propStateUncovered,
						schema:        specFoo.Definitions["variant2_def"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/variant2_def": struct{}{},
						},
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "variant2_def")),
						},
						swaggerURI:      specFooPath,
						isDiscriminator: true,
					},
				},
				swaggerURI: specFooPath,
				swagger:    specFoo,
			},
		},
		// 10
		{
			name:           `Expand "def_to_def_base" (property is a discriminator)`,
			swaggerRelPath: "foo.json",
			schemaName:     "def_to_def_base",
			expandAddrs: []addr.Property{
				addr.MustParseProperty("prop"),
			},
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
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
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_variant1")),
						},
						swaggerURI:      specFooPath,
						isDiscriminator: true,
					},
					"prop{def_variant2}": {
						PropertyState: propStateUncovered,
						schema:        specFoo.Definitions["variant2_def"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_to_def_base": struct{}{},
							specFooPath + "#/definitions/variant2_def":    struct{}{},
						},
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "variant2_def")),
						},
						swaggerURI:      specFooPath,
						isDiscriminator: true,
					},
				},
				swaggerURI: specFooPath,
				swagger:    specFoo,
			},
		},
		// 11
		{
			name:           `Expand "def_to_def_base" variant property (property is a discriminator)`,
			swaggerRelPath: "foo.json",
			schemaName:     "def_to_def_base",
			expandAddrs: []addr.Property{
				addr.MustParseProperty("prop"),
				addr.MustParseProperty("prop{def_variant1}"),
			},
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "foo.json",
					Name:           "def_to_def_base",
				},
				Properties: map[string]*Property{
					"prop{def_variant1}/type": {
						PropertyState: propStateUncovered,
						schema:        specFoo.Definitions["def_base"].Properties["type"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_to_def_base": struct{}{},
							specFooPath + "#/definitions/def_variant1":    struct{}{},
						},
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_base")),
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
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "variant2_def")),
						},
						swaggerURI:      specFooPath,
						isDiscriminator: true,
					},
				},
				swaggerURI: specFooPath,
				swagger:    specFoo,
			},
		},
		// 12
		{
			name:           `Expand "def_to_def_base_template" (property is a discriminator)`,
			swaggerRelPath: "foo.json",
			schemaName:     "def_to_def_base_template",
			expandAddrs: []addr.Property{
				addr.MustParseProperty("prop"),
			},
			repeat: 10, // ensure the result is consistent
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "foo.json",
					Name:           "def_to_def_base_template",
				},
				Properties: map[string]*Property{
					"prop{def_variant1}": {
						PropertyState: propStateUncovered,
						schema:        specFoo.Definitions["def_variant1_template"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_to_def_base_template": struct{}{},
							specFooPath + "#/definitions/def_variant1_template":    struct{}{},
						},
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_variant1_template")),
						},
						swaggerURI:      specFooPath,
						isDiscriminator: true,
					},
					"prop{def_variant2}": {
						PropertyState: propStateUncovered,
						schema:        specFoo.Definitions["def_variant2_template"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_to_def_base_template": struct{}{},
							specFooPath + "#/definitions/def_variant2_template":    struct{}{},
						},
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_variant2_template")),
						},
						swaggerURI:      specFooPath,
						isDiscriminator: true,
					},
				},
				swaggerURI: specFooPath,
				swagger:    specFoo,
			},
		},
		// 13
		{
			name:           `Expand "def_parent" (schema itself is a discriminator)`,
			swaggerRelPath: "foo.json",
			schemaName:     "def_parent",
			expandAddrs: []addr.Property{
				addr.MustParseProperty("{def_child1}"),
			},
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "foo.json",
					Name:           "def_parent",
				},
				Properties: map[string]*Property{
					"{def_child1}/type": {
						PropertyState: propStateUncovered,
						schema:        specFoo.Definitions["def_parent"].Properties["type"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_child1": struct{}{},
						},
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_parent")),
						},
						swaggerURI: specFooPath,
					},
					"{def_child2}": {
						PropertyState: propStateUncovered,
						schema:        specFoo.Definitions["def_child2"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_child2": struct{}{},
						},
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_child2")),
						},
						swaggerURI:      specFooPath,
						isDiscriminator: true,
					},
				},
				swaggerURI: specFooPath,
				swagger:    specFoo,
			},
		},
		// 14
		{
			name:           `Expand "def_parent" (schema itself is a discriminator)`,
			swaggerRelPath: "foo.json",
			schemaName:     "def_parent",
			expandAddrs: []addr.Property{
				addr.MustParseProperty("{def_child1}"),
			},
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "foo.json",
					Name:           "def_parent",
				},
				Properties: map[string]*Property{
					"{def_child1}/type": {
						PropertyState: propStateUncovered,
						schema:        specFoo.Definitions["def_parent"].Properties["type"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_child1": struct{}{},
						},
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_parent")),
						},
						swaggerURI: specFooPath,
					},
					"{def_child2}": {
						PropertyState: propStateUncovered,
						schema:        specFoo.Definitions["def_child2"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_child2": struct{}{},
						},
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_child2")),
						},
						swaggerURI:      specFooPath,
						isDiscriminator: true,
					},
				},
				swaggerURI: specFooPath,
				swagger:    specFoo,
			},
		},
		// 15
		{
			name:           `Expand "def_b" (schema is a in-direct discriminator)`,
			swaggerRelPath: "nested_discriminator.json",
			schemaName:     "def_b",
			expandAddrs: []addr.Property{
				addr.MustParseProperty("{def_c}"),
			},
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "nested_discriminator.json",
					Name:           "def_b",
				},
				Properties: map[string]*Property{
					"type": {
						PropertyState: propStateUncovered,
						schema:        specNestedDiscriminator.Definitions["def_a"].Properties["type"],
						resolvedRefs:  map[string]interface{}{},
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPathLocal, "def_a")),
						},
						swaggerURI: specNestedDiscriminatorPathLocal,
					},
					"{def_c}/type": {
						PropertyState: propStateUncovered,
						schema:        specNestedDiscriminator.Definitions["def_a"].Properties["type"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPathLocal + "#/definitions/def_c": struct{}{},
						},
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPathLocal, "def_a")),
						},
						swaggerURI: specNestedDiscriminatorPathLocal,
					},
				},
				swaggerURI: specNestedDiscriminatorPathLocal,
				swagger:    specNestedDiscriminator,
			},
		},
		// 16
		{
			name:           `Expand "def_wrap_a" (property is a multi-level discriminator)`,
			swaggerRelPath: "nested_discriminator.json",
			schemaName:     "def_wrap_a",
			expandAddrs: []addr.Property{
				addr.MustParseProperty("a"),
			},
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "nested_discriminator.json",
					Name:           "def_wrap_a",
				},
				Properties: map[string]*Property{
					"a{def_b}": {
						PropertyState: propStateUncovered,
						schema:        specNestedDiscriminator.Definitions["def_b"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPathLocal + "#/definitions/def_b":      struct{}{},
							specNestedDiscriminatorPathLocal + "#/definitions/def_wrap_a": struct{}{},
						},
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPathLocal, "def_b")),
						},
						swaggerURI:      specNestedDiscriminatorPathLocal,
						isDiscriminator: true,
						childRefSet: map[string]struct{}{
							fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPathLocal, "def_c"): {},
						},
					},
					"a{def_c}": {
						PropertyState: propStateUncovered,
						schema:        specNestedDiscriminator.Definitions["def_c"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPathLocal + "#/definitions/def_c":      struct{}{},
							specNestedDiscriminatorPathLocal + "#/definitions/def_wrap_a": struct{}{},
						},
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPathLocal, "def_c")),
						},
						swaggerURI:      specNestedDiscriminatorPathLocal,
						isDiscriminator: true,
					},
				},
				swaggerURI: specNestedDiscriminatorPathLocal,
				swagger:    specNestedDiscriminator,
			},
		},
		// 17
		{
			name:           `Expand "def_wrap_a" to its variant property (property is a multi-level discriminator)`,
			swaggerRelPath: "nested_discriminator.json",
			schemaName:     "def_wrap_a",
			expandAddrs: []addr.Property{
				addr.MustParseProperty("a"),
				addr.MustParseProperty("a{def_b}"),
			},
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "nested_discriminator.json",
					Name:           "def_wrap_a",
				},
				Properties: map[string]*Property{
					"a{def_b}/type": {
						PropertyState: propStateUncovered,
						schema:        specNestedDiscriminator.Definitions["def_a"].Properties["type"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPathLocal + "#/definitions/def_b":      struct{}{},
							specNestedDiscriminatorPathLocal + "#/definitions/def_wrap_a": struct{}{},
						},
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPathLocal, "def_a")),
						},
						swaggerURI: specNestedDiscriminatorPathLocal,
					},
					"a{def_c}": {
						PropertyState: propStateUncovered,
						schema:        specNestedDiscriminator.Definitions["def_c"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPathLocal + "#/definitions/def_c":      struct{}{},
							specNestedDiscriminatorPathLocal + "#/definitions/def_wrap_a": struct{}{},
						},
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPathLocal, "def_c")),
						},
						swaggerURI:      specNestedDiscriminatorPathLocal,
						isDiscriminator: true,
					},
				},
				swaggerURI: specNestedDiscriminatorPathLocal,
				swagger:    specNestedDiscriminator,
			},
		},
		// 18
		{
			name:           `Expand "def_wrap_a" to its variant property (property is a multi-level discriminator)`,
			swaggerRelPath: "nested_discriminator.json",
			schemaName:     "def_wrap_a",
			expandAddrs: []addr.Property{
				addr.MustParseProperty("a"),
				addr.MustParseProperty("a{def_c}"),
			},
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "nested_discriminator.json",
					Name:           "def_wrap_a",
				},
				Properties: map[string]*Property{
					"a{def_b}": {
						PropertyState: propStateUncovered,
						schema:        specNestedDiscriminator.Definitions["def_b"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPathLocal + "#/definitions/def_b":      struct{}{},
							specNestedDiscriminatorPathLocal + "#/definitions/def_wrap_a": struct{}{},
						},
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPathLocal, "def_b")),
						},
						swaggerURI:      specNestedDiscriminatorPathLocal,
						isDiscriminator: true,
						childRefSet: map[string]struct{}{
							fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPathLocal, "def_c"): {},
						},
					},
					"a{def_c}/type": {
						PropertyState: propStateUncovered,
						schema:        specNestedDiscriminator.Definitions["def_a"].Properties["type"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPathLocal + "#/definitions/def_c":      struct{}{},
							specNestedDiscriminatorPathLocal + "#/definitions/def_wrap_a": struct{}{},
						},
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPathLocal, "def_a")),
						},
						swaggerURI: specNestedDiscriminatorPathLocal,
					},
				},
				swaggerURI: specNestedDiscriminatorPathLocal,
				swagger:    specNestedDiscriminator,
			},
		},
		// 19
		{
			name:           `Expand "def_wrap_a" to its variant property (property is a multi-level discriminator)`,
			swaggerRelPath: "nested_discriminator.json",
			schemaName:     "def_wrap_a",
			expandAddrs: []addr.Property{
				addr.MustParseProperty("a"),
				addr.MustParseProperty("a{def_c}"),
				addr.MustParseProperty("a{def_b}"),
			},
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "nested_discriminator.json",
					Name:           "def_wrap_a",
				},
				Properties: map[string]*Property{
					"a{def_b}/type": {
						PropertyState: propStateUncovered,
						schema:        specNestedDiscriminator.Definitions["def_a"].Properties["type"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPathLocal + "#/definitions/def_b":      struct{}{},
							specNestedDiscriminatorPathLocal + "#/definitions/def_wrap_a": struct{}{},
						},
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPathLocal, "def_a")),
						},
						swaggerURI: specNestedDiscriminatorPathLocal,
					},
					"a{def_c}/type": {
						PropertyState: propStateUncovered,
						schema:        specNestedDiscriminator.Definitions["def_a"].Properties["type"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPathLocal + "#/definitions/def_c":      struct{}{},
							specNestedDiscriminatorPathLocal + "#/definitions/def_wrap_a": struct{}{},
						},
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPathLocal, "def_a")),
						},
						swaggerURI: specNestedDiscriminatorPathLocal,
					},
				},
				swaggerURI: specNestedDiscriminatorPathLocal,
				swagger:    specNestedDiscriminator,
			},
		},
		// 20
		{
			name:           `Expand "def_wrap_b" (property is a indirect discriminator)`,
			swaggerRelPath: "nested_discriminator.json",
			schemaName:     "def_wrap_b",
			expandAddrs: []addr.Property{
				addr.MustParseProperty("b"),
			},
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "nested_discriminator.json",
					Name:           "def_wrap_b",
				},
				Properties: map[string]*Property{
					"b/type": {
						PropertyState: propStateUncovered,
						schema:        specNestedDiscriminator.Definitions["def_a"].Properties["type"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPathLocal + "#/definitions/def_wrap_b": struct{}{},
						},
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPathLocal, "def_a")),
						},
						swaggerURI: specNestedDiscriminatorPathLocal,
					},
					"b{def_c}": {
						PropertyState: propStateUncovered,
						schema:        specNestedDiscriminator.Definitions["def_c"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPathLocal + "#/definitions/def_wrap_b": struct{}{},
							specNestedDiscriminatorPathLocal + "#/definitions/def_c":      struct{}{},
						},
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPathLocal, "def_c")),
						},
						swaggerURI:      specNestedDiscriminatorPathLocal,
						isDiscriminator: true,
					},
				},
				swaggerURI: specNestedDiscriminatorPathLocal,
				swagger:    specNestedDiscriminator,
			},
		},
		// 21
		{
			name:           `Expand "def_wrap_b" to its variant property (property is a indirect discriminator)`,
			swaggerRelPath: "nested_discriminator.json",
			schemaName:     "def_wrap_b",
			expandAddrs: []addr.Property{
				addr.MustParseProperty("b"),
				addr.MustParseProperty("b{def_c}"),
			},
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "nested_discriminator.json",
					Name:           "def_wrap_b",
				},
				Properties: map[string]*Property{
					"b/type": {
						PropertyState: propStateUncovered,
						schema:        specNestedDiscriminator.Definitions["def_a"].Properties["type"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPathLocal + "#/definitions/def_wrap_b": struct{}{},
						},
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPathLocal, "def_a")),
						},
						swaggerURI: specNestedDiscriminatorPathLocal,
					},
					"b{def_c}/type": {
						PropertyState: propStateUncovered,
						schema:        specNestedDiscriminator.Definitions["def_a"].Properties["type"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPathLocal + "#/definitions/def_wrap_b": struct{}{},
							specNestedDiscriminatorPathLocal + "#/definitions/def_c":      struct{}{},
						},
						fromRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPathLocal, "def_a")),
						},
						swaggerURI: specNestedDiscriminatorPathLocal,
					},
				},
				swaggerURI: specNestedDiscriminatorPathLocal,
				swagger:    specNestedDiscriminator,
			},
		},
	}

	for idx, c := range cases {
		//if idx != 19 {
		//	continue
		//}
		for i := 0; i <= c.repeat; i++ {
			swgschema, err := NewEmptySchema(specBasePath, c.swaggerRelPath, c.schemaName)
			require.Equal(t, c.err, err, c.name, idx)
			if err == nil {
				for _, addr := range c.expandAddrs {
					require.NoError(t, swgschema.expandPropertyOneLevelDeep(addr), c.name, idx)
				}
				require.Equal(t, c.expect, *swgschema, c.name, idx)
			}
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
			name:           "Add same property twice should take no diff (re-entry)",
			swaggerRelPath: "foo.json",
			schemaName:     "def_cov",
			steps: []step{
				{
					swgPropAddr: addr.MustParseProperty("p1"),
					expect: Schema{
						SchemaMetaInfo: SchemaMetaInfo{
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
								fromRef: &openapispec.Ref{
									Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_cov")),
								},
								swaggerURI: specFooPath,
							},
							"p2": {
								PropertyState: propStateUncovered,
								schema:        specFoo.Definitions["def_cov"].Properties["p2"],
								resolvedRefs: map[string]interface{}{
									specFooPath + "#/definitions/def_cov": struct{}{},
								},
								fromRef: &openapispec.Ref{
									Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_cov")),
								},
								swaggerURI: specFooPath,
							},
						},
						swaggerURI: specFooPath,
						swagger:    specFoo,
					},
				},
				// Add a second tf link to the same swg property should have no change
				{
					swgPropAddr: addr.MustParseProperty("p1"),
					expect: Schema{
						SchemaMetaInfo: SchemaMetaInfo{
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
								fromRef: &openapispec.Ref{
									Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_cov")),
								},
								swaggerURI: specFooPath,
							},
							"p2": {
								PropertyState: propStateUncovered,
								schema:        specFoo.Definitions["def_cov"].Properties["p2"],
								resolvedRefs: map[string]interface{}{
									specFooPath + "#/definitions/def_cov": struct{}{},
								},
								fromRef: &openapispec.Ref{
									Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_cov")),
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
			name:           "Add tf link to swagger child property first, then Add another tf link to swagger parent property",
			swaggerRelPath: "foo.json",
			schemaName:     "def_cov",
			steps: []step{
				{
					swgPropAddr: addr.MustParseProperty("p2/p2_1"),
					expect: Schema{
						SchemaMetaInfo: SchemaMetaInfo{
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
								fromRef: &openapispec.Ref{
									Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_cov")),
								},
								swaggerURI: specFooPath,
							},
							"p2/p2_1": {
								PropertyState: propStateCovered,
								schema:        specFoo.Definitions["def_cov"].Properties["p2"].Properties["p2_1"],
								resolvedRefs: map[string]interface{}{
									specFooPath + "#/definitions/def_cov": struct{}{},
								},
								fromRef: &openapispec.Ref{
									Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_cov")),
								},
								swaggerURI: specFooPath,
							},
							"p2/p2_2": {
								PropertyState: propStateUncovered,
								schema:        specFoo.Definitions["def_cov"].Properties["p2"].Properties["p2_2"],
								resolvedRefs: map[string]interface{}{
									specFooPath + "#/definitions/def_cov": struct{}{},
								},
								fromRef: &openapispec.Ref{
									Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_cov")),
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
						SchemaMetaInfo: SchemaMetaInfo{
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
								fromRef: &openapispec.Ref{
									Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_cov")),
								},
								swaggerURI: specFooPath,
							},
							"p2/p2_1": {
								PropertyState: propStateCovered,
								schema:        specFoo.Definitions["def_cov"].Properties["p2"].Properties["p2_1"],
								resolvedRefs: map[string]interface{}{
									specFooPath + "#/definitions/def_cov": struct{}{},
								},
								fromRef: &openapispec.Ref{
									Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_cov")),
								},
								swaggerURI: specFooPath,
							},
							"p2/p2_2": {
								PropertyState: propStateCovered,
								schema:        specFoo.Definitions["def_cov"].Properties["p2"].Properties["p2_2"],
								resolvedRefs: map[string]interface{}{
									specFooPath + "#/definitions/def_cov": struct{}{},
								},
								fromRef: &openapispec.Ref{
									Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_cov")),
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
			name:           "Add tf link to swagger parent property first, then Add another tf link to swagger child property",
			swaggerRelPath: "foo.json",
			schemaName:     "def_cov",
			steps: []step{
				{
					swgPropAddr: addr.MustParseProperty("p2"),
					expect: Schema{
						SchemaMetaInfo: SchemaMetaInfo{
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
								fromRef: &openapispec.Ref{
									Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_cov")),
								},
								swaggerURI: specFooPath,
							},
							"p2": {
								PropertyState: propStateCovered,
								schema:        specFoo.Definitions["def_cov"].Properties["p2"],
								resolvedRefs: map[string]interface{}{
									specFooPath + "#/definitions/def_cov": struct{}{},
								},
								fromRef: &openapispec.Ref{
									Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_cov")),
								},
								swaggerURI: specFooPath,
							},
						},
						swaggerURI: specFooPath,
						swagger:    specFoo,
					},
				},
				{
					swgPropAddr: addr.MustParseProperty("p2/p2_1"),
					expect: Schema{
						SchemaMetaInfo: SchemaMetaInfo{
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
								fromRef: &openapispec.Ref{
									Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_cov")),
								},
								swaggerURI: specFooPath,
							},
							"p2/p2_1": {
								PropertyState: propStateCovered,
								schema:        specFoo.Definitions["def_cov"].Properties["p2"].Properties["p2_1"],
								resolvedRefs: map[string]interface{}{
									specFooPath + "#/definitions/def_cov": struct{}{},
								},
								fromRef: &openapispec.Ref{
									Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_cov")),
								},
								swaggerURI: specFooPath,
							},
							"p2/p2_2": {
								PropertyState: propStateCovered,
								schema:        specFoo.Definitions["def_cov"].Properties["p2"].Properties["p2_2"],
								resolvedRefs: map[string]interface{}{
									specFooPath + "#/definitions/def_cov": struct{}{},
								},
								fromRef: &openapispec.Ref{
									Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_cov")),
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
			name:           "Add discriminator property",
			swaggerRelPath: "foo.json",
			schemaName:     "def_base",
			steps: []step{
				{
					swgPropAddr: addr.MustParseProperty("{def_variant1}/type"),
					expect: Schema{
						SchemaMetaInfo: SchemaMetaInfo{
							SwaggerRelPath: "foo.json",
							Name:           "def_base",
						},
						Properties: map[string]*Property{
							"{def_variant1}/type": {
								PropertyState: propStateCovered,
								schema:        specFoo.Definitions["def_base"].Properties["type"],
								resolvedRefs: map[string]interface{}{
									specFooPath + "#/definitions/def_variant1": struct{}{},
								},
								fromRef: &openapispec.Ref{
									Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_base")),
								},
								swaggerURI: specFooPath,
							},
							"{def_variant2}": {
								PropertyState: propStateUncovered,
								schema:        specFoo.Definitions["variant2_def"],
								resolvedRefs: map[string]interface{}{
									specFooPath + "#/definitions/variant2_def": struct{}{},
								},
								fromRef: &openapispec.Ref{
									Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "variant2_def")),
								},
								swaggerURI:      specFooPath,
								isDiscriminator: true,
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
    "swagger_rel_path": "foo.json",
    "name": "def_regular",
    "properties": {
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
    "swagger_rel_path": "foo.json",
    "name": "def_prop_in_file_ref",
    "properties": {
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
    "swagger_rel_path": "foo.json",
    "name": "def_all_of",
    "properties": {
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
				swgPropAddr := addr.MustParseProperty("p2/p2_1")
				require.NoError(t, schema.AddCovProperty(swgPropAddr))
			},
			expect: fmt.Sprintf(`{
    "swagger_rel_path": "foo.json",
    "name": "def_cov",
    "properties": {
		"p1": {
		  "state": "Uncovered"
		},
		"p2/p2_1": {
		  "state": "Covered"
		},
		"p2/p2_2": {
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
   "swagger_rel_path": "foo.json",
   "name": "def_cov",
   "version": "2020-12-01",
   "properties": {
		"p1": {
		  "state": "Uncovered"
		},
		"p2/p2_1": {
		  "state": "Covered"
		},
		"p2/p2_2": {
		  "state": "Uncovered"
		}
    }
}`,
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "foo.json",
					Name:           "def_cov",
				},
				Properties: Properties{
					"p1": &Property{
						PropertyState: propStateUncovered,
					},
					"p2/p2_1": &Property{
						PropertyState: propStateCovered,
					},
					"p2/p2_2": &Property{
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

func TestSchema_Merge(t *testing.T) {
	cases := []struct {
		schemas []Schema
		expect  Schema
	}{
		// 0
		{
			schemas: []Schema{
				{
					SchemaMetaInfo: SchemaMetaInfo{
						SwaggerRelPath: "foo.json",
						Name:           "def_base",
					},
					Properties: map[string]*Property{
						"{def_variant1}/type": {
							PropertyState: propStateCovered,
						},
						"{def_variant2}/type": {
							PropertyState: propStateUncovered,
						},
					},
				},
				{
					SchemaMetaInfo: SchemaMetaInfo{
						SwaggerRelPath: "foo.json",
						Name:           "def_base",
					},
					Properties: map[string]*Property{
						"{def_variant1}/type": {
							PropertyState: propStateUncovered,
						},
						"{def_variant2}/type": {
							PropertyState: propStateCovered,
						},
					},
				},
			},
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "foo.json",
					Name:           "def_base",
				},
				Properties: map[string]*Property{
					"{def_variant1}/type": {
						PropertyState: propStateCovered,
					},
					"{def_variant2}/type": {
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
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "path",
					Name:           "schema1",
				},
				Properties: map[string]*Property{
					"prop1": {},
				},
			},
			grants: SchemaGrants{
				SchemaMetaInfo{
					SwaggerRelPath: "path",
					Name:           "schema1",
				}.Ref(): {
					Comment: "granted",
				},
			},
			expectSchema: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "path",
					Name:           "schema1",
				},
				Properties: map[string]*Property{
					"prop1": {},
				},
				IsGranted:    true,
				GrantComment: "granted",
			},
		},
		{
			schema: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "path",
					Name:           "schema1",
				},
				Properties: map[string]*Property{
					"prop1":   {},
					"prop2":   {},
					"prop2_2": {},
					"prop3":   {},
				},
			},
			grants: map[string]SchemaGrant{
				SchemaMetaInfo{
					SwaggerRelPath: "path",
					Name:           "schema1",
				}.Ref(): {
					Properties: map[string]string{
						"prop2_2": "granted",
						"prop3":   "granted",
					},
				},
			},
			expectSchema: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "path",
					Name:           "schema1",
				},
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
