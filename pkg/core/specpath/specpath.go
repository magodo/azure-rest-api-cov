package specpath

import (
	"fmt"
	"os"
	"regexp"
)

const schemaFolderPatternStr = `^([^%[1]s]+(%[1]sresource-manager(%[1]sMicrosoft.\w+(%[1]s(preview|stable)(%[1]s([^%[1]s]+))?)?)?)?)?$`
const schemaPatternStr = `^([^%[1]s]+)%[1]sresource-manager%[1]sMicrosoft.\w+%[1]s(preview|stable)%[1]s(.+?)%[1]s.+?\.json$`

type Path string

type PathInfo struct {
	ResourceProvider string
	IsPreview        bool
	Version          string
}

func (p Path) PathInfo() (*PathInfo, error) {
	schemaPattern := regexp.MustCompile(fmt.Sprintf(schemaPatternStr, regexp.QuoteMeta(string(os.PathSeparator))))
	match := schemaPattern.FindStringSubmatch(string(p))
	if len(match) == 0 {
		return nil, fmt.Errorf("path %q is not a swagger spec path", p)
	}
	return &PathInfo{
		ResourceProvider: match[1],
		IsPreview:        match[2] == "preview",
		Version:          match[3],
	}, nil
}

// IsSpecFolder returns true if the path is any level of parent folder to the swagger spec.
func (p Path) IsSpecFolder() bool {
	schemaFolderPattern := regexp.MustCompile(fmt.Sprintf(schemaFolderPatternStr, regexp.QuoteMeta(string(os.PathSeparator))))
	return schemaFolderPattern.MatchString(string(p))
}
