package workspace

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/magodo/azure-rest-api-cov/pkg/core/swagger"

	openapispec "github.com/go-openapi/spec"
	"github.com/magodo/azure-rest-api-cov/pkg/core/swaggerloader"
)

const schemaFolderPatternStr = `^([^%[1]s]+(%[1]sresource-manager(%[1]sMicrosoft.\w+(%[1]s(preview|stable)(%[1]s([^%[1]s]+))?)?)?)?)?$`
const schemaPatternStr = `^([^%[1]s]+)%[1]sresource-manager%[1]sMicrosoft.\w+%[1]s(preview|stable)%[1]s(.+?)%[1]s.+?\.json$`

var (
	ErrFailedToLocateSwagger = errors.New("can't locate swagger spec")
)

type OperationKind string

const (
	OperationKindGet   OperationKind = "GET"
	OperationKindPut                 = "PUT"
	OperationKindPost                = "POST"
	OperationKindPatch               = "PATCH"
)

type SpecLocator struct {
	ApiVersion string `json:"api_version"`
	ApiPath    string `json:"api_path"`
}

type SchemaLocator struct {
	SpecLocator
	Operation OperationKind `json:"operation"`
}

type AzureRepoIndex struct {
	SpecRootDir       string                 `json:"spec_root_dir"`
	Commit            string                 `json:"commit"`
	ResourceProviders ResourceProviders      `json:"resource_providers"`
	specLookupView    map[SpecLocator]string // value is the spec rel path
}

type ResourceProviders map[ResourceProvider]APIVersions

type ResourceProvider string

type APIVersions map[APIVersion]Specs

type APIVersion string

type Specs map[SpecPath]APIPaths

type SpecPath string

type APIPaths map[APIPath]APIOperations

type APIPath string

type APIOperations map[OperationKind]SchemaInfo

type SchemaInfo struct {
	Name        string `json:"name"`
	SpecRelPath string `json:"spec_rel_path"` // The relative path to the spec that define this schema, which might be different that the SpecPath (defines the API Path)
}

func NewAzureRepoIndex(specRootDir string, commit string) (*AzureRepoIndex, error) {
	index := &AzureRepoIndex{
		SpecRootDir:       specRootDir,
		Commit:            commit,
		ResourceProviders: ResourceProviders{},
	}
	schemaFolderPattern := regexp.MustCompile(fmt.Sprintf(schemaFolderPatternStr, regexp.QuoteMeta(string(os.PathSeparator))))
	schemaPattern := regexp.MustCompile(fmt.Sprintf(schemaPatternStr, regexp.QuoteMeta(string(os.PathSeparator))))
	if err := filepath.Walk(specRootDir,
		func(p string, info os.FileInfo, err error) error {
			relPath := strings.TrimPrefix(strings.TrimPrefix(p, specRootDir), string(os.PathSeparator))
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

			rp := ResourceProvider(match[1])

			if err := index.recordSpec(SpecPath(relPath), rp); err != nil {
				return err
			}
			return nil
		}); err != nil {
		return nil, err
	}
	return index, nil
}

func (index *AzureRepoIndex) recordSpec(specPathDefAPIPath SpecPath, rp ResourceProvider) error {
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
		apiPath := APIPath(path)
		operations := APIOperations{}
		apiPaths[apiPath] = operations

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

func getSchemaInfosFromAPIPath(specRootPath, specRelPathDefAPI string, apiPath APIPath, pathItem openapispec.PathItem) (map[OperationKind]SchemaInfo, error) {
	specAbsPathDefAPI := filepath.Join(specRootPath, specRelPathDefAPI)
	swg, err := swaggerloader.LoadSwagger(specAbsPathDefAPI)
	if err != nil {
		return nil, err
	}

	ops := map[OperationKind]*openapispec.Operation{
		OperationKindGet:   pathItem.Get,
		OperationKindPut:   pathItem.Put,
		OperationKindPost:  pathItem.Post,
		OperationKindPatch: pathItem.Patch,
	}
	infos := map[OperationKind]SchemaInfo{}
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
		default:
			return nil, fmt.Errorf("invalid operation kind %q for api path %q", kind, apiPath)
		}
		if err != nil {
			return nil, err
		}
		infos[kind] = *info
	}
	return infos, nil
}

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

		name, refPath := swagger.SchemaInfoFromRef(param.Schema.Ref)
		relPath := specRelPathDefAPI
		if refPath != "" {
			if filepath.IsAbs(refPath) {
				return nil, fmt.Errorf("un-expected absolute swagger path appears in json ref: %s", refPath)
			}
			relPath = filepath.Clean(filepath.Join(filepath.Dir(relPath), refPath))
		}
		// The name might be empty string for those swg with inlined schema definitions. Currently, we don't support schema definitions with no name
		// (since SDK will not generate corresponding data model).
		// E.g. The in-body parameter of "ApiOperationsPolicy_CreateOrUpdate" in: specification/apimanagement/resource-manager/Microsoft.ApiManagement/stable/2016-10-10/apimapis.json.
		return &SchemaInfo{
			Name:        name,
			SpecRelPath: relPath,
		}, nil
	}
	return nil, fmt.Errorf("no in-body parameter found for operation %q of api path %q", operation.Description, apiPath)
}

func getSchemaInfoFromResponse(swg *openapispec.Swagger, specRootPath, specRelPathDefAPI string, apiPath APIPath, operation openapispec.Operation) (*SchemaInfo, error) {
	specAbsPathDefAPI := filepath.Join(specRootPath, specRelPathDefAPI)
	resp, ok := operation.Responses.ResponsesProps.StatusCodeResponses[200]
	if !ok {
		return nil, fmt.Errorf("can't find response definition for code 200 in operation %q of api path %q", operation.Description, apiPath)
	}
	// The resp is a ref, we simply deref it for once.
	if resp.Ref.String() != "" {
		refRespones, err := openapispec.ResolveResponseWithBase(swg, resp.Ref, &openapispec.ExpandOptions{RelativeBase: specAbsPathDefAPI})
		if err != nil {
			return nil, fmt.Errorf("resolve response ref %q: %v", resp.Ref.String(), err)
		}
		resp = *refRespones
	}

	name, refPath := swagger.SchemaInfoFromRef(resp.Schema.Ref)
	relPath := specRelPathDefAPI
	if refPath != "" {
		if filepath.IsAbs(refPath) {
			return nil, fmt.Errorf("un-expected absolute swagger path appears in json ref: %s", refPath)
		}
		relPath = filepath.Clean(filepath.Join(filepath.Dir(relPath), refPath))
	}
	return &SchemaInfo{
		Name:        name,
		SpecRelPath: relPath,
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
	index.specLookupView = map[SpecLocator]string{}
	ambiguousLocators := map[SpecLocator]struct{}{}
	for _, versions := range index.ResourceProviders {
		for version, specs := range versions {
			for specPath, apiPaths := range specs {
				for apiPath := range apiPaths {
					locator := SpecLocator{
						ApiVersion: string(version),
						ApiPath:    string(apiPath),
					}
					// The locator should end up with one and only one match. If there are ambiguous ones,
					// we will just ignore them for now.
					if _, ok := index.specLookupView[locator]; ok {
						ambiguousLocators[locator] = struct{}{}
						continue
					}
					index.specLookupView[locator] = string(specPath)
				}
			}
		}
	}
	// For those ambiguous locators, we have to avoid process them.
	for loc := range ambiguousLocators {
		log.Printf("[WARNING] removing ambiguous spec locator: (%q, %q)", loc.ApiPath, loc.ApiVersion)
		delete(index.specLookupView, loc)
	}
	return
}

func (index *AzureRepoIndex) SpecLookupView() map[SpecLocator]string {
	if index.specLookupView != nil {
		return index.specLookupView
	}
	index.buildSpecLookupView()
	return index.specLookupView
}

// LocateSpec locates the Swagger Spec that defines the API path of the API version as defined in the locator.
// On success, it returns the relative path to the swagger spec.
func (index *AzureRepoIndex) LocateSpec(locator SpecLocator) (string, error) {
	p, ok := index.SpecLookupView()[locator]
	if !ok {
		return "", fmt.Errorf("can't locate swagger spec by (%q, %q), this might because this tuple ends up with multiple swagger specs: %w",
			locator.ApiVersion, locator.ApiPath, ErrFailedToLocateSwagger)
	}
	return p, nil
}

// LocateSchema locates the schema name and the Swagger Spec that defines the "in-body" parameter (for PUT/POST/PATCH operations)
// or 200 response (for GET operation) on the API path of the API version as defined in the locator.
// The spec path might be different that the one that defines the API path.
// On success, it returns the schema name and the relative path to the swagger spec.
func (index *AzureRepoIndex) LocateSchema(locator SchemaLocator) (schemaName, swaggerRelPath string, err error) {
	swaggerRelPathDefApiPath, err := index.LocateSpec(locator.SpecLocator)
	if err != nil {
		return "", "", err
	}

	swaggerPathDefApiPath := filepath.Join(index.SpecRootDir, swaggerRelPathDefApiPath)

	swg, err := swaggerloader.LoadSwagger(swaggerPathDefApiPath)
	if err != nil {
		return "", "", err
	}

	pathItem := swg.Paths.Paths[locator.ApiPath]
	schemaInfos, err := getSchemaInfosFromAPIPath(index.SpecRootDir, swaggerRelPathDefApiPath, APIPath(locator.ApiPath), pathItem)
	if err != nil {
		return "", "", err
	}

	info, ok := schemaInfos[locator.Operation]
	if !ok {
		return "", "", fmt.Errorf("no schema info found for schema locator: %v", locator)
	}

	return info.Name, info.SpecRelPath, nil
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
