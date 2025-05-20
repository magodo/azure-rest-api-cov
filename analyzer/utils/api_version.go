package utils

import "strings"

func ApiVersionIsPreview(v string) bool {
	return strings.HasSuffix(v, "preview")
}
