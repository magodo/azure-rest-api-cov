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

func (schemaInput SchemaInput) ToSchema(index *workspace.AzureRepoIndex, strict bool) (*swagger.Schema, error) {
	name, relPath, err := index.LocateSchema(schemaInput.SchemaLocator)
	if err != nil {
		return nil, err
	}

	schema, err := swagger.ProcessCoverageForSchema(index.SpecRootDir, string(relPath), name, schemaInput.Variant, schemaInput.Properties, schemaInput.Operation == workspace.OperationKindGet, strict)
	if err != nil {
		return nil, err
	}
	return schema, nil
}
