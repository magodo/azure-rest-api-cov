package specpath

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Path string

type PathInfo struct {
	ResourceProvider   string
	ResourceProviderMS string
	IsPreview          bool
	Version            string
	SpecName           string
}

// PathInfo returns the PathInfo of the Path, given the Path is a relative path to a swagger spec.
// e.g.
// - compute/resource-manager/Microsoft.Compute/stable/2020-01-01/compute.json
// - compute/resource-manager/Microsoft.Compute/stable/2022-03-02/DiskRP/disk.json
func (p Path) PathInfo() (*PathInfo, error) {
	segs := strings.Split(string(p), string(os.PathSeparator))
	if len(segs) < 6 {
		return nil, fmt.Errorf("swagger spec path expects more nesting level, but %q has only level %d", p, len(segs))
	}
	filename := filepath.Base(string(p))
	if filepath.Ext(filename) != ".json" {
		return nil, fmt.Errorf("expect a json file, got %s", filename)
	}
	return &PathInfo{
		ResourceProvider:   segs[0],
		ResourceProviderMS: segs[2],
		IsPreview:          segs[3] == "preview",
		Version:            segs[4],
		SpecName:           filename,
	}, nil
}

func (info PathInfo) ToPath() Path {
	previewOrStable := "stable"
	if info.IsPreview {
		previewOrStable = "preview"
	}
	return Path(fmt.Sprintf("%[2]s%[1]sresource-manager%[1]s%[3]s%[1]s%[4]s%[1]s%[5]s%[1]s%[6]s",
		string(os.PathSeparator), info.ResourceProvider, info.ResourceProviderMS, previewOrStable, info.Version, info.SpecName))
}
