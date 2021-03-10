package addr

import (
	"errors"
	"fmt"
	"sync"
)

var propertyCache = struct {
	sync.Mutex
	m map[string]Property
}{
	sync.Mutex{},
	map[string]Property{},
}

func ParseProperty(input string) (Property, error) {
	propertyCache.Lock()
	defer propertyCache.Unlock()

	if prop, ok := propertyCache.m[input]; ok {
		return prop, nil
	}

	var prop Property
	var segment PropertySegment
	emptySegment := PropertySegment{}

	l := lex(input)
loop:
	for {
		item := l.nextItem()
		switch item.typ {
		case itemEOF:
			if segment != emptySegment {
				prop = append(prop, segment)
			}
			break loop
		case itemError:
			return nil, errors.New(item.val)
		case itemPropertyDelim:
			prop = append(prop, segment)
			segment = emptySegment
		case itemPropertyName:
			segment.Name = UnescapePropertyName(item.val)
		case itemDiscriminatorValue:
			segment.DiscriminatorValue = UnescapeDiscriminatorValue(item.val)
		}
	}
	propertyCache.m[input] = prop
	return prop, nil
}

func ParsePropertySegment(input string) (PropertySegment, error) {
	var segment PropertySegment

	l := lex(input)

loop:
	for {
		item := l.nextItem()
		switch item.typ {
		case itemEOF:
			break loop
		case itemError:
			return PropertySegment{}, errors.New(item.val)
		case itemPropertyDelim:
			return PropertySegment{}, fmt.Errorf("unexpected property deliminator")
		case itemPropertyName:
			segment.Name = UnescapePropertyName(item.val)
		case itemDiscriminatorValue:
			segment.DiscriminatorValue = UnescapeDiscriminatorValue(item.val)
		}
	}

	return segment, nil
}
