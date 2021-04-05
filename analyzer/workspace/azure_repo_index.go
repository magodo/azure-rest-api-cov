package workspace

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/magodo/azure-rest-api-cov/analyzer/specpath"

	"github.com/magodo/azure-rest-api-cov/analyzer/swagger"

	openapispec "github.com/go-openapi/spec"
	"github.com/magodo/azure-rest-api-cov/analyzer/swaggerloader"
)

var (
	ErrFailedToLocateSwagger = errors.New("can't locate swagger spec")
	ErrFailedToLocateSchema  = errors.New("can't locate schema")
)

type OperationKind string

const (
	OperationKindGet     OperationKind = "GET"
	OperationKindPut                   = "PUT"
	OperationKindPost                  = "POST"
	OperationKindDelete                = "DELETE"
	OperationKindOptions               = "OPTIONS"
	OperationKindHead                  = "HEAD"
	OperationKindPatch                 = "PATCH"
	OperationKindGlob                  = "*"
)

type SpecLocator struct {
	ApiVersion string `json:"api_version"`
	ApiPath    string `json:"api_path"`
}

type SchemaLocator struct {
	SpecLocator
	Operation OperationKind `json:"operation"`
}

type SchemaLocators []SchemaLocator

func (locators SchemaLocators) Len() int {
	return len(locators)
}

func (locators SchemaLocators) Less(i, j int) bool {
	if locators[i].ApiPath < locators[j].ApiPath {
		return true
	}
	if locators[i].ApiPath > locators[j].ApiPath {
		return false
	}
	if locators[i].ApiVersion < locators[j].ApiVersion {
		return true
	}
	if locators[i].ApiVersion > locators[j].ApiVersion {
		return false
	}
	return locators[i].Operation < locators[j].Operation
}

func (locators SchemaLocators) Swap(i, j int) {
	locators[i], locators[j] = locators[j], locators[i]
}

// NewSchemaLocatorFromPattern create a SchemaLocator from the pattern in the format of: <API Path>:<Operation>:<API Version>
func NewSchemaLocatorFromPattern(p string) (SchemaLocator, error) {
	segments := strings.Split(p, ":")
	if len(segments) != 3 {
		return SchemaLocator{}, fmt.Errorf("malformed format of SchemaLocator pattern %q, expecting: <API Path>:<Operation>:<API Version>", p)
	}

	path, op, ver := segments[0], segments[1], segments[2]
	return SchemaLocator{
		SpecLocator: SpecLocator{
			ApiVersion: ver,
			ApiPath:    path,
		},
		Operation: OperationKind(op),
	}, nil
}

func (loc SchemaLocator) Match(oloc SchemaLocator) bool {
	if loc.Operation != OperationKindGlob {
		if loc.Operation != oloc.Operation {
			return false
		}
	}

	if loc.ApiPath != "*" {
		if loc.ApiPath != oloc.ApiPath {
			return false
		}
	}

	if loc.ApiVersion != "*" {
		if loc.ApiVersion != oloc.ApiVersion {
			return false
		}
	}

	return true
}

type AzureRepoIndex struct {
	SpecRootDir       string                        `json:"spec_root_dir"`
	Commit            string                        `json:"commit"`
	ResourceProviders ResourceProviders             `json:"resource_providers"`
	specLookupView    map[SpecLocator]specpath.Path // value is the spec rel path
}

type ResourceProviders map[ResourceProvider]APIVersions

type ResourceProvider string

type APIVersions map[APIVersion]Specs

type APIVersion string

type Specs map[specpath.Path]APIPaths

type APIPaths map[string]APIOperations

type APIOperations map[OperationKind]*SchemaInfo

type APIPath struct {
	raw string
}

func NewAPIPath(path string) APIPath {
	return APIPath{raw: path}
}

func (p APIPath) String() string {
	return p.raw
}

// Provider returns the first provider component value that is defined in its API path
func (p APIPath) Provider() (string, error) {
	path := p.raw

	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, "/")

	components := strings.Split(path, "/")

	if len(components)%2 != 0 {
		components = components[:len(components)-1]
	}

	for current := 0; current < len(components); current += 2 {
		key := components[current]
		value := components[current+1]

		// Check key/value for empty strings.
		if key == "" || value == "" {
			return "", fmt.Errorf("Key/Value cannot be empty strings. Key: '%s', Value: '%s'", key, value)
		}

		// Catch the subscriptionID before it can be overwritten by another "subscriptions"
		// value in the ID which is the case for the Service Bus subscription resource
		if key == "providers" {
			return value, nil
		}
	}
	return "", fmt.Errorf("API path %q has no `provider` component", path)
}

type SchemaInfo struct {
	Name        string        `json:"name"`
	SpecRelPath specpath.Path `json:"spec_rel_path"` // The relative path to the spec that define this schema, which might be different that the SpecPath (defines the API Path)
}

func NewAzureRepoIndex(specRootDir string, commit string) (*AzureRepoIndex, error) {
	index := &AzureRepoIndex{
		SpecRootDir:       specRootDir,
		Commit:            commit,
		ResourceProviders: ResourceProviders{},
	}
	if err := filepath.Walk(specRootDir,
		func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			relPath := strings.TrimPrefix(strings.TrimPrefix(p, specRootDir), string(os.PathSeparator))
			if info == nil {
				return nil
			}

			path := specpath.Path(relPath)

			// Skip directories not match the schema folder patterns
			if info.IsDir() {
				if !path.IsSpecFolder() {
					return filepath.SkipDir
				}
				return nil
			}

			// Skip files not match the schema file patterns
			pathInfo, err := path.PathInfo()
			if err != nil {
				return nil
			}

			rp := ResourceProvider(pathInfo.ResourceProvider)

			if err := index.recordSpec(specpath.Path(relPath), rp); err != nil {
				return err
			}
			return nil
		}); err != nil {
		return nil, err
	}
	return index, nil
}

func (index *AzureRepoIndex) recordSpec(specPathDefAPIPath specpath.Path, rp ResourceProvider) error {
	apiVersions, ok := index.ResourceProviders[rp]
	if !ok {
		apiVersions = APIVersions{}
		index.ResourceProviders[rp] = apiVersions
	}

	specAbsPath := filepath.Join(index.SpecRootDir, string(specPathDefAPIPath))
	swg, err := swaggerloader.LoadSwagger(specAbsPath)
	if err != nil {
		return err
	}

	apiVersion := APIVersion(swg.Info.Version)

	specs, ok := apiVersions[apiVersion]
	if !ok {
		apiVersions[apiVersion] = Specs{}
		specs = apiVersions[apiVersion]
	}

	if swg.Paths == nil {
		return nil
	}

	apiPaths := APIPaths{}
	specs[specPathDefAPIPath] = apiPaths

	for path, pathItem := range swg.Paths.Paths {
		apiPath := NewAPIPath(path)
		operations := APIOperations{}
		apiPaths[apiPath.String()] = operations

		infos, err := getSchemaInfosFromAPIPath(index.SpecRootDir, string(specPathDefAPIPath), apiPath, pathItem)
		if err != nil {
			return err
		}
		for kind, info := range infos {
			operations[kind] = info
		}
	}

	return nil
}

func getSchemaInfosFromAPIPath(specRootPath, specRelPathDefAPI string, apiPath APIPath, pathItem openapispec.PathItem) (map[OperationKind]*SchemaInfo, error) {
	specAbsPathDefAPI := filepath.Join(specRootPath, specRelPathDefAPI)
	swg, err := swaggerloader.LoadSwagger(specAbsPathDefAPI)
	if err != nil {
		return nil, err
	}

	ops := map[OperationKind]*openapispec.Operation{
		OperationKindGet:     pathItem.Get,
		OperationKindPut:     pathItem.Put,
		OperationKindPost:    pathItem.Post,
		OperationKindDelete:  pathItem.Delete,
		OperationKindOptions: pathItem.Options,
		OperationKindHead:    pathItem.Head,
		OperationKindPatch:   pathItem.Patch,
	}
	infos := map[OperationKind]*SchemaInfo{}
	for kind, op := range ops {
		if op == nil {
			continue
		}

		var (
			info *SchemaInfo
			err  error
		)
		switch kind {
		case OperationKindPost,
			OperationKindPut,
			OperationKindPatch:
			info, err = getSchemaInfoFromParameter(swg, specRootPath, specRelPathDefAPI, apiPath, *op)
		case OperationKindGet:
			info, err = getSchemaInfoFromResponse(swg, specRootPath, specRelPathDefAPI, apiPath, *op)
		case OperationKindDelete, OperationKindHead, OperationKindOptions:
		default:
			return nil, fmt.Errorf("invalid operation kind %q for api path %q", kind, apiPath)
		}
		if err != nil {
			return nil, err
		}

		// For schema-less operations, only set the spec relative path.
		if info == nil {
			info = &SchemaInfo{
				Name:        "",
				SpecRelPath: specpath.Path(specRelPathDefAPI),
			}
		}
		infos[kind] = info
	}
	return infos, nil
}

//  getSchemaInfoFromParameter get the information for in-body parameter if any. Otherwise, it returns nil.
func getSchemaInfoFromParameter(swg *openapispec.Swagger, specRootPath, specRelPathDefAPI string, apiPath APIPath, operation openapispec.Operation) (*SchemaInfo, error) {
	specAbsPathDefAPI := filepath.Join(specRootPath, specRelPathDefAPI)
	for _, param := range operation.Parameters {
		// The param is a ref, we simply deref it for once.
		if param.Ref.String() != "" {
			refParam, err := openapispec.ResolveParameterWithBase(swg, param.Ref, &openapispec.ExpandOptions{RelativeBase: specAbsPathDefAPI})
			if err != nil {
				return nil, fmt.Errorf("resolve param ref %q: %v", param.Ref.String(), err)
			}
			param = *refParam
		}

		if param.In != "body" {
			continue
		}

		if param.Schema == nil {
			return nil, nil
		}

		name, refPath := swagger.SchemaInfoFromRef(param.Schema.Ref)
		if name == "" {
			// The name might be empty string for those swg with inlined schema definitions. Currently, we don't support schema definitions with no name
			// (since SDK will not generate corresponding data model).
			// E.g. The in-body parameter of "ApiOperationsPolicy_CreateOrUpdate" in: specification/apimanagement/resource-manager/Microsoft.ApiManagement/stable/2016-10-10/apimapis.json.
			return nil, nil
		}

		relPath := specRelPathDefAPI
		if refPath != "" {
			if filepath.IsAbs(refPath) {
				return nil, fmt.Errorf("un-expected absolute swagger path appears in json ref: %s", refPath)
			}
			relPath = filepath.Clean(filepath.Join(filepath.Dir(relPath), refPath))
		}
		return &SchemaInfo{
			Name:        name,
			SpecRelPath: specpath.Path(relPath),
		}, nil
	}
	return nil, nil
}

//  getSchemaInfoFromResponse get the information for the response of code 200 if any. Otherwise, it returns nil.
func getSchemaInfoFromResponse(swg *openapispec.Swagger, specRootPath, specRelPathDefAPI string, apiPath APIPath, operation openapispec.Operation) (*SchemaInfo, error) {
	specAbsPathDefAPI := filepath.Join(specRootPath, specRelPathDefAPI)
	resp, ok := operation.Responses.ResponsesProps.StatusCodeResponses[200]
	if !ok {
		return nil, nil
	}

	// The resp is a ref, we simply deref it for once.
	if resp.Ref.String() != "" {
		refRespones, err := openapispec.ResolveResponseWithBase(swg, resp.Ref, &openapispec.ExpandOptions{RelativeBase: specAbsPathDefAPI})
		if err != nil {
			return nil, fmt.Errorf("resolve response ref %q: %v", resp.Ref.String(), err)
		}
		resp = *refRespones
	}

	if resp.Schema == nil {
		return nil, nil
	}

	name, refPath := swagger.SchemaInfoFromRef(resp.Schema.Ref)
	if name == "" {
		return nil, nil
	}
	relPath := specRelPathDefAPI
	if refPath != "" {
		if filepath.IsAbs(refPath) {
			return nil, fmt.Errorf("un-expected absolute swagger path appears in json ref: %s", refPath)
		}
		relPath = filepath.Clean(filepath.Join(filepath.Dir(relPath), refPath))
	}
	return &SchemaInfo{
		Name:        name,
		SpecRelPath: specpath.Path(relPath),
	}, nil
}

func ImportAzureRepoIndex(path string) (*AzureRepoIndex, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var index AzureRepoIndex
	if err := json.Unmarshal(b, &index); err != nil {
		return nil, err
	}
	return &index, nil
}

func (index AzureRepoIndex) Export(path string) error {
	b, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, b, 0644)
}

func (index *AzureRepoIndex) buildSpecLookupView() {
	index.specLookupView = map[SpecLocator]specpath.Path{}
	ambiguousLocators := map[SpecLocator][]specpath.Path{}
	for _, versions := range index.ResourceProviders {
		for version, specs := range versions {
			for specPath, apiPaths := range specs {
				for apiPath := range apiPaths {
					locator := SpecLocator{
						ApiVersion: string(version),
						ApiPath:    apiPath,
					}
					// The locator should end up with one and only one match. If there are ambiguous ones,
					// we will just ignore them for now.
					if existPath, ok := index.specLookupView[locator]; ok {
						if len(ambiguousLocators[locator]) == 0 {
							ambiguousLocators[locator] = []specpath.Path{existPath}
						}
						ambiguousLocators[locator] = append(ambiguousLocators[locator], specPath)
						continue
					}
					index.specLookupView[locator] = specPath
				}
			}
		}
	}
	// For those ambiguous locators, try to automatically find the origin swagger file by comparing the resource provider
	// and the one defined in the path, together with the version.
	for loc, ambigPaths := range ambiguousLocators {
		locProvider, err := NewAPIPath(loc.ApiPath).Provider()
		locVersion := loc.ApiVersion
		if err != nil {
			log.Warnln(err)
			continue
		}

		var ambigPathStrings []string
		var maybeOriginPath []specpath.Path
		for _, ambigPath := range ambigPaths {
			ambigPathStrings = append(ambigPathStrings, string(ambigPath))

			ambigPathInfo, _ := ambigPath.PathInfo() // guaranteed during index build phase
			if strings.EqualFold(ambigPathInfo.ResourceProviderMS, locProvider) &&
				strings.EqualFold(ambigPathInfo.Version, locVersion) {
				maybeOriginPath = append(maybeOriginPath, ambigPath)
			}
		}

		if len(maybeOriginPath) == 1 {
			index.specLookupView[loc] = maybeOriginPath[0]
			continue
		}

		sort.Strings(ambigPathStrings)
		log.Warnf(` removing ambiguous spec locator: (%q, %q), appears in:
%s`, loc.ApiPath, loc.ApiVersion, strings.Join(ambigPathStrings, "\n"))
		delete(index.specLookupView, loc)
	}
	return
}

func (index *AzureRepoIndex) SpecLookupView() map[SpecLocator]specpath.Path {
	if index.specLookupView != nil {
		return index.specLookupView
	}
	index.buildSpecLookupView()
	return index.specLookupView
}

// LocateSpec locates the Swagger Spec that defines the API path of the API version as defined in the locator.
// On success, it returns the relative path to the swagger spec.
func (index *AzureRepoIndex) LocateSpec(locator SpecLocator) (specpath.Path, error) {
	p, ok := index.SpecLookupView()[locator]
	if !ok {
		return "", fmt.Errorf("Can't locate swagger spec by (%q, %q). This might either because this tuple ends up with multiple swagger specs, or the API version you're predicting dones't contain this api path: %w",
			locator.ApiVersion, locator.ApiPath, ErrFailedToLocateSwagger)
	}
	return p, nil
}

// LocateSchema locates the schema info that defines the "in-body" parameter (for PUT/POST/PATCH operations)
// or 200 response (for GET operation) on the API path of the API version as defined in the locator.
// For other operations, it returns a info with an empty name and a repl path to the swagger file defining the api path.
// The spec path might be different that the one that defines the API path.
func (index *AzureRepoIndex) LocateSchema(locator SchemaLocator) (info *SchemaInfo, err error) {
	swaggerRelPathDefApiPath, err := index.LocateSpec(locator.SpecLocator)
	if err != nil {
		return nil, err
	}

	swg, err := swaggerloader.LoadSwagger(filepath.Join(index.SpecRootDir, string(swaggerRelPathDefApiPath)))
	if err != nil {
		return nil, err
	}

	pathItem := swg.Paths.Paths[locator.ApiPath]
	schemaInfos, err := getSchemaInfosFromAPIPath(index.SpecRootDir, string(swaggerRelPathDefApiPath), NewAPIPath(locator.ApiPath), pathItem)
	if err != nil {
		return nil, err
	}

	info, ok := schemaInfos[locator.Operation]
	if !ok {
		return nil, fmt.Errorf("no schema info found for schema locator %v: %w", locator, ErrFailedToLocateSchema)
	}

	return info, nil
}

//func (index *AzureRepoIndex) Refresh() error {
//	newIndex := AzureRepoIndex{
//		SpecRootDir:       index.SpecRootDir,
//		ResourceProviders: ResourceProviders{},
//	}
//	schemaFolderPattern := regexp.MustCompile(fmt.Sprintf(schemaFolderPatternStr, regexp.QuoteMeta(string(os.PathSeparator))))
//	schemaPattern := regexp.MustCompile(fmt.Sprintf(schemaPatternStr, regexp.QuoteMeta(string(os.PathSeparator))))
//	if err := filepath.Walk(index.SpecRootDir,
//		func(p string, info os.FileInfo, err error) error {
//			relPath := strings.TrimPrefix(strings.TrimPrefix(p, index.SpecRootDir), string(os.PathSeparator))
//			if err != nil {
//				return err
//			}
//			if info == nil {
//				return nil
//			}
//
//			// Skip directories not match the schema folder patterns
//			if info.IsDir() {
//				if !schemaFolderPattern.MatchString(relPath) {
//					return filepath.SkipDir
//				}
//				return nil
//			}
//
//			// Skip files not match the schema file patterns
//			match := schemaPattern.FindStringSubmatch(relPath)
//			if len(match) == 0 {
//				return nil
//			}
//
//			rpName, apiState, apiVersion := match[1], match[2], match[3]
//			rp := ResourceProvider{Name: rpName}
//			api := API{
//				Version:   apiVersion,
//				IsPreview: apiState == "preview",
//			}
//
//			if _, ok := index.SpecSetView()[relPath]; ok {
//				apis, ok := newIndex.ResourceProviders[rp]
//				if !ok {
//					apis = APIs{}
//					newIndex.ResourceProviders[rp] = apis
//				}
//				specs, ok := apis[api]
//				if !ok {
//					specs = Specs{}
//					apis[api] = specs
//				}
//				specs[relPath] = index.ResourceProviders[rp][api][relPath]
//				return nil
//			}
//
//			if err := newIndex.recordSpec(relPath, rpName, api); err != nil {
//				return err
//			}
//			return nil
//		}); err != nil {
//		return err
//	}
//
//	// rebuild views
//	newIndex.buildSpecSetView()
//	newIndex.buildSpecLookupView()
//
//	*index = newIndex
//	return nil
//}
