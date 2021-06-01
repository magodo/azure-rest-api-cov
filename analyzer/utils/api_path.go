package utils

import "strings"

func NormalizeAPIPath(p string) string {
	segs := strings.Split(p, "/")
	out := make([]string, 0, len(segs))
	for _, seg := range segs {
		if strings.HasPrefix(seg, "{") && strings.HasSuffix(seg, "}") {
			out = append(out, "{}")
			continue
		}
		out = append(out, strings.ToUpper(seg))
	}
	return strings.Join(out, "/")
}
