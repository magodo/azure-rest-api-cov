package swagger

import (
	"encoding/json"
	"fmt"

	"github.com/magodo/azure-rest-api-cov/pkg/core/swaggerloader"

	openapispec "github.com/go-openapi/spec"
)

type PropertyState interface {
	IsCovered() bool
	IsUncovered() bool
	IsGranted() bool
}

type PropState string

const (
	PropStateCovered   PropState = "Covered"
	PropStateUncovered           = "Uncovered"
	PropStateGranted             = "Granted"
)

var (
	PropertyStateUncovered           = NewPropertyState(PropStateUncovered, "")
	PropertyStateCovered             = NewPropertyState(PropStateCovered, "")
	PropertyStateGrantedForReqROProp = NewPropertyState(PropStateGranted, "read-only property wouldn't show up in the request payload")
)

func NewPropertyState(state PropState, grantMsg string) PropertyState {
	switch state {
	case PropStateCovered:
		return propertyStateCovered{State: state}
	case PropStateUncovered:
		return propertyStateUncovered{State: state}
	case PropStateGranted:
		return propertyStateGranted{State: state, GrantComment: grantMsg}
	default:
		panic("will not happen")
	}
}

type propertyStateCovered struct {
	State PropState `json:",omitempty"`
}

func (s propertyStateCovered) IsCovered() bool      { return true }
func (s propertyStateCovered) IsUncovered() bool    { return false }
func (s propertyStateCovered) IsGranted() bool      { return false }
func (s propertyStateCovered) GrantMessage() string { panic("Covered property has no grant message") }

type propertyStateUncovered struct {
	State PropState `json:",omitempty"`
}

func (s propertyStateUncovered) IsCovered() bool   { return false }
func (s propertyStateUncovered) IsUncovered() bool { return true }
func (s propertyStateUncovered) IsGranted() bool   { return false }
func (s propertyStateUncovered) GrantMessage() string {
	panic("Uncovered property has no grant message")
}

type propertyStateGranted struct {
	State        PropState `json:",omitempty"`
	GrantComment string    `json:",omitempty"`
}

func (s propertyStateGranted) IsCovered() bool      { return false }
func (s propertyStateGranted) IsUncovered() bool    { return false }
func (s propertyStateGranted) IsGranted() bool      { return true }
func (s propertyStateGranted) GrantMessage() string { return s.GrantComment }

type Property struct {
	PropertyState `json:"state"`

	// The schemas of this property
	schema openapispec.Schema

	// The normalized ref of the schema (in definition section) that defines/owns this property.
	ownerSchemaRef *openapispec.Ref

	// The resolved URI refs along the way to this schemas, each is an normalized reference.
	// This is used to avoid cyclic reference.
	// Especially for discriminator and variant schemas:
	// - Pure discriminator: will never appear in this set since it is never "really" used
	// - Non-leaf variant: will be added into this map only when it is expanded to its variants
	// - Leaf variant: will be added into this map when its ref got resolved
	resolvedRefs map[string]interface{}

	// The URI of the schema file
	swaggerURI string

	// Whether this property is part of a response payload.
	// If false, we will automatically grant the read-only properties, since they are meaningless
	// in context of a request payload.
	// See: https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#fixed-fields-13
	forResponse bool

	// Whether the schema of this property is a discriminator (either pure or derived schema)
	isDiscriminator bool

	// Whether the schema of this property is a pure discriminator
	isPureDiscriminator bool

	// Normalized reference to variant schemas of the schema of this property. It is nil if `isDiscriminator` is false.
	// Note that: Non-leaf variant schema has itself as its variant schema (i.e. both pure discriminator and leaf variant
	//            do not contain itself in the variantRefSet).
	//            This is because non-leaf variant schema is both a discriminator and a variant (can be instantiate), when
	//            expanding it to its variant, we will need to both instantiate itself and its children variants.
	variantRefSet map[string]struct{}
}

type Properties map[string]*Property // the key is relative property addr string representation

func (p Properties) Merge(o Properties) {
	for k, v := range o {
		p[k] = v
	}
}

// propagateProperty propagates a new property which is directly defined by the schema
// of the current property.
func (prop Property) propagateProperty(schema openapispec.Schema) *Property {
	newResolvedRefs := map[string]interface{}{}
	if len(prop.resolvedRefs) != 0 {
		for k, v := range prop.resolvedRefs {
			newResolvedRefs[k] = v
		}
	}

	state := prop.PropertyState
	if !prop.forResponse && schema.ReadOnly {
		state = PropertyStateGrantedForReqROProp
	}
	return &Property{
		PropertyState:  state,
		schema:         schema,
		ownerSchemaRef: prop.ownerSchemaRef,
		resolvedRefs:   newResolvedRefs,
		swaggerURI:     prop.swaggerURI,
		forResponse:    prop.forResponse,
	}
}

// buildProperty build a new property by providing all the needed information about this property.
// It is useful when initiating the root property or building a property by following a reference to a new schema.
// The "resolvedRefs" keeps track about resolved refs at this point. In case that the schema of this property is
// a non-discriminator schema or leaf variant, current schema's ref will be added to resolvedRefs of the new property.
func buildProperty(state PropertyState, resolvedRefs map[string]interface{}, normSchemaRef openapispec.Ref, forResponse bool) (*Property, error) {
	schema, err := openapispec.ResolveRefWithBase(nil, &normSchemaRef, nil)
	if err != nil {
		return nil, err
	}
	swaggerURI := normSchemaRef.GetURL().Path

	newResolvedRefs := map[string]interface{}{}
	if len(resolvedRefs) != 0 {
		for k, v := range resolvedRefs {
			newResolvedRefs[k] = v
		}
	}

	if !forResponse && schema.ReadOnly {
		state = PropertyStateGrantedForReqROProp
	}
	prop := &Property{
		PropertyState:  state,
		schema:         *schema,
		ownerSchemaRef: &normSchemaRef,
		resolvedRefs:   newResolvedRefs,
		swaggerURI:     swaggerURI,
		forResponse:    forResponse,
	}

	isDiscriminator, pure, err := checkDiscriminatorSchema(normSchemaRef)
	if err != nil {
		return nil, err
	}
	if !isDiscriminator {
		// In case current schema is not a discriminator, we will mark it as ref-ed.
		prop.resolvedRefs[normSchemaRef.String()] = struct{}{}
		return prop, nil
	}

	prop.isDiscriminator = isDiscriminator
	prop.isPureDiscriminator = pure

	prop.variantRefSet, err = getVariantRefSet(normSchemaRef)
	if err != nil {
		return nil, err
	}

	// In case current schema is a leaf variant, we will mark it as ref-ed.
	if prop.isLeafVariant() {
		prop.resolvedRefs[normSchemaRef.String()] = struct{}{}
	}

	return prop, nil
}

func getVariantRefSet(normSchemaRef openapispec.Ref) (map[string]struct{}, error) {
	is, pure, err := checkDiscriminatorSchema(normSchemaRef)
	if err != nil {
		return nil, err
	}
	if !is {
		return nil, nil
	}

	swaggerURI := normSchemaRef.GetURL().Path
	allAllOfs, err := swaggerAllAllOfs(swaggerURI)
	if err != nil {
		return nil, err
	}

	// Find all the child variants which derives from this schema.
	childRefSet := map[string]struct{}{normSchemaRef.String(): {}}
	oldChildRefSet := map[string]struct{}{}
	for len(childRefSet) != len(oldChildRefSet) {
		oldChildRefSet = childRefSet
		childRefSet = make(map[string]struct{}, len(oldChildRefSet))
		for k := range oldChildRefSet {
			childRefSet[k] = struct{}{}
		}

		for ref := range oldChildRefSet {
			// Collect its children
			for childSchemaRef, allOfs := range allAllOfs {
				if _, ok := allOfs[ref]; ok {
					childRefSet[childSchemaRef] = struct{}{}
				}
			}
		}
	}
	// Keep this schema ref in the child ref only if this schema is a non-leaf variant and non-pure discriminator.
	if pure || len(childRefSet) == 1 {
		delete(childRefSet, normSchemaRef.String())
	}
	if len(childRefSet) == 0 {
		return nil, nil
	}

	return childRefSet, nil
}

// Construct all the "allOf"s of the schemas in current spec.
// {normRef: {normRef1, normRef2,...}, ...}
// TODO: cache it
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
		ref := SchemaNameToRef(name, swaggerURI)
		allOfRefs := map[string]struct{}{}
		allAllOfs[ref.String()] = allOfRefs

		for _, allOf := range schema.AllOf {
			if ref := allOf.Ref; ref.String() != "" {
				normalizedRef := normalizeFileRef(&ref, swaggerURI)
				allOfRefs[normalizedRef.String()] = struct{}{}
			}
		}
	}
	return allAllOfs, nil
}

// checkDiscriminatorSchema tells whether the specified schema is a discriminator (either pure discriminator or derived schema).
// Note that in Azure rest api spec (OpenAPI 2), a schema is a discriminator when itself directly defines "discriminator" or it is a derived schema (variant),
// as derived schema also inherits the "discriminator" property.
// This function goes backwards the "allOf" chain to find the nearest parent who explicitly defines the discriminator. If it finds any such parent schema,
// `is` is `true`. Additionally, if the property itself corresponds to a direct discriminator, `pure` is set to `true`.
func checkDiscriminatorSchema(normSchemaRef openapispec.Ref) (is bool, pure bool, err error) {
	schema, err := openapispec.ResolveRefWithBase(nil, &normSchemaRef, nil)
	if err != nil {
		return false, false, err
	}

	swaggerURI := normSchemaRef.GetURL().Path
	swagger, err := swaggerloader.LoadSwagger(swaggerURI)
	if err != nil {
		return false, false, err
	}

	allAllOfs, err := swaggerAllAllOfs(swaggerURI)
	if err != nil {
		return false, false, err
	}

	if schema.Discriminator != "" {
		return true, true, nil
	}

	baseRefSet := map[string]struct{}{normSchemaRef.String(): {}}
	for len(baseRefSet) != 0 {
		newBaseRefSet := make(map[string]struct{})
		for normRef := range baseRefSet {
			// Collect its parents
			allOfs, ok := allAllOfs[normRef]
			if ok {
				for k := range allOfs {
					newBaseRefSet[k] = struct{}{}
				}
			}
		}
		// Check whether there is any base schema has explicitly set the discriminator
		for ref := range newBaseRefSet {
			apiRef := openapispec.MustCreateRef(ref)
			schema, err := openapispec.ResolveRefWithBase(swagger, &apiRef, &openapispec.ExpandOptions{RelativeBase: swaggerURI})
			if err != nil {
				return false, false, err
			}
			if schema.Discriminator != "" {
				return true, false, nil
			}
		}
		baseRefSet = newBaseRefSet
	}
	return false, false, nil
}

// expandRefPropertyInPlace expand a property itself IN-PLACE until either it is a concrete schemas (i.e. not a ref) or hit a cyclic ref.
func (prop *Property) expandRefPropertyInPlace() (isCyclic bool, err error) {
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

	// If current ref has already been derefed, meaning a cyclic ref is hit, we will return.
	if _, ok := prop.resolvedRefs[normalizedRef.String()]; ok {
		return true, nil
	}

	refProp, err := buildProperty(prop.PropertyState, prop.resolvedRefs, *normalizedRef, prop.forResponse)
	if err != nil {
		return false, err
	}

	*prop = *refProp

	return prop.expandRefPropertyInPlace()
}

func (prop *Property) isLeafVariant() bool {
	if !prop.isDiscriminator {
		return false
	}
	if prop.isPureDiscriminator {
		return false
	}
	return prop.variantRefSet == nil
}

func (props Properties) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{}
	for k, prop := range props {
		mm := map[string]interface{}{}
		m[k] = mm
		switch state := prop.PropertyState.(type) {
		case propertyStateCovered:
			mm["state"] = state.State
		case propertyStateUncovered:
			mm["state"] = state.State
		case propertyStateGranted:
			mm["state"] = state.State
			mm["grant_comment"] = state.GrantComment
		default:
			panic("will not happen")
		}
	}
	return json.Marshal(m)
}

func (props *Properties) UnmarshalJSON(b []byte) error {
	*props = map[string]*Property{}
	m := map[string]interface{}{}
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	for k, v := range m {
		v, ok := v.(map[string]interface{})
		if !ok {
			return fmt.Errorf("failed to convert internal object to string map: %v", string(b))
		}
		stateRaw, ok := v["state"]
		if !ok {
			return fmt.Errorf(`no "state" in the raw byte`)
		}
		state, ok := stateRaw.(string)
		if !ok {
			return fmt.Errorf(`the "state" in the raw byte is not string`)
		}
		var grantComment string
		grantCommentRaw, ok := v["grant_comment"]
		if ok {
			grantComment, ok = grantCommentRaw.(string)
			if !ok {
				return fmt.Errorf(`the "grant_comment" in the raw byte is not string`)
			}
		}
		(*props)[k] = &Property{PropertyState: NewPropertyState(PropState(state), grantComment)}
	}
	return nil
}
