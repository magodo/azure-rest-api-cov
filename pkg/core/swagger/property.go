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
	//GrantMessage() string
	//UnmarshalJSON([]byte) error
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

//func MarshalPropertyState(state PropertyState) ([]byte, error) {
//	m := map[string]interface{}{}
//	switch state := state.(type) {
//	case *PropertyStateCovered:
//		m["state"] = state.State
//	case *PropertyStateUncovered:
//		m["state"] = state.State
//	case *PropertyStateGranted:
//		m["state"] = state.State
//		m["grant_comment"] = state.GrantComment
//	default:
//		panic("will not happen")
//	}
//	return json.Marshal(m)
//}
//
//func UnmarshalPropertyState(s PropertyState, b []byte) error {
//	m := map[string]string{}
//	if err := json.Unmarshal(b, &m); err != nil {
//		return err
//	}
//	state, ok := m["state"]
//	if !ok {
//		return fmt.Errorf(`no "state" in the raw byte`)
//	}
//	pstate := PropState(state)
//	switch s := s.(type) {
//	case *PropertyStateCovered:
//		if pstate != PropStateCovered {
//			return fmt.Errorf(`cannot unmarshall covered state to non-covered PropertyState`)
//		}
//		s.State = pstate
//		return nil
//	case *PropertyStateUncovered:
//		if pstate != PropStateUncovered {
//			return fmt.Errorf(`cannot unmarshall uncovered state to non-uncovered PropertyState`)
//		}
//		s.State = pstate
//		return nil
//	case *PropertyStateGranted:
//		if pstate != PropStateGranted {
//			return fmt.Errorf(`cannot unmarshall granted state to non-granted PropertyState`)
//		}
//		s.State = pstate
//		s.GrantComment = m["grant_comment"]
//		return nil
//	default:
//		panic("will not happen")
//	}
//}

type PropertyStateCovered struct {
	State PropState `json:",omitempty"`
}

func (s PropertyStateCovered) IsCovered() bool      { return true }
func (s PropertyStateCovered) IsUncovered() bool    { return false }
func (s PropertyStateCovered) IsGranted() bool      { return false }
func (s PropertyStateCovered) GrantMessage() string { panic("Covered property has no grant message") }

//func (s *PropertyStateCovered) MarshalJSON() ([]byte, error) { return MarshalPropertyState(s) }
//func (s *PropertyStateCovered) UnmarshalJSON(b []byte) error { return UnmarshalPropertyState(s, b) }

type PropertyStateUncovered struct {
	State PropState `json:",omitempty"`
}

func (s PropertyStateUncovered) IsCovered() bool   { return false }
func (s PropertyStateUncovered) IsUncovered() bool { return true }
func (s PropertyStateUncovered) IsGranted() bool   { return false }
func (s PropertyStateUncovered) GrantMessage() string {
	panic("Uncovered property has no grant message")
}

//func (s *PropertyStateUncovered) MarshalJSON() ([]byte, error) { return MarshalPropertyState(s) }
//func (s *PropertyStateUncovered) UnmarshalJSON(b []byte) error { return UnmarshalPropertyState(s, b) }

type PropertyStateGranted struct {
	State        PropState `json:",omitempty"`
	GrantComment string    `json:",omitempty"`
}

func (s PropertyStateGranted) IsCovered() bool      { return false }
func (s PropertyStateGranted) IsUncovered() bool    { return false }
func (s PropertyStateGranted) IsGranted() bool      { return true }
func (s PropertyStateGranted) GrantMessage() string { return s.GrantComment }

//func (s *PropertyStateGranted) MarshalJSON() ([]byte, error) { return MarshalPropertyState(s) }
//func (s *PropertyStateGranted) UnmarshalJSON(b []byte) error { return UnmarshalPropertyState(s, b) }

type Property struct {
	PropertyState `json:"state"`

	// The schemas of this swagger schemas property
	schema openapispec.Schema

	// The resolved URI refs along the way to this schemas, each is an absolute/normalized reference.
	resolvedRefs map[string]interface{}

	// The URI of the schema file
	swaggerURI string
}

func NewProperty(state PropertyState, schema openapispec.Schema, resolvedRefs map[string]interface{}, schemaURI string) *Property {
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

func NewProperties() Properties {
	return map[string]*Property{}
}

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
