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
}

func buildProperty(state PropertyState, schema openapispec.Schema, resolvedRefs map[string]interface{}, schemaURI string, fromRef *openapispec.Ref) *Property {
	newResolvedRefs := map[string]interface{}{}
	if len(resolvedRefs) != 0 {
		for k, v := range resolvedRefs {
			newResolvedRefs[k] = v
		}
	}
	return &Property{
		PropertyState: state,
		schema:        schema,
		resolvedRefs:  newResolvedRefs,
		swaggerURI:    schemaURI,
		fromRef:       fromRef,
	}
}

type Properties map[string]*Property // the key is relative property addr string representation

func (p Properties) Merge(o Properties) {
	for k, v := range o {
		p[k] = v
	}
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

	// Keep track of the resolved reference to avoid cyclic ref.
	// Specifically, we need to avoid track the discriminator base schemas as they will not be exactly referenced
	// (only its variants are referenced). This allows us to reference the base schema's properties in the
	// `allOf` from the variant's schema.
	if schema.Discriminator == "" {
		prop.resolvedRefs[normalizedRefURI] = struct{}{}
	}

	// update the schemas by the de-referred schemas
	prop.schema = *schema

	// update the fromRef
	prop.fromRef = normalizedRef

	// update the swaggerURI to using the referenced swagger
	prop.swaggerURI = normalizedRef.GetURL().Path

	return prop.expandRefPropertyInPlace()
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
