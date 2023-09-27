package utils

import (
	"fmt"
	"strings"

	openapispec "github.com/go-openapi/spec"
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
func SpecializeAPIPath(p string, swg *openapispec.Swagger, swaggerAbsPath string, params []openapispec.Parameter) (string, error) {
	pm := map[string]string{}
	for _, param := range params {
		// The param is a ref, we simply deref it for once.
		if param.Ref.String() != "" {
			refParam, err := openapispec.ResolveParameterWithBase(swg, param.Ref, &openapispec.ExpandOptions{RelativeBase: swaggerAbsPath})
			if err != nil {
				return "", fmt.Errorf("resolve param ref %q: %v", param.Ref.String(), err)
			}
			param = *refParam
		}
		if param.In == "path" && param.Required && param.Type == "string" && len(param.Enum) == 1 {
			pm[param.Name] = param.Enum[0].(string)
		}
	}

	segs := strings.Split(p, "/")
	out := make([]string, 0, len(segs))
	for _, seg := range segs {
		if strings.HasPrefix(seg, "{") && strings.HasSuffix(seg, "}") {
			name := seg[1 : len(seg)-1]
			if v, ok := pm[name]; ok {
				out = append(out, v)
			} else {
				out = append(out, "{}")
			}
		}
		out = append(out, seg)
	}
	return strings.ToUpper(strings.Join(out, "/")), nil
}
