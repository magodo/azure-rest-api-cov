package addr

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

const PropertyAddrSep = "/"
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
	potentialSplits := strings.Split(addr, PropertyAddrSep)
	var splits []string
	var crossProp string
	for i, prop := range potentialSplits {
		p := regexp.MustCompile(`(\\+)$`)
		if m := p.FindStringSubmatch(prop); i != len(potentialSplits)-1 && len(m) == 2 && len(m[1])%2 == 1 {
			crossProp += prop + `/`
			continue
		}
		splits = append(splits, crossProp+prop)
		crossProp = ""
	}
	var props Property
	for _, prop := range splits {
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

	return strings.Join(props, PropertyAddrSep)
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

	findDiscriminatorMarkIndex := func(mark string) (int, error) {
		p1 := regexp.MustCompile(fmt.Sprintf(`(([^%[1]s\\]*)|(\\+))(%[1]s)`, mark))
		matches := p1.FindAllStringSubmatchIndex(addr, -1)
		discStartIdx := -1
		for _, m := range matches {
			idx := -1
			if m[6] != -1 {
				backSlashes := addr[m[6]:m[7]]
				if len(backSlashes)%2 == 0 {
					idx = m[8]
				}
			} else {
				idx = m[8]
			}
			if idx != -1 {
				if discStartIdx != -1 {
					return 0, fmt.Errorf(`invalid property notation: multiple unescaped %ss: %q`,
						addr, mark)
				}
				discStartIdx = idx
			}
		}
		return discStartIdx, nil
	}

	discStartIdx, err := findDiscriminatorMarkIndex(propertyDiscriminatorStartMark)
	if err != nil {
		return PropertySegment{}, err
	}
	discEndIdx, err := findDiscriminatorMarkIndex(propertyDiscriminatorEndMark)
	if err != nil {
		return PropertySegment{}, err
	}
	if discStartIdx == -1 && discEndIdx == -1 {
		return PropertySegment{name: unEscapeProperty(addr)}, nil
	}

	if discStartIdx == -1 {
		return PropertySegment{}, fmt.Errorf(`invalid property notation: no paring discriminator start mark: %q`, addr)
	}
	if discEndIdx == -1 {
		return PropertySegment{}, fmt.Errorf(`invalid property notation: no paring discriminator end mark: %q`, addr)
	}
	if discEndIdx != len(addr)-1 {
		return PropertySegment{}, fmt.Errorf(`invalid property notation: content after discriminator end mark: %q`, addr)
	}
	if discStartIdx >= discEndIdx {
		return PropertySegment{}, fmt.Errorf(`invalid property notation: discriminator start mark (%d) after discriminator end mark (%d): %q`,
			discStartIdx, discEndIdx, addr)
	}

	return PropertySegment{
		name:               unEscapeProperty(unEscapeProperty(addr[:discStartIdx])),
		discriminatorValue: unEscapeProperty(unEscapeProperty(addr[discStartIdx+1 : discEndIdx])),
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
	v := EscapeProperty(prop.name)
	if prop.discriminatorValue != "" {
		v += EncloseVariant(EscapeProperty(prop.discriminatorValue))
	}
	return v
}

func EncloseVariant(v string) string {
	return propertyDiscriminatorStartMark + v + propertyDiscriminatorEndMark
}

func EscapeProperty(v string) string {
	v = strings.ReplaceAll(v, `\`, `\\`)
	v = strings.ReplaceAll(v, propertyDiscriminatorStartMark, `\`+propertyDiscriminatorStartMark)
	v = strings.ReplaceAll(v, propertyDiscriminatorEndMark, `\`+propertyDiscriminatorEndMark)
	v = strings.ReplaceAll(v, PropertyAddrSep, `\`+PropertyAddrSep)
	return v
}

func unEscapeProperty(v string) string {
	p1 := regexp.MustCompile(`([^\\]|^)(\\)` + propertyDiscriminatorStartMark)
	p2 := regexp.MustCompile(`([^\\]|^)(\\)` + propertyDiscriminatorEndMark)
	p3 := regexp.MustCompile(`([^\\]|^)(\\)` + PropertyAddrSep)
	v = p1.ReplaceAllString(v, "$1"+propertyDiscriminatorStartMark)
	v = p2.ReplaceAllString(v, "$1"+propertyDiscriminatorEndMark)
	v = p3.ReplaceAllString(v, "$1"+PropertyAddrSep)
	v = strings.ReplaceAll(v, `\\`, `\`)
	return v
}
