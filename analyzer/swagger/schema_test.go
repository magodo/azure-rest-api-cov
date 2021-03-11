package swagger

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/magodo/azure-rest-api-cov/analyzer/specpath"

	"github.com/go-openapi/jsonreference"
	openapispec "github.com/go-openapi/spec"

	"github.com/magodo/azure-rest-api-cov/analyzer/swaggerloader"

	"github.com/magodo/azure-rest-api-cov/analyzer/addr"
	"github.com/stretchr/testify/require"
)

var (
	// Spec Paths
	specBasePath                string
	specFooPath                 string
	specBarPath                 string
	specBazPath                 string
	specNestedDiscriminatorPath string
	specSiblingVariantPath      string
	specRemoteDiscPath          string

	specRemotePath string
	specRemoteURL  string

	// Swaggers
	swaggerFoo                 *openapispec.Swagger
	swaggerBar                 *openapispec.Swagger
	swaggerBaz                 *openapispec.Swagger
	swaggerRemote              *openapispec.Swagger
	swaggerNestedDiscriminator *openapispec.Swagger
	swaggerSiblingVariant      *openapispec.Swagger
	swaggerRemoteDisc          *openapispec.Swagger
)

func TestMain(m *testing.M) {
	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	specBasePath = filepath.Join(filepath.Dir(pwd), "testdata", "swagger")
	specFooPath = filepath.Join(specBasePath, "foo.json")
	specBarPath = filepath.Join(specBasePath, "bar.json")
	specBazPath = filepath.Join(specBasePath, "baz", "baz.json")
	specNestedDiscriminatorPath = filepath.Join(specBasePath, "nested_discriminator.json")
	specSiblingVariantPath = filepath.Join(specBasePath, "sibling_variant.json")
	specRemoteDiscPath = filepath.Join(specBasePath, "remote", "discriminator.json")

	specRemotePath = filepath.Join(specBasePath, "remote.json")
	specRemoteURL = "https://gist.githubusercontent.com/magodo/f054bb1c2e7a1c74fd78f65eb42a17bb/raw/8308d9295287bc590be9b431ea0353b626da7ada/remote.json"

	swaggerFoo, err = swaggerloader.LoadSwagger(specFooPath)
	if err != nil {
		log.Fatal(err)
	}
	swaggerBar, err = swaggerloader.LoadSwagger(specBarPath)
	if err != nil {
		log.Fatal(err)
	}
	swaggerBaz, err = swaggerloader.LoadSwagger(specBazPath)
	if err != nil {
		log.Fatal(err)
	}
	swaggerRemote, err = swaggerloader.LoadSwagger(specRemotePath)
	if err != nil {
		log.Fatal(err)
	}
	swaggerNestedDiscriminator, err = swaggerloader.LoadSwagger(specNestedDiscriminatorPath)
	if err != nil {
		log.Fatal(err)
	}

	swaggerSiblingVariant, err = swaggerloader.LoadSwagger(specSiblingVariantPath)
	if err != nil {
		log.Fatal(err)
	}
	swaggerRemoteDisc, err = swaggerloader.LoadSwagger(specRemoteDiscPath)
	if err != nil {
		log.Fatal(err)
	}

	_ = swaggerFoo
	_ = swaggerBar
	_ = swaggerBaz
	_ = swaggerRemote
	_ = swaggerNestedDiscriminator
	_ = swaggerSiblingVariant
	_ = swaggerRemoteDisc

	os.Exit(m.Run())
}

func TestSchema(t *testing.T) {
	cases := []struct {
		specBaseURL string
		specRelPath string
		schemaName  string
		forResponse bool
		repeat      int
		err         bool
		expect      Schema
	}{
		// 0
		{
			specBaseURL: specBasePath,
			specRelPath: "foo.json",
			schemaName:  "def_foo",
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "foo.json",
					Name:           "def_foo",
				},
				Properties: map[string]*Property{
					"foo": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["def_foo"].Properties["foo"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_foo": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_foo")),
						},
						swaggerURI: specFooPath,
					},
				},
				swaggerURI: specFooPath,
			},
		},
		// 1
		{
			specBaseURL: specBasePath,
			specRelPath: "foo.json",
			schemaName:  "def_regular",
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "foo.json",
					Name:           "def_regular",
				},
				Properties: map[string]*Property{
					"prop_primitive": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["def_regular"].Properties["prop_primitive"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_regular": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_regular")),
						},
						swaggerURI: specFooPath,
					},
					"prop_array_of_primitive": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["def_regular"].Properties["prop_array_of_primitive"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_regular": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_regular")),
						},
						swaggerURI: specFooPath,
					},
					"prop_array_of_object": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["def_regular"].Properties["prop_array_of_object"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_regular": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_regular")),
						},
						swaggerURI: specFooPath,
					},
					"prop_object": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["def_regular"].Properties["prop_object"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_regular": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_regular")),
						},
						swaggerURI: specFooPath,
					},
				},
				swaggerURI: specFooPath,
			},
		},
		// 2
		{
			specBaseURL: specBasePath,
			specRelPath: "foo.json",
			schemaName:  "def_prop_in_file_ref",
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "foo.json",
					Name:           "def_prop_in_file_ref",
				},
				Properties: map[string]*Property{
					"prop": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["def_prop_in_file_ref"].Properties["prop"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_prop_in_file_ref": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_prop_in_file_ref")),
						},
						swaggerURI: specFooPath,
					},
				},
				swaggerURI: specFooPath,
			},
		},
		// 3
		{
			specBaseURL: specBasePath,
			specRelPath: "foo.json",
			schemaName:  "def_prop_cross_file_ref",
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "foo.json",
					Name:           "def_prop_cross_file_ref",
				},
				Properties: map[string]*Property{
					"prop": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["def_prop_cross_file_ref"].Properties["prop"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_prop_cross_file_ref": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_prop_cross_file_ref")),
						},
						swaggerURI: specFooPath,
					},
				},
				swaggerURI: specFooPath,
			},
		},
		// 4
		{
			specBaseURL: specBasePath,
			specRelPath: "foo.json",
			schemaName:  "def_prop_self_ref",
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "foo.json",
					Name:           "def_prop_self_ref",
				},
				Properties: map[string]*Property{
					"prop": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["def_prop_self_ref"].Properties["prop"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_prop_self_ref": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_prop_self_ref")),
						},
						swaggerURI: specFooPath,
					},
				},
				swaggerURI: specFooPath,
			},
		},
		// 5
		{
			specBaseURL: specBasePath,
			specRelPath: "foo.json",
			schemaName:  "def_array_ref",
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "foo.json",
					Name:           "def_array_ref",
				},
				Properties: map[string]*Property{
					"foo": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["def_foo"].Properties["foo"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_foo":       struct{}{},
							specFooPath + "#/definitions/def_array_ref": struct{}{},
						},
						swaggerURI: specFooPath,
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_foo")),
						},
					},
				},
				swaggerURI: specFooPath,
			},
		},
		// 6
		{
			specBaseURL: specBasePath,
			specRelPath: "foo.json",
			schemaName:  "def_in_file_ref",
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "foo.json",
					Name:           "def_in_file_ref",
				},
				Properties: map[string]*Property{
					"foo": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["def_foo"].Properties["prop"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_in_file_ref": struct{}{},
							specFooPath + "#/definitions/def_foo":         struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_foo")),
						},
						swaggerURI: specFooPath,
					},
				},
				swaggerURI: specFooPath,
			},
		},
		// 7
		{
			specBaseURL: specBasePath,
			specRelPath: "foo.json",
			schemaName:  "def_cross_file_ref",
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "foo.json",
					Name:           "def_cross_file_ref",
				},
				Properties: map[string]*Property{
					"bar": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerBar.Definitions["def_bar"].Properties["bar"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_cross_file_ref": struct{}{},
							specBarPath + "#/definitions/def_bar":            struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specBarPath, "def_bar")),
						},
						swaggerURI: specBarPath,
					},
				},
				swaggerURI: specFooPath,
			},
		},
		// 8
		{
			specBaseURL: specBasePath,
			specRelPath: "foo.json",
			schemaName:  "def_self_ref",
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "foo.json",
					Name:           "def_self_ref",
				},
				Properties: map[string]*Property{},
				swaggerURI: specFooPath,
			},
		},
		// 9
		{
			specBaseURL: specBasePath,
			specRelPath: "foo.json",
			schemaName:  "def_all_of",
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "foo.json",
					Name:           "def_all_of",
				},
				Properties: map[string]*Property{
					"prop_nested": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["def_all_of"].AllOf[0].Properties["prop_nested"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_all_of": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_all_of")),
						},
						swaggerURI: specFooPath,
					},
					"foo": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["def_foo"].Properties["foo"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_all_of":  struct{}{},
							specFooPath + "#/definitions/def_all_of2": struct{}{},
							specFooPath + "#/definitions/def_foo":     struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_foo")),
						},
						swaggerURI: specFooPath,
					},
					"bar": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerBar.Definitions["def_bar"].Properties["bar"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_all_of": struct{}{},
							specBarPath + "#/definitions/def_bar":    struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specBarPath, "def_bar")),
						},
						swaggerURI: specBarPath,
					},
					"baz": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerBaz.Definitions["def_baz"].Properties["baz"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_all_of": struct{}{},
							specBazPath + "#/definitions/def_all_of": struct{}{},
							specBazPath + "#/definitions/def_baz":    struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specBazPath, "def_baz")),
						},
						swaggerURI: specBazPath,
					},
					"p1": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["def_all_of"].Properties["p1"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_all_of": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_all_of")),
						},
						swaggerURI: specFooPath,
					},
				},
				swaggerURI: specFooPath,
			},
		},
		// 10
		{
			specBaseURL: specBasePath,
			specRelPath: "foo.json",
			schemaName:  "def_base",
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "foo.json",
					Name:           "def_base",
				},
				Properties: map[string]*Property{
					"{def_variant1}": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["def_variant1"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_variant1": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_variant1")),
						},
						swaggerURI:          specFooPath,
						isDiscriminator:     true,
						directVariantRefSet: map[string]struct{}{},
						rootVariantRefSet: map[string]struct{}{
							fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_variant1"): {},
							fmt.Sprintf("%s#/definitions/%s", specFooPath, "variant2_def"): {},
						},
					},
					"{def_variant2}": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["variant2_def"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/variant2_def": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "variant2_def")),
						},
						swaggerURI:          specFooPath,
						isDiscriminator:     true,
						directVariantRefSet: map[string]struct{}{},
						rootVariantRefSet: map[string]struct{}{
							fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_variant1"): {},
							fmt.Sprintf("%s#/definitions/%s", specFooPath, "variant2_def"): {},
						},
					},
				},
				swaggerURI: specFooPath,
			},
		},
		// 11
		{
			specBaseURL: specBasePath,
			specRelPath: "foo.json",
			schemaName:  "def_to_def_base",
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "foo.json",
					Name:           "def_to_def_base",
				},
				Properties: map[string]*Property{
					"prop": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["def_to_def_base"].Properties["prop"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_to_def_base": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_to_def_base")),
						},
						swaggerURI: specFooPath,
					},
				},
				swaggerURI: specFooPath,
			},
		},
		// 12
		{
			specBaseURL: specBasePath,
			specRelPath: "foo.json",
			schemaName:  "non_existent_schema",
			err:         true,
		},
		// 13
		// The "def_base_template" uses the same discriminator values as "def_base",
		// the difference is that the variant schemas are inheriting from different base schemas (via allOf).
		{
			specBaseURL: specBasePath,
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
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["def_variant1_template"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_variant1_template": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_variant1_template")),
						},
						swaggerURI:          specFooPath,
						isDiscriminator:     true,
						directVariantRefSet: map[string]struct{}{},
						rootVariantRefSet: map[string]struct{}{
							fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_variant1_template"): {},
							fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_variant2_template"): {},
						},
					},
					"{def_variant2}": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["def_variant2_template"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_variant2_template": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_variant2_template")),
						},
						swaggerURI:          specFooPath,
						isDiscriminator:     true,
						directVariantRefSet: map[string]struct{}{},
						rootVariantRefSet: map[string]struct{}{
							fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_variant1_template"): {},
							fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_variant2_template"): {},
						},
					},
				},
				swaggerURI: specFooPath,
			},
		},
		// 14
		{
			specBaseURL: specBasePath,
			specRelPath: "foo.json",
			schemaName:  "def_parent",
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "foo.json",
					Name:           "def_parent",
				},
				Properties: map[string]*Property{
					"{def_child1}": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["def_child1"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_child1": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_child1")),
						},
						swaggerURI:          specFooPath,
						isDiscriminator:     true,
						directVariantRefSet: map[string]struct{}{},
						rootVariantRefSet: map[string]struct{}{
							fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_child1"): {},
							fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_child2"): {},
						},
					},
					"{def_child2}": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["def_child2"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_child2": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_child2")),
						},
						swaggerURI:          specFooPath,
						isDiscriminator:     true,
						directVariantRefSet: map[string]struct{}{},
						rootVariantRefSet: map[string]struct{}{
							fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_child1"): {},
							fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_child2"): {},
						},
					},
				},
				swaggerURI: specFooPath,
			},
		},
		// 15
		{
			specBaseURL: specBasePath,
			specRelPath: "nested_discriminator.json",
			schemaName:  "def_a",
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "nested_discriminator.json",
					Name:           "def_a",
				},
				Properties: map[string]*Property{
					"{def_b}": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerNestedDiscriminator.Definitions["def_b"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPath + "#/definitions/def_b": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_b")),
						},
						swaggerURI:      specNestedDiscriminatorPath,
						isDiscriminator: true,
						directVariantRefSet: map[string]struct{}{
							fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_b"): {},
							fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_c"): {},
						},
						rootVariantRefSet: map[string]struct{}{
							fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_b"): {},
							fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_c"): {},
						},
					},
					"{def_c}": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerNestedDiscriminator.Definitions["def_c"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPath + "#/definitions/def_c": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_c")),
						},
						swaggerURI:          specNestedDiscriminatorPath,
						isDiscriminator:     true,
						directVariantRefSet: map[string]struct{}{},
						rootVariantRefSet: map[string]struct{}{
							fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_b"): {},
							fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_c"): {},
						},
					},
				},
				swaggerURI: specNestedDiscriminatorPath,
			},
		},
		// 16
		{
			specBaseURL: specBasePath,
			specRelPath: "nested_discriminator.json",
			schemaName:  "def_b",
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "nested_discriminator.json",
					Name:           "def_b",
				},
				Properties: map[string]*Property{
					"{def_b}": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerNestedDiscriminator.Definitions["def_b"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPath + "#/definitions/def_b": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_b")),
						},
						swaggerURI:      specNestedDiscriminatorPath,
						isDiscriminator: true,
						directVariantRefSet: map[string]struct{}{
							fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_b"): {},
							fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_c"): {},
						},
						rootVariantRefSet: map[string]struct{}{
							fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_b"): {},
							fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_c"): {},
						},
					},
					"{def_c}": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerNestedDiscriminator.Definitions["def_c"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPath + "#/definitions/def_c": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_c")),
						},
						swaggerURI:          specNestedDiscriminatorPath,
						isDiscriminator:     true,
						directVariantRefSet: map[string]struct{}{},
						rootVariantRefSet: map[string]struct{}{
							fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_b"): {},
							fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_c"): {},
						},
					},
				},
				swaggerURI: specNestedDiscriminatorPath,
			},
		},
		// 17
		{
			specBaseURL: specBasePath,
			specRelPath: "nested_discriminator.json",
			schemaName:  "def_wrap_a",
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "nested_discriminator.json",
					Name:           "def_wrap_a",
				},
				Properties: map[string]*Property{
					"a": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerNestedDiscriminator.Definitions["def_wrap_a"].Properties["a"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPath + "#/definitions/def_wrap_a": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_wrap_a")),
						},
						swaggerURI: specNestedDiscriminatorPath,
					},
				},
				swaggerURI: specNestedDiscriminatorPath,
			},
		},
		// 18
		{
			specBaseURL: specBasePath,
			specRelPath: "nested_discriminator.json",
			schemaName:  "def_wrap_b",
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "nested_discriminator.json",
					Name:           "def_wrap_b",
				},
				Properties: map[string]*Property{
					"b": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerNestedDiscriminator.Definitions["def_wrap_b"].Properties["b"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPath + "#/definitions/def_wrap_b": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_wrap_b")),
						},
						swaggerURI: specNestedDiscriminatorPath,
					},
				},
				swaggerURI: specNestedDiscriminatorPath,
			},
		},
		// 19
		{
			specBaseURL: specBasePath,
			specRelPath: "nested_discriminator.json",
			schemaName:  "def_c",
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "nested_discriminator.json",
					Name:           "def_c",
				},
				Properties: map[string]*Property{
					"prop_b": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerNestedDiscriminator.Definitions["def_b"].Properties["prop_b"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPath + "#/definitions/def_c": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_b")),
						},
						swaggerURI: specNestedDiscriminatorPath,
					},
					"prop_c": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerNestedDiscriminator.Definitions["def_c"].Properties["prop_c"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPath + "#/definitions/def_c": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_c")),
						},
						swaggerURI: specNestedDiscriminatorPath,
					},
					"type": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerNestedDiscriminator.Definitions["def_a"].Properties["type"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPath + "#/definitions/def_c": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_a")),
						},
						swaggerURI: specNestedDiscriminatorPath,
					},
				},
				swaggerURI: specNestedDiscriminatorPath,
			},
		},
		// 20
		{
			specBaseURL: specBasePath,
			specRelPath: "foo.json",
			schemaName:  "def_to_remote_discriminator",
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "foo.json",
					Name:           "def_to_remote_discriminator",
				},
				Properties: map[string]*Property{
					"properties": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["def_to_remote_discriminator"].Properties["properties"],
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_to_remote_discriminator")),
						},
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_to_remote_discriminator": struct{}{},
						},
						swaggerURI: specFooPath,
					},
				},
				swaggerURI: specFooPath,
			},
		},
		// 21
		{
			specBaseURL: specBasePath,
			specRelPath: "foo.json",
			schemaName:  "def_contain_ro",
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "foo.json",
					Name:           "def_contain_ro",
				},
				Properties: map[string]*Property{
					"p_primitive": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["def_contain_ro"].Properties["p_primitive"],
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(specFooPath + "#/definitions/def_contain_ro"),
						},
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_contain_ro": struct{}{},
						},
						swaggerURI: specFooPath,
					},
					"p_primitive_ro": {
						PropertyState: PropertyStateIgnoredForReqROProp,
						schema:        swaggerFoo.Definitions["def_contain_ro"].Properties["p_primitive_ro"],
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(specFooPath + "#/definitions/def_contain_ro"),
						},
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_contain_ro": struct{}{},
						},
						swaggerURI: specFooPath,
					},
					"p_obj_array": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["def_contain_ro"].Properties["p_obj_array"],
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(specFooPath + "#/definitions/def_contain_ro"),
						},
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_contain_ro": struct{}{},
						},
						swaggerURI: specFooPath,
					},
					"p_obj_array_ro": {
						PropertyState: PropertyStateIgnoredForReqROProp,
						schema:        swaggerFoo.Definitions["def_contain_ro"].Properties["p_obj_array_ro"],
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(specFooPath + "#/definitions/def_contain_ro"),
						},
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_contain_ro": struct{}{},
						},
						swaggerURI: specFooPath,
					},
					"p_obj": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["def_contain_ro"].Properties["p_obj"],
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(specFooPath + "#/definitions/def_contain_ro"),
						},
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_contain_ro": struct{}{},
						},
						swaggerURI: specFooPath,
					},
					"p_obj_ro": {
						PropertyState: PropertyStateIgnoredForReqROProp,
						schema:        swaggerFoo.Definitions["def_contain_ro"].Properties["p_obj_ro"],
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(specFooPath + "#/definitions/def_contain_ro"),
						},
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_contain_ro": struct{}{},
						},
						swaggerURI: specFooPath,
					},
				},
				swaggerURI: specFooPath,
			},
		},
		// 22: non-leaf variant will cause all possible variants (including siblings) of its root discriminator expanded.
		{
			specBaseURL: specBasePath,
			specRelPath: "sibling_variant.json",
			schemaName:  "def_b",
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "sibling_variant.json",
					Name:           "def_b",
				},
				Properties: map[string]*Property{
					"{def_a}": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerSiblingVariant.Definitions["def_a"],
						resolvedRefs: map[string]interface{}{
							specSiblingVariantPath + "#/definitions/def_a": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specSiblingVariantPath, "def_a")),
						},
						swaggerURI:      specSiblingVariantPath,
						isDiscriminator: true,
						rootVariantRefSet: map[string]struct{}{
							fmt.Sprintf("%s#/definitions/%s", specSiblingVariantPath, "def_a"):       {},
							fmt.Sprintf("%s#/definitions/%s", specSiblingVariantPath, "def_b"):       {},
							fmt.Sprintf("%s#/definitions/%s", specSiblingVariantPath, "def_b_child"): {},
						},
						directVariantRefSet: map[string]struct{}{},
					},
					"{def_b}": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerSiblingVariant.Definitions["def_b"],
						resolvedRefs: map[string]interface{}{
							specSiblingVariantPath + "#/definitions/def_b": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specSiblingVariantPath, "def_b")),
						},
						swaggerURI:      specSiblingVariantPath,
						isDiscriminator: true,
						rootVariantRefSet: map[string]struct{}{
							fmt.Sprintf("%s#/definitions/%s", specSiblingVariantPath, "def_a"):       {},
							fmt.Sprintf("%s#/definitions/%s", specSiblingVariantPath, "def_b"):       {},
							fmt.Sprintf("%s#/definitions/%s", specSiblingVariantPath, "def_b_child"): {},
						},
						directVariantRefSet: map[string]struct{}{
							fmt.Sprintf("%s#/definitions/%s", specSiblingVariantPath, "def_b"):       {},
							fmt.Sprintf("%s#/definitions/%s", specSiblingVariantPath, "def_b_child"): {},
						},
					},
					"{def_b_child}": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerSiblingVariant.Definitions["def_b_child"],
						resolvedRefs: map[string]interface{}{
							specSiblingVariantPath + "#/definitions/def_b_child": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specSiblingVariantPath, "def_b_child")),
						},
						swaggerURI:      specSiblingVariantPath,
						isDiscriminator: true,
						rootVariantRefSet: map[string]struct{}{
							fmt.Sprintf("%s#/definitions/%s", specSiblingVariantPath, "def_a"):       {},
							fmt.Sprintf("%s#/definitions/%s", specSiblingVariantPath, "def_b"):       {},
							fmt.Sprintf("%s#/definitions/%s", specSiblingVariantPath, "def_b_child"): {},
						},
						directVariantRefSet: map[string]struct{}{},
					},
				},
				swaggerURI: specSiblingVariantPath,
			},
		},
		// 23: leaf variant will only cause itself expanded, as a normal property.
		{
			specBaseURL: specBasePath,
			specRelPath: "sibling_variant.json",
			schemaName:  "def_b_child",
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "sibling_variant.json",
					Name:           "def_b_child",
				},
				Properties: map[string]*Property{
					"type": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerSiblingVariant.Definitions["def_root"].Properties["type"],
						resolvedRefs: map[string]interface{}{
							specSiblingVariantPath + "#/definitions/def_b_child": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specSiblingVariantPath, "def_root")),
						},
						swaggerURI: specSiblingVariantPath,
					},
				},
				swaggerURI: specSiblingVariantPath,
			},
		},
	}

	for idx, c := range cases {
		msg := fmt.Sprintf("(%d) %s: %s", idx, filepath.Join(c.specBaseURL, c.specRelPath), c.schemaName)
		for i := 0; i <= c.repeat; i++ {
			actual, err := NewEmptySchema(c.specBaseURL, c.specRelPath, c.schemaName, c.forResponse, false)
			if c.err {
				require.Errorf(t, err, msg)
				continue
			}
			require.NoErrorf(t, err, msg)
			require.Equal(t, c.expect, *actual, msg)
		}
	}
}

func TestSchema_ExpandProperties(t *testing.T) {
	cases := []struct {
		name           string
		swaggerRelPath string
		schemaName     string
		forResponse    bool
		repeat         int
		expect         Properties
	}{
		// 0
		{
			name:           `Expand "def_foo"`,
			swaggerRelPath: "foo.json",
			schemaName:     "def_foo",
			expect: map[string]*Property{
				"foo": {
					PropertyState: PropertyStateUncovered,
					schema:        swaggerFoo.Definitions["def_foo"].Properties["foo"],
					resolvedRefs: map[string]interface{}{
						specFooPath + "#/definitions/def_foo": struct{}{},
					},
					ownerSchemaRef: &openapispec.Ref{
						Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_foo")),
					},
					swaggerURI: specFooPath,
				},
			},
		},
		// 1
		{
			name:           `Expand "def_regular"`,
			swaggerRelPath: "foo.json",
			schemaName:     "def_regular",
			expect: map[string]*Property{
				"prop_primitive": {
					PropertyState: PropertyStateUncovered,
					schema:        swaggerFoo.Definitions["def_regular"].Properties["prop_primitive"],
					resolvedRefs: map[string]interface{}{
						specFooPath + "#/definitions/def_regular": struct{}{},
					},
					ownerSchemaRef: &openapispec.Ref{
						Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_regular")),
					},
					swaggerURI: specFooPath,
				},
				"prop_array_of_primitive": {
					PropertyState: PropertyStateUncovered,
					schema:        swaggerFoo.Definitions["def_regular"].Properties["prop_array_of_primitive"],
					resolvedRefs: map[string]interface{}{
						specFooPath + "#/definitions/def_regular": struct{}{},
					},
					ownerSchemaRef: &openapispec.Ref{
						Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_regular")),
					},
					swaggerURI: specFooPath,
				},
				"prop_array_of_object/prop_nested": {
					PropertyState: PropertyStateUncovered,
					schema:        swaggerFoo.Definitions["def_regular"].Properties["prop_array_of_object"].Items.Schema.Properties["prop_nested"],
					resolvedRefs: map[string]interface{}{
						specFooPath + "#/definitions/def_regular": struct{}{},
					},
					ownerSchemaRef: &openapispec.Ref{
						Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_regular")),
					},
					swaggerURI: specFooPath,
				},
				"prop_object/prop_nested": {
					PropertyState: PropertyStateUncovered,
					schema:        swaggerFoo.Definitions["def_regular"].Properties["prop_object"].Properties["prop_nested"],
					resolvedRefs: map[string]interface{}{
						specFooPath + "#/definitions/def_regular": struct{}{},
					},
					ownerSchemaRef: &openapispec.Ref{
						Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_regular")),
					},
					swaggerURI: specFooPath,
				},
			},
		},
		// 2
		{
			name:           `Expand "def_prop_in_file_ref"`,
			swaggerRelPath: "foo.json",
			schemaName:     "def_prop_in_file_ref",
			expect: map[string]*Property{
				"prop/foo": {
					PropertyState: PropertyStateUncovered,
					schema:        swaggerFoo.Definitions["def_foo"].Properties["foo"],
					resolvedRefs: map[string]interface{}{
						specFooPath + "#/definitions/def_prop_in_file_ref": struct{}{},
						specFooPath + "#/definitions/def_foo":              struct{}{},
					},
					ownerSchemaRef: &openapispec.Ref{
						Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_foo")),
					},
					swaggerURI: specFooPath,
				},
			},
		},
		// 3
		{
			name:           `Expand "def_prop_self_ref"`,
			swaggerRelPath: "foo.json",
			schemaName:     "def_prop_self_ref",
			expect: map[string]*Property{
				"prop": {
					PropertyState: PropertyStateUncovered,
					schema:        swaggerFoo.Definitions["def_prop_self_ref"].Properties["prop"],
					resolvedRefs: map[string]interface{}{
						specFooPath + "#/definitions/def_prop_self_ref": struct{}{},
					},
					ownerSchemaRef: &openapispec.Ref{
						Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_prop_self_ref")),
					},
					swaggerURI: specFooPath,
				},
			},
		},
		// 4
		{
			name:           `Expand "def_prop_cross_file_ref"`,
			swaggerRelPath: "foo.json",
			schemaName:     "def_prop_cross_file_ref",
			expect: map[string]*Property{
				"prop/bar": {
					PropertyState: PropertyStateUncovered,
					schema:        swaggerBar.Definitions["def_bar"].Properties["bar"],
					resolvedRefs: map[string]interface{}{
						specFooPath + "#/definitions/def_prop_cross_file_ref": struct{}{},
						specBarPath + "#/definitions/def_bar":                 struct{}{},
					},
					ownerSchemaRef: &openapispec.Ref{
						Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specBarPath, "def_bar")),
					},
					swaggerURI: specBarPath,
				},
			},
		},
		// 5
		{
			name:           `Expand "def_in_file_ref"`,
			swaggerRelPath: "foo.json",
			schemaName:     "def_in_file_ref",
			expect: map[string]*Property{
				"foo": {
					PropertyState: PropertyStateUncovered,
					schema:        swaggerFoo.Definitions["def_foo"].Properties["foo"],
					resolvedRefs: map[string]interface{}{
						specFooPath + "#/definitions/def_in_file_ref": struct{}{},
						specFooPath + "#/definitions/def_foo":         struct{}{},
					},
					ownerSchemaRef: &openapispec.Ref{
						Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_foo")),
					},
					swaggerURI: specFooPath,
				},
			},
		},
		// 6
		{
			name:           `Expand "def_cross_file_ref"`,
			swaggerRelPath: "foo.json",
			schemaName:     "def_cross_file_ref",
			expect: map[string]*Property{
				"bar": {
					PropertyState: PropertyStateUncovered,
					schema:        swaggerBar.Definitions["def_bar"].Properties["bar"],
					resolvedRefs: map[string]interface{}{
						specFooPath + "#/definitions/def_cross_file_ref": struct{}{},
						specBarPath + "#/definitions/def_bar":            struct{}{},
					},
					ownerSchemaRef: &openapispec.Ref{
						Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specBarPath, "def_bar")),
					},
					swaggerURI: specBarPath,
				},
			},
		},
		// 7
		{
			name:           `Expand "def_self_ref"`,
			swaggerRelPath: "foo.json",
			schemaName:     "def_self_ref",
			expect:         map[string]*Property{},
		},
		// 8
		{
			name:           `Expand "def_base" (schema itself is a discriminator)`,
			swaggerRelPath: "foo.json",
			schemaName:     "def_base",
			expect: map[string]*Property{
				"{def_variant1}/type": {
					PropertyState: PropertyStateUncovered,
					schema:        swaggerFoo.Definitions["def_base"].Properties["type"],
					resolvedRefs: map[string]interface{}{
						specFooPath + "#/definitions/def_variant1": struct{}{},
					},
					ownerSchemaRef: &openapispec.Ref{
						Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_base")),
					},
					swaggerURI: specFooPath,
				},
				"{def_variant2}/type": {
					PropertyState: PropertyStateUncovered,
					schema:        swaggerFoo.Definitions["def_base"].Properties["type"],
					resolvedRefs: map[string]interface{}{
						specFooPath + "#/definitions/variant2_def": struct{}{},
					},
					ownerSchemaRef: &openapispec.Ref{
						Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_base")),
					},
					swaggerURI: specFooPath,
				},
			},
		},
		// 9
		{
			name:           `Expand "def_to_def_base" (property is a discriminator)`,
			swaggerRelPath: "foo.json",
			schemaName:     "def_to_def_base",
			expect: map[string]*Property{
				"prop{def_variant1}/type": {
					PropertyState: PropertyStateUncovered,
					schema:        swaggerFoo.Definitions["def_base"].Properties["type"],
					resolvedRefs: map[string]interface{}{
						specFooPath + "#/definitions/def_to_def_base": struct{}{},
						specFooPath + "#/definitions/def_variant1":    struct{}{},
					},
					ownerSchemaRef: &openapispec.Ref{
						Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_base")),
					},
					swaggerURI: specFooPath,
				},
				"prop{def_variant2}/type": {
					PropertyState: PropertyStateUncovered,
					schema:        swaggerFoo.Definitions["def_base"].Properties["type"],
					resolvedRefs: map[string]interface{}{
						specFooPath + "#/definitions/def_to_def_base": struct{}{},
						specFooPath + "#/definitions/variant2_def":    struct{}{},
					},
					ownerSchemaRef: &openapispec.Ref{
						Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_base")),
					},
					swaggerURI: specFooPath,
				},
			},
		},
		// 10
		{
			name:           `Expand "def_to_def_base_template" (property is a discriminator)`,
			swaggerRelPath: "foo.json",
			schemaName:     "def_to_def_base_template",
			repeat:         10, // ensure the result is consistent
			expect: map[string]*Property{
				"prop{def_variant1}/type": {
					PropertyState: PropertyStateUncovered,
					schema:        swaggerFoo.Definitions["def_base_template"].Properties["type"],
					resolvedRefs: map[string]interface{}{
						specFooPath + "#/definitions/def_to_def_base_template": struct{}{},
						specFooPath + "#/definitions/def_variant1_template":    struct{}{},
					},
					ownerSchemaRef: &openapispec.Ref{
						Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_base_template")),
					},
					swaggerURI: specFooPath,
				},
				"prop{def_variant2}/type": {
					PropertyState: PropertyStateUncovered,
					schema:        swaggerFoo.Definitions["def_base_template"].Properties["type"],
					resolvedRefs: map[string]interface{}{
						specFooPath + "#/definitions/def_to_def_base_template": struct{}{},
						specFooPath + "#/definitions/def_variant2_template":    struct{}{},
					},
					ownerSchemaRef: &openapispec.Ref{
						Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_base_template")),
					},
					swaggerURI: specFooPath,
				},
			},
		},
		// 11
		{
			name:           `Expand "def_parent" (schema itself is a discriminator, but didn't specify the accepted variant via enum)`,
			swaggerRelPath: "foo.json",
			schemaName:     "def_parent",
			expect: map[string]*Property{
				"{def_child1}/type": {
					PropertyState: PropertyStateUncovered,
					schema:        swaggerFoo.Definitions["def_parent"].Properties["type"],
					resolvedRefs: map[string]interface{}{
						specFooPath + "#/definitions/def_child1": struct{}{},
					},
					ownerSchemaRef: &openapispec.Ref{
						Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_parent")),
					},
					swaggerURI: specFooPath,
				},
				"{def_child2}/type": {
					PropertyState: PropertyStateUncovered,
					schema:        swaggerFoo.Definitions["def_parent"].Properties["type"],
					resolvedRefs: map[string]interface{}{
						specFooPath + "#/definitions/def_child2": struct{}{},
					},
					ownerSchemaRef: &openapispec.Ref{
						Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_parent")),
					},
					swaggerURI: specFooPath,
				},
			},
		},
		// 12
		{
			name:           `Expand "def_b" (schema is a in-direct discriminator)`,
			swaggerRelPath: "nested_discriminator.json",
			schemaName:     "def_b",
			expect: map[string]*Property{
				"{def_b}/type": {
					PropertyState: PropertyStateUncovered,
					schema:        swaggerNestedDiscriminator.Definitions["def_a"].Properties["type"],
					resolvedRefs: map[string]interface{}{
						specNestedDiscriminatorPath + "#/definitions/def_b": struct{}{},
					},
					ownerSchemaRef: &openapispec.Ref{
						Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_a")),
					},
					swaggerURI: specNestedDiscriminatorPath,
				},
				"{def_b}/prop_b": {
					PropertyState: PropertyStateUncovered,
					schema:        swaggerNestedDiscriminator.Definitions["def_b"].Properties["prop_b"],
					resolvedRefs: map[string]interface{}{
						specNestedDiscriminatorPath + "#/definitions/def_b": struct{}{},
					},
					ownerSchemaRef: &openapispec.Ref{
						Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_b")),
					},
					swaggerURI: specNestedDiscriminatorPath,
				},
				"{def_c}/type": {
					PropertyState: PropertyStateUncovered,
					schema:        swaggerNestedDiscriminator.Definitions["def_a"].Properties["type"],
					resolvedRefs: map[string]interface{}{
						specNestedDiscriminatorPath + "#/definitions/def_c": struct{}{},
					},
					ownerSchemaRef: &openapispec.Ref{
						Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_a")),
					},
					swaggerURI: specNestedDiscriminatorPath,
				},
				"{def_c}/prop_b": {
					PropertyState: PropertyStateUncovered,
					schema:        swaggerNestedDiscriminator.Definitions["def_b"].Properties["prop_b"],
					resolvedRefs: map[string]interface{}{
						specNestedDiscriminatorPath + "#/definitions/def_c": struct{}{},
					},
					ownerSchemaRef: &openapispec.Ref{
						Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_b")),
					},
					swaggerURI: specNestedDiscriminatorPath,
				},
				"{def_c}/prop_c": {
					PropertyState: PropertyStateUncovered,
					schema:        swaggerNestedDiscriminator.Definitions["def_c"].Properties["prop_c"],
					resolvedRefs: map[string]interface{}{
						specNestedDiscriminatorPath + "#/definitions/def_c": struct{}{},
					},
					ownerSchemaRef: &openapispec.Ref{
						Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_c")),
					},
					swaggerURI: specNestedDiscriminatorPath,
				},
			},
		},
		// 13
		{
			name:           `Expand "def_wrap_a" (property is a multi-level discriminator)`,
			swaggerRelPath: "nested_discriminator.json",
			schemaName:     "def_wrap_a",
			expect: map[string]*Property{
				"a{def_b}/type": {
					PropertyState: PropertyStateUncovered,
					schema:        swaggerNestedDiscriminator.Definitions["def_a"].Properties["type"],
					resolvedRefs: map[string]interface{}{
						specNestedDiscriminatorPath + "#/definitions/def_wrap_a": struct{}{},
						specNestedDiscriminatorPath + "#/definitions/def_b":      struct{}{},
					},
					ownerSchemaRef: &openapispec.Ref{
						Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_a")),
					},
					swaggerURI: specNestedDiscriminatorPath,
				},
				"a{def_b}/prop_b": {
					PropertyState: PropertyStateUncovered,
					schema:        swaggerNestedDiscriminator.Definitions["def_b"].Properties["prop_b"],
					resolvedRefs: map[string]interface{}{
						specNestedDiscriminatorPath + "#/definitions/def_wrap_a": struct{}{},
						specNestedDiscriminatorPath + "#/definitions/def_b":      struct{}{},
					},
					ownerSchemaRef: &openapispec.Ref{
						Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_b")),
					},
					swaggerURI: specNestedDiscriminatorPath,
				},
				"a{def_c}/type": {
					PropertyState: PropertyStateUncovered,
					schema:        swaggerNestedDiscriminator.Definitions["def_a"].Properties["type"],
					resolvedRefs: map[string]interface{}{
						specNestedDiscriminatorPath + "#/definitions/def_wrap_a": struct{}{},
						specNestedDiscriminatorPath + "#/definitions/def_c":      struct{}{},
					},
					ownerSchemaRef: &openapispec.Ref{
						Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_a")),
					},
					swaggerURI: specNestedDiscriminatorPath,
				},
				"a{def_c}/prop_b": {
					PropertyState: PropertyStateUncovered,
					schema:        swaggerNestedDiscriminator.Definitions["def_b"].Properties["prop_b"],
					resolvedRefs: map[string]interface{}{
						specNestedDiscriminatorPath + "#/definitions/def_wrap_a": struct{}{},
						specNestedDiscriminatorPath + "#/definitions/def_c":      struct{}{},
					},
					ownerSchemaRef: &openapispec.Ref{
						Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_b")),
					},
					swaggerURI: specNestedDiscriminatorPath,
				},
				"a{def_c}/prop_c": {
					PropertyState: PropertyStateUncovered,
					schema:        swaggerNestedDiscriminator.Definitions["def_c"].Properties["prop_c"],
					resolvedRefs: map[string]interface{}{
						specNestedDiscriminatorPath + "#/definitions/def_wrap_a": struct{}{},
						specNestedDiscriminatorPath + "#/definitions/def_c":      struct{}{},
					},
					ownerSchemaRef: &openapispec.Ref{
						Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_c")),
					},
					swaggerURI: specNestedDiscriminatorPath,
				},
			},
		},
		// 14
		{
			name:           `Expand "def_wrap_b" (property is a indirect discriminator)`,
			swaggerRelPath: "nested_discriminator.json",
			schemaName:     "def_wrap_b",
			expect: map[string]*Property{
				"b{def_b}/type": {
					PropertyState: PropertyStateUncovered,
					schema:        swaggerNestedDiscriminator.Definitions["def_a"].Properties["type"],
					resolvedRefs: map[string]interface{}{
						specNestedDiscriminatorPath + "#/definitions/def_wrap_b": struct{}{},
						specNestedDiscriminatorPath + "#/definitions/def_b":      struct{}{},
					},
					ownerSchemaRef: &openapispec.Ref{
						Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_a")),
					},
					swaggerURI: specNestedDiscriminatorPath,
				},
				"b{def_b}/prop_b": {
					PropertyState: PropertyStateUncovered,
					schema:        swaggerNestedDiscriminator.Definitions["def_b"].Properties["prop_b"],
					resolvedRefs: map[string]interface{}{
						specNestedDiscriminatorPath + "#/definitions/def_wrap_b": struct{}{},
						specNestedDiscriminatorPath + "#/definitions/def_b":      struct{}{},
					},
					ownerSchemaRef: &openapispec.Ref{
						Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_b")),
					},
					swaggerURI: specNestedDiscriminatorPath,
				},
				"b{def_c}/type": {
					PropertyState: PropertyStateUncovered,
					schema:        swaggerNestedDiscriminator.Definitions["def_a"].Properties["type"],
					resolvedRefs: map[string]interface{}{
						specNestedDiscriminatorPath + "#/definitions/def_wrap_b": struct{}{},
						specNestedDiscriminatorPath + "#/definitions/def_c":      struct{}{},
					},
					ownerSchemaRef: &openapispec.Ref{
						Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_a")),
					},
					swaggerURI: specNestedDiscriminatorPath,
				},
				"b{def_c}/prop_b": {
					PropertyState: PropertyStateUncovered,
					schema:        swaggerNestedDiscriminator.Definitions["def_b"].Properties["prop_b"],
					resolvedRefs: map[string]interface{}{
						specNestedDiscriminatorPath + "#/definitions/def_wrap_b": struct{}{},
						specNestedDiscriminatorPath + "#/definitions/def_c":      struct{}{},
					},
					ownerSchemaRef: &openapispec.Ref{
						Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_b")),
					},
					swaggerURI: specNestedDiscriminatorPath,
				},
				"b{def_c}/prop_c": {
					PropertyState: PropertyStateUncovered,
					schema:        swaggerNestedDiscriminator.Definitions["def_c"].Properties["prop_c"],
					resolvedRefs: map[string]interface{}{
						specNestedDiscriminatorPath + "#/definitions/def_wrap_b": struct{}{},
						specNestedDiscriminatorPath + "#/definitions/def_c":      struct{}{},
					},
					ownerSchemaRef: &openapispec.Ref{
						Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_c")),
					},
					swaggerURI: specNestedDiscriminatorPath,
				},
			},
		},
		// 15
		{
			name:           `Expand "def_wrap_c" all variant properties (property is a leaf variant)`,
			swaggerRelPath: "nested_discriminator.json",
			schemaName:     "def_wrap_c",
			expect: map[string]*Property{
				"c/prop_b": {
					PropertyState: PropertyStateUncovered,
					schema:        swaggerNestedDiscriminator.Definitions["def_b"].Properties["prop_b"],
					resolvedRefs: map[string]interface{}{
						specNestedDiscriminatorPath + "#/definitions/def_wrap_c": struct{}{},
						specNestedDiscriminatorPath + "#/definitions/def_c":      struct{}{},
					},
					ownerSchemaRef: &openapispec.Ref{
						Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_b")),
					},
					swaggerURI: specNestedDiscriminatorPath,
				},
				"c/prop_c": {
					PropertyState: PropertyStateUncovered,
					schema:        swaggerNestedDiscriminator.Definitions["def_c"].Properties["prop_c"],
					resolvedRefs: map[string]interface{}{
						specNestedDiscriminatorPath + "#/definitions/def_wrap_c": struct{}{},
						specNestedDiscriminatorPath + "#/definitions/def_c":      struct{}{},
					},
					ownerSchemaRef: &openapispec.Ref{
						Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_c")),
					},
					swaggerURI: specNestedDiscriminatorPath,
				},
				"c/type": {
					PropertyState: PropertyStateUncovered,
					schema:        swaggerNestedDiscriminator.Definitions["def_a"].Properties["type"],
					resolvedRefs: map[string]interface{}{
						specNestedDiscriminatorPath + "#/definitions/def_wrap_c": struct{}{},
						specNestedDiscriminatorPath + "#/definitions/def_c":      struct{}{},
					},
					ownerSchemaRef: &openapispec.Ref{
						Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_a")),
					},
					swaggerURI: specNestedDiscriminatorPath,
				},
			},
		},
		// 16
		{
			name:           `Expand "def_a" (schema is itself a discriminator)`,
			swaggerRelPath: "nested_discriminator.json",
			schemaName:     "def_a",
			expect: map[string]*Property{
				"{def_b}/type": {
					PropertyState: PropertyStateUncovered,
					schema:        swaggerNestedDiscriminator.Definitions["def_a"].Properties["type"],
					resolvedRefs: map[string]interface{}{
						specNestedDiscriminatorPath + "#/definitions/def_b": struct{}{},
					},
					ownerSchemaRef: &openapispec.Ref{
						Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_a")),
					},
					swaggerURI: specNestedDiscriminatorPath,
				},
				"{def_b}/prop_b": {
					PropertyState: PropertyStateUncovered,
					schema:        swaggerNestedDiscriminator.Definitions["def_b"].Properties["prop_b"],
					resolvedRefs: map[string]interface{}{
						specNestedDiscriminatorPath + "#/definitions/def_b": struct{}{},
					},
					ownerSchemaRef: &openapispec.Ref{
						Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_b")),
					},
					swaggerURI: specNestedDiscriminatorPath,
				},
				"{def_c}/type": {
					PropertyState: PropertyStateUncovered,
					schema:        swaggerNestedDiscriminator.Definitions["def_a"].Properties["type"],
					resolvedRefs: map[string]interface{}{
						specNestedDiscriminatorPath + "#/definitions/def_c": struct{}{},
					},
					ownerSchemaRef: &openapispec.Ref{
						Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_a")),
					},
					swaggerURI: specNestedDiscriminatorPath,
				},
				"{def_c}/prop_b": {
					PropertyState: PropertyStateUncovered,
					schema:        swaggerNestedDiscriminator.Definitions["def_b"].Properties["prop_b"],
					resolvedRefs: map[string]interface{}{
						specNestedDiscriminatorPath + "#/definitions/def_c": struct{}{},
					},
					ownerSchemaRef: &openapispec.Ref{
						Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_b")),
					},
					swaggerURI: specNestedDiscriminatorPath,
				},
				"{def_c}/prop_c": {
					PropertyState: PropertyStateUncovered,
					schema:        swaggerNestedDiscriminator.Definitions["def_c"].Properties["prop_c"],
					resolvedRefs: map[string]interface{}{
						specNestedDiscriminatorPath + "#/definitions/def_c": struct{}{},
					},
					ownerSchemaRef: &openapispec.Ref{
						Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_c")),
					},
					swaggerURI: specNestedDiscriminatorPath,
				},
			},
		},
		// 17
		{
			name:           `Expand to ref to remote discriminator schema`,
			swaggerRelPath: "foo.json",
			schemaName:     "def_to_remote_discriminator",
			expect: map[string]*Property{
				"properties{1}/p1": {
					PropertyState: PropertyStateUncovered,
					schema:        swaggerRemoteDisc.Definitions["v1"].Properties["p1"],
					ownerSchemaRef: &openapispec.Ref{
						Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specRemoteDiscPath, "v1")),
					},
					resolvedRefs: map[string]interface{}{
						specFooPath + "#/definitions/def_to_remote_discriminator": struct{}{},
						specRemoteDiscPath + "#/definitions/v1":                   struct{}{},
					},
					swaggerURI: specRemoteDiscPath,
				},
				"properties{1}/type": {
					PropertyState: PropertyStateUncovered,
					schema:        swaggerRemoteDisc.Definitions["base"].Properties["type"],
					ownerSchemaRef: &openapispec.Ref{
						Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specRemoteDiscPath, "base")),
					},
					resolvedRefs: map[string]interface{}{
						specFooPath + "#/definitions/def_to_remote_discriminator": struct{}{},
						specRemoteDiscPath + "#/definitions/v1":                   struct{}{},
					},
					swaggerURI: specRemoteDiscPath,
				},
			},
		},
		// 18
		{
			name:           `Expand all properties that contains read-only property`,
			swaggerRelPath: "foo.json",
			schemaName:     "def_contain_ro",
			expect: map[string]*Property{
				"p_primitive": {
					PropertyState: PropertyStateUncovered,
					schema:        swaggerFoo.Definitions["def_contain_ro"].Properties["p_primitive"],
					ownerSchemaRef: &openapispec.Ref{
						Ref: jsonreference.MustCreateRef(specFooPath + "#/definitions/def_contain_ro"),
					},
					resolvedRefs: map[string]interface{}{
						specFooPath + "#/definitions/def_contain_ro": struct{}{},
					},
					swaggerURI: specFooPath,
				},
				"p_primitive_ro": {
					PropertyState: PropertyStateIgnoredForReqROProp,
					schema:        swaggerFoo.Definitions["def_contain_ro"].Properties["p_primitive_ro"],
					ownerSchemaRef: &openapispec.Ref{
						Ref: jsonreference.MustCreateRef(specFooPath + "#/definitions/def_contain_ro"),
					},
					resolvedRefs: map[string]interface{}{
						specFooPath + "#/definitions/def_contain_ro": struct{}{},
					},
					swaggerURI: specFooPath,
				},
				"p_obj_array/foo": {
					PropertyState: PropertyStateUncovered,
					schema:        swaggerFoo.Definitions["def_foo"].Properties["foo"],
					ownerSchemaRef: &openapispec.Ref{
						Ref: jsonreference.MustCreateRef(specFooPath + "#/definitions/def_foo"),
					},
					resolvedRefs: map[string]interface{}{
						specFooPath + "#/definitions/def_contain_ro": struct{}{},
						specFooPath + "#/definitions/def_foo":        struct{}{},
					},
					swaggerURI: specFooPath,
				},
				"p_obj_array_ro/foo": {
					PropertyState: PropertyStateIgnoredForReqROProp,
					schema:        swaggerFoo.Definitions["def_foo"].Properties["foo"],
					ownerSchemaRef: &openapispec.Ref{
						Ref: jsonreference.MustCreateRef(specFooPath + "#/definitions/def_foo"),
					},
					resolvedRefs: map[string]interface{}{
						specFooPath + "#/definitions/def_contain_ro": struct{}{},
						specFooPath + "#/definitions/def_foo":        struct{}{},
					},
					swaggerURI: specFooPath,
				},
				"p_obj/foo": {
					PropertyState: PropertyStateUncovered,
					schema:        swaggerFoo.Definitions["def_foo"].Properties["foo"],
					ownerSchemaRef: &openapispec.Ref{
						Ref: jsonreference.MustCreateRef(specFooPath + "#/definitions/def_foo"),
					},
					resolvedRefs: map[string]interface{}{
						specFooPath + "#/definitions/def_contain_ro": struct{}{},
						specFooPath + "#/definitions/def_foo":        struct{}{},
					},
					swaggerURI: specFooPath,
				},
				"p_obj_ro/foo": {
					PropertyState: PropertyStateIgnoredForReqROProp,
					schema:        swaggerFoo.Definitions["def_foo"].Properties["foo"],
					ownerSchemaRef: &openapispec.Ref{
						Ref: jsonreference.MustCreateRef(specFooPath + "#/definitions/def_foo"),
					},
					resolvedRefs: map[string]interface{}{
						specFooPath + "#/definitions/def_contain_ro": struct{}{},
						specFooPath + "#/definitions/def_foo":        struct{}{},
					},
					swaggerURI: specFooPath,
				},
			},
		},
		// 19
		{
			name:           `Expand Map/Dict properties`,
			swaggerRelPath: "foo.json",
			schemaName:     "def_contain_map",
			expect: map[string]*Property{
				"tags": {
					PropertyState: PropertyStateUncovered,
					schema:        swaggerFoo.Definitions["def_contain_map"].Properties["tags"],
					ownerSchemaRef: &openapispec.Ref{
						Ref: jsonreference.MustCreateRef(specFooPath + "#/definitions/def_contain_map"),
					},
					resolvedRefs: map[string]interface{}{
						specFooPath + "#/definitions/def_contain_map": struct{}{},
					},
					swaggerURI: specFooPath,
				},
			},
		},
		// 20
		{
			name:           `Expand cyclic discriminator`,
			swaggerRelPath: "foo.json",
			schemaName:     "def_activity",
			expect: map[string]*Property{
				"{def_activity_a}/type": {
					PropertyState: PropertyStateUncovered,
					schema:        swaggerFoo.Definitions["def_activity"].Properties["type"],
					ownerSchemaRef: &openapispec.Ref{
						Ref: jsonreference.MustCreateRef(specFooPath + "#/definitions/def_activity"),
					},
					resolvedRefs: map[string]interface{}{
						specFooPath + "#/definitions/def_activity_a": struct{}{},
					},
					swaggerURI: specFooPath,
				},
				"{def_activity_a}/actor{def_actor_a}/activity": {
					PropertyState: PropertyStateUncovered,
					schema:        swaggerFoo.Definitions["def_activity"],
					ownerSchemaRef: &openapispec.Ref{
						Ref: jsonreference.MustCreateRef(specFooPath + "#/definitions/def_activity"),
					},
					resolvedRefs: map[string]interface{}{
						specFooPath + "#/definitions/def_actor_a":    struct{}{},
						specFooPath + "#/definitions/def_activity_a": struct{}{},
					},
					swaggerURI:          specFooPath,
					isDiscriminator:     true,
					isPureDiscriminator: true,
					rootVariantRefSet: map[string]struct{}{
						fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_activity_a"): {},
					},
					directVariantRefSet: map[string]struct{}{
						fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_activity_a"): {},
					},
				},
				"{def_activity_a}/actor{def_actor_a}/type": {
					PropertyState: PropertyStateUncovered,
					schema:        swaggerFoo.Definitions["def_actor"].Properties["type"],
					ownerSchemaRef: &openapispec.Ref{
						Ref: jsonreference.MustCreateRef(specFooPath + "#/definitions/def_actor"),
					},
					resolvedRefs: map[string]interface{}{
						specFooPath + "#/definitions/def_actor_a":    struct{}{},
						specFooPath + "#/definitions/def_activity_a": struct{}{},
					},
					swaggerURI: specFooPath,
				},
			},
		},
	}

	for idx, c := range cases {
		for i := 0; i <= c.repeat; i++ {
			schema, err := NewEmptySchema(specBasePath, c.swaggerRelPath, c.schemaName, c.forResponse, false)
			require.NoError(t, err, i)

			require.NoError(t, schema.ExpandProperties(), c.name, idx)
			require.Equal(t, c.expect, schema.Properties, idx)
		}
	}
}

func TestSchema_ExpandPropertyOneLevelDeep(t *testing.T) {
	cases := []struct {
		name           string
		swaggerRelPath string
		schemaName     string
		forResponse    bool
		expandAddrs    []addr.Property
		repeat         int
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
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["def_foo"].Properties["foo"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_foo": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_foo")),
						},
						swaggerURI: specFooPath,
					},
				},
				swaggerURI: specFooPath,
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
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "foo.json",
					Name:           "def_foo",
				},
				Properties: map[string]*Property{
					"foo": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["def_foo"].Properties["foo"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_foo": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_foo")),
						},
						swaggerURI: specFooPath,
					},
				},
				swaggerURI: specFooPath,
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
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "foo.json",
					Name:           "def_regular",
				},
				Properties: map[string]*Property{
					"prop_primitive": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["def_regular"].Properties["prop_primitive"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_regular": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_regular")),
						},
						swaggerURI: specFooPath,
					},
					"prop_array_of_primitive": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["def_regular"].Properties["prop_array_of_primitive"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_regular": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_regular")),
						},
						swaggerURI: specFooPath,
					},
					"prop_array_of_object/prop_nested": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["def_regular"].Properties["prop_array_of_object"].Items.Schema.Properties["prop_nested"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_regular": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_regular")),
						},
						swaggerURI: specFooPath,
					},
					"prop_object/prop_nested": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["def_regular"].Properties["prop_object"].Properties["prop_nested"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_regular": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_regular")),
						},
						swaggerURI: specFooPath,
					},
				},
				swaggerURI: specFooPath,
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
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["def_foo"].Properties["foo"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_prop_in_file_ref": struct{}{},
							specFooPath + "#/definitions/def_foo":              struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_foo")),
						},
						swaggerURI: specFooPath,
					},
				},
				swaggerURI: specFooPath,
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
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["def_prop_self_ref"].Properties["prop"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_prop_self_ref": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_prop_self_ref")),
						},
						swaggerURI: specFooPath,
					},
				},
				swaggerURI: specFooPath,
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
						PropertyState: PropertyStateUncovered,
						schema:        swaggerBar.Definitions["def_bar"].Properties["bar"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_prop_cross_file_ref": struct{}{},
							specBarPath + "#/definitions/def_bar":                 struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specBarPath, "def_bar")),
						},
						swaggerURI: specBarPath,
					},
				},
				swaggerURI: specFooPath,
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
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["def_foo"].Properties["foo"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_in_file_ref": struct{}{},
							specFooPath + "#/definitions/def_foo":         struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_foo")),
						},
						swaggerURI: specFooPath,
					},
				},
				swaggerURI: specFooPath,
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
						PropertyState: PropertyStateUncovered,
						schema:        swaggerBar.Definitions["def_bar"].Properties["bar"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_cross_file_ref": struct{}{},
							specBarPath + "#/definitions/def_bar":            struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specBarPath, "def_bar")),
						},
						swaggerURI: specBarPath,
					},
				},
				swaggerURI: specFooPath,
			},
		},
		// 8
		{
			name:           `Expand "def_self_ref"`,
			swaggerRelPath: "foo.json",
			schemaName:     "def_self_ref",
			expandAddrs:    []addr.Property{},
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "foo.json",
					Name:           "def_self_ref",
				},
				Properties: map[string]*Property{},
				swaggerURI: specFooPath,
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
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["def_base"].Properties["type"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_variant1": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_base")),
						},
						swaggerURI: specFooPath,
					},
					"{def_variant2}": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["variant2_def"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/variant2_def": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "variant2_def")),
						},
						swaggerURI:          specFooPath,
						isDiscriminator:     true,
						directVariantRefSet: map[string]struct{}{},
						rootVariantRefSet: map[string]struct{}{
							fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_variant1"): {},
							fmt.Sprintf("%s#/definitions/%s", specFooPath, "variant2_def"): {},
						},
					},
				},
				swaggerURI: specFooPath,
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
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["def_variant1"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_to_def_base": struct{}{},
							specFooPath + "#/definitions/def_variant1":    struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_variant1")),
						},
						swaggerURI:          specFooPath,
						isDiscriminator:     true,
						directVariantRefSet: map[string]struct{}{},
						rootVariantRefSet: map[string]struct{}{
							fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_variant1"): {},
							fmt.Sprintf("%s#/definitions/%s", specFooPath, "variant2_def"): {},
						},
					},
					"prop{def_variant2}": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["variant2_def"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_to_def_base": struct{}{},
							specFooPath + "#/definitions/variant2_def":    struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "variant2_def")),
						},
						swaggerURI:          specFooPath,
						isDiscriminator:     true,
						directVariantRefSet: map[string]struct{}{},
						rootVariantRefSet: map[string]struct{}{
							fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_variant1"): {},
							fmt.Sprintf("%s#/definitions/%s", specFooPath, "variant2_def"): {},
						},
					},
				},
				swaggerURI: specFooPath,
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
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["def_base"].Properties["type"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_to_def_base": struct{}{},
							specFooPath + "#/definitions/def_variant1":    struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_base")),
						},
						swaggerURI: specFooPath,
					},
					"prop{def_variant2}": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["variant2_def"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_to_def_base": struct{}{},
							specFooPath + "#/definitions/variant2_def":    struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "variant2_def")),
						},
						swaggerURI:          specFooPath,
						isDiscriminator:     true,
						directVariantRefSet: map[string]struct{}{},
						rootVariantRefSet: map[string]struct{}{
							fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_variant1"): {},
							fmt.Sprintf("%s#/definitions/%s", specFooPath, "variant2_def"): {},
						},
					},
				},
				swaggerURI: specFooPath,
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
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["def_variant1_template"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_to_def_base_template": struct{}{},
							specFooPath + "#/definitions/def_variant1_template":    struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_variant1_template")),
						},
						swaggerURI:          specFooPath,
						isDiscriminator:     true,
						directVariantRefSet: map[string]struct{}{},
						rootVariantRefSet: map[string]struct{}{
							fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_variant1_template"): {},
							fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_variant2_template"): {},
						},
					},
					"prop{def_variant2}": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["def_variant2_template"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_to_def_base_template": struct{}{},
							specFooPath + "#/definitions/def_variant2_template":    struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_variant2_template")),
						},
						swaggerURI:          specFooPath,
						isDiscriminator:     true,
						directVariantRefSet: map[string]struct{}{},
						rootVariantRefSet: map[string]struct{}{
							fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_variant1_template"): {},
							fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_variant2_template"): {},
						},
					},
				},
				swaggerURI: specFooPath,
			},
		},
		// 13
		{
			name:           `Expand "def_parent" (schema itself is a discriminator, but didn't specify the accepted variant via enum)`,
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
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["def_parent"].Properties["type"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_child1": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_parent")),
						},
						swaggerURI: specFooPath,
					},
					"{def_child2}": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["def_child2"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_child2": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_child2")),
						},
						swaggerURI:          specFooPath,
						isDiscriminator:     true,
						directVariantRefSet: map[string]struct{}{},
						rootVariantRefSet: map[string]struct{}{
							fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_child1"): {},
							fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_child2"): {},
						},
					},
				},
				swaggerURI: specFooPath,
			},
		},
		// 14
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
					"{def_b}": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerNestedDiscriminator.Definitions["def_b"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPath + "#/definitions/def_b": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_b")),
						},
						swaggerURI:      specNestedDiscriminatorPath,
						isDiscriminator: true,
						directVariantRefSet: map[string]struct{}{
							fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_b"): {},
							fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_c"): {},
						},
						rootVariantRefSet: map[string]struct{}{
							fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_b"): {},
							fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_c"): {},
						},
					},
					"{def_c}/type": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerNestedDiscriminator.Definitions["def_a"].Properties["type"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPath + "#/definitions/def_c": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_a")),
						},
						swaggerURI: specNestedDiscriminatorPath,
					},
					"{def_c}/prop_b": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerNestedDiscriminator.Definitions["def_b"].Properties["prop_b"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPath + "#/definitions/def_c": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_b")),
						},
						swaggerURI: specNestedDiscriminatorPath,
					},
					"{def_c}/prop_c": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerNestedDiscriminator.Definitions["def_c"].Properties["prop_c"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPath + "#/definitions/def_c": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_c")),
						},
						swaggerURI: specNestedDiscriminatorPath,
					},
				},
				swaggerURI: specNestedDiscriminatorPath,
			},
		},
		// 15
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
						PropertyState: PropertyStateUncovered,
						schema:        swaggerNestedDiscriminator.Definitions["def_b"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPath + "#/definitions/def_b":      struct{}{},
							specNestedDiscriminatorPath + "#/definitions/def_wrap_a": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_b")),
						},
						swaggerURI:      specNestedDiscriminatorPath,
						isDiscriminator: true,
						directVariantRefSet: map[string]struct{}{
							fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_b"): {},
							fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_c"): {},
						},
						rootVariantRefSet: map[string]struct{}{
							fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_b"): {},
							fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_c"): {},
						},
					},
					"a{def_c}": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerNestedDiscriminator.Definitions["def_c"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPath + "#/definitions/def_c":      struct{}{},
							specNestedDiscriminatorPath + "#/definitions/def_wrap_a": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_c")),
						},
						swaggerURI:          specNestedDiscriminatorPath,
						isDiscriminator:     true,
						directVariantRefSet: map[string]struct{}{},
						rootVariantRefSet: map[string]struct{}{
							fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_b"): {},
							fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_c"): {},
						},
					},
				},
				swaggerURI: specNestedDiscriminatorPath,
			},
		},
		// 16
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
						PropertyState: PropertyStateUncovered,
						schema:        swaggerNestedDiscriminator.Definitions["def_a"].Properties["type"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPath + "#/definitions/def_b":      struct{}{},
							specNestedDiscriminatorPath + "#/definitions/def_wrap_a": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_a")),
						},
						swaggerURI: specNestedDiscriminatorPath,
					},
					"a{def_b}/prop_b": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerNestedDiscriminator.Definitions["def_b"].Properties["prop_b"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPath + "#/definitions/def_b":      struct{}{},
							specNestedDiscriminatorPath + "#/definitions/def_wrap_a": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_b")),
						},
						swaggerURI: specNestedDiscriminatorPath,
					},
					"a{def_c}": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerNestedDiscriminator.Definitions["def_c"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPath + "#/definitions/def_c":      struct{}{},
							specNestedDiscriminatorPath + "#/definitions/def_wrap_a": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_c")),
						},
						swaggerURI:          specNestedDiscriminatorPath,
						isDiscriminator:     true,
						directVariantRefSet: map[string]struct{}{},
						rootVariantRefSet: map[string]struct{}{
							fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_b"): {},
							fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_c"): {},
						},
					},
				},
				swaggerURI: specNestedDiscriminatorPath,
			},
		},
		// 17
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
						PropertyState: PropertyStateUncovered,
						schema:        swaggerNestedDiscriminator.Definitions["def_b"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPath + "#/definitions/def_b":      struct{}{},
							specNestedDiscriminatorPath + "#/definitions/def_wrap_a": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_b")),
						},
						swaggerURI:      specNestedDiscriminatorPath,
						isDiscriminator: true,
						directVariantRefSet: map[string]struct{}{
							fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_b"): {},
							fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_c"): {},
						},
						rootVariantRefSet: map[string]struct{}{
							fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_b"): {},
							fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_c"): {},
						},
					},
					"a{def_c}/prop_b": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerNestedDiscriminator.Definitions["def_b"].Properties["prop_b"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPath + "#/definitions/def_c":      struct{}{},
							specNestedDiscriminatorPath + "#/definitions/def_wrap_a": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_b")),
						},
						swaggerURI: specNestedDiscriminatorPath,
					},
					"a{def_c}/prop_c": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerNestedDiscriminator.Definitions["def_c"].Properties["prop_c"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPath + "#/definitions/def_c":      struct{}{},
							specNestedDiscriminatorPath + "#/definitions/def_wrap_a": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_c")),
						},
						swaggerURI: specNestedDiscriminatorPath,
					},
					"a{def_c}/type": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerNestedDiscriminator.Definitions["def_a"].Properties["type"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPath + "#/definitions/def_c":      struct{}{},
							specNestedDiscriminatorPath + "#/definitions/def_wrap_a": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_a")),
						},
						swaggerURI: specNestedDiscriminatorPath,
					},
				},
				swaggerURI: specNestedDiscriminatorPath,
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
				addr.MustParseProperty("a{def_b}"),
			},
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "nested_discriminator.json",
					Name:           "def_wrap_a",
				},
				Properties: map[string]*Property{
					"a{def_b}/type": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerNestedDiscriminator.Definitions["def_a"].Properties["type"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPath + "#/definitions/def_b":      struct{}{},
							specNestedDiscriminatorPath + "#/definitions/def_wrap_a": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_a")),
						},
						swaggerURI: specNestedDiscriminatorPath,
					},
					"a{def_b}/prop_b": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerNestedDiscriminator.Definitions["def_b"].Properties["prop_b"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPath + "#/definitions/def_b":      struct{}{},
							specNestedDiscriminatorPath + "#/definitions/def_wrap_a": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_b")),
						},
						swaggerURI: specNestedDiscriminatorPath,
					},
					"a{def_c}/prop_b": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerNestedDiscriminator.Definitions["def_c"].Properties["prop_b"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPath + "#/definitions/def_c":      struct{}{},
							specNestedDiscriminatorPath + "#/definitions/def_wrap_a": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_b")),
						},
						swaggerURI: specNestedDiscriminatorPath,
					},
					"a{def_c}/prop_c": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerNestedDiscriminator.Definitions["def_c"].Properties["prop_c"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPath + "#/definitions/def_c":      struct{}{},
							specNestedDiscriminatorPath + "#/definitions/def_wrap_a": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_c")),
						},
						swaggerURI: specNestedDiscriminatorPath,
					},
					"a{def_c}/type": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerNestedDiscriminator.Definitions["def_a"].Properties["type"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPath + "#/definitions/def_c":      struct{}{},
							specNestedDiscriminatorPath + "#/definitions/def_wrap_a": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_a")),
						},
						swaggerURI: specNestedDiscriminatorPath,
					},
				},
				swaggerURI: specNestedDiscriminatorPath,
			},
		},
		// 19
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
					"b{def_b}": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerNestedDiscriminator.Definitions["def_b"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPath + "#/definitions/def_wrap_b": struct{}{},
							specNestedDiscriminatorPath + "#/definitions/def_b":      struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_b")),
						},
						swaggerURI:      specNestedDiscriminatorPath,
						isDiscriminator: true,
						directVariantRefSet: map[string]struct{}{
							fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_b"): {},
							fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_c"): {},
						},
						rootVariantRefSet: map[string]struct{}{
							fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_b"): {},
							fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_c"): {},
						},
					},
					"b{def_c}": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerNestedDiscriminator.Definitions["def_c"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPath + "#/definitions/def_wrap_b": struct{}{},
							specNestedDiscriminatorPath + "#/definitions/def_c":      struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_c")),
						},
						swaggerURI:          specNestedDiscriminatorPath,
						isDiscriminator:     true,
						directVariantRefSet: map[string]struct{}{},
						rootVariantRefSet: map[string]struct{}{
							fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_b"): {},
							fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_c"): {},
						},
					},
				},
				swaggerURI: specNestedDiscriminatorPath,
			},
		},
		// 20
		{
			name:           `Expand "def_wrap_b" to its leaf variant property (property is a indirect discriminator)`,
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
					"b{def_b}": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerNestedDiscriminator.Definitions["def_b"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPath + "#/definitions/def_wrap_b": struct{}{},
							specNestedDiscriminatorPath + "#/definitions/def_b":      struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_b")),
						},
						swaggerURI:      specNestedDiscriminatorPath,
						isDiscriminator: true,
						directVariantRefSet: map[string]struct{}{
							fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_b"): {},
							fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_c"): {},
						},
						rootVariantRefSet: map[string]struct{}{
							fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_b"): {},
							fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_c"): {},
						},
					},
					"b{def_c}/prop_b": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerNestedDiscriminator.Definitions["def_c"].Properties["prop_b"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPath + "#/definitions/def_c":      struct{}{},
							specNestedDiscriminatorPath + "#/definitions/def_wrap_b": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_b")),
						},
						swaggerURI: specNestedDiscriminatorPath,
					},
					"b{def_c}/prop_c": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerNestedDiscriminator.Definitions["def_c"].Properties["prop_c"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPath + "#/definitions/def_c":      struct{}{},
							specNestedDiscriminatorPath + "#/definitions/def_wrap_b": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_c")),
						},
						swaggerURI: specNestedDiscriminatorPath,
					},
					"b{def_c}/type": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerNestedDiscriminator.Definitions["def_a"].Properties["type"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPath + "#/definitions/def_wrap_b": struct{}{},
							specNestedDiscriminatorPath + "#/definitions/def_c":      struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_a")),
						},
						swaggerURI: specNestedDiscriminatorPath,
					},
				},
				swaggerURI: specNestedDiscriminatorPath,
			},
		},
		// 21
		{
			name:           `Expand "def_wrap_b" to its middle variant property (property is a indirect discriminator)`,
			swaggerRelPath: "nested_discriminator.json",
			schemaName:     "def_wrap_b",
			expandAddrs: []addr.Property{
				addr.MustParseProperty("b"),
				addr.MustParseProperty("b{def_b}"),
			},
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "nested_discriminator.json",
					Name:           "def_wrap_b",
				},
				Properties: map[string]*Property{
					"b{def_b}/prop_b": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerNestedDiscriminator.Definitions["def_b"].Properties["prop_b"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPath + "#/definitions/def_wrap_b": struct{}{},
							specNestedDiscriminatorPath + "#/definitions/def_b":      struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_b")),
						},
						swaggerURI: specNestedDiscriminatorPath,
					},
					"b{def_b}/type": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerNestedDiscriminator.Definitions["def_a"].Properties["type"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPath + "#/definitions/def_wrap_b": struct{}{},
							specNestedDiscriminatorPath + "#/definitions/def_b":      struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_a")),
						},
						swaggerURI: specNestedDiscriminatorPath,
					},
					"b{def_c}": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerNestedDiscriminator.Definitions["def_c"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPath + "#/definitions/def_wrap_b": struct{}{},
							specNestedDiscriminatorPath + "#/definitions/def_c":      struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_c")),
						},
						swaggerURI:          specNestedDiscriminatorPath,
						isDiscriminator:     true,
						directVariantRefSet: map[string]struct{}{},
						rootVariantRefSet: map[string]struct{}{
							fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_b"): {},
							fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_c"): {},
						},
					},
				},
				swaggerURI: specNestedDiscriminatorPath,
			},
		},
		// 22
		{
			name:           `Expand "def_wrap_b" all variant properties (property is a indirect discriminator)`,
			swaggerRelPath: "nested_discriminator.json",
			schemaName:     "def_wrap_b",
			expandAddrs: []addr.Property{
				addr.MustParseProperty("b"),
				addr.MustParseProperty("b{def_b}"),
				addr.MustParseProperty("b{def_c}"),
			},
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "nested_discriminator.json",
					Name:           "def_wrap_b",
				},
				Properties: map[string]*Property{
					"b{def_b}/prop_b": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerNestedDiscriminator.Definitions["def_b"].Properties["prop_b"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPath + "#/definitions/def_wrap_b": struct{}{},
							specNestedDiscriminatorPath + "#/definitions/def_b":      struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_b")),
						},
						swaggerURI: specNestedDiscriminatorPath,
					},
					"b{def_b}/type": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerNestedDiscriminator.Definitions["def_a"].Properties["type"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPath + "#/definitions/def_wrap_b": struct{}{},
							specNestedDiscriminatorPath + "#/definitions/def_b":      struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_a")),
						},
						swaggerURI: specNestedDiscriminatorPath,
					},
					"b{def_c}/prop_b": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerNestedDiscriminator.Definitions["def_c"].Properties["prop_b"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPath + "#/definitions/def_c":      struct{}{},
							specNestedDiscriminatorPath + "#/definitions/def_wrap_b": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_b")),
						},
						swaggerURI: specNestedDiscriminatorPath,
					},
					"b{def_c}/prop_c": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerNestedDiscriminator.Definitions["def_c"].Properties["prop_c"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPath + "#/definitions/def_c":      struct{}{},
							specNestedDiscriminatorPath + "#/definitions/def_wrap_b": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_c")),
						},
						swaggerURI: specNestedDiscriminatorPath,
					},
					"b{def_c}/type": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerNestedDiscriminator.Definitions["def_a"].Properties["type"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPath + "#/definitions/def_wrap_b": struct{}{},
							specNestedDiscriminatorPath + "#/definitions/def_c":      struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_a")),
						},
						swaggerURI: specNestedDiscriminatorPath,
					},
				},
				swaggerURI: specNestedDiscriminatorPath,
			},
		},
		// 23
		{
			name:           `Expand "def_wrap_c" all variant properties (property is a leaf variant)`,
			swaggerRelPath: "nested_discriminator.json",
			schemaName:     "def_wrap_c",
			expandAddrs: []addr.Property{
				addr.MustParseProperty("c"),
			},
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "nested_discriminator.json",
					Name:           "def_wrap_c",
				},
				Properties: map[string]*Property{
					"c/prop_b": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerNestedDiscriminator.Definitions["def_b"].Properties["prop_b"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPath + "#/definitions/def_wrap_c": struct{}{},
							specNestedDiscriminatorPath + "#/definitions/def_c":      struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_b")),
						},
						swaggerURI: specNestedDiscriminatorPath,
					},
					"c/prop_c": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerNestedDiscriminator.Definitions["def_c"].Properties["prop_c"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPath + "#/definitions/def_wrap_c": struct{}{},
							specNestedDiscriminatorPath + "#/definitions/def_c":      struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_c")),
						},
						swaggerURI: specNestedDiscriminatorPath,
					},
					"c/type": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerNestedDiscriminator.Definitions["def_a"].Properties["type"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPath + "#/definitions/def_wrap_c": struct{}{},
							specNestedDiscriminatorPath + "#/definitions/def_c":      struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_a")),
						},
						swaggerURI: specNestedDiscriminatorPath,
					},
				},
				swaggerURI: specNestedDiscriminatorPath,
			},
		},
		// 24
		{
			name:           `Expand "def_a" (schema is itself a discriminator)`,
			swaggerRelPath: "nested_discriminator.json",
			schemaName:     "def_a",
			expandAddrs: []addr.Property{
				addr.MustParseProperty("{def_b}"),
				addr.MustParseProperty("{def_c}"),
			},
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "nested_discriminator.json",
					Name:           "def_a",
				},
				Properties: map[string]*Property{
					"{def_b}/type": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerNestedDiscriminator.Definitions["def_a"].Properties["type"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPath + "#/definitions/def_b": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_a")),
						},
						swaggerURI: specNestedDiscriminatorPath,
					},
					"{def_b}/prop_b": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerNestedDiscriminator.Definitions["def_b"].Properties["prop_b"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPath + "#/definitions/def_b": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_b")),
						},
						swaggerURI: specNestedDiscriminatorPath,
					},
					"{def_c}/type": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerNestedDiscriminator.Definitions["def_a"].Properties["type"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPath + "#/definitions/def_c": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_a")),
						},
						swaggerURI: specNestedDiscriminatorPath,
					},
					"{def_c}/prop_b": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerNestedDiscriminator.Definitions["def_b"].Properties["prop_b"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPath + "#/definitions/def_c": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_b")),
						},
						swaggerURI: specNestedDiscriminatorPath,
					},
					"{def_c}/prop_c": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerNestedDiscriminator.Definitions["def_c"].Properties["prop_c"],
						resolvedRefs: map[string]interface{}{
							specNestedDiscriminatorPath + "#/definitions/def_c": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specNestedDiscriminatorPath, "def_c")),
						},
						swaggerURI: specNestedDiscriminatorPath,
					},
				},
				swaggerURI: specNestedDiscriminatorPath,
			},
		},
		// 25
		{
			name:           `Expand to ref to remote discriminator schema`,
			swaggerRelPath: "foo.json",
			schemaName:     "def_to_remote_discriminator",
			expandAddrs: []addr.Property{
				addr.MustParseProperty("properties"),
			},
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "foo.json",
					Name:           "def_to_remote_discriminator",
				},
				Properties: map[string]*Property{
					"properties{1}": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerRemoteDisc.Definitions["v1"],
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specRemoteDiscPath, "v1")),
						},
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_to_remote_discriminator": struct{}{},
							specRemoteDiscPath + "#/definitions/v1":                   struct{}{},
						},
						swaggerURI:          specRemoteDiscPath,
						isDiscriminator:     true,
						directVariantRefSet: map[string]struct{}{},
						rootVariantRefSet: map[string]struct{}{
							fmt.Sprintf("%s#/definitions/%s", specRemoteDiscPath, "v1"): {},
						},
					},
				},
				swaggerURI: specFooPath,
			},
		},
		// 26
		{
			name:           `Expand all of ref to remote discriminator schema`,
			swaggerRelPath: "foo.json",
			schemaName:     "def_to_remote_discriminator",
			expandAddrs: []addr.Property{
				addr.MustParseProperty("properties"),
				addr.MustParseProperty("properties{1}"),
			},
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "foo.json",
					Name:           "def_to_remote_discriminator",
				},
				Properties: map[string]*Property{
					"properties{1}/type": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerRemoteDisc.Definitions["base"].Properties["type"],
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specRemoteDiscPath, "base")),
						},
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_to_remote_discriminator": struct{}{},
							specRemoteDiscPath + "#/definitions/v1":                   struct{}{},
						},
						swaggerURI: specRemoteDiscPath,
					},
					"properties{1}/p1": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerRemoteDisc.Definitions["v1"].Properties["p1"],
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specRemoteDiscPath, "v1")),
						},
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_to_remote_discriminator": struct{}{},
							specRemoteDiscPath + "#/definitions/v1":                   struct{}{},
						},
						swaggerURI: specRemoteDiscPath,
					},
				},
				swaggerURI: specFooPath,
			},
		},
		// 27
		{
			name:           `Expand all properties that contains read-only property`,
			swaggerRelPath: "foo.json",
			schemaName:     "def_contain_ro",
			expandAddrs: []addr.Property{
				addr.MustParseProperty("p_obj_array"),
				addr.MustParseProperty("p_obj_array_ro"),
				addr.MustParseProperty("p_obj"),
				addr.MustParseProperty("p_obj_ro"),
			},
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "foo.json",
					Name:           "def_contain_ro",
				},
				Properties: map[string]*Property{
					"p_primitive": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["def_contain_ro"].Properties["p_primitive"],
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(specFooPath + "#/definitions/def_contain_ro"),
						},
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_contain_ro": struct{}{},
						},
						swaggerURI: specFooPath,
					},
					"p_primitive_ro": {
						PropertyState: PropertyStateIgnoredForReqROProp,
						schema:        swaggerFoo.Definitions["def_contain_ro"].Properties["p_primitive_ro"],
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(specFooPath + "#/definitions/def_contain_ro"),
						},
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_contain_ro": struct{}{},
						},
						swaggerURI: specFooPath,
					},
					"p_obj_array/foo": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["def_foo"].Properties["foo"],
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(specFooPath + "#/definitions/def_foo"),
						},
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_contain_ro": struct{}{},
							specFooPath + "#/definitions/def_foo":        struct{}{},
						},
						swaggerURI: specFooPath,
					},
					"p_obj_array_ro/foo": {
						PropertyState: PropertyStateIgnoredForReqROProp,
						schema:        swaggerFoo.Definitions["def_foo"].Properties["foo"],
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(specFooPath + "#/definitions/def_foo"),
						},
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_contain_ro": struct{}{},
							specFooPath + "#/definitions/def_foo":        struct{}{},
						},
						swaggerURI: specFooPath,
					},
					"p_obj/foo": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["def_foo"].Properties["foo"],
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(specFooPath + "#/definitions/def_foo"),
						},
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_contain_ro": struct{}{},
							specFooPath + "#/definitions/def_foo":        struct{}{},
						},
						swaggerURI: specFooPath,
					},
					"p_obj_ro/foo": {
						PropertyState: PropertyStateIgnoredForReqROProp,
						schema:        swaggerFoo.Definitions["def_foo"].Properties["foo"],
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(specFooPath + "#/definitions/def_foo"),
						},
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_contain_ro": struct{}{},
							specFooPath + "#/definitions/def_foo":        struct{}{},
						},
						swaggerURI: specFooPath,
					},
				},
				swaggerURI: specFooPath,
			},
		},
	}

	for idx, c := range cases {
		for i := 0; i <= c.repeat; i++ {
			swgschema, err := NewEmptySchema(specBasePath, c.swaggerRelPath, c.schemaName, c.forResponse, false)
			require.NoError(t, err, i)
			for _, addr := range c.expandAddrs {
				_, err := swgschema.expandPropertyOneLevelDeep(addr)
				require.NoError(t, err, c.name, idx)
			}
			require.Equal(t, c.expect, *swgschema, idx)
		}
	}
}

func TestSchema_AddCovProperty(t *testing.T) {
	cases := []struct {
		name           string
		swaggerRelPath string
		schemaName     string
		forResponse    bool
		addrs          []addr.Property
		expect         Schema
	}{
		// 0
		{
			name:           "Normal case",
			swaggerRelPath: "foo.json",
			schemaName:     "def_cov",
			addrs: []addr.Property{
				addr.MustParseProperty("p2/p2_1"),
			},
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "foo.json",
					Name:           "def_cov",
				},
				Properties: map[string]*Property{
					"p1": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["def_cov"].Properties["p1"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_cov": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_cov")),
						},
						swaggerURI: specFooPath,
					},
					"p2/p2_1": {
						PropertyState: PropertyStateCovered,
						schema:        swaggerFoo.Definitions["def_cov"].Properties["p2"].Properties["p2_1"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_cov": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_cov")),
						},
						swaggerURI: specFooPath,
					},
					"p2/p2_2": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["def_cov"].Properties["p2"].Properties["p2_2"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_cov": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_cov")),
						},
						swaggerURI: specFooPath,
					},
					"tags": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["def_cov"].Properties["tags"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_cov": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_cov")),
						},
						swaggerURI: specFooPath,
					},
				},
				swaggerURI: specFooPath,
			},
		},
		// 1
		{
			name:           "Add tf link to swagger non-leaf property will not be counted as covered",
			swaggerRelPath: "foo.json",
			schemaName:     "def_cov",
			addrs: []addr.Property{
				addr.MustParseProperty("p2"),
			},
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "foo.json",
					Name:           "def_cov",
				},
				Properties: map[string]*Property{
					"p1": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["def_cov"].Properties["p1"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_cov": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_cov")),
						},
						swaggerURI: specFooPath,
					},
					"p2/p2_1": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["def_cov"].Properties["p2"].Properties["p2_1"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_cov": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_cov")),
						},
						swaggerURI: specFooPath,
					},
					"p2/p2_2": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["def_cov"].Properties["p2"].Properties["p2_2"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_cov": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_cov")),
						},
						swaggerURI: specFooPath,
					},
					"tags": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["def_cov"].Properties["tags"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_cov": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_cov")),
						},
						swaggerURI: specFooPath,
					},
				},
				swaggerURI: specFooPath,
			},
		},
		// 2
		{
			name:           "Add tf link to model with Map/Dict properties should be counted as covered",
			swaggerRelPath: "foo.json",
			schemaName:     "def_cov",
			addrs: []addr.Property{
				addr.MustParseProperty("tags"),
			},
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "foo.json",
					Name:           "def_cov",
				},
				Properties: map[string]*Property{
					"p1": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["def_cov"].Properties["p1"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_cov": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_cov")),
						},
						swaggerURI: specFooPath,
					},
					"p2/p2_1": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["def_cov"].Properties["p2"].Properties["p2_1"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_cov": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_cov")),
						},
						swaggerURI: specFooPath,
					},
					"p2/p2_2": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["def_cov"].Properties["p2"].Properties["p2_2"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_cov": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_cov")),
						},
						swaggerURI: specFooPath,
					},
					"tags": {
						PropertyState: PropertyStateCovered,
						schema:        swaggerFoo.Definitions["def_cov"].Properties["tags"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_cov": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_cov")),
						},
						swaggerURI: specFooPath,
					},
				},
				swaggerURI: specFooPath,
			},
		},
		// 3
		{
			name:           "Add property of discriminator will cover",
			swaggerRelPath: "foo.json",
			schemaName:     "def_base",
			addrs: []addr.Property{
				addr.MustParseProperty("{def_variant1}/type"),
			},
			expect: Schema{
				SchemaMetaInfo: SchemaMetaInfo{
					SwaggerRelPath: "foo.json",
					Name:           "def_base",
				},
				Properties: map[string]*Property{
					"{def_variant1}/type": {
						PropertyState: PropertyStateCovered,
						schema:        swaggerFoo.Definitions["def_base"].Properties["type"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/def_variant1": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_base")),
						},
						swaggerURI: specFooPath,
					},
					"{def_variant2}/type": {
						PropertyState: PropertyStateUncovered,
						schema:        swaggerFoo.Definitions["def_base"].Properties["type"],
						resolvedRefs: map[string]interface{}{
							specFooPath + "#/definitions/variant2_def": struct{}{},
						},
						ownerSchemaRef: &openapispec.Ref{
							Ref: jsonreference.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", specFooPath, "def_base")),
						},
						swaggerURI: specFooPath,
					},
				},
				swaggerURI: specFooPath,
			},
		},
	}

	for idx, c := range cases {
		schema, err := NewEmptySchema(specBasePath, c.swaggerRelPath, c.schemaName, c.forResponse, false)
		require.NoError(t, err, idx)
		require.NoError(t, schema.ExpandProperties(), idx)
		if err == nil {
			for iidx, addr := range c.addrs {
				err := schema.AddCovProperty(addr, true)
				require.NoError(t, err, fmt.Sprintf("%s: %d.%d", c.name, idx, iidx))
				require.Equal(t, c.expect, *schema, fmt.Sprintf("%d.%d", idx, iidx))
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
		forResponse    bool
		process        process
		expect         string
	}{
		{
			swaggerRelPath: "foo.json",
			schemaName:     "def_regular",
			expect: fmt.Sprintf(`{
    "swagger_rel_path": "foo.json",
    "name": "def_regular",
	"active": true,
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
	"active": true,
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
	"active": true,
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
				require.NoError(t, schema.ExpandProperties())
				require.NoError(t, schema.AddCovProperty(swgPropAddr, true))
			},
			expect: fmt.Sprintf(`{
    "swagger_rel_path": "foo.json",
    "name": "def_cov",
	"active": true,
    "properties": {
		"p1": {
		  "state": "Uncovered"
		},
		"p2/p2_1": {
		  "state": "Covered"
		},
		"p2/p2_2": {
		  "state": "Uncovered"
		},
        "tags": {
          "state": "Uncovered"
        }
    }
}`),
		},
	}

	for idx, c := range cases {
		schema, err := NewEmptySchema(specBasePath, c.swaggerRelPath, c.schemaName, c.forResponse, true)
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
						PropertyState: PropertyStateUncovered,
					},
					"p2/p2_1": &Property{
						PropertyState: PropertyStateCovered,
					},
					"p2/p2_2": &Property{
						PropertyState: PropertyStateUncovered,
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
							PropertyState: PropertyStateCovered,
						},
						"{def_variant2}/type": {
							PropertyState: PropertyStateUncovered,
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
							PropertyState: PropertyStateUncovered,
						},
						"{def_variant2}/type": {
							PropertyState: PropertyStateCovered,
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
						PropertyState: PropertyStateCovered,
					},
					"{def_variant2}/type": {
						PropertyState: PropertyStateCovered,
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
	schemaMetaInfo := SchemaMetaInfo{
		SwaggerRelPath: specpath.Path("path"),
		Name:           "schema1",
	}
	var cases = []struct {
		schema       Schema
		grant        SchemaGrant
		expectSchema Schema
		hasError     bool
	}{
		// 0: schema is granted
		{
			schema: Schema{
				SchemaMetaInfo: schemaMetaInfo,
				Properties: map[string]*Property{
					"prop1": {},
				},
			},
			grant: SchemaGrant{
				Comment: "granted",
			},
			expectSchema: Schema{
				SchemaMetaInfo: schemaMetaInfo,
				IsGranted:      true,
				GrantComment:   "granted",
			},
		},
		// 1: property granted
		{
			schema: Schema{
				SchemaMetaInfo: schemaMetaInfo,
				Properties: map[string]*Property{
					"prop1":   {},
					"prop2":   {},
					"prop2_2": {},
					"prop3":   {},
				},
			},
			grant: SchemaGrant{
				Properties: map[string]string{
					"prop2_2": "granted",
					"prop3":   "granted",
				},
			},
			expectSchema: Schema{
				SchemaMetaInfo: schemaMetaInfo,
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
		// 2: grant unknown property cause error
		{
			schema: Schema{
				SchemaMetaInfo: schemaMetaInfo,
				Properties: map[string]*Property{
					"prop1":   {},
					"prop2":   {},
					"prop2_2": {},
					"prop3":   {},
				},
			},
			grant: SchemaGrant{
				Properties: map[string]string{
					"prop2_2":      "granted",
					"prop3":        "granted",
					"unknown_prop": "granted",
				},
			},
			hasError: true,
		},
		// 3: grant property of non-uncovered state cause error
		{
			schema: Schema{
				SchemaMetaInfo: schemaMetaInfo,
				Properties: map[string]*Property{
					"prop1": {
						PropertyState: NewPropertyState(PropStateCovered, ""),
					},
				},
			},
			grant: SchemaGrant{
				Properties: map[string]string{
					"prop1": "granted",
				},
			},
			hasError: true,
		},
	}

	for idx, c := range cases {
		err := c.schema.Grant(c.grant, true)
		if c.hasError {
			require.Error(t, err, idx)
			continue
		}
		require.NoError(t, err, idx)
		require.Equal(t, c.expectSchema, c.schema, idx)
	}
}
