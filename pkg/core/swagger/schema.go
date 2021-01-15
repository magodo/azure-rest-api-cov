package swagger

import (
	"fmt"
	"strings"

	"github.com/magodo/azure-rest-api-cov/pkg/core/specpath"

	"github.com/magodo/azure-rest-api-cov/pkg/core/swaggerloader"

	openapispec "github.com/go-openapi/spec"
	"github.com/magodo/azure-rest-api-cov/pkg/core/addr"
)

const swaggerExtensionMSDiscriminatorValue = "x-ms-discriminator-value"

const rootProp = ""

type SchemaMetaInfo struct {
	Name           string        `json:"name"`
	SwaggerRelPath specpath.Path `json:"swagger_rel_path"`
}

// Ref returns a hypothetical schema reference. It is mainly used as a unique ID of the schema, to be used in log message or map key.
func (i SchemaMetaInfo) Ref() string {
	return string(i.SwaggerRelPath) + "#/definitions/" + i.Name
}

type Schema struct {
	SchemaMetaInfo
	Properties Properties `json:"properties"`

	// Whether this property is granted to be not to implement in Terraform
	IsGranted    bool   `json:",omitempty"`
	GrantComment string `json:",omitempty"`

	// The full URI of the Swagger file that defines this schema.
	swaggerURI string
}

func NewEmptySchema(swaggerBaseURL, swaggerRelPath string, schemaName string, forResponse bool) (*Schema, error) {
	swaggerURI := swaggerBaseURL + "/" + swaggerRelPath

	ref := SchemaNameToRef(schemaName, "")
	normalizedSchemaRef := normalizePaths(ref.String(), swaggerURI)

	swgSchema := &Schema{
		SchemaMetaInfo: SchemaMetaInfo{
			Name:           schemaName,
			SwaggerRelPath: specpath.Path(swaggerRelPath),
		},
		swaggerURI: swaggerURI,
	}

	rootProperty, err := buildProperty(
		PropertyStateUncovered,
		map[string]interface{}{},
		openapispec.MustCreateRef(normalizedSchemaRef),
		forResponse,
	)
	if err != nil {
		return nil, err
	}

	swgSchema.Properties = map[string]*Property{
		rootProp: rootProperty,
	}

	// Expand the root level properties of the schemas
	rootAddr := addr.Property{}
	_, err = swgSchema.expandPropertyOneLevelDeep(rootAddr)
	if err != nil {
		return nil, fmt.Errorf("expanding schemas %s (%s): %w", schemaName, swaggerURI, err)
	}

	// Remove the rootProp no matter expand one level succeed or not.
	// (If succeeded, the rootProp was already deleted. Otherwise, it means the schema is empty, we just delete the rootProp)
	delete(swgSchema.Properties, rootProp)

	return swgSchema, nil
}

// expandPropertyOneLevelDeep expand the specified swagger schemas property one level deep, with any allOf and $ref taken into consideration.
func (s *Schema) expandPropertyOneLevelDeep(propAddr addr.Property) (end bool, err error) {
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
		return false, fmt.Errorf("property %s does not exist in Schema %s (%s)", propAddr, s.Name, s.swaggerURI)
	}

	isCyclic, err := prop.expandRefPropertyInPlace()
	if err != nil {
		return false, fmt.Errorf("dereferencing property %s in Schema %s (%s): %w", propAddr, s.Name, s.swaggerURI, err)
	}

	// If the property to be expanded is a cyclic reference, we will do nothing but keep that property
	if isCyclic {
		return true, nil
	}

	// temporarily store the Properties expanded in this level
	levelProperties := Properties{}

	// expand direct top level properties
	directTopSWGProperties, err := s.expandSubProperties(propAddr, prop)
	if err != nil {
		return false, fmt.Errorf("failed to expand direct top level properties for %q: %v", propAddr.String(), err)
	}

	// expand AllOf properties
	allOfSWGProperties, err := s.expandAllOfProperties(propAddr, prop)
	if err != nil {
		return false, fmt.Errorf("failed to expand allOf properties for %q: %v", propAddr.String(), err)
	}

	levelProperties.Merge(directTopSWGProperties)
	levelProperties.Merge(allOfSWGProperties)

	// For non-pure discriminator, because itself can be instantiated, so we will add the expanded properties also.

	// In case current property is a discriminator schema, we will expand it into its variants/children.
	// Otherwise, we will simply return here and do not do the expanding to variants below.
	if !prop.isDiscriminator {
		s.addProperties(levelProperties)
		end := len(levelProperties) == 0
		return end, nil
	}

	switch {
	case prop.isPureDiscriminator:
		// continue
	case prop.isLeafVariant():
		s.addProperties(levelProperties)
		end := len(levelProperties) == 0
		return end, nil
	default:
		// A non-leaf variant, we shall check whether it has been resolved yet.
		// If it is resolved, it means the variant has been expanded along the expansion of current property,
		// then we shall stop expanding its children variants and set its properties here.
		_, resolved := prop.resolvedRefs[prop.ownerSchemaRef.String()]
		if resolved {
			s.addProperties(levelProperties)
			end := len(levelProperties) == 0
			return end, nil
		}
	}

	for variantRef := range prop.variantRefSet {
		if _, resolved := prop.resolvedRefs[variantRef]; resolved {
			continue
		}
		ref := openapispec.MustCreateRef(variantRef)
		name, _ := SchemaInfoFromRef(ref)

		descValue := name
		// If the 'x-ms-discriminator-value' is defined for the variant schemas, then it will take priority.
		swagger, err := swaggerloader.LoadSwagger(prop.swaggerURI)
		if err != nil {
			return false, err
		}
		if v, ok := swagger.Definitions[name].Extensions[swaggerExtensionMSDiscriminatorValue].(string); ok {
			descValue = v
		}

		// Since we removed the discriminator base schema from the "resolvedRefs" when we creat the Property, we should in turn Add the
		// exact variant schema expanded to the "resolvedRefs".
		resolvedRefs := map[string]interface{}{}
		for k, v := range prop.resolvedRefs {
			resolvedRefs[k] = v
		}
		normVariantRef := openapispec.MustCreateRef(variantRef)
		resolvedRefs[variantRef] = struct{}{}

		p, err := buildProperty(prop.PropertyState, resolvedRefs, normVariantRef, prop.forResponse)
		if err != nil {
			return false, err
		}
		s.addProperty(propAddr.AsVariant(descValue), *p)
	}

	return false, nil
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
		p := prop.copy()
		p.transformToSubProp(propSchema)
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
		tmpProp := prop.copy()
		tmpProp.transformToSubProp(schema)

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
// In case the property doesn't exist in the Swagger definition, it will error out if in the strict mode.
// Otherwise, it will just ignore this property.
// NOTE: This has to be called before the schema goes under granting process.
func (s *Schema) AddCovProperty(propAddr addr.Property, strict bool) error {
	if s.HasGrantedAny() {
		return fmt.Errorf("AddCovProperty() has to be called on schema (%q) before granting", s.Ref())
	}
	for existRAddr, prop := range s.Properties {
		existPropAddr := addr.MustParseProperty(existRAddr)

		if !propAddr.Contains(existPropAddr) && !existPropAddr.Contains(propAddr) && !existPropAddr.Equals(propAddr) {
			continue
		}

		// If the property to Add is a parent property of any existing properties.
		// We will throw that property since we will only regard leaf property as covered.
		if propAddr.Contains(existPropAddr) {
			return nil
		}

		if existPropAddr.Equals(propAddr) {
			// Ensure the property is a leaf property.
			isLeaf, err := prop.isLeafProperty()
			if err != nil {
				return err
			}
			if isLeaf {
				// If the property state is already set, do nothing.
				// Otherwise, mark its state as covered.
				if prop.PropertyState == PropertyStateIgnoredForReqROProp {
					return nil
				}
				prop.PropertyState = PropertyStateCovered
			}
			return nil
		}

		// The property we're seeking is a member of the property under iteration, we will first expand it one level deep.
		end, err := s.expandPropertyOneLevelDeep(existPropAddr)
		if err != nil {
			return fmt.Errorf("expanding top level property for %s: %w", existPropAddr, err)
		}

		if end {
			continue
		}

		return s.AddCovProperty(propAddr, strict)
	}

	if strict {
		return fmt.Errorf("property %s doesn't belong to schemas %s (%s)", propAddr, s.Name, s.swaggerURI)
	}
	return nil
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
			if err := s.AddCovProperty(addr.MustParseProperty(pAddr), true); err != nil {
				return err
			}
		}
	}
	return nil
}

// Grant inquiries the SchemaGrants to Add the granting information onto the Schema
// In strict, we will ensure the properties appear in granting file exist in the current schema.
// Normally strict should always be true, except you are trying to predict the coverage for newer API version and there is
// no granting file for that version yet.
func (c *Schema) Grant(grant SchemaGrant, strict bool) error {
	if grant.IsSchemaGranted() {
		c.IsGranted = true
		c.GrantComment = grant.Comment
		c.Properties = nil
		return nil
	}

	for propAddr, grantComment := range grant.Properties {
		property, ok := c.Properties[propAddr]
		if !ok {
			if strict {
				return fmt.Errorf(`property %q that is to be granted doesn't exist in Swagger schema %q'`, propAddr, c.Ref())
			}
			continue
		}
		property.PropertyState = NewPropertyState(PropStateGranted, grantComment)

		// Following logic is to add the ability to grant a parent property. This is commented out
		// since it seems to be an invalid requirement.

		//// Check whether there is child properties of the property to grant.
		//// If any is found, we will remove those child properties from the schema,
		//// and put the property to grant into the schema, and mark it as granted.
		//var isAny bool
		//pAddr := addr.MustParseProperty(propAddr)
		//for existPropAddr := range c.Properties {
		//	epAddr := addr.MustParseProperty(existPropAddr)
		//	if pAddr.Contains(epAddr) {
		//		isAny = true
		//		delete(c.Properties, existPropAddr)
		//	}
		//}
		//if !isAny {
		//	return fmt.Errorf(`property %q that is to be granted doesn't exist in Swagger schema %q'`, propAddr, c.Ref())
		//}
		//c.Properties[propAddr] = &Property{
		//	PropertyState: NewPropertyState(PropStateGranted, grantComment),
		//}
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
