package swaggerloader

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/magodo/azure-rest-api-cov/analyzer/utils"

	"github.com/go-openapi/loads"
	openapispec "github.com/go-openapi/spec"
)

type SwaggerCache struct {
	sync.Mutex
	m map[string]*openapispec.Swagger
}

// swaggerCache caches the swagger document (spec) using swagger file absolute path as key.
var swaggerCache = SwaggerCache{
	Mutex: sync.Mutex{},
	m:     map[string]*openapispec.Swagger{},
}

// LoadSwagger load a certain swagger spec (document) from either file or http
func LoadSwagger(swaggerURI string) (*openapispec.Swagger, error) {
	var err error
	swaggerURI, err = filepath.Abs(swaggerURI)
	if err != nil {
		return nil, err
	}

	swaggerCache.Lock()
	defer swaggerCache.Unlock()

	// construct key
	if schema, ok := swaggerCache.m[swaggerURI]; ok {
		return schema, nil
	}

	doc, err := loads.Spec(swaggerURI)
	if err != nil {
		return nil, fmt.Errorf("loading swagger spec %s: %w", swaggerURI, err)
	}

	spec := doc.Spec()

	paths := map[string]openapispec.PathItem{}
	for k, v := range spec.Paths.Paths {
		paths[k] = v
	}

	// Normalize the API path
	if spec.Paths != nil {
		for p, item := range paths {
			delete(spec.Paths.Paths, p)
			normAPIPath := utils.NormalizeAPIPath(p)
			spec.Paths.Paths[normAPIPath] = item

			pathParams := item.PathItemProps.Parameters
			if pathParams == nil {
				// This is not technically correct since different operations might have different definition for the API path,
				// but is fine for Azure api for now.
				// This is an unfortunate easy hack to adopt Pandora SDK's API path mangling.
				if item.Get != nil && item.Get.Parameters != nil {
					pathParams = item.Get.Parameters
				} else if item.Put != nil && item.Put.Parameters != nil {
					pathParams = item.Put.Parameters
				} else if item.Patch != nil && item.Patch.Parameters != nil {
					pathParams = item.Patch.Parameters
				} else if item.Delete != nil && item.Delete.Parameters != nil {
					pathParams = item.Delete.Parameters
				} else if item.Post != nil && item.Post.Parameters != nil {
					pathParams = item.Post.Parameters
				}
			}
			if pathParams != nil {
				specialAPIPath, err := utils.SpecializeAPIPath(p, spec, swaggerURI, pathParams)
				if err != nil {
					return nil, fmt.Errorf("specializing api path %s in %s", p, swaggerURI)
				}
				if specialAPIPath != normAPIPath {
					spec.Paths.Paths[specialAPIPath] = item
				}
			}
		}
	}

	swaggerCache.m[swaggerURI] = doc.Spec()
	return doc.Spec(), nil
}
