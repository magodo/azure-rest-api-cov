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

func NewPropertyState(state PropState, grantMsg string) PropertyState {
	switch state {
	case PropStateCovered:
		return PropertyStateCovered{State: state}
	case PropStateUncovered:
		return PropertyStateUncovered{State: state}
	case PropStateGranted:
		return PropertyStateGranted{State: state, GrantComment: grantMsg}
	default:
		panic("will not happen")
	}
}

type PropertyStateCovered struct {
	State PropState `json:",omitempty"`
}

func (s PropertyStateCovered) IsCovered() bool      { return true }
func (s PropertyStateCovered) IsUncovered() bool    { return false }
func (s PropertyStateCovered) IsGranted() bool      { return false }
func (s PropertyStateCovered) GrantMessage() string { panic("Covered property has no grant message") }

type PropertyStateUncovered struct {
	State PropState `json:",omitempty"`
}

func (s PropertyStateUncovered) IsCovered() bool   { return false }
func (s PropertyStateUncovered) IsUncovered() bool { return true }
func (s PropertyStateUncovered) IsGranted() bool   { return false }
func (s PropertyStateUncovered) GrantMessage() string {
	panic("Uncovered property has no grant message")
}

type PropertyStateGranted struct {
	State        PropState `json:",omitempty"`
	GrantComment string    `json:",omitempty"`
}

func (s PropertyStateGranted) IsCovered() bool      { return false }
func (s PropertyStateGranted) IsUncovered() bool    { return false }
func (s PropertyStateGranted) IsGranted() bool      { return true }
func (s PropertyStateGranted) GrantMessage() string { return s.GrantComment }

type Property struct {
	PropertyState `json:"state"`

	// The schemas of this swagger schemas property
	schema openapispec.Schema

	// The normalized ref of the schema directly defines the concrete schema of this property.
	fromRef *openapispec.Ref

	// The resolved URI refs along the way to this schemas, each is an absolute/normalized reference.
	resolvedRefs map[string]interface{}

	// The URI of the schema file
	swaggerURI string

	isDiscriminator bool

	isPureDiscriminator bool

	childRefSet map[string]struct{} // Key is normalized ref to the direct or indirect child(derived) schemas. Non-nil only if isDiscriminator is true
}

type Properties map[string]*Property // the key is relative property addr string representation

func (p Properties) Merge(o Properties) {
	for k, v := range o {
		p[k] = v
	}
}

func buildProperty(state PropertyState, swaggerURI string, schema openapispec.Schema, resolvedRefs map[string]interface{}, fromRef *openapispec.Ref) (*Property, error) {
	newResolvedRefs := map[string]interface{}{}
	if len(resolvedRefs) != 0 {
		for k, v := range resolvedRefs {
			newResolvedRefs[k] = v
		}
	}

	prop := &Property{
		PropertyState: state,
		schema:        schema,
		resolvedRefs:  newResolvedRefs,
		swaggerURI:    swaggerURI,
		fromRef:       fromRef,
	}

	allAllOfs, err := swaggerAllAllOfs(swaggerURI)
	if err != nil {
		return nil, err
	}
	isDiscriminator, pure, err := prop.checkDiscriminator(swaggerURI, allAllOfs)
	if err != nil {
		return nil, err
	}
	prop.isDiscriminator = isDiscriminator
	prop.isPureDiscriminator = pure

	thisSchemaRef := prop.fromRef.String()

	// Find all the child variants which derives from this property.
	if isDiscriminator {
		childRefSet := map[string]struct{}{thisSchemaRef: {}}
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
		// Remove the this schema ref from the child ref set.
		// The remaining potential base refs represent the complete set of variant schemas' ref.
		if prop.isPureDiscriminator {
			delete(childRefSet, thisSchemaRef)
		}

		if len(childRefSet) != 0 {
			prop.childRefSet = childRefSet
		}
	}

	return prop, nil
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
	normalizedRefURI := normalizedRef.String()

	// If current ref has already been derefed, meaning a cyclic ref is hit, we will return.
	if _, ok := prop.resolvedRefs[normalizedRefURI]; ok {
		return true, nil
	}

	swagger, err := swaggerloader.LoadSwagger(prop.swaggerURI)
	if err != nil {
		return false, err
	}
	schema, err := openapispec.ResolveRefWithBase(swagger, &ref, &openapispec.ExpandOptions{RelativeBase: prop.swaggerURI})
	if err != nil {
		return false, fmt.Errorf("resolve reference %s: %w", ref.String(), err)
	}

	refProp, err := buildProperty(prop.PropertyState, normalizedRef.GetURL().Path, *schema, prop.resolvedRefs, normalizedRef)
	if err != nil {
		return false, err
	}

	// Specifically, we need to avoid track the discriminator base schemas as they will not be exactly referenced
	// (only its variants are referenced). This allows us to reference the base schema's properties in the
	// `allOf` from the variant's schema.
	if !refProp.isDiscriminator || refProp.isLeafVariant() {
		refProp.resolvedRefs[normalizedRefURI] = struct{}{}
	}
	*prop = *refProp

	return prop.expandRefPropertyInPlace()
}

func (prop *Property) isLeafVariant() bool {
	if prop.childRefSet == nil {
		return false
	}
	if prop.isPureDiscriminator {
		return false
	}
	return len(prop.childRefSet) == 1
}

// checkDiscriminator tells whether the schema corresponds to this property is a discriminator.
// Note that in Azure rest api spec (OpenAPI 2), a schema is a discriminator when itself directly defines "discriminator" or it is a derived schema.
// (Since derived schema also inherits the "discriminator" property)
// This function goes backwards the "allOf" chain to find the nearest parent who explicitly defines the discriminator. If it finds any such parent schema,
// `is` is `true`. Additionally, if the property itself corresponds to a direct discriminator, `pure` is set to `true`.
func (prop *Property) checkDiscriminator(swaggerURI string, allAllOfs map[string]map[string]struct{}) (is bool, pure bool, err error) {
	swagger, err := swaggerloader.LoadSwagger(swaggerURI)
	if err != nil {
		return false, false, err
	}

	thisSchemaRef := prop.fromRef.String()
	if prop.schema.Discriminator != "" {
		return true, true, nil
	}

	baseRefSet := map[string]struct{}{thisSchemaRef: {}}
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
			schema, err := openapispec.ResolveRefWithBase(swagger, &apiRef, &openapispec.ExpandOptions{RelativeBase: prop.swaggerURI})
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

func (props Properties) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{}
	for k, prop := range props {
		mm := map[string]interface{}{}
		m[k] = mm
		switch state := prop.PropertyState.(type) {
		case PropertyStateCovered:
			mm["state"] = state.State
		case PropertyStateUncovered:
			mm["state"] = state.State
		case PropertyStateGranted:
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
