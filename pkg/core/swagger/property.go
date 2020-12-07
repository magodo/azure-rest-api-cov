package swagger

import (
	"encoding/json"
	"fmt"

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

	// The resolved URI refs along the way to this schemas, each is an absolute/normalized reference.
	resolvedRefs map[string]interface{}

	// The URI of the schema file
	swaggerURI string
}

func buildProperty(state PropertyState, schema openapispec.Schema, resolvedRefs map[string]interface{}, schemaURI string) *Property {
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
	}
}

type Properties map[string]*Property // the key is relative property addr string representation

func (p Properties) Merge(o Properties) {
	for k, v := range o {
		p[k] = v
	}
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
