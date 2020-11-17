package swagger

import (
	"fmt"
	openapispec "github.com/go-openapi/spec"
	"github.com/magodo/azure-rest-api-cov/pkg/core/addr"
)

const swaggerExtensionMSDiscriminatorValue = "x-ms-discriminator-value"

type Schema struct {
	SwaggerRelPath string
	Name           string
	Properties     Properties

	// Whether this property is granted to be not to implement in Terraform
	IsGranted    bool   `json:",omitempty"`
	GrantComment string `json:",omitempty"`

	swagger       *openapispec.Swagger
	swaggerURI    string
}

func NewSchema(swaggerBaseURL, swaggerRelPath string, schemaName string) (*Schema, error) {
	swaggerURI := swaggerBaseURL + "/" + swaggerRelPath
	swagger, err := LoadSwagger(swaggerURI)
	if err != nil {
		return nil, err
	}

	schema, ok := swagger.Definitions[schemaName]
	if !ok {
		return nil, fmt.Errorf("schema %q is not found in swagger spec %q", schemaName, swaggerURI)
	}

	swgSchema := &Schema{
		SwaggerRelPath: swaggerRelPath,
		Name:           schemaName,
		Properties: map[string]*Property{
			"": {
				PropertyState: NewPropertyState(PropStateUncovered, ""),
				schema:        schema,
				swaggerURI:    swaggerURI,
			},
		},
		swaggerURI: swaggerURI,
		swagger:    swagger,
	}

	// Consider this schemas itself as resolved reference only when it is not a discriminator schema
	if schema.Discriminator == "" {
		swgSchema.Properties[""].resolvedRefs = map[string]interface{}{
			normalizePaths("#/definitions/"+schemaName, swaggerURI): struct{}{},
		}
	}

	// Expand the root level properties of the schemas
	rootAddr := addr.NewProperty(schemaName, addr.RelProperty{})
	err = swgSchema.ExpandPropertyOneLevelDeep(rootAddr)
	if err != nil {
		return nil, fmt.Errorf("expanding schemas %s (%s): %w", schemaName, swaggerURI, err)
	}
	return swgSchema, nil
}

// ExpandPropertyOneLevelDeep expand the specified swagger schemas property one level deep, with any allOf and $ref taken into consideration.
func (s *Schema) ExpandPropertyOneLevelDeep(propAddr addr.Property) error {
	raddr := propAddr.Property.String()

	defer func() {
		// We have to check whether we added any child property of this property.
		// If yes, then we need to remove this property from the Schema property map.
		for currentRAddr := range s.Properties {
			currentAddr := addr.NewProperty(s.Name, addr.MustParseRelProperty(currentRAddr))
			if propAddr.Contains(currentAddr) {
				delete(s.Properties, raddr)
				return
			}
		}
	}()

	prop, ok := s.Properties[raddr]
	if !ok {
		return fmt.Errorf("property %s does not exist in Schema %s (%s)", propAddr, s.Name, s.swaggerURI)
	}

	isCyclic, err := s.expandRefPropertyInPlace(prop)
	if err != nil {
		return fmt.Errorf("dereferencing property %s in Schema %s (%s): %w", propAddr, s.Name, s.swaggerURI, err)
	}

	// If the property to be expanded is a cyclic reference, we will do nothing but keep that property
	if isCyclic {
		return nil
	}

	// temporarily store the Properties expanded in this level
	levelProperties := NewProperties()

	// expand direct top level properties
	directTopSWGProperties := s.expandSubProperties(propAddr, prop)

	// expand AllOf properties
	allOfSWGProperties, err := s.expandAllOfProperties(propAddr, prop)
	if err != nil {
		return fmt.Errorf("failed to expand allOf properties for %q: %v", propAddr.String(), err)
	}

	levelProperties.Merge(directTopSWGProperties)
	levelProperties.Merge(allOfSWGProperties)

	discriminator := prop.schema.Discriminator
	if discriminator == "" {
		s.addProperties(levelProperties)
		return nil
	}

	// If the property to be expanded is a discriminator, we will expand it into its variants
	levelNamedProperties := map[string]openapispec.Schema{}
	for raddrStr, prop := range levelProperties {
		raddr := addr.MustParseRelProperty(raddrStr)
		propName := raddr[len(raddr)-1].String()
		levelNamedProperties[propName] = prop.schema
	}

	if discriminatorProp := levelNamedProperties[discriminator]; discriminatorProp.Enum != nil {
		// TODO: optimize to using map
		for _, variantRaw := range discriminatorProp.Enum {
			variant, ok := variantRaw.(string)
			if !ok {
				panic(fmt.Sprintf("failed to find variant dscSchema who implements discriminator %q in %q", discriminator, propAddr.String()))
			}
			var dscSchemaName string

			for name, schema := range s.swagger.Definitions {
				if v, ok := schema.Extensions[swaggerExtensionMSDiscriminatorValue].(string); ok && v == variant {
					dscSchemaName = name
					break
				}
			}

			// For some malformed swagger, the 'x-ms-discriminator-value' is not defined for the variant schemas,
			// then we will fallback to try a schema name match
			if dscSchemaName == "" {
				if _, ok := s.swagger.Definitions[variant]; ok {
					dscSchemaName = variant
				}
			}

			if dscSchemaName == "" {
				return fmt.Errorf("variant schema with discriminator set to %q is not found", variant)
			}

			// Since we removed the discriminator base schema from the "resolvedRefs" when we creat the Property, we should in turn add the
			// exact variant schema expanded to the "resolvedRefs".
			resolvedRefs := map[string]interface{}{}
			for k, v := range prop.resolvedRefs {
				resolvedRefs[k] = v
			}
			resolvedRefs[normalizePaths("#/definitions/"+dscSchemaName, s.swaggerURI)] = struct{}{}

			p := NewProperty(prop.PropertyState, s.swagger.Definitions[dscSchemaName], resolvedRefs, prop.swaggerURI)
			s.addProperty(propAddr.AsVariant(variant), *p)
		}
	}

	return nil
}

// expandSubProperties expand the directly contained sub-properties of the property (prop) in the specified address (addr)
// Especially, if the property is an array to object, it will expand to the sub-properties of the object item instead.
func (s *Schema) expandSubProperties(propAddr addr.Property, prop *Property) Properties {
	output := NewProperties()
	var properties openapispec.SchemaProperties
	if prop.schema.Items != nil {
		properties = prop.schema.Items.Schema.Properties
	} else {
		properties = prop.schema.Properties
	}
	for propName, propSchema := range properties {
		p := NewProperty(prop.PropertyState, propSchema, prop.resolvedRefs, prop.swaggerURI)
		propAddr := propAddr.Append(addr.MustParseRelProperty(propName))
		output[propAddr.Property.String()] = p
	}
	return output
}

// expandAllOfProperties recursively expand containing "allOf" properties for property (prop) in the specified address (addr)
func (s *Schema) expandAllOfProperties(propAddr addr.Property, prop *Property) (Properties, error) {
	output := NewProperties()
	for _, schema := range prop.schema.AllOf {

		// We construct a temp Property here (as it has no object/property related) to expand it into a concrete schemas.
		// Then we will iterate that schemas's property which by concept is the top level property of this parent property.
		tmpProp := NewProperty(prop.PropertyState, schema, prop.resolvedRefs, prop.swaggerURI)

		// AllOf contains refs, then need to expand the reference properties first.
		if tmpProp.schema.Ref.String() != "" {
			isCyclic, err := s.expandRefPropertyInPlace(tmpProp)
			if err != nil {
				return nil, fmt.Errorf("dereferencing property %s in Schema %s (%s): %w", propAddr, s.Name, s.swaggerURI, err)
			}
			// Ignore... as there is no better way to handle this (since it has no object/property related, or shall we panic?)
			if isCyclic {
				continue
			}
		}

		// Each allOf property itself might contains allOf properties and direct properties, here we recursively retrieve those
		// nested allOf/directed properties.
		nestedAllOfProperties, err := s.expandAllOfProperties(propAddr, tmpProp)
		if err != nil {
			return nil, fmt.Errorf("failed to expand nested allOf Properties for %q: %v", propAddr.String(), err)
		}
		output.Merge(nestedAllOfProperties)
		output.Merge(s.expandSubProperties(propAddr, tmpProp))
	}

	return output, nil
}

// expandRefPropertyInPlace expand a property itself IN-PLACE until either it is a concrete schemas (i.e. not a ref) or hit a cyclic ref.
func (s *Schema) expandRefPropertyInPlace(prop *Property) (isCyclic bool, err error) {
	ref := prop.schema.Ref
	if ref.String() == "" {
		// Specially, if current schema is an array and the items is a ref, we need to go on expand it.
		if prop.schema.Items == nil {
			return false, nil
		}

		if prop.schema.Items.Schema == nil || len(prop.schema.Items.Schemas) != 0 {
			return false, nil
		}

		var schema openapispec.Schema
		if prop.schema.Items.Schema != nil {
			schema = *prop.schema.Items.Schema
		} else {
			schema = prop.schema.Items.Schemas[0]
		}
		if schema.Ref.String() == "" {
			return false, nil
		}
		// continue expanding the ref of the array item
		ref = schema.Ref
	}

	normalizedRef := normalizeFileRef(&ref, prop.swaggerURI)
	normalizedRefURI := normalizedRef.String()

	// If current ref has already been derefed, meaning a cyclic ref is hit, we will return.
	if _, ok := prop.resolvedRefs[normalizedRefURI]; ok {
		return true, nil
	}

	swagger, err := LoadSwagger(prop.swaggerURI)
	if err != nil {
		return false, err
	}
	schema, err := openapispec.ResolveRefWithBase(swagger, &ref, &openapispec.ExpandOptions{RelativeBase: prop.swaggerURI})
	if err != nil {
		return false, fmt.Errorf("resolve reference %s: %w", ref.String(), err)
	}

	// Keep track of the resolved reference to avoid cyclic ref.
	// Specifically, we need to avoid track the discriminator base schemas as they will not be exactly referenced
	// (only its variants are referenced). This allows us to reference the base schema's properties in the
	// `allOf` from the variant's schema.
	if schema.Discriminator == "" {
		prop.resolvedRefs[normalizedRefURI] = struct{}{}
	}

	// update the stored schemas by the derefed schemas
	prop.schema = *schema

	// update the swaggerURI to using the referenced swagger
	prop.swaggerURI = normalizedRef.GetURL().Path

	return s.expandRefPropertyInPlace(prop)
}

// addProperty adds a new Property to the Schema.
func (s *Schema) addProperty(propAddr addr.Property, prop Property) {
	s.Properties[propAddr.Property.String()] = &prop
}

// addProperties merges a new Properties to the Schema's Properties.
func (s *Schema) addProperties(props Properties) {
	s.Properties.Merge(props)
}

// AddCovProperty adds a covered property to the schema.
func (s *Schema) AddCovProperty(propAddr addr.Property) error {
	var isExpandToChildProperties bool
	for existRAddr, prop := range s.Properties {
		existPropAddr := addr.NewProperty(s.Name, addr.MustParseRelProperty(existRAddr))

		if !propAddr.Contains(existPropAddr) && !existPropAddr.Contains(propAddr) && !existPropAddr.Equals(propAddr) {
			continue
		}

		// If the property to add is a parent property of any existing properties.
		// Then we will mark the existing properties as covered also.
		if propAddr.Contains(existPropAddr) {
			isExpandToChildProperties = true
			prop.PropertyState = NewPropertyState(PropStateCovered, "")
			continue
		}

		if existPropAddr.Equals(propAddr) {
			prop.PropertyState = NewPropertyState(PropStateCovered, "")
			return nil
		}

		// The property we're seeking is a member of the property under iteration, we will first expand it one level deep.
		if err := s.ExpandPropertyOneLevelDeep(existPropAddr); err != nil {
			return fmt.Errorf("expanding top level property for %s: %w", existPropAddr, err)
		}
		return s.AddCovProperty(propAddr)
	}
	if isExpandToChildProperties {
		return nil
	}
	return fmt.Errorf("property %s doesn't belong to schemas %s (%s)", propAddr, s.Name, s.swaggerURI)
}

// CalcCoverage calculates the property coverage (<=1) of the schema/property.
// Those granted properties are not counted during the calculation.
func (s *Schema) CalcCoverage() (*PropertyCoverageStore, error) {
	store := NewPropertyCoverageStore()
	for propRAddr, prop := range s.Properties {
		if err := store.Add(addr.MustParseRelProperty(propRAddr), *prop); err != nil {
			return nil, fmt.Errorf("adding property %q: %v", propRAddr, err)
		}
	}
	return &store, nil
}
