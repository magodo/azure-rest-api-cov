package schemainput

import (
	"github.com/magodo/azure-rest-api-cov/pkg/core/addr"
	"github.com/magodo/azure-rest-api-cov/pkg/core/swagger"
	"github.com/magodo/azure-rest-api-cov/pkg/core/workspace"
)

type SchemaInput struct {
	workspace.SchemaLocator
	Variant    string   `json:"variant,omitempty"`
	Properties []string `json:"properties"`
}

func (schemaInput SchemaInput) ToSchema(index *workspace.AzureRepoIndex) (*swagger.Schema, error) {
	name, relPath, err := index.LocateSchema(schemaInput.SchemaLocator)
	if err != nil {
		return nil, err
	}

	schema, err := swagger.NewEmptySchema(index.SpecRootDir, relPath, name)
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
