package swagger

import (
	"github.com/magodo/azure-rest-api-cov/pkg/core/addr"
)

type SchemaInput struct {
	SpecLocator
	Name       string   `json:"name"`
	Variant    string   `json:"variant,omitempty"`
	Properties []string `json:"properties"`
}

func (schemaInput SchemaInput) ToSchema(swaggerRepoSpecBasePath string) (*Schema, error) {
	swaggerRelPath, err := LookupSpecPath(swaggerRepoSpecBasePath, SpecLocator{
		ApiVersion: schemaInput.ApiVersion,
		ApiPath:    schemaInput.ApiPath,
	})
	if err != nil {
		return nil, err
	}

	schema, err := NewEmptySchema(swaggerRepoSpecBasePath, swaggerRelPath, schemaInput.Name)
	if err != nil {
		return nil, err
	}

	rootProp := addr.MustParseProperty("").AsVariant(schemaInput.Variant)
	for _, prop := range schemaInput.Properties {
		if err := schema.AddCovProperty(rootProp.Append(addr.MustParseProperty(prop))); err != nil {
			return nil, err
		}
	}
	return schema, nil
}
