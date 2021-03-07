package swagger

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/magodo/azure-rest-api-cov/analyzer/swaggerloader"

	openapispec "github.com/go-openapi/spec"
)

type PropertyState interface {
	IsCovered() bool
	IsUncovered() bool
	IsGranted() bool
	IsIgnored() bool
	Message() string
	GetState() PropState
}

type PropState string

const (
	PropStateCovered   PropState = "Covered"
	PropStateUncovered           = "Uncovered"
	PropStateGranted             = "Granted"
	PropStateIgnored             = "Ignored"
)

var (
	PropertyStateUncovered = NewPropertyState(PropStateUncovered, "")
	PropertyStateCovered   = NewPropertyState(PropStateCovered, "")
	// read-only property wouldn't show up in the request payload
	PropertyStateIgnoredForReqROProp = NewPropertyState(PropStateIgnored, "")
)

func NewPropertyState(state PropState, msg string) PropertyState {
	switch state {
	case PropStateCovered:
		return propertyStateCovered{State: state}
	case PropStateUncovered:
		return propertyStateUncovered{State: state}
	case PropStateGranted:
		return propertyStateGranted{State: state, GrantComment: msg}
	case PropStateIgnored:
		return propertyStateIgnored{State: state}
	default:
		panic("will not happen")
	}
}

type propertyStateCovered struct {
	State PropState `json:",omitempty"`
}

func (s propertyStateCovered) IsCovered() bool     { return true }
func (s propertyStateCovered) IsUncovered() bool   { return false }
func (s propertyStateCovered) IsGranted() bool     { return false }
func (s propertyStateCovered) IsIgnored() bool     { return false }
func (s propertyStateCovered) Message() string     { return "" }
func (s propertyStateCovered) GetState() PropState { return s.State }

type propertyStateUncovered struct {
	State PropState `json:",omitempty"`
}

func (s propertyStateUncovered) IsCovered() bool     { return false }
func (s propertyStateUncovered) IsUncovered() bool   { return true }
func (s propertyStateUncovered) IsGranted() bool     { return false }
func (s propertyStateUncovered) IsIgnored() bool     { return false }
func (s propertyStateUncovered) Message() string     { return "" }
func (s propertyStateUncovered) GetState() PropState { return s.State }

type propertyStateGranted struct {
	State        PropState `json:",omitempty"`
	GrantComment string    `json:",omitempty"`
}

func (s propertyStateGranted) IsCovered() bool     { return false }
func (s propertyStateGranted) IsUncovered() bool   { return false }
func (s propertyStateGranted) IsGranted() bool     { return true }
func (s propertyStateGranted) IsIgnored() bool     { return false }
func (s propertyStateGranted) Message() string     { return s.GrantComment }
func (s propertyStateGranted) GetState() PropState { return s.State }

type propertyStateIgnored struct {
	State PropState `json:",omitempty"`
}

func (s propertyStateIgnored) IsCovered() bool     { return false }
func (s propertyStateIgnored) IsUncovered() bool   { return false }
func (s propertyStateIgnored) IsGranted() bool     { return false }
func (s propertyStateIgnored) IsIgnored() bool     { return true }
func (s propertyStateIgnored) Message() string     { return "" }
func (s propertyStateIgnored) GetState() PropState { return s.State }

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
	//            do not contain itself in the rootVariantRefSet).
	//            This is because non-leaf variant schema is both a discriminator and a variant (can be instantiate), when
	//            expanding it to its variant, we will need to both instantiate itself and its children variants.
	rootVariantRefSet map[string]struct{}

	// Similar to rootVariantRefSet, but only count for the variants under current schema of this property in turns of allOf hierarchy.
	directVariantRefSet map[string]struct{}
}

type Properties map[string]*Property // the key is relative property addr string representation

func (p Properties) Merge(o Properties) {
	for k, v := range o {
		p[k] = v
	}
}

func (prop Property) copy() *Property {
	newResolvedRefs := map[string]interface{}{}
	for k, v := range prop.resolvedRefs {
		newResolvedRefs[k] = v
	}

	newRootVariantRefSet := map[string]struct{}{}
	for k, v := range prop.rootVariantRefSet {
		newRootVariantRefSet[k] = v
	}

	newDirectVariantRefSet := map[string]struct{}{}
	for k, v := range prop.directVariantRefSet {
		newDirectVariantRefSet[k] = v
	}

	return &Property{
		PropertyState:       prop.PropertyState,
		schema:              prop.schema,
		ownerSchemaRef:      prop.ownerSchemaRef,
		resolvedRefs:        newResolvedRefs,
		swaggerURI:          prop.swaggerURI,
		forResponse:         prop.forResponse,
		isDiscriminator:     prop.isDiscriminator,
		isPureDiscriminator: prop.isPureDiscriminator,
		rootVariantRefSet:   newRootVariantRefSet,
		directVariantRefSet: newDirectVariantRefSet,
	}
}

// propagateSubPropOfSchema propagates a sub property with the given schema. The sub-property contains
// all the property level information from the parent property, except the discriminator related ones as the sub-property
// will not inherit them in the syntax of swagger.
func (prop *Property) propagateSubPropOfSchema(schema openapispec.Schema) *Property {
	newResolvedRefs := map[string]interface{}{}
	for k, v := range prop.resolvedRefs {
		newResolvedRefs[k] = v
	}

	newProp := &Property{
		PropertyState:  prop.PropertyState,
		schema:         schema,
		ownerSchemaRef: prop.ownerSchemaRef,
		resolvedRefs:   newResolvedRefs,
		swaggerURI:     prop.swaggerURI,
		forResponse:    prop.forResponse,
		// Ignore the discriminator related properties here as the new sub-property will not be a discriminator.
	}

	if !prop.forResponse && schema.ReadOnly {
		newProp.PropertyState = PropertyStateIgnoredForReqROProp
	}
	return newProp
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
	for k, v := range resolvedRefs {
		newResolvedRefs[k] = v
	}

	if !forResponse && schema.ReadOnly {
		state = PropertyStateIgnoredForReqROProp
	}
	prop := &Property{
		PropertyState:  state,
		schema:         *schema,
		ownerSchemaRef: &normSchemaRef,
		resolvedRefs:   newResolvedRefs,
		swaggerURI:     swaggerURI,
		forResponse:    forResponse,
	}

	isDiscriminator, pure, _, err := checkDiscriminatorSchema(normSchemaRef)
	if err != nil {
		return nil, err
	}
	if !isDiscriminator {
		// In case the current schema is not a discriminator, we will mark it as ref-ed.
		prop.resolvedRefs[normSchemaRef.String()] = struct{}{}
		return prop, nil
	}

	// Following applies to discriminator property
	prop.isDiscriminator = isDiscriminator
	prop.isPureDiscriminator = pure

	prop.rootVariantRefSet, prop.directVariantRefSet, err = getVariantRefSet(normSchemaRef)
	if err != nil {
		return nil, err
	}

	// In case the current schema is a leaf variant, we will mark it as ref-ed.
	if prop.isLeafVariant() {
		prop.resolvedRefs[normSchemaRef.String()] = struct{}{}
	}

	return prop, nil
}

// getVariantRefSet gets the variant set for both "root" and "direct".
func getVariantRefSet(normSchemaRef openapispec.Ref) (map[string]struct{}, map[string]struct{}, error) {
	is, pure, rootDiscriminatorRef, err := checkDiscriminatorSchema(normSchemaRef)
	if err != nil {
		return nil, nil, err
	}
	if !is {
		return nil, nil, nil
	}

	swaggerURI := normSchemaRef.GetURL().Path
	allAllOfs, err := swaggerAllAllOfs(swaggerURI)
	if err != nil {
		return nil, nil, err
	}

	fillInChildRefSet := func(in *map[string]struct{}) {
		childRefSet := *in
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
		*in = childRefSet
	}

	// Find all the child variants which derives from this schema.
	directChildRefSet := map[string]struct{}{normSchemaRef.String(): {}}
	fillInChildRefSet(&directChildRefSet)

	// Find all the child variants which derives from the root discriminator schema.
	//
	// This is because currently in the Azure Go SDK (track1), all the derived concrete types will implement the root discriminator interface.
	// Whilst in the cov-terraform tool, when analyzing the variants of a field of type interface, we will check all its possible implementors
	// (including the directly inherited models and sibling models in Swagger).
	// Though this is currently observed in Go SDK, since Go SDK is based from autorest modeler, which is common base for other generators.
	// Therefore, this behavior should be the same for other SDKs.
	rootChildRefSet := map[string]struct{}{rootDiscriminatorRef.String(): {}}
	fillInChildRefSet(&rootChildRefSet)

	// Keep this schema ref in the child ref only if this schema is a non-leaf variant and non-pure discriminator (i.e. an intermediate discrminator).
	if pure || len(directChildRefSet) == 1 {
		delete(directChildRefSet, normSchemaRef.String())
	}

	delete(rootChildRefSet, rootDiscriminatorRef.String())

	return rootChildRefSet, directChildRefSet, nil
}

type AllAllOfsCache struct {
	sync.Mutex
	m map[string]map[string]map[string]struct{}
}

var allAllOfsCache = AllAllOfsCache{
	Mutex: sync.Mutex{},
	m:     map[string]map[string]map[string]struct{}{},
}

// Construct all the "allOf"s of the schemas in current spec.
// {normRef: {normRef1, normRef2,...}, ...}
func swaggerAllAllOfs(swaggerURI string) (map[string]map[string]struct{}, error) {
	var err error
	swaggerURI, err = filepath.Abs(swaggerURI)
	if err != nil {
		return nil, err
	}

	allAllOfsCache.Lock()
	defer allAllOfsCache.Unlock()

	if allAllOfs, ok := allAllOfsCache.m[swaggerURI]; ok {
		return allAllOfs, nil
	}

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
	allAllOfsCache.m[swaggerURI] = allAllOfs
	return allAllOfs, nil
}

// checkDiscriminatorSchema tells whether the specified schema is a discriminator (either pure discriminator or derived schema).
// Note that in Azure rest api spec (OpenAPI 2), a schema is a discriminator when itself directly defines "discriminator", or it is a derived schema (variant),
// as derived schema also inherits the "discriminator" property.
// This function goes backwards the "allOf" chain to find the nearest parent who explicitly defines the discriminator. If it finds any such parent schema,
// `is` is `true`. Additionally, if the property itself corresponds to a direct discriminator, `pure` is set to `true`. This pure discriminator parent will also be returned.
func checkDiscriminatorSchema(normSchemaRef openapispec.Ref) (is bool, pure bool, rootDiscriminatorRef *openapispec.Ref, err error) {
	schema, err := openapispec.ResolveRefWithBase(nil, &normSchemaRef, nil)
	if err != nil {
		return false, false, nil, err
	}

	swaggerURI := normSchemaRef.GetURL().Path
	swagger, err := swaggerloader.LoadSwagger(swaggerURI)
	if err != nil {
		return false, false, nil, err
	}

	allAllOfs, err := swaggerAllAllOfs(swaggerURI)
	if err != nil {
		return false, false, nil, err
	}

	if schema.Discriminator != "" {
		return true, true, &normSchemaRef, nil
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
				return false, false, nil, err
			}
			if schema.Discriminator != "" {
				return true, false, &apiRef, nil
			}
		}
		baseRefSet = newBaseRefSet
	}
	return false, false, nil, nil
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

func (prop Property) isLeafVariant() bool {
	if !prop.isDiscriminator {
		return false
	}
	if prop.isPureDiscriminator {
		return false
	}
	return len(prop.directVariantRefSet) == 0
}

func (prop Property) isLeafProperty() (bool, error) {
	p := prop.copy()
	cyclic, err := p.expandRefPropertyInPlace()
	if err != nil {
		return false, err
	}
	if cyclic {
		return false, nil
	}

	// Model with Map/Dictionary properties needs special handling
	// See https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#model-with-mapdictionary-properties
	schema := p.schema
	if schema.Type.Contains("object") {
		if _, ok := schema.Properties["additionalProperties"]; ok {
			return true, nil
		}
	}

	return len(p.schema.AllOf) == 0 && p.schema.Properties == nil, nil
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
