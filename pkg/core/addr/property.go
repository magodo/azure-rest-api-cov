package addr

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

const propertySchemaSep = ":"
const propertyAddrSep = "."
const propertyDiscriminatorStartMark = "{"
const propertyDiscriminatorEndMark = "}"

type Property struct {
	Schema   string
	Property RelProperty
}

func MustParseProperty(addr string) Property {
	a, err := ParseProperty(addr)
	if err != nil {
		panic(err)
	}

	return a
}

func ParseProperty(addr string) (Property, error) {
	if strings.Trim(addr, " ") == "" {
		return Property{}, nil
	}

	p := strings.Split(addr, propertySchemaSep)
	var (
		schemaName string
		propAddr   string
	)
	switch len(p) {
	case 1:
		propAddr = p[0]
	case 2:
		schemaName = p[0]
		propAddr = p[1]
	default:
		return Property{}, fmt.Errorf("invalid Swagger Property Address: %s", addr)
	}

	props, err := ParseRelProperty(propAddr)
	if err != nil {
		return Property{}, err
	}

	return Property{
		Schema:   schemaName,
		Property: props,
	}, nil
}

func NewProperty(schemaName string, relProp RelProperty) Property {
	return Property{
		Schema:   schemaName,
		Property: relProp,
	}
}

func (addr Property) Contains(oaddr Property) bool {
	if addr.Schema != oaddr.Schema {
		return false
	}

	if len(addr.Property) > len(oaddr.Property) {
		return false
	}

	if len(addr.Property) == len(oaddr.Property) {
		lastIdx := len(addr.Property) - 1
		if lastIdx == -1 {
			return false
		}
		addrLastProp := addr.Property[lastIdx]
		oaddrLastProp := oaddr.Property[lastIdx]
		if addrLastProp.discriminatorValue != "" || oaddrLastProp.discriminatorValue == "" {
			return false
		}
	}

	for i := range addr.Property {
		addrProp := addr.Property[i]
		oaddrProp := oaddr.Property[i]
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
	if addr.Schema != oaddr.Schema {
		return false
	}

	if len(addr.Property) != len(oaddr.Property) {
		return false
	}

	for i := range addr.Property {
		addrProp := addr.Property[i]
		oaddrProp := oaddr.Property[i]
		if addrProp.name != oaddrProp.name {
			return false
		}
		if addrProp.discriminatorValue != oaddrProp.discriminatorValue {
			return false
		}
	}

	return true
}

func (addr Property) Append(oprop RelProperty) Property {
	if len(oprop) == 0 {
		return addr
	}

	newPropAddrs := make(RelProperty, len(addr.Property)+len(oprop))
	copy(newPropAddrs, addr.Property)
	copy(newPropAddrs[len(addr.Property):], oprop)
	return Property{
		Schema:   addr.Schema,
		Property: newPropAddrs,
	}
}

// AsVariant will create a new Property and turn it into the specified discriminator variant.
func (addr Property) AsVariant(v string) Property {
	newAddr := addr.Copy()
	if len(newAddr.Property) == 0 {
		newAddr.Property = RelProperty{
			{
				discriminatorValue: v,
			},
		}
		return newAddr
	}
	newAddr.Property[len(newAddr.Property)-1].discriminatorValue = v
	return newAddr
}

// Copy makes a deep copy on the Property.
func (addr Property) Copy() Property {
	newAddr := Property{
		Schema:   addr.Schema,
		Property: make(RelProperty, len(addr.Property)),
	}
	copy(newAddr.Property, addr.Property)
	return newAddr
}

func (addr Property) String() string {
	if addr.Schema == "" {
		return addr.Property.String()
	}
	return addr.Schema + propertySchemaSep + addr.Property.String()
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

type RelProperty []PropertySegment

func (addr RelProperty) String() string {
	props := []string{}
	for _, prop := range addr {
		props = append(props, prop.String())
	}

	return strings.Join(props, propertyAddrSep)
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

func ParseRelProperty(addr string) (RelProperty, error) {
	if strings.Trim(addr, " ") == "" {
		return nil, nil
	}
	var props RelProperty
	for _, prop := range strings.Split(addr, propertyAddrSep) {
		segment, err := ParsePropertySegment(prop)
		if err != nil {
			return RelProperty{}, err
		}
		props = append(props, segment)
	}
	return props, nil
}

func MustParseRelProperty(addr string) RelProperty {
	a, err := ParseRelProperty(addr)
	if err != nil {
		panic(err)
	}

	return a
}
