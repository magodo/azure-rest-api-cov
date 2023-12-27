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

	AppPropertyMaps []AppPropertyMap `json:"app_property_maps,omitempty"`

	// The schemas of this property
	schema openapispec.Schema

	// The normalized ref of the schema (in definition section) that defines/owns this property.
	ownerSchemaRef *openapispec.Ref

	// The resolved URI refs along the way to this schemas, each is an normalized reference.
	// This is used to avoid cyclic reference.
	resolvedRefs map[string]interface{}

	// The URI of the schema file
	swaggerURI string

	// Whether this property is part of a response payload.
	// If false, we will automatically grant the read-only properties, since they are meaningless
	// in context of a request payload.
	// See: https://github.com/OAI/OpenAPI-Specification/blob/master/versions/2.0.md#fixed-fields-13
	forResponse bool

	// Whether the schema of this property is a discriminator
	isDiscriminator bool

	// Normalized reference to variant schemas of this property's schema of this property. It is nil if `isDiscriminator` is false.
	variantRefSet map[string]struct{}
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

	newVariantRefSet := map[string]struct{}{}
	for k, v := range prop.variantRefSet {
		newVariantRefSet[k] = v
	}

	return &Property{
		PropertyState:   prop.PropertyState,
		schema:          prop.schema,
		ownerSchemaRef:  prop.ownerSchemaRef,
		resolvedRefs:    newResolvedRefs,
		swaggerURI:      prop.swaggerURI,
		forResponse:     prop.forResponse,
		isDiscriminator: prop.isDiscriminator,
		variantRefSet:   newVariantRefSet,
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

type RefFollowReason int

const (
	// RefFollowReasonRef indicates the ref is a direct referencing for a property
	RefFollowReasonRef RefFollowReason = iota
	// RefFollowReasonInherit indicates the ref is inside allOf
	RefFollowReasonInherit
	// RefFollowReasonVarianting indicates the a variant ref
	RefFollowReasonVarianting
)

// followRefOneStep build a new property by its ref, together with all the needed information provided.
// It is used when initiating the root property or building a property by following a reference to a new schema.
// The "resolvedRefs" keeps track about resolved refs at this point.
func followRefOneStep(state PropertyState, resolvedRefs map[string]interface{}, normSchemaRef openapispec.Ref, forResponse bool, reason RefFollowReason) (*Property, error) {
	schema, err := openapispec.ResolveRefWithBase(nil, &normSchemaRef, nil)
	if err != nil {
		return nil, err
	}
	swaggerURI := normSchemaRef.GetURL().Path

	if !forResponse && schema.ReadOnly {
		state = PropertyStateIgnoredForReqROProp
	}

	newResolvedRefs := map[string]interface{}{}
	for k, v := range resolvedRefs {
		newResolvedRefs[k] = v
	}

	prop := &Property{
		PropertyState:  state,
		schema:         *schema,
		ownerSchemaRef: &normSchemaRef,
		resolvedRefs:   newResolvedRefs,
		swaggerURI:     swaggerURI,
		forResponse:    forResponse,
	}

	isDiscriminator, err := isDiscriminator(normSchemaRef)
	if err != nil {
		return nil, err
	}

	if !isDiscriminator {
		switch reason {
		case RefFollowReasonRef:
			prop.resolvedRefs[normSchemaRef.String()] = struct{}{}
		case RefFollowReasonInherit:
			prop.resolvedRefs[normSchemaRef.String()] = struct{}{}
		case RefFollowReasonVarianting:
			prop.resolvedRefs[normSchemaRef.String()] = struct{}{}
		}
		return prop, nil
	}

	// Following applies to discriminator property
	prop.isDiscriminator = isDiscriminator
	prop.variantRefSet, err = GetVariantRefSet(normSchemaRef)
	if err != nil {
		return nil, err
	}

	switch reason {
	case RefFollowReasonRef:
		// Variant ref: when the ref is a discriminator, it is not marked as resolved, as this discriminator will be later used for expanding variants as itself, or allOf by its variants.
		// Except this variant is a leaf variant, where there is no variants to expand.
		if len(prop.variantRefSet) == 0 {
			prop.resolvedRefs[normSchemaRef.String()] = struct{}{}
		}
	case RefFollowReasonInherit:
		prop.resolvedRefs[normSchemaRef.String()] = struct{}{}
	case RefFollowReasonVarianting:
		prop.resolvedRefs[normSchemaRef.String()] = struct{}{}
	}

	return prop, nil
}

// GetVariantRefSet gets the variant set for both "root" and "direct".
func GetVariantRefSet(normSchemaRef openapispec.Ref) (map[string]struct{}, error) {
	is, err := isDiscriminator(normSchemaRef)
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
	childRefSet := map[string]struct{}{normSchemaRef.String(): {}}
	fillInChildRefSet(&childRefSet)
	delete(childRefSet, normSchemaRef.String())

	return childRefSet, nil
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
		ref := SchemaNameToRef(swaggerURI, name)
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

// isDiscriminator tells whether the specified schema is a discriminator (either pure discriminator or derived schema).
// Note that in Azure rest api spec (OpenAPI 2), a schema is a discriminator when itself directly defines "discriminator", or it is a derived schema (variant),
// as derived schema also inherits the "discriminator" property.
// This function goes backwards the "allOf" chain to find the nearest parent who explicitly defines the discriminator.
func isDiscriminator(normSchemaRef openapispec.Ref) (is bool, err error) {
	schema, err := openapispec.ResolveRefWithBase(nil, &normSchemaRef, nil)
	if err != nil {
		return false, err
	}
	if schema.Discriminator != "" {
		return true, nil
	}

	swaggerURI := normSchemaRef.GetURL().Path
	swagger, err := swaggerloader.LoadSwagger(swaggerURI)
	if err != nil {
		return false, err
	}

	allAllOfs, err := swaggerAllAllOfs(swaggerURI)
	if err != nil {
		return false, err
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
				return false, err
			}
			if schema.Discriminator != "" {
				return true, nil
			}
		}
		baseRefSet = newBaseRefSet
	}
	return false, nil
}

// expandRefPropertyInPlace recursively expand a property itself IN-PLACE until either it is a concrete schemas (i.e. not a ref) or hit a cyclic ref.
func (prop *Property) expandRefPropertyInPlace(reason RefFollowReason) (isCyclic bool, err error) {
	ref := prop.schema.Ref
	if ref.String() == "" {
		if prop.schema.Items == nil {
			return false, nil
		}

		// Specially, if current schema is an array, and the items is a ref, we need to go on expand it.
		if prop.schema.Items.Schema == nil && len(prop.schema.Items.Schemas) == 0 {
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

	refProp, err := followRefOneStep(prop.PropertyState, prop.resolvedRefs, *normalizedRef, prop.forResponse, reason)
	if err != nil {
		return false, err
	}

	*prop = *refProp

	return prop.expandRefPropertyInPlace(reason)
}

// isLeafProperty tells whether the prop is a leaf property, meaning it has no unresolved ref, sub-properties, or allOf(s).
func (prop Property) isLeafProperty() (bool, error) {
	p := prop.copy()
	cyclic, err := p.expandRefPropertyInPlace(RefFollowReasonRef)
	if err != nil {
		return false, err
	}
	if cyclic {
		return false, nil
	}

	schema := p.schema

	if len(schema.AllOf) != 0 || (schema.Properties != nil && len(schema.Properties) != 0) {
		return false, nil
	}

	// In case this property is an array, we will further check whether its item has defined non-primary type schema.
	if schema.Items != nil {
		var s openapispec.Schema
		if schema.Items.Schema != nil {
			s = *schema.Items.Schema
		} else {
			s = schema.Items.Schemas[0]
		}
		return prop.propagateSubPropOfSchema(s).isLeafProperty()
	}

	// TODO: In case this is a map, we should follow the schema if the map value.
	return true, nil

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
		case propertyStateIgnored:
			mm["state"] = state.State
		default:
			panic("will not happen")
		}

		if len(prop.AppPropertyMaps) != 0 {
			b, err := json.Marshal(prop.AppPropertyMaps)
			if err != nil {
				return nil, err
			}
			var maps []interface{}
			if err := json.Unmarshal(b, &maps); err != nil {
				return nil, err
			}
			mm["app_property_maps"] = maps
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

		var appPropMaps []AppPropertyMap
		appPropMapsRaw, ok := v["app_property_maps"]
		if ok {
			b, err := json.Marshal(appPropMapsRaw)
			if err != nil {
				return err
			}
			if err := json.Unmarshal(b, &appPropMaps); err != nil {
				return err
			}
		}

		(*props)[k] = &Property{PropertyState: NewPropertyState(PropState(state), grantComment), AppPropertyMaps: appPropMaps}
	}
	return nil
}
