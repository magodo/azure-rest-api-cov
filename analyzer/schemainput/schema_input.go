package schemainput

import (
	"github.com/magodo/azure-rest-api-cov/analyzer/swagger"
	"github.com/magodo/azure-rest-api-cov/analyzer/workspace"
)

type SchemaInput struct {
	workspace.SchemaLocator
	Variant    string   `json:"variant,omitempty"`
	Properties []string `json:"properties"`
}

func (schemaInput SchemaInput) ToSchema(index *workspace.AzureRepoIndex, strict, active bool) (*swagger.Schema, error) {
	info, err := index.LocateSchema(schemaInput.SchemaLocator)
	if err != nil {
		return nil, err
	}

	// For schema-less operations, e.g. DELETE, just return nil.
	if info.Name == "" {
		return &swagger.Schema{
			SchemaMetaInfo: swagger.SchemaMetaInfo{
				Name:           info.Name,
				SwaggerRelPath: info.SpecRelPath,
				Active:         active,
			},
		}, nil
	}

	name, relPath := info.Name, info.SpecRelPath
	schema, err := swagger.ProcessCoverageForSchema(index.SpecRootDir, string(relPath), name, schemaInput.Variant, schemaInput.Properties, schemaInput.Operation == workspace.OperationKindGet, strict, active)
	if err != nil {
		return nil, err
	}
	return schema, nil
}
