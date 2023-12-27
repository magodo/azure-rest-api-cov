package schemainput

import (
	"encoding/json"
	"fmt"

	"github.com/magodo/azure-rest-api-cov/analyzer/operation"
	"github.com/magodo/azure-rest-api-cov/analyzer/propmatch"
	"github.com/magodo/azure-rest-api-cov/analyzer/swagger"
	"github.com/magodo/azure-rest-api-cov/analyzer/workspace"
)

type SchemaInput struct {
	workspace.OperationLocator
	Variant    string                `json:"variant,omitempty"`
	Properties []SchemaInputProperty `json:"properties"`
}

type SchemaInputProperty struct {
	Addr            string           `json:"addr"`
	AppPropertyMaps []AppPropertyMap `json:"app_property_maps"`
}

func (prop SchemaInputProperty) ToSwaggerInputProperty() swagger.InputProperty {
	var appPropMap []swagger.AppPropertyMap
	for _, m := range prop.AppPropertyMaps {
		appPropMap = append(appPropMap, swagger.AppPropertyMap{
			Name: m.Name,
			Addr: m.Addr,
		})
	}
	return swagger.InputProperty{
		Addr:            prop.Addr,
		AppPropertyMaps: appPropMap,
	}
}

func FromSwaggerProperty(addr string, prop swagger.Property) SchemaInputProperty {
	var appPropMap []AppPropertyMap
	for _, m := range prop.AppPropertyMaps {
		appPropMap = append(appPropMap, AppPropertyMap{
			Name: m.Name,
			Addr: m.Addr,
		})
	}
	return SchemaInputProperty{
		Addr:            addr,
		AppPropertyMaps: appPropMap,
	}
}

type AppPropertyMap struct {
	Name string `json:"name"`
	Addr string `json:"addr"`
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
		APIPath:        info.APIPath,
		APIVersion:     opLoc.ApiVersion,
		IsPreview:      pathInfo.IsPreview,
		IsActive:       active,
		Operation:      opLoc.Operation,
		SwaggerRelPath: info.SpecRelPath,
	}

	var inputProps []swagger.InputProperty
	for _, prop := range schemaInput.Properties {
		inputProps = append(inputProps, prop.ToSwaggerInputProperty())
	}

	if info.Model != nil {
		name, relPath := info.Model.Name, info.Model.SpecRelPath
		schema, err := swagger.ProcessCoverageForSchema(index.SpecRootDir, relPath, name, schemaInput.Variant, inputProps, schemaInput.Operation == workspace.OperationKindGet, strict, allowMismatchPropMatchers)
		if err != nil {
			return nil, err
		}
		out.Model = schema
	}

	return out, nil
}

type SchemaInputPropertyForEncode SchemaInputProperty

func (input SchemaInputProperty) MarshalJSON() ([]byte, error) {
	if len(input.AppPropertyMaps) == 0 {
		return json.Marshal(input.Addr)
	}
	return json.Marshal(SchemaInputPropertyForEncode(input))
}

func (input *SchemaInputProperty) UnmarshalJSON(b []byte) error {
	var ie SchemaInputPropertyForEncode
	if err := json.Unmarshal(b, &ie); err == nil {
		*input = SchemaInputProperty{
			Addr:            ie.Addr,
			AppPropertyMaps: ie.AppPropertyMaps,
		}
		return nil
	}
	return json.Unmarshal(b, &input.Addr)
}
