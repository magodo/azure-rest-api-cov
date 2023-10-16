package workspace

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/go-openapi/jsonreference"

	"github.com/magodo/azure-rest-api-cov/analyzer/specpath"
	"github.com/magodo/azure-rest-api-cov/analyzer/utils"
	"github.com/magodo/azure-rest-api-index/azidx"

	"github.com/magodo/azure-rest-api-cov/analyzer/swagger"

	openapispec "github.com/go-openapi/spec"
	"github.com/magodo/azure-rest-api-cov/analyzer/swaggerloader"
)

var (
	ErrFailedToLocateSwagger = errors.New("can't locate swagger spec")
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

type OperationLocator struct {
	SpecLocator
	Operation OperationKind `json:"operation"`
}

func (s OperationLocator) String() string {
	return fmt.Sprintf("%s %s %s", s.ApiPath, s.Operation, s.ApiVersion)
}

type SchemaLocators []OperationLocator

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

// NewSchemaLocatorFromPattern create a OperationLocator from the pattern in the format of: <API Path>:<Operation>:<API Version>
func NewSchemaLocatorFromPattern(p string) (OperationLocator, error) {
	segments := strings.Split(p, ":")
	if len(segments) != 3 {
		return OperationLocator{}, fmt.Errorf("malformed format of OperationLocator pattern %q, expecting: <API Path>:<Operation>:<API Version>", p)
	}

	path, op, ver := segments[0], segments[1], segments[2]
	return OperationLocator{
		SpecLocator: SpecLocator{
			ApiVersion: ver,
			ApiPath:    path,
		},
		Operation: OperationKind(op),
	}, nil
}

func (loc OperationLocator) Match(oloc OperationLocator) bool {
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
	SpecRootDir       string
	ResourceProviders ResourceProviders
	index             *azidx.Index
}

type ResourceProviders map[ResourceProvider]APIVersions

type ResourceProvider string

type APIVersions map[APIVersion]Specs

type APIVersion string

type Specs map[specpath.Path]APIPaths

type APIPaths map[string]APIOperations

type APIOperations map[OperationKind]bool

type OperationInfo struct {
	SpecRelPath specpath.Path
	APIVersion  string
	APIPath     string
	Model       *ModelInfo
}

type ModelInfo struct {
	Name        string
	SpecRelPath string // The relative path to the spec that define this schema, which might be different than the SpecPath (defines the API Path)
}

type APIPath struct {
	raw string
}

func NewAPIPath(path string) APIPath {
	return APIPath{raw: path}
}

func (p APIPath) String() string {
	return p.raw
}

func NewAzureRepoIndex(specRootDir string) (*AzureRepoIndex, error) {
	idx, err := azidx.BuildIndex(specRootDir, "")
	if err != nil {
		return nil, err
	}

	index := &AzureRepoIndex{
		index:             idx,
		SpecRootDir:       specRootDir,
		ResourceProviders: buildResourceProvidersFromIndex(idx),
	}

	return index, nil
}

func (idx AzureRepoIndex) Commit() string {
	return idx.index.Commit
}

func buildResourceProvidersFromIndex(idx *azidx.Index) ResourceProviders {
	recordToRPs := func(rps ResourceProviders, version string, method azidx.OperationKind, ref jsonreference.Ref) {
		filepath := specpath.Path(ref.GetURL().Path)
		pinfo, _ := filepath.PathInfo()
		rp := pinfo.ResourceProvider
		path := utils.NormalizeAPIPath(ref.GetPointer().DecodedTokens()[1])

		apiVersions, ok := rps[ResourceProvider(rp)]
		if !ok {
			apiVersions = APIVersions{}
			rps[ResourceProvider(rp)] = apiVersions
		}

		apiSpecs, ok := apiVersions[APIVersion(version)]
		if !ok {
			apiSpecs = Specs{}
			apiVersions[APIVersion(version)] = apiSpecs
		}

		apiPaths, ok := apiSpecs[filepath]
		if !ok {
			apiPaths = APIPaths{}
			apiSpecs[filepath] = apiPaths
		}

		apiOperations, ok := apiPaths[path]
		if !ok {
			apiOperations = APIOperations{}
			apiPaths[path] = apiOperations
		}

		apiOperations[OperationKind(method)] = true
	}

	rps := ResourceProviders{}
	for _, versions := range idx.ResourceProviders {
		for version, methods := range versions {
			for method, rts := range methods {
				for _, opinfo := range rts {
					for _, refs := range opinfo.Actions {
						for _, ref := range refs {
							recordToRPs(rps, version, method, ref)
						}
					}
					for _, ref := range opinfo.OperationRefs {
						recordToRPs(rps, version, method, ref)
					}
				}
			}
		}
	}
	return rps
}

func getOpInfosFromAPIPath(specRootPath, specRelPathDefAPI string, apiPath string, pathItem openapispec.PathItem) (map[OperationKind]*OperationInfo, error) {
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
	infos := map[OperationKind]*OperationInfo{}
	for kind, op := range ops {
		if op == nil {
			continue
		}

		var (
			info *ModelInfo
			err  error
		)
		switch kind {
		case OperationKindPost,
			OperationKindPut,
			OperationKindPatch:
			info, err = getModelInfoFromParameter(swg, specRootPath, specRelPathDefAPI, *op)
		case OperationKindGet:
			info, err = getModelInfoFromResponse(swg, specRootPath, specRelPathDefAPI, *op)
		case OperationKindDelete, OperationKindHead, OperationKindOptions:
		default:
			return nil, fmt.Errorf("invalid operation kind %q for api path %q", kind, apiPath)
		}
		if err != nil {
			return nil, err
		}

		opInfo := &OperationInfo{
			SpecRelPath: specpath.Path(specRelPathDefAPI),
			APIPath:     apiPath,
			Model:       info,
		}
		infos[kind] = opInfo
	}
	return infos, nil
}

// getModelInfoFromParameter get the information for in-body parameter if any. Otherwise, it returns nil.
func getModelInfoFromParameter(swg *openapispec.Swagger, specRootPath, specRelPathDefAPI string, operation openapispec.Operation) (*ModelInfo, error) {
	specAbsPathDefAPI := filepath.Join(specRootPath, specRelPathDefAPI)
	for _, param := range operation.Parameters {
		relPath := specRelPathDefAPI
		// The param is a ref, we simply deref it for once.
		if param.Ref.String() != "" {
			refParam, err := openapispec.ResolveParameterWithBase(swg, param.Ref, &openapispec.ExpandOptions{RelativeBase: specAbsPathDefAPI})
			if err != nil {
				return nil, fmt.Errorf("resolve param ref %q: %v", param.Ref.String(), err)
			}
			// Update the relPath to using the swagger spec which indeed defines that parameter schema.
			_, refPath := swagger.SchemaInfoFromRef(param.Ref)
			if refPath != "" {
				if filepath.IsAbs(refPath) {
					return nil, fmt.Errorf("un-expected absolute swagger path appears in json ref: %s", refPath)
				}
				relPath = filepath.Clean(filepath.Join(filepath.Dir(relPath), refPath))
			}

			// Update the param
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
		if refPath != "" {
			if filepath.IsAbs(refPath) {
				return nil, fmt.Errorf("un-expected absolute swagger path appears in json ref: %s", refPath)
			}
			relPath = filepath.Clean(filepath.Join(filepath.Dir(relPath), refPath))
		}
		return &ModelInfo{
			Name:        name,
			SpecRelPath: relPath,
		}, nil
	}
	return nil, nil
}

// getModelInfoFromResponse get the information for the response of code 200 if any. Otherwise, it returns nil.
func getModelInfoFromResponse(swg *openapispec.Swagger, specRootPath, specRelPathDefAPI string, operation openapispec.Operation) (*ModelInfo, error) {
	specAbsPathDefAPI := filepath.Join(specRootPath, specRelPathDefAPI)
	relPath := specRelPathDefAPI
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

		// Update the relPath to using the swagger spec which indeed defines that response schema.
		_, refPath := swagger.SchemaInfoFromRef(resp.Ref)
		if refPath != "" {
			if filepath.IsAbs(refPath) {
				return nil, fmt.Errorf("un-expected absolute swagger path appears in json ref: %s", refPath)
			}
			relPath = filepath.Clean(filepath.Join(filepath.Dir(relPath), refPath))
		}

		// Update the resp
		resp = *refRespones
	}

	if resp.Schema == nil {
		return nil, nil
	}

	name, refPath := swagger.SchemaInfoFromRef(resp.Schema.Ref)
	if name == "" {
		return nil, nil
	}
	if refPath != "" {
		if filepath.IsAbs(refPath) {
			return nil, fmt.Errorf("un-expected absolute swagger path appears in json ref: %s", refPath)
		}
		relPath = filepath.Clean(filepath.Join(filepath.Dir(relPath), refPath))
	}
	return &ModelInfo{
		Name:        name,
		SpecRelPath: relPath,
	}, nil
}

func ImportAzureRepoIndex(specRootDir, path string) (*AzureRepoIndex, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var idx azidx.Index
	if err := json.Unmarshal(b, &idx); err != nil {
		return nil, err
	}

	index := AzureRepoIndex{
		SpecRootDir:       specRootDir,
		index:             &idx,
		ResourceProviders: buildResourceProvidersFromIndex(&idx),
	}

	return &index, nil
}

func (index *AzureRepoIndex) Export(path string) error {
	b, err := json.MarshalIndent(index.index, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, b, 0644)
}

// LocateOperation locates the schema info that defines the "in-body" parameter (for PUT/POST/PATCH operations)
// or 200 response (for GET operation) on the API path of the API version as defined in the locator.
// For other operations, it returns a info with an empty name and a repl path to the swagger file defining the api path.
// The spec path might be different that the one that defines the API path.
func (index *AzureRepoIndex) LocateOperation(locator OperationLocator) (info *OperationInfo, err error) {
	uRL := url.URL{
		Path:     utils.SpecializeAPIPath(locator.ApiPath),
		RawQuery: "api-version=" + locator.ApiVersion,
	}

	ref, err := index.index.Lookup(string(locator.Operation), uRL)
	if err != nil {
		return nil, fmt.Errorf("%v: %w", err, ErrFailedToLocateSwagger)
	}

	swaggerRelPathDefApiPath := ref.GetURL().Path
	swg, err := swaggerloader.LoadSwagger(filepath.Join(index.SpecRootDir, swaggerRelPathDefApiPath))
	if err != nil {
		return nil, err
	}
	path := ref.GetPointer().DecodedTokens()[1]
	normpath := utils.NormalizeAPIPath(path)
	pathItem := swg.Paths.Paths[path]
	opInfos, err := getOpInfosFromAPIPath(index.SpecRootDir, string(swaggerRelPathDefApiPath), normpath, pathItem)
	if err != nil {
		return nil, fmt.Errorf("get operation information from api path %q: %v", locator.ApiPath, err)
	}

	info, ok := opInfos[locator.Operation]
	if !ok {
		return nil, fmt.Errorf("no schema info found for schema locator %v", locator)
	}

	return info, nil
}
