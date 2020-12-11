package schemainput

import (
	"github.com/magodo/azure-rest-api-cov/pkg/core/addr"
	"github.com/magodo/azure-rest-api-cov/pkg/core/swagger"
	"github.com/magodo/azure-rest-api-cov/pkg/core/swaggerlocator"
)

type SchemaInput struct {
	swaggerlocator.SpecLocator
	Name       string   `json:"name"`
	Variant    string   `json:"variant,omitempty"`
	Properties []string `json:"properties"`
}

func (schemaInput SchemaInput) ToSchema(swaggerRepoSpecBasePath string, swaggerRelPath string) (*swagger.Schema, error) {
	schema, err := swagger.NewEmptySchema(swaggerRepoSpecBasePath, swaggerRelPath, schemaInput.Name)
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
