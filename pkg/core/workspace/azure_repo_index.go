package workspace

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	openapispec "github.com/go-openapi/spec"
	"github.com/magodo/azure-rest-api-cov/pkg/core/swaggerloader"
)

const apiSep = ":"

type AzureRepoIndex struct {
	SpecRootDir       string            `json:"spec_root_dir"`
	ResourceProviders ResourceProviders `json:"resource_providers"`
	specSetView       map[string]struct{}
}

type ResourceProviders map[ResourceProvider]APIs

type ResourceProvider struct {
	Name string
}

type APIs map[API]APISpecs

type API struct {
	Version   string
	IsPreview bool
}

type APISpecs map[string]APISpec // The key is the spec relative path from spec root dir

type APISpec struct {
	Path           string   `json:"path"`
	APIPaths       []string `json:"api_paths"`
	APIRootSchemas []string `json:"api_put_schemas"`
}

func (index *AzureRepoIndex) buildSpecSetView() {
	index.specSetView = map[string]struct{}{}
	for _, apis := range index.ResourceProviders {
		for _, specs := range apis {
			for _, spec := range specs {
				index.specSetView[spec.Path] = struct{}{}
			}
		}
	}
	return
}

func (index *AzureRepoIndex) SpecSetView() map[string]struct{} {
	if index.specSetView != nil {
		return index.specSetView
	}
	index.buildSpecSetView()
	return index.specSetView
}

func (index *AzureRepoIndex) recordSpec(relPath string, rpName string, api API) error {
	rp := ResourceProvider{Name: rpName}
	apis, ok := index.ResourceProviders[rp]
	if !ok {
		apis = map[API]APISpecs{}
		index.ResourceProviders[rp] = apis
	}

	apiSpecs, ok := apis[api]
	if !ok {
		apis[api] = make(map[string]APISpec)
		apiSpecs = apis[api]
	}

	if _, ok := apiSpecs[relPath]; ok {
		return nil
	}

	specAbsPath := filepath.Join(index.SpecRootDir, relPath)
	spec, err := swaggerloader.LoadSwagger(specAbsPath)
	if err != nil {
		return err
	}
	if spec.Paths == nil {
		return nil
	}

	var apiPaths []string
	for p := range spec.Paths.Paths {
		apiPaths = append(apiPaths, p)
	}

	schemaNameSet := map[string]struct{}{}
	for _, path := range spec.Paths.Paths {
		// We only consider resource contains GET, PUT and DELETE methods.
		if path.Put == nil || path.Delete == nil || path.Get == nil {
			continue
		}
		for _, param := range path.Put.Parameters {
			// The param is a ref, we simply deref it for once.
			if param.Ref.String() != "" {
				refParam, err := openapispec.ResolveParameterWithBase(spec, param.Ref, &openapispec.ExpandOptions{RelativeBase: specAbsPath})
				if err != nil {
					return fmt.Errorf("resolve param ref %q: %v", param.Ref.String(), err)
				}
				param = *refParam
			}

			if param.In != "body" {
				continue
			}

			refString := param.Schema.Ref.GetPointer().String()
			refStringPattern := regexp.MustCompile(`^/definitions/([^/]+)$`)
			matches := refStringPattern.FindStringSubmatch(refString)
			if len(matches) != 2 {
				// This might happen for those spec with inlined schema definitions. Currently, we don't support schema definitions with no name
				// (since SDK will not generate corresponding data model).
				// E.g. The in-body parameter of "ApiOperationsPolicy_CreateOrUpdate" in: specification/apimanagement/resource-manager/Microsoft.ApiManagement/stable/2016-10-10/apimapis.json.
				continue
			}

			schemaNameSet[matches[1]] = struct{}{}
		}
	}
	schemaNames := make([]string, 0, len(schemaNameSet))
	for schemaName := range schemaNameSet {
		schemaNames = append(schemaNames, schemaName)
	}

	apiSpec := APISpec{
		Path:           relPath,
		APIPaths:       apiPaths,
		APIRootSchemas: schemaNames,
	}

	apiSpecs[relPath] = apiSpec
	return nil
}

func (index *AzureRepoIndex) Refresh() error {
	newIndex := AzureRepoIndex{
		SpecRootDir:       index.SpecRootDir,
		ResourceProviders: ResourceProviders{},
	}
	schemaFolderPattern := regexp.MustCompile(fmt.Sprintf(`^(\w+(%[1]sresource-manager(%[1]sMicrosoft.\w+(%[1]s(preview|stable)(%[1]s([^%[1]s]+))?)?)?)?)?$`, regexp.QuoteMeta(string(os.PathSeparator))))
	schemaPattern := regexp.MustCompile(fmt.Sprintf(`^(\w+)%[1]sresource-manager%[1]sMicrosoft.\w+%[1]s(preview|stable)%[1]s(.+?)%[1]s.+?\.json$`, regexp.QuoteMeta(string(os.PathSeparator))))
	if err := filepath.Walk(index.SpecRootDir,
		func(p string, info os.FileInfo, err error) error {
			relPath := strings.TrimPrefix(strings.TrimPrefix(p, index.SpecRootDir), string(os.PathSeparator))
			if err != nil {
				return err
			}
			if info == nil {
				return nil
			}

			// Skip directories not match the schema folder patterns
			if info.IsDir() {
				if !schemaFolderPattern.MatchString(relPath) {
					return filepath.SkipDir
				}
				return nil
			}

			// Skip files not match the schema file patterns
			match := schemaPattern.FindStringSubmatch(relPath)
			if len(match) == 0 {
				return nil
			}

			rpName, apiState, apiVersion := match[1], match[2], match[3]
			rp := ResourceProvider{Name: rpName}
			api := API{
				Version:   apiVersion,
				IsPreview: apiState == "preview",
			}

			if _, ok := index.SpecSetView()[relPath]; ok {
				apis, ok := newIndex.ResourceProviders[rp]
				if !ok {
					apis = APIs{}
					newIndex.ResourceProviders[rp] = apis
				}
				specs, ok := apis[api]
				if !ok {
					specs = APISpecs{}
					apis[api] = specs
				}
				specs[relPath] = index.ResourceProviders[rp][api][relPath]
				return nil
			}

			if err := newIndex.recordSpec(relPath, rpName, api); err != nil {
				return err
			}
			return nil
		}); err != nil {
		return err
	}

	// rebuild views
	newIndex.buildSpecSetView()

	*index = newIndex
	return nil
}

func (rps ResourceProviders) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{}
	for rp, apis := range rps {
		m[rp.String()] = apis
	}
	return json.Marshal(m)
}

func (rps *ResourceProviders) UnmarshalJSON(b []byte) error {
	var m map[string]APIs
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	tmpRps := ResourceProviders{}
	for k, v := range m {
		tmpRps[ParseResourceProvider(k)] = v
	}
	*rps = tmpRps
	return nil
}

func (rp ResourceProvider) String() string {
	return rp.Name
}

func ParseResourceProvider(s string) ResourceProvider {
	return ResourceProvider{Name: s}
}

func (apis APIs) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{}
	for version, spec := range apis {
		m[version.String()] = spec
	}
	return json.Marshal(m)
}

func (apis *APIs) UnmarshalJSON(b []byte) error {
	var m map[string]APISpecs
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	tmpApis := APIs{}
	for k, v := range m {
		version, err := ParseAPI(k)
		if err != nil {
			return err
		}
		tmpApis[*version] = v
	}
	*apis = tmpApis
	return nil
}

func (v API) String() string {
	s := v.Version + apiSep
	if v.IsPreview {
		s += "preview"
	} else {
		s += "stable"
	}
	return s
}

func ParseAPI(s string) (*API, error) {
	p := strings.Split(s, apiSep)
	if len(p) != 2 {
		return nil, fmt.Errorf("malformed API %q, expected: <version>%s<preview|stable>", s, apiSep)
	}
	return &API{
		Version:   p[0],
		IsPreview: p[1] == "preview",
	}, nil
}
