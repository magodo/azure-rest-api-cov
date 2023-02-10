package swagger

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/magodo/azure-rest-api-cov/analyzer/propmatch"
	"github.com/magodo/azure-rest-api-cov/analyzer/swaggerloader"

	openapispec "github.com/go-openapi/spec"
	"github.com/magodo/azure-rest-api-cov/analyzer/addr"
)

const swaggerExtensionMSDiscriminatorValue = "x-ms-discriminator-value"

const rootProp = ""

type SchemaMetaInfo struct {
	// In case this is a schema-less operation's meta info, the Name is empty string.
	Name string `json:"name"`

	// The rel path to the Swagger defining this schema. Especially, for schema-less operation's meta info, this is the path to the Swagger defining the API path.
	SwaggerRelPath string `json:"swagger_rel_path"`
}

// Ref returns a hypothetical schema reference. It is mainly used as a unique ID of the schema, to be used in log message or map key.
func (i SchemaMetaInfo) Ref() string {
	ref := SchemaNameToRef(i.SwaggerRelPath, i.Name)
	return ref.String()
}

func ParseSchemaMetaInfoRef(input string) (*SchemaMetaInfo, error) {
	parts := strings.Split(input, "#/definitions/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("malformed format of schema reference")
	}
	return &SchemaMetaInfo{
		Name:           parts[1],
		SwaggerRelPath: parts[0],
	}, nil
}

// Schema records the information for a schema. Especially, it is also used for storing meta information for schema-less operations.
type Schema struct {
	SchemaMetaInfo
	Properties Properties `json:"properties"`

	// The full URI of the Swagger file that defines this schema.
	swaggerURI string
}

func NewEmptySchema(swaggerBaseURL, swaggerRelPath, schemaName string, forResponse bool) (*Schema, error) {
	swaggerURI := swaggerBaseURL + "/" + swaggerRelPath

	swgSchema := &Schema{
		SchemaMetaInfo: SchemaMetaInfo{
			Name:           schemaName,
			SwaggerRelPath: swaggerRelPath,
		},
		swaggerURI: swaggerURI,
	}

	normSchemaRef := SchemaNameToRef(swaggerURI, schemaName)
	rootProperty, err := followRefOneStep(
		PropertyStateUncovered,
		map[string]interface{}{},
		normSchemaRef,
		forResponse,
		RefFollowReasonRef,
	)
	if err != nil {
		return nil, err
	}

	swgSchema.Properties = map[string]*Property{
		rootProp: rootProperty,
	}

	// Expand the root level properties of the schemas
	rootAddr := addr.Property{}
	_, err = swgSchema.expandPropertyOneLevelDeep(rootAddr, false)
	if err != nil {
		return nil, fmt.Errorf("expanding schemas %s (%s): %w", schemaName, swaggerURI, err)
	}

	// Remove the rootProp no matter expand one level succeed or not.
	// (If succeeded, the rootProp was already deleted. Otherwise, it means the schema is empty, we just delete the rootProp)
	delete(swgSchema.Properties, rootProp)

	return swgSchema, nil
}

// ProcessCoverageForSchema will create a new empty Schema for the input schema, then process it to generate the property coverages based on the coveredProps.
// If "forResponse" is specified, the read-only properties will  be ignored in the final coverage result.
// If "strict" is specified, the coveredProps should guarantee to exist in the designated schema. Otherwise, an error will return.
func ProcessCoverageForSchema(swaggerBaseDir, swaggerRelPath string, schemaName string, schemaVariantValue string, coveredProps []string, forResponse bool, strict bool, allowMismatchPropMatchers []propmatch.PropMatcher) (*Schema, error) {
	schema, err := NewEmptySchema(swaggerBaseDir, swaggerRelPath, schemaName, forResponse)
	if err != nil {
		return nil, err
	}

	if err := schema.ExpandProperties(); err != nil {
		return nil, err
	}

	for _, propStr := range coveredProps {
		prop := addr.MustParseProperty(propStr)
		if schemaVariantValue != "" {
			prop = addr.MustParseProperty("").AsVariant(schemaVariantValue).Append(prop)
		}
		if err := schema.AddCovProperty(prop, strict, allowMismatchPropMatchers); err != nil {
			return nil, err
		}
	}

	return schema, nil
}

// ExpandProperties expands properties to the "leaf", with expanding the ref, allOf, subProperties all the way along.
// NOTE: This has to be called before the schema goes under granting process.
func (s *Schema) ExpandProperties() error {
	if s.HasGrantedAny() {
		return fmt.Errorf("ExpandProperties() has to be called on schema (%q) before granting", s.Ref())
	}

	propsExpandedToEnd := map[string]bool{}
	propsToExpand := map[string]bool{} // Key is the prop addr, value is whether the prop is a discriminator

	log.Debugf("Expanding properties for schema: %s", s.Ref())
	for {
		propsToExpand = map[string]bool{}
		for k, v := range s.Properties {
			if !propsExpandedToEnd[k] {
				propsToExpand[k] = v.isDiscriminator
			}
		}

		if len(propsToExpand) == 0 {
			break
		}

		log.Tracef("Total props: %d, Props to expand: %d", len(s.Properties), len(propsToExpand))
		for existRAddr, isDiscriminator := range propsToExpand {
			existPropAddr := addr.MustParseProperty(existRAddr)
			end, err := s.expandPropertyOneLevelDeep(existPropAddr, isDiscriminator)
			if err != nil {
				return fmt.Errorf("expanding top property for %s: %w", existPropAddr, err)
			}

			if end {
				propsExpandedToEnd[existRAddr] = true
			}
		}
	}
	return nil
}

// AddCovProperty marks a property as covered.
// In case the property doesn't exist in the Swagger definition, it will error out if in the strict mode. Otherwise, it will just ignore this property.
// Furthermore, if the property is a parent of some existing properties, it will always be skipped. Since we only count coverage for leaf properties.
// NOTE: You shall ensure to call after expanding properties. Also, this has to be called before the schema goes under granting process.
func (s *Schema) AddCovProperty(propAddr addr.Property, strict bool, allowMismatchPropMatchers []propmatch.PropMatcher) error {
	if s.HasGrantedAny() {
		return fmt.Errorf("AddCovProperty() has to be called on schema (%q) before granting", s.Ref())
	}
	for existRAddr, prop := range s.Properties {
		existPropAddr := addr.MustParseProperty(existRAddr)

		if propAddr.Contains(existPropAddr) {
			return nil
		}

		if !existPropAddr.Equals(propAddr) {
			continue
		}

		// Ensure the property is a leaf property.
		isLeaf, err := prop.isLeafProperty()
		if err != nil {
			return err
		}
		if isLeaf {
			// Ignore in case this property is already marked to be ignored.
			if prop.PropertyState == PropertyStateIgnoredForReqROProp {
				return nil
			}
			prop.PropertyState = PropertyStateCovered
		}
		return nil
	}

	if strict {
		for _, m := range allowMismatchPropMatchers {
			o := propmatch.PropMatcher{
				SwaggerURI: s.swaggerURI,
				ModelName:  s.Name,
				PropAddr:   propAddr,
			}
			if m.Match(o) {
				return nil
			}
		}
		return fmt.Errorf("property %s doesn't belong to schemas %s (%s)", propAddr, s.Name, s.swaggerURI)
	}
	return nil
}

// expandPropertyOneLevelDeep expand the specified swagger property one level deep, with any ref, allOf, sub-properties taken into consideration.
func (s *Schema) expandPropertyOneLevelDeep(propAddr addr.Property, isVariant bool) (end bool, err error) {
	raddr := propAddr.String()

	defer func() {
		// We have to check whether we added any child property of this property.
		// If yes, then we need to remove this property from the Schema property map.
		if !end {
			delete(s.Properties, raddr)
		}
	}()

	prop, ok := s.Properties[raddr]
	if !ok {
		return false, fmt.Errorf("property %s does not exist in Schema %s (%s)", propAddr, s.Name, s.swaggerURI)
	}

	isCyclic, err := prop.expandRefPropertyInPlace(RefFollowReasonRef)
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

	// In case:
	// - current property is not a discriminator schema
	// - or it is already a expanded variant
	// - or it is a leaf variant
	// we will simply return here and do not do the expanding to variants below.
	if !prop.isDiscriminator || isVariant || len(prop.variantRefSet) == 0 {
		s.addProperties(levelProperties)
		return len(levelProperties) == 0, nil
	}

	// We are expanding a non-leaf discriminator for all of its variants, together with the expanding schema itself.
	skipped := 0

	refset := map[string]struct{}{
		prop.ownerSchemaRef.String(): {},
	}
	for ref := range prop.variantRefSet {
		refset[ref] = struct{}{}
	}

	for variantRef := range refset {
		if _, resolved := prop.resolvedRefs[variantRef]; resolved {
			skipped++
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

		normVariantRef := openapispec.MustCreateRef(variantRef)
		p, err := followRefOneStep(prop.PropertyState, prop.resolvedRefs, normVariantRef, prop.forResponse, RefFollowReasonVarianting)
		if err != nil {
			return false, err
		}

		s.addProperty(propAddr.AsVariant(descValue), *p)
	}

	// In case all the variants are skipped, we will potentially mark this expansion as end as there is no new properties got expanded.
	if skipped == len(refset) {
		return true, nil
	}

	return false, nil
}

// expandSubProperties expand the directly contained sub-properties of the property (prop) in the specified address (addr), given it is not a leaf property.
// Especially, if the property is an array to object, it will expand to the sub-properties of the object item instead.
func (s *Schema) expandSubProperties(propAddr addr.Property, prop *Property) (Properties, error) {
	output := Properties{}
	isLeaf, err := prop.isLeafProperty()
	if err != nil {
		return nil, err
	}
	if isLeaf {
		return output, nil
	}

	var properties openapispec.SchemaProperties
	if prop.schema.Items != nil {
		properties = prop.schema.Items.Schema.Properties
	} else {
		properties = prop.schema.Properties
	}
	for propName, propSchema := range properties {
		p := prop.propagateSubPropOfSchema(propSchema)
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
		// Then we will iterate that schema's property which by the concept is the top level property of this parent property.
		tmpProp := prop.propagateSubPropOfSchema(schema)

		// AllOf contains refs, then need to expand the reference properties first.
		if tmpProp.schema.Ref.String() != "" {
			isCyclic, err := tmpProp.expandRefPropertyInPlace(RefFollowReasonInherit)
			if err != nil {
				return nil, fmt.Errorf("dereferencing property %s in Schema %s (%s): %w", propAddr, s.Name, s.swaggerURI, err)
			}
			// Ignore... as there is no better way to handle this (since it has no object/property related, or shall we panic?)
			if isCyclic {
				continue
			}
		}

		// Each allOf property itself might contain allOf properties and direct properties, here we recursively retrieve those
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
	for _, prop := range s.Properties {
		if prop.IsGranted() {
			return true
		}
	}
	return false
}

// Merge merges two schema with same address in turns of covered properties.
// NOTE: This has to be called before the schema goes under granting process.
func (s *Schema) Merge(os Schema) error {
	if s.HasGrantedAny() {
		return fmt.Errorf("Merge() has to be called on schema (%q) before granting", s.Ref())
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

// NOTE: In case the ref is nil, it will return empty strings.
func SchemaInfoFromRef(ref openapispec.Ref) (name string, path string) {
	if ref.GetURL() == nil {
		return "", ""
	}
	fragments := strings.Split(ref.GetURL().Fragment, "/")
	return fragments[len(fragments)-1], ref.GetURL().Path
}

func SchemaNameToRef(path string, schemaName string) openapispec.Ref {
	return openapispec.MustCreateRef(fmt.Sprintf("%s#/definitions/%s", path, schemaName))
}
