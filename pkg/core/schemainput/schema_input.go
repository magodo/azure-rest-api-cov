package schemainput

import (
	"errors"
	"fmt"
	"path/filepath"

	openapispec "github.com/go-openapi/spec"
	"github.com/magodo/azure-rest-api-cov/pkg/core/addr"
	"github.com/magodo/azure-rest-api-cov/pkg/core/swagger"
	"github.com/magodo/azure-rest-api-cov/pkg/core/swaggerloader"
	"github.com/magodo/azure-rest-api-cov/pkg/core/swaggerlocator"
	"github.com/magodo/azure-rest-api-cov/pkg/core/workspace"
)

type OperationKind string

const (
	OperationKindPut   OperationKind = "PUT"
	OperationKindPost                = "POST"
	OperationKindPatch               = "PATCH"
)

type SchemaInput struct {
	swaggerlocator.SpecLocator
	Operation  OperationKind `json:"operation"`
	Variant    string        `json:"variant,omitempty"`
	Properties []string      `json:"properties"`
}

var (
	ErrFailtoLocateSwagger = errors.New("can't locate swagger spec")
)

func (schemaInput SchemaInput) ToSchema(swaggerRepoSpecBasePath string, index *workspace.AzureRepoIndex) (*swagger.Schema, error) {
	swaggerRelPath, ok := index.SpecLookupView()[schemaInput.SpecLocator]
	if !ok {
		return nil, fmt.Errorf("can't locate swagger spec by (%q, %q), this might because this tuple ends up with multiple swagger specs: %w",
			schemaInput.SpecLocator.ApiVersion, schemaInput.SpecLocator.ApiPath, ErrFailtoLocateSwagger)
	}
	swaggerPath := filepath.Join(swaggerRepoSpecBasePath, swaggerRelPath)

	// Looking for the schema name and the path defining this schema, based on the spec locator and the operation kind.
	swg, err := swaggerloader.LoadSwagger(swaggerPath)
	if err != nil {
		return nil, err
	}

	path := swg.Paths.Paths[schemaInput.ApiPath]
	params := []openapispec.Parameter{}
	switch schemaInput.Operation {
	case OperationKindPut:
		params = path.Put.Parameters
	case OperationKindPost:
		params = path.Post.Parameters
	case OperationKindPatch:
		params = path.Patch.Parameters
	default:
		return nil, fmt.Errorf("unknown operation kind: %q", schemaInput.Operation)
	}

	var param *openapispec.Parameter
	for _, p := range params {
		// In case the parameter itself is a ref, we will simply deref it for once.
		if p.Ref.String() != "" {
			pparam, err := openapispec.ResolveParameterWithBase(swg, p.Ref, &openapispec.ExpandOptions{RelativeBase: swaggerPath})
			if err != nil {
				return nil, err
			}
			p = *pparam
		}

		if p.In == "body" {
			param = &p
			break
		}
	}
	if param == nil {
		return nil, fmt.Errorf("no in body parameter found")
	}

	name, swaggerPath := swagger.SchemaNameFromRef(param.Schema.Ref)
	if swaggerPath != "" {
		if filepath.IsAbs(swaggerPath) {
			return nil, fmt.Errorf("un-expected absolute swagger path: %s", swaggerPath)
		}
		swaggerRelPath = filepath.Clean(filepath.Join(filepath.Dir(swaggerRelPath), swaggerPath))
	}

	schema, err := swagger.NewEmptySchema(swaggerRepoSpecBasePath, swaggerRelPath, name)
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
