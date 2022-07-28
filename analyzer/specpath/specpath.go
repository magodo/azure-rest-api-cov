package specpath

import (
	"fmt"
	"os"
	"regexp"
)

const SchemaFolderPatternStr = `^([^%[1]s]+(%[1]sresource-manager(%[1]s\w+\.\w+(%[1]s(preview|stable)(%[1]s([^%[1]s]+))?)?)?)?)?$`
const SchemaPatternStr = `^([^%[1]s]+)%[1]sresource-manager%[1]s(\w+\.\w+)%[1]s(preview|stable)%[1]s(.+?)%[1]s(.+?\.json)$`

type Path string

type PathInfo struct {
	ResourceProvider   string
	ResourceProviderMS string
	IsPreview          bool
	Version            string
	SpecName           string
}

func (p Path) PathInfo() (*PathInfo, error) {
	schemaPattern := regexp.MustCompile(fmt.Sprintf(SchemaPatternStr, regexp.QuoteMeta(string(os.PathSeparator))))
	match := schemaPattern.FindStringSubmatch(string(p))
	if len(match) == 0 {
		return nil, fmt.Errorf("path %q is not a swagger spec path", p)
	}
	return &PathInfo{
		ResourceProvider:   match[1],
		ResourceProviderMS: match[2],
		IsPreview:          match[3] == "preview",
		Version:            match[4],
		SpecName:           match[5],
	}, nil
}

// IsSpecFolder returns true if the path is any level of parent folder to the swagger spec.
func (p Path) IsSpecFolder() bool {
	schemaFolderPattern := regexp.MustCompile(fmt.Sprintf(SchemaFolderPatternStr, regexp.QuoteMeta(string(os.PathSeparator))))
	return schemaFolderPattern.MatchString(string(p))
}

func (p Path) IsSpec() bool {
	schemaPattern := regexp.MustCompile(fmt.Sprintf(SchemaPatternStr, regexp.QuoteMeta(string(os.PathSeparator))))
	return schemaPattern.MatchString(string(p))
}

func (info PathInfo) ToPath() Path {
	previewOrStable := "stable"
	if info.IsPreview {
		previewOrStable = "preview"
	}
	return Path(fmt.Sprintf("%[2]s%[1]sresource-manager%[1]s%[3]s%[1]s%[4]s%[1]s%[5]s%[1]s%[6]s",
		string(os.PathSeparator), info.ResourceProvider, info.ResourceProviderMS, previewOrStable, info.Version, info.SpecName))
}
