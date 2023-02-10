package propmatch

import (
	"fmt"
	"strings"

	"github.com/magodo/azure-rest-api-cov/analyzer/addr"
)

func ParsePropMatcher(s string) (*PropMatcher, error) {
	segs := strings.Split(s, ":")
	if len(segs) != 3 {
		return nil, fmt.Errorf("expect 3 segments, but got %d", len(segs))
	}
	propAddr, err := addr.ParseProperty(segs[2])
	if err != nil {
		return nil, err
	}
	return &PropMatcher{
		SwaggerURI: segs[0],
		ModelName:  segs[1],
		PropAddr:   propAddr,
	}, nil
}

func MustParsePropMatcher(s string) PropMatcher {
	m, err := ParsePropMatcher(s)
	if err != nil {
		panic(err)
	}
	return *m
}

type PropMatcher struct {
	SwaggerURI string
	ModelName  string
	PropAddr   addr.Property
}

func (m PropMatcher) Match(o PropMatcher) bool {
	return m.SwaggerURI == o.SwaggerURI && m.ModelName == o.ModelName && propMatch(m.PropAddr, o.PropAddr)
}

const (
	wildcardPropSegment    = "*"
	anyWildcardPropSegment = "**"
)

func propMatch(prop, propPattern addr.Property) bool {
	var idxPat, idxProp, nextIdxPat, nextIdxProp int
	for idxProp < len(prop) {
		if idxPat < len(propPattern) {
			seg := propPattern[idxPat]
			if seg.Name == "*" || (idxProp < len(prop) && seg == prop[idxProp]) {
				idxPat++
				idxProp++
				continue
			}

			if seg.Name == "**" {
				nextIdxPat, nextIdxProp = idxPat, idxProp+1
				idxPat++
				continue
			}
		}

		if nextIdxProp > 0 {
			idxPat, idxProp = nextIdxPat, nextIdxProp
			continue
		}
		return false
	}
	for _, seg := range propPattern[idxPat:] {
		if seg.Name != "**" {
			return false
		}
	}
	return true
}
