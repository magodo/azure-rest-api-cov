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

	for _, propStr := range schemaInput.Properties {
		prop := addr.MustParseProperty(propStr)
		if schemaInput.Variant != "" {
			prop = addr.MustParseProperty("").AsVariant(schemaInput.Variant).Append(prop)
		}
		if err := schema.AddCovProperty(prop); err != nil {
			return nil, err
		}
	}
	return schema, nil
}
