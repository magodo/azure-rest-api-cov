package addr

import (
	"encoding/json"
	"strings"
)

type Property []PropertySegment

func MustParseProperty(addr string) Property {
	a, err := ParseProperty(addr)
	if err != nil {
		panic(err)
	}

	return a
}

func (addr Property) Contains(oaddr Property) bool {
	if len(addr) > len(oaddr) {
		return false
	}

	if len(addr) == len(oaddr) {
		lastIdx := len(addr) - 1
		if lastIdx == -1 {
			return false
		}
		addrLastProp := addr[lastIdx]
		oaddrLastProp := oaddr[lastIdx]
		if addrLastProp.DiscriminatorValue != "" || oaddrLastProp.DiscriminatorValue == "" {
			return false
		}
	}

	for i := range addr {
		addrProp := addr[i]
		oaddrProp := oaddr[i]
		if addrProp.Name != oaddrProp.Name {
			return false
		}

		if addrProp.DiscriminatorValue != "" && oaddrProp.DiscriminatorValue == "" ||
			addrProp.DiscriminatorValue != "" && oaddrProp.DiscriminatorValue != "" && addrProp.DiscriminatorValue != oaddrProp.DiscriminatorValue {
			return false
		}
	}

	return true
}

func (addr Property) Equals(oaddr Property) bool {
	if len(addr) != len(oaddr) {
		return false
	}

	for i := range addr {
		addrProp := addr[i]
		oaddrProp := oaddr[i]
		if addrProp.Name != oaddrProp.Name {
			return false
		}
		if addrProp.DiscriminatorValue != oaddrProp.DiscriminatorValue {
			return false
		}
	}

	return true
}

func (addr Property) Append(oprop Property) Property {
	if len(oprop) == 0 {
		return addr
	}

	newProp := make(Property, len(addr)+len(oprop))
	copy(newProp, addr)
	copy(newProp[len(addr):], oprop)
	return newProp
}

// AsVariant will create a new Property and turn it into the specified discriminator variant.
func (addr Property) AsVariant(v string) Property {
	newAddr := addr.Copy()
	if len(newAddr) == 0 {
		newAddr = Property{
			{
				DiscriminatorValue: v,
			},
		}
		return newAddr
	}
	newAddr[len(newAddr)-1].DiscriminatorValue = v
	return newAddr
}

// Copy makes a deep copy on the Property.
func (addr Property) Copy() Property {
	newAddr := make(Property, len(addr))
	copy(newAddr, addr)
	return newAddr
}

func (addr Property) String() string {
	props := []string{}
	for _, prop := range addr {
		props = append(props, prop.String())
	}

	return strings.Join(props, string(propertyDelim))
}

func (addr Property) MarshalJSON() ([]byte, error) {
	return json.Marshal(addr.String())
}

func (addr *Property) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	var err error
	*addr, err = ParseProperty(s)
	return err
}

type PropertySegment struct {
	Name string

	// DiscriminatorValue is non-empty only when the property is a derived model.
	DiscriminatorValue string
}

func MustParsePropertySegment(addr string) PropertySegment {
	a, err := ParsePropertySegment(addr)
	if err != nil {
		panic(err)
	}

	return a
}

func (prop PropertySegment) String() string {
	v := EscapePropertyName(prop.Name)
	if prop.DiscriminatorValue != "" {
		v += EncloseVariant(EscapeDiscriminatorValue(prop.DiscriminatorValue))
	}
	return v
}

func EncloseVariant(v string) string {
	return string(leftDiscriminatorDelim) + v + string(rightDiscriminatorDelim)
}

func EncloseProperties(props []string) string {
	return strings.Join(props, string(propertyDelim))
}

func EscapeDiscriminatorValue(v string) string {
	v = strings.ReplaceAll(v, string(escapeCharacter), strings.Repeat(string(escapeCharacter), 2))
	v = strings.ReplaceAll(v, string(rightDiscriminatorDelim), string(escapeCharacter)+string(rightDiscriminatorDelim))
	return v
}

func UnescapeDiscriminatorValue(v string) string {
	v = strings.ReplaceAll(v, string(escapeCharacter)+string(rightDiscriminatorDelim), string(rightDiscriminatorDelim))
	v = strings.ReplaceAll(v, strings.Repeat(string(escapeCharacter), 2), string(escapeCharacter))
	return v
}

func EscapePropertyName(v string) string {
	v = strings.ReplaceAll(v, string(escapeCharacter), strings.Repeat(string(escapeCharacter), 2))
	v = strings.ReplaceAll(v, string(leftDiscriminatorDelim), string(escapeCharacter)+string(leftDiscriminatorDelim))
	v = strings.ReplaceAll(v, string(propertyDelim), string(escapeCharacter)+string(propertyDelim))
	return v
}

func UnescapePropertyName(v string) string {
	v = strings.ReplaceAll(v, string(escapeCharacter)+string(leftDiscriminatorDelim), string(leftDiscriminatorDelim))
	v = strings.ReplaceAll(v, string(escapeCharacter)+string(propertyDelim), string(propertyDelim))
	v = strings.ReplaceAll(v, strings.Repeat(string(escapeCharacter), 2), string(escapeCharacter))
	return v
}
