package utils

import (
	"strings"
)

// NormalizeAPIPath converts path from "/foos/{name}" to "/FOOS/{}"
func NormalizeAPIPath(p string) string {
	segs := strings.Split(p, "/")
	out := make([]string, 0, len(segs))
	for _, seg := range segs {
		if strings.HasPrefix(seg, "{") && strings.HasSuffix(seg, "}") {
			out = append(out, "{}")
			continue
		}
		out = append(out, seg)
	}
	return strings.ToUpper(strings.Join(out, "/"))
}

// SpecializeAPIPath converts path from "/foos/{fooname}/bars/{barname}" to "/FOOS/FOO1/BARS/{}", given "{fooname}" is an enum whose only value is "foo1", and "{barname}" is variable.
func SpecializeAPIPath(p string, enumMap map[string]string) string {
	segs := strings.Split(p, "/")
	out := make([]string, 0, len(segs))
	for _, seg := range segs {
		if strings.HasPrefix(seg, "{") && strings.HasSuffix(seg, "}") {
			name := seg[1 : len(seg)-1]
			if v, ok := enumMap[name]; ok {
				out = append(out, v)
			} else {
				out = append(out, "{}")
			}
			continue
		}
		out = append(out, seg)
	}
	return strings.ToUpper(strings.Join(out, "/"))
}
