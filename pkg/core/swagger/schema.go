package swagger

import (
	"fmt"
	"strings"

	"github.com/go-openapi/jsonreference"

	"github.com/magodo/azure-rest-api-cov/pkg/core/swaggerloader"

	openapispec "github.com/go-openapi/spec"
	"github.com/magodo/azure-rest-api-cov/pkg/core/addr"
)

const swaggerExtensionMSDiscriminatorValue = "x-ms-discriminator-value"

const rootProp = ""

type SchemaMetaInfo struct {
	Name           string `json:"name"`
	SwaggerRelPath string `json:"swagger_rel_path"`
}

// Ref returns a hypothetical schema reference. It is mainly used as a unique ID of the schema, to be used in log message or map key.
func (i SchemaMetaInfo) Ref() string {
	return i.SwaggerRelPath + "#/definitions/" + i.Name
}

type Schema struct {
	SchemaMetaInfo
	Properties Properties `json:"properties"`

	// Whether this property is granted to be not to implement in Terraform
	IsGranted    bool   `json:",omitempty"`
	GrantComment string `json:",omitempty"`

	swagger    *openapispec.Swagger
	swaggerURI string
}

func NewEmptySchema(swaggerBaseURL, swaggerRelPath string, schemaName string) (*Schema, error) {
	swaggerURI := swaggerBaseURL + "/" + swaggerRelPath
	swagger, err := swaggerloader.LoadSwagger(swaggerURI)
	if err != nil {
		return nil, err
	}

	schema, ok := swagger.Definitions[schemaName]
	if !ok {
		return nil, fmt.Errorf("schema %q is not found in swagger spec %q", schemaName, swaggerURI)
	}

	ref := SchemaNameToRef(schemaName, "")
	normalizedSchemaRef := normalizePaths(ref.String(), swaggerURI)

	swgSchema := &Schema{
		SchemaMetaInfo: SchemaMetaInfo{
			Name:           schemaName,
			SwaggerRelPath: swaggerRelPath,
		},
		swaggerURI: swaggerURI,
		swagger:    swagger,
	}

	rootProperty, err := buildProperty(
		NewPropertyState(PropStateUncovered, ""),
		swaggerURI,
		schema,
		map[string]interface{}{},
		&openapispec.Ref{
			Ref: jsonreference.MustCreateRef(normalizedSchemaRef),
		})
	if err != nil {
		return nil, err
	}

	swgSchema.Properties = map[string]*Property{
		rootProp: rootProperty,
	}

	allAllOfs, err := swaggerAllAllOfs(swaggerURI)
	if err != nil {
		return nil, err
	}
	isDiscriminator, _, err := rootProperty.checkDiscriminator(swaggerURI, allAllOfs)
	if err != nil {
		return nil, err
	}

	// Consider this schemas itself as resolved reference only when it is not a discriminator schema
	if !isDiscriminator {
		swgSchema.Properties[""].resolvedRefs = map[string]interface{}{
			normalizedSchemaRef: struct{}{},
		}
	}

	// Expand the root level properties of the schemas
	rootAddr := addr.Property{}
	err = swgSchema.expandPropertyOneLevelDeep(rootAddr)
	if err != nil {
		return nil, fmt.Errorf("expanding schemas %s (%s): %w", schemaName, swaggerURI, err)
	}
	return swgSchema, nil
}

// expandPropertyOneLevelDeep expand the specified swagger schemas property one level deep, with any allOf and $ref taken into consideration.
func (s *Schema) expandPropertyOneLevelDeep(propAddr addr.Property) error {
	raddr := propAddr.String()

	defer func() {
		// We have to check whether we added any child property of this property.
		// If yes, then we need to remove this property from the Schema property map.
		for currentRAddr := range s.Properties {
			currentAddr := addr.MustParseProperty(currentRAddr)
			if propAddr.Contains(currentAddr) {
				delete(s.Properties, raddr)
			}
		}
	}()

	prop, ok := s.Properties[raddr]
	if !ok {
		return fmt.Errorf("property %s does not exist in Schema %s (%s)", propAddr, s.Name, s.swaggerURI)
	}

	isCyclic, err := prop.expandRefPropertyInPlace()
	if err != nil {
		return fmt.Errorf("dereferencing property %s in Schema %s (%s): %w", propAddr, s.Name, s.swaggerURI, err)
	}

	// If the property to be expanded is a cyclic reference, we will do nothing but keep that property
	if isCyclic {
		return nil
	}

	// temporarily store the Properties expanded in this level
	levelProperties := Properties{}

	// expand direct top level properties
	directTopSWGProperties, err := s.expandSubProperties(propAddr, prop)
	if err != nil {
		return fmt.Errorf("failed to expand direct top level properties for %q: %v", propAddr.String(), err)
	}

	// expand AllOf properties
	allOfSWGProperties, err := s.expandAllOfProperties(propAddr, prop)
	if err != nil {
		return fmt.Errorf("failed to expand allOf properties for %q: %v", propAddr.String(), err)
	}

	levelProperties.Merge(directTopSWGProperties)
	levelProperties.Merge(allOfSWGProperties)

	// For non-pure discriminator, because itself can be instantiated, so we will add the expanded properties also.
	if !prop.isPureDiscriminator {
		s.addProperties(levelProperties)
	}

	// In case current property is a discriminator schema, we will expand it into its variants/children.
	// Otherwise, we will simply return here and do not doing the expanding to variants below.
	if !prop.isDiscriminator {
		return nil
	}

	_, resolved := prop.resolvedRefs[prop.fromRef.String()]
	if resolved {
		return nil
	}

	for variantRef := range prop.childRefSet {
		ref := openapispec.MustCreateRef(variantRef)
		name, _ := SchemaInfoFromRef(ref)

		descValue := name
		// If the 'x-ms-discriminator-value' is defined for the variant schemas, then it will take priority.
		if v, ok := s.swagger.Definitions[name].Extensions[swaggerExtensionMSDiscriminatorValue].(string); ok {
			descValue = v
		}

		// Since we removed the discriminator base schema from the "resolvedRefs" when we creat the Property, we should in turn Add the
		// exact variant schema expanded to the "resolvedRefs".
		resolvedRefs := map[string]interface{}{}
		for k, v := range prop.resolvedRefs {
			resolvedRefs[k] = v
		}
		normalizedRef := openapispec.Ref{
			Ref: jsonreference.MustCreateRef(normalizePaths(variantRef, prop.swaggerURI)),
		}
		resolvedRefs[normalizedRef.String()] = struct{}{}

		p, err := buildProperty(prop.PropertyState, prop.swaggerURI, s.swagger.Definitions[name], resolvedRefs, &normalizedRef)
		if err != nil {
			return err
		}
		s.addProperty(propAddr.AsVariant(descValue), *p)
	}

	return nil
}

// Construct all the "allOf"s of the schemas in current spec.
// {normRef: {normRef1, normRef2,...}, ...}
func swaggerAllAllOfs(swaggerURI string) (map[string]map[string]struct{}, error) {
	swagger, err := swaggerloader.LoadSwagger(swaggerURI)
	if err != nil {
		return nil, err
	}
	allAllOfs := map[string]map[string]struct{}{}
	for name, schema := range swagger.Definitions {
		if len(schema.AllOf) == 0 {
			continue
		}
		ref := SchemaNameToRef(name, "")
		normRef := normalizeFileRef(&ref, swaggerURI).String()
		allOfRefs := map[string]struct{}{}
		allAllOfs[normRef] = allOfRefs

		for _, allOf := range schema.AllOf {
			if ref := allOf.Ref; ref.String() != "" {
				normalizedRef := normalizeFileRef(&ref, swaggerURI)
				allOfRefs[normalizedRef.String()] = struct{}{}
			}
		}
	}
	return allAllOfs, nil
}

// expandSubProperties expand the directly contained sub-properties of the property (prop) in the specified address (addr)
// Especially, if the property is an array to object, it will expand to the sub-properties of the object item instead.
func (s *Schema) expandSubProperties(propAddr addr.Property, prop *Property) (Properties, error) {
	output := Properties{}
	var properties openapispec.SchemaProperties
	if prop.schema.Items != nil {
		properties = prop.schema.Items.Schema.Properties
	} else {
		properties = prop.schema.Properties
	}
	for propName, propSchema := range properties {
		p, err := buildProperty(prop.PropertyState, prop.swaggerURI, propSchema, prop.resolvedRefs, prop.fromRef)
		if err != nil {
			return nil, err
		}
		propAddr := propAddr.Append(addr.MustParseProperty(propName))
		output[propAddr.String()] = p
	}
	return output, nil
}

// expandAllOfProperties recursively expand containing "allOf" properties for property (prop) in the specified address (addr)
func (s *Schema) expandAllOfProperties(propAddr addr.Property, prop *Property) (Properties, error) {
	output := Properties{}
	for _, schema := range prop.schema.AllOf {

		// We construct a temp Property here (as it has no object/property related) to expand it into a concrete schemas.
		// Then we will iterate that schemas's property which by concept is the top level property of this parent property.
		tmpProp, err := buildProperty(prop.PropertyState, prop.swaggerURI, schema, prop.resolvedRefs, prop.fromRef)
		if err != nil {
			return nil, err
		}

		// AllOf contains refs, then need to expand the reference properties first.
		if tmpProp.schema.Ref.String() != "" {
			isCyclic, err := tmpProp.expandRefPropertyInPlace()
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
		subProps, err := s.expandSubProperties(propAddr, tmpProp)
		if err != nil {
			return nil, err
		}
		output.Merge(subProps)
	}

	return output, nil
}

// addProperty adds a new Property to the Schema.
func (s *Schema) addProperty(propAddr addr.Property, prop Property) {
	s.Properties[propAddr.String()] = &prop
}

// addProperties merges a new Properties to the Schema's Properties.
func (s *Schema) addProperties(props Properties) {
	s.Properties.Merge(props)
}

func (s Schema) HasGrantedAny() bool {
	if s.IsGranted {
		return true
	}
	for _, prop := range s.Properties {
		if prop.IsGranted() {
			return true
		}
	}
	return false
}

// AddCovProperty adds a covered property to the schema.
// NOTE: This has to be called before the schema goes under granting process.
func (s *Schema) AddCovProperty(propAddr addr.Property) error {
	if s.HasGrantedAny() {
		return fmt.Errorf("AddCovProperty() has to be called on schema (%q) before granting", s.Ref())
	}
	var propagateToChild bool
	for existRAddr, prop := range s.Properties {
		existPropAddr := addr.MustParseProperty(existRAddr)

		if !propAddr.Contains(existPropAddr) && !existPropAddr.Contains(propAddr) && !existPropAddr.Equals(propAddr) {
			continue
		}

		// If the property to Add is a parent property of any existing properties.
		// Then we will mark the existing properties as covered also.
		if propAddr.Contains(existPropAddr) {
			propagateToChild = true
			prop.PropertyState = NewPropertyState(PropStateCovered, "")
			continue
		}

		if existPropAddr.Equals(propAddr) {
			prop.PropertyState = NewPropertyState(PropStateCovered, "")
			return nil
		}

		// The property we're seeking is a member of the property under iteration, we will first expand it one level deep.
		if err := s.expandPropertyOneLevelDeep(existPropAddr); err != nil {
			return fmt.Errorf("expanding top level property for %s: %w", existPropAddr, err)
		}
		return s.AddCovProperty(propAddr)
	}
	if propagateToChild {
		return nil
	}
	return fmt.Errorf("property %s doesn't belong to schemas %s (%s)", propAddr, s.Name, s.swaggerURI)
}

// Merge merges two schema with same address in turns of covered properties.
// NOTE: This has to be called before the schema goes under granting process.
func (s *Schema) Merge(os Schema) error {
	if s.HasGrantedAny() {
		return fmt.Errorf("AddCovProperty() has to be called on schema (%q) before granting", s.Ref())
	}
	if s.Ref() != os.Ref() {
		return fmt.Errorf("trying to merge different schemas: %s and %s", s.Ref(), os.Ref())
	}

	for pAddr, prop := range os.Properties {
		if prop.IsCovered() {
			if err := s.AddCovProperty(addr.MustParseProperty(pAddr)); err != nil {
				return err
			}
		}
	}
	return nil
}

// Grant inquiries the SchemaGrants to Add the granting information onto the Schema
func (c *Schema) Grant(grants SchemaGrants) error {
	grant, ok := grants[c.Ref()]
	if !ok {
		// There is no grants for this schema
		return nil
	}
	if grant.IsSchemaGranted() {
		c.IsGranted = true
		c.GrantComment = grant.Comment
		return nil
	}

	for propRelAddr, grantComment := range grant.Properties {
		property, ok := c.Properties[propRelAddr]
		if !ok {
			return fmt.Errorf(`property %q that is to be granted doesn't exist in Swagger schema %q'`, propRelAddr, c.Ref())
		}
		property.PropertyState = NewPropertyState(PropStateGranted, grantComment)
	}

	return nil
}

// NOTE: In case the ref is nil, it will return empty strings.
func SchemaInfoFromRef(ref openapispec.Ref) (name string, path string) {
	if ref.GetURL() == nil {
		return "", ""
	}
	fragments := strings.Split(ref.GetURL().Fragment, "/")
	return fragments[len(fragments)-1], ref.GetURL().Path
}

func SchemaNameToRef(schemaName string, path string) openapispec.Ref {
	return openapispec.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", path, schemaName))
}
