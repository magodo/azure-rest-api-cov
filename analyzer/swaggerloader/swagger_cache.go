package swaggerloader

import (
	"fmt"
	"path/filepath"
	"sync"

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

	swaggerCache.m[swaggerURI] = doc.Spec()
	return doc.Spec(), nil
}
