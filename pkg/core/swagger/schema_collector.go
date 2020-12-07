package swagger

import openapispec "github.com/go-openapi/spec"

type SchemaCollector func(swagger *openapispec.Swagger) (schemaNames []string)

// CollectSchemas collects the schemas from a swagger spec
func CollectSchemas(swaggerBaseURL, swaggerRelPath string, collector SchemaCollector) ([]Schema, error) {
	swaggerURI := swaggerBaseURL + "/" + swaggerRelPath
	swagger, err := LoadSwagger(swaggerURI)
	if err != nil {
		return nil, err
	}
	schemas := collector(swagger)

	out := make([]Schema, 0, len(schemas))
	for _, schemaName := range schemas {
		schema, err := NewEmptySchema(swaggerBaseURL, swaggerRelPath, schemaName)
		if err != nil {
			return nil, err
		}
		out = append(out, *schema)
	}
	return out, nil
}
