package addr

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

const propertyAddrSep = "."
const propertyDiscriminatorStartMark = "{"
const propertyDiscriminatorEndMark = "}"

type Property []PropertySegment

func MustParseProperty(addr string) Property {
	a, err := ParseProperty(addr)
	if err != nil {
		panic(err)
	}

	return a
}

func ParseProperty(addr string) (Property, error) {
	if strings.Trim(addr, " ") == "" {
		return nil, nil
	}
	var props Property
	for _, prop := range strings.Split(addr, propertyAddrSep) {
		segment, err := ParsePropertySegment(prop)
		if err != nil {
			return Property{}, err
		}
		props = append(props, segment)
	}
	return props, nil
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
		if addrLastProp.discriminatorValue != "" || oaddrLastProp.discriminatorValue == "" {
			return false
		}
	}

	for i := range addr {
		addrProp := addr[i]
		oaddrProp := oaddr[i]
		if addrProp.name != oaddrProp.name {
			return false
		}

		if addrProp.discriminatorValue != "" && oaddrProp.discriminatorValue == "" ||
			addrProp.discriminatorValue != "" && oaddrProp.discriminatorValue != "" && addrProp.discriminatorValue != oaddrProp.discriminatorValue {
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
		if addrProp.name != oaddrProp.name {
			return false
		}
		if addrProp.discriminatorValue != oaddrProp.discriminatorValue {
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
				discriminatorValue: v,
			},
		}
		return newAddr
	}
	newAddr[len(newAddr)-1].discriminatorValue = v
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

	return strings.Join(props, propertyAddrSep)
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
	name string

	// discriminatorValue is non-empty only when the property is a derived model.
	discriminatorValue string
}

func ParsePropertySegment(addr string) (PropertySegment, error) {
	discriminatorPattern := regexp.MustCompile(fmt.Sprintf(`^(.*)\%s(.+)\%s$`, propertyDiscriminatorStartMark, propertyDiscriminatorEndMark))
	if !strings.HasSuffix(addr, propertyDiscriminatorEndMark) {
		return PropertySegment{name: addr}, nil
	}

	m := discriminatorPattern.FindStringSubmatch(addr)
	if len(m) != 3 {
		return PropertySegment{}, fmt.Errorf(`invalid discriminator property notation: %q (expected format: "prop%svariant%s")`,
			addr, propertyDiscriminatorStartMark, propertyDiscriminatorEndMark)
	}
	return PropertySegment{
		name:               m[1],
		discriminatorValue: m[2],
	}, nil
}

func MustParsePropertySegment(addr string) PropertySegment {
	a, err := ParsePropertySegment(addr)
	if err != nil {
		panic(err)
	}

	return a
}

func (prop PropertySegment) String() string {
	v := prop.name
	if prop.discriminatorValue != "" {
		v += propertyDiscriminatorStartMark + prop.discriminatorValue + propertyDiscriminatorEndMark
	}
	return v
}
