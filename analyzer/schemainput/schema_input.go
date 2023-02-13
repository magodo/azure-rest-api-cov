package schemainput

import (
	"fmt"

	"github.com/magodo/azure-rest-api-cov/analyzer/operation"
	"github.com/magodo/azure-rest-api-cov/analyzer/propmatch"
	"github.com/magodo/azure-rest-api-cov/analyzer/swagger"
	"github.com/magodo/azure-rest-api-cov/analyzer/workspace"
)

type SchemaInput struct {
	workspace.OperationLocator
	Variant    string   `json:"variant,omitempty"`
	Properties []string `json:"properties"`
}

func (schemaInput SchemaInput) ToOperationInfo(index *workspace.AzureRepoIndex, strict, active bool, allowMismatchPropMatchers []propmatch.PropMatcher) (*operation.Info, error) {
	opLoc := schemaInput.OperationLocator
	info, err := index.LocateOperation(opLoc)
	if err != nil {
		return nil, err
	}

	pathInfo, err := info.SpecRelPath.PathInfo()
	if err != nil {
		return nil, fmt.Errorf("getting path info for %q: %v", info.SpecRelPath, err)
	}

	out := &operation.Info{
		RP:             pathInfo.ResourceProvider,
		APIPath:        opLoc.ApiPath,
		APIVersion:     opLoc.ApiVersion,
		IsPreview:      pathInfo.IsPreview,
		IsActive:       active,
		Operation:      opLoc.Operation,
		SwaggerRelPath: info.SpecRelPath,
	}

	if info.Model != nil {
		name, relPath := info.Model.Name, info.Model.SpecRelPath
		schema, err := swagger.ProcessCoverageForSchema(index.SpecRootDir, relPath, name, schemaInput.Variant, schemaInput.Properties, schemaInput.Operation == workspace.OperationKindGet, strict, allowMismatchPropMatchers)
		if err != nil {
			return nil, err
		}
		out.Model = schema
	}

	return out, nil
}
