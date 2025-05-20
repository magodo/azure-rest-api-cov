package utils

import (
	"os"
	"strings"
)

// SpecRelPathRP returns the "top" resoruce provider name given the spec relative path.
// The value is the first path segment.
func SpecRelPathRP(p string) string {
	return strings.Split(p, string(os.PathSeparator))[0]
}
