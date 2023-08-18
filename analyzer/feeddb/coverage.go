package feeddb

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/magodo/aztfq/aztfq"
	"github.com/magodo/azure-rest-api-cov/analyzer/operation"
	"github.com/magodo/azure-rest-api-cov/analyzer/propmatch"

	"github.com/magodo/azure-rest-api-cov/analyzer/schemainput"
	"github.com/magodo/azure-rest-api-cov/analyzer/swagger"
	"github.com/magodo/azure-rest-api-cov/analyzer/utils"
	"github.com/magodo/azure-rest-api-cov/analyzer/workspace"
	log "github.com/sirupsen/logrus"
)

type Coverage struct {
	Ref               string            `json:"ref"`
	Comment           string            `json:"comment"`
	ResourceProviders ResourceProviders `json:"resourceProviders"`
	TFLookupTable     aztfq.LookupTable `json:"-"`
}

type ResourceProviders map[string]ResourceProvider

type ResourceProvider struct {
	Name  string   `json:"name"`
	Paths APIPaths `json:"paths"`
}

type APIPaths map[string]APIPath

type APIPath struct {
	Name       string     `json:"name"`
	Operations Operations `json:"operations"`
}

type Operations map[string]Operation

type Operation struct {
	Name     string      `json:"name"`
	Versions APIVersions `json:"versions"`
}

type APIVersions map[string]APIVersion

type APIVersion struct {
	Version   string    `json:"version"`
	IsActive  bool      `json:"isActive"`
	IsPreview bool      `json:"isPreview"`
	Model     *APIModel `json:"model,omitempty"`
}

func NewEmptyCoverage(index *workspace.AzureRepoIndex, comment string, bridge []byte) Coverage {
	coverage := Coverage{
		Ref:               index.Commit(),
		Comment:           comment,
		ResourceProviders: ResourceProviders{},
	}

	for rp, versions := range index.ResourceProviders {
		rpName := string(rp)
		paths := APIPaths{}
		coverage.ResourceProviders[rpName] = ResourceProvider{
			Name:  string(rp),
			Paths: paths,
		}
		for version, specs := range versions {
			versionName := string(version)
			for apiPath, apiPaths := range specs {
				for pathName, ops := range apiPaths {
					if _, ok := paths[pathName]; !ok {
						paths[pathName] = APIPath{
							Name:       pathName,
							Operations: Operations{},
						}
					}

					for op := range ops {
						opName := string(op)
						if _, ok := paths[pathName].Operations[opName]; !ok {
							paths[pathName].Operations[opName] = Operation{
								Name:     opName,
								Versions: APIVersions{},
							}
						}

						pathInfo, _ := apiPath.PathInfo()
						paths[pathName].Operations[opName].Versions[versionName] = APIVersion{
							Version:   versionName,
							IsActive:  false,
							IsPreview: pathInfo.IsPreview,
						}
					}
				}
			}
		}
	}

	if len(bridge) > 0 {
		tfTable, err := aztfq.BuildLookupTable(bridge)
		if err != nil {
			log.Warnf("Failed to build aztfq lookup table: %v", err)
		} else {
			coverage.TFLookupTable = tfTable
		}
	}

	return coverage
}

func (cov *Coverage) AddOperation(opInfo *operation.Info) error {
	rpName := opInfo.RP
	rp, ok := cov.ResourceProviders[rpName]
	if !ok {
		return fmt.Errorf("no resource provider %q found", rpName)
	}
	return rp.AddOperation(opInfo, cov.TFLookupTable)
}

func (rp *ResourceProvider) AddOperation(opInfo *operation.Info, tfLookupTable aztfq.LookupTable) error {
	path, ok := rp.Paths[opInfo.APIPath]
	if !ok {
		return fmt.Errorf("no API path %q found for resource provider %q", opInfo.APIPath, rp.Name)
	}

	op, ok := path.Operations[string(opInfo.Operation)]
	if !ok {
		return fmt.Errorf("no operation %q found for API path %q, resource provider %q", opInfo.Operation, opInfo.APIPath, rp.Name)
	}

	version, ok := op.Versions[opInfo.APIVersion]
	if !ok {
		return fmt.Errorf("no version %q found for operation %q, API path %q, resource provider %q", opInfo.APIVersion, opInfo.Operation, opInfo.APIPath, rp.Name)
	}

	v := APIVersion{
		Version:   version.Version,
		IsActive:  opInfo.IsActive,
		IsPreview: opInfo.IsPreview,
	}

	if opInfo.Model != nil {
		model, err := NewAPIModel(*opInfo.Model)
		if err != nil {
			return err
		}
		v.Model = model
	}

	tfPropTable := map[string][]aztfq.TFResult{}
	if len(tfLookupTable) > 0 {
		tfPropTable = lookupAztfqByPath(tfLookupTable, opInfo.APIPath, opInfo.APIVersion, workspace.OperationKindPut)
	}

	if len(tfLookupTable) > 0 && v.Model != nil {
		v.Model.MapTFAddr(tfPropTable)
	}

	rp.Paths[opInfo.APIPath].Operations[string(opInfo.Operation)].Versions[opInfo.APIVersion] = v

	return nil
}

func (rp ResourceProvider) Copy() ResourceProvider {
	return ResourceProvider{
		Name:  rp.Name,
		Paths: rp.Paths,
	}
}

// GenerateWithUncoveredOperationModels walks through the coverage data and gather the schemas for those uncovered operations. For each RP, merge both covered and uncovered schemas and generate in json format.
// This means for operations that none of the available versions is covered, then all of those versions of this operation will be converted into schema inputs.
// This can be some API path has covered some operations but not the others. This can also happen when some API paths are not covered at all.
func (cov *Coverage) GenerateWithUncoveredOperationModels(idx *workspace.AzureRepoIndex, ignoreSchema []string, allowMismatchPropMatchers []propmatch.PropMatcher) (string, error) {
	log.Info("Generating with uncovered operation models...")
	var ignorePatterns []workspace.OperationLocator
	for _, p := range ignoreSchema {
		loc, err := workspace.NewSchemaLocatorFromPattern(p)
		if err != nil {
			return "", err
		}
		ignorePatterns = append(ignorePatterns, loc)
	}

	for rpName, rp := range cov.ResourceProviders {
		log.Debugf("Generating for %q...", rpName)

		// Set up the work pool
		uncoveredOpInfos := []*operation.Info{}
		wp := utils.NewWorkPool(runtime.NumCPU())
		wp.Run(func(i interface{}) error {
			uncoveredOpInfos = append(uncoveredOpInfos, i.(*operation.Info))
			return nil
		})

		for pathName, path := range rp.Paths {
			for opName, op := range path.Operations {
				// Iterate all the versions to see whether there is an active one, if so it means this operation is covered already, we don't need to expand the uncovered ones.
				// Since in this case we only need to expand the newer API versions, which has been done in the "coverage" stage (with "-predict").
				var hasActive bool
				for _, version := range op.Versions {
					if version.IsActive {
						hasActive = true
						break
					}
				}
				if hasActive {
					continue
				}

				// There is no version covered for this operation, we will need to expand the model of this operation.
			VersionLoop:
				for version := range op.Versions {
					loc := workspace.OperationLocator{
						SpecLocator: workspace.SpecLocator{
							ApiVersion: version,
							ApiPath:    pathName,
						},
						Operation: workspace.OperationKind(opName),
					}

					for _, igloc := range ignorePatterns {
						if igloc.Match(loc) {
							log.Infof("schema %q is ignored", loc)
							continue VersionLoop
						}
					}

					wp.AddTask(func() (interface{}, error) {
						log.Debugf("Expanding uncovered schema for %q", loc)
						opInfo, err := getUncoveredOperation(loc, idx, allowMismatchPropMatchers)
						if err != nil {
							return nil, err
						}
						if opInfo == nil {
							return nil, utils.ErrSkipTask
						}
						return opInfo, nil
					})
				}
			}
		}

		if err := wp.Done(); err != nil {
			return "", err
		}

		// Apply the uncovered schemas to the rp cov, which contained the covered schemas' coverage data.
		for _, opInfo := range uncoveredOpInfos {
			if err := rp.AddOperation(opInfo, aztfq.LookupTable{}); err != nil {
				return "", err
			}
		}
	}

	rpCov, err := json.MarshalIndent(cov, "", "  ")
	return string(rpCov), err
}

// getUncoveredOperation search the index for an uncovered operation. If that operation has a model, further expands it to its end.
// Note that the returned operation info might be nil when the index failed to look up the locator, which indicates that locator is ambiguous in most cases.
func getUncoveredOperation(loc workspace.OperationLocator, idx *workspace.AzureRepoIndex, allowMismatchPropMatchers []propmatch.PropMatcher) (*operation.Info, error) {
	log.Infof("Processing (uncovered) %q", loc)
	opInfo, err := idx.LocateOperation(loc)
	if err != nil {
		if errors.Is(err, workspace.ErrFailedToLocateSwagger) {
			log.Warn(err)
			return nil, nil
		}
		return nil, err
	}

	// Schemaless operations
	if opInfo.Model == nil {
		si := schemainput.SchemaInput{
			OperationLocator: loc,
			Properties:       []string{},
		}
		out, err := si.ToOperationInfo(idx, true, false, allowMismatchPropMatchers)
		if err != nil {
			return nil, fmt.Errorf("converting schema input to opInfo for %q: %v", loc, err)
		}
		return out, nil
	}

	// Operations with schemas
	// In most case we can simpy initialize a schema input here, however, we should take care about the discriminator schemas (i.e., the schema itself is a discriminator)
	schemaRef := swagger.SchemaNameToRef(filepath.Join(idx.SpecRootDir, opInfo.Model.SpecRelPath), opInfo.Model.Name)
	variantRefSet, err := swagger.GetVariantRefSet(schemaRef)
	if err != nil {
		return nil, fmt.Errorf("getting variant ref set: %v", err)
	}

	// The schema is not a discriminator, we can simply convert it to schema input and process on
	if len(variantRefSet) == 0 {
		si := schemainput.SchemaInput{
			OperationLocator: loc,
			Properties:       []string{},
		}
		out, err := si.ToOperationInfo(idx, true, false, allowMismatchPropMatchers)
		if err != nil {
			return nil, fmt.Errorf("converting schema input to opInfo for %q: %v", loc, err)
		}
		return out, nil
	}

	// The schema is a discriminator, we will need to expand all its variants and merge them into one schema.
	var out *operation.Info
	for ref := range variantRefSet {
		si := schemainput.SchemaInput{
			OperationLocator: loc,
			Variant:          ref,
			Properties:       []string{},
		}
		info, err := si.ToOperationInfo(idx, true, false, allowMismatchPropMatchers)
		if err != nil {
			return nil, fmt.Errorf("converting schema input to opInfo for %q: %v", loc, err)
		}
		if out == nil {
			out = info
		} else {
			if err := out.Model.Merge(*info.Model); err != nil {
				return nil, fmt.Errorf("merging schema for %q: %v", loc, err)
			}
		}
	}
	return out, nil
}

func ParseRTbyPath(path string, ignoreSegCnt bool) string {
	idx := strings.LastIndex(path, "/PROVIDERS/")
	if idx == -1 {
		return ""
	}
	path = path[idx+1:]
	segs := strings.Split(path, "/")

	rtSegs := segs[2:]

	if len(rtSegs)%2 != 0 {
		if ignoreSegCnt {
			rtSegs = rtSegs[0 : len(rtSegs)-1]
		} else {
			return ""
		}
	}

	rts := []string{segs[1]}
	for i := 0; i < len(rtSegs); i += 2 {
		rts = append(rts, rtSegs[i])
	}

	return strings.ToUpper(strings.Join(rts, "/"))
}

// lookupAztfqByPath looks up the aztfq lookup table by operation path
// aztfq uses the full name of resource provider such as MICROSOFT.ENTERPRISEKNOWLEDGEGRAPH as index, which is different from the `rpName` in coverage tool
func lookupAztfqByPath(table aztfq.LookupTable, path string, version string, method workspace.OperationKind) (result map[string][]aztfq.TFResult) {
	result = make(map[string][]aztfq.TFResult)
	rtTable, ok := table[ParseRTbyPath(path, method == workspace.OperationKindPut)]
	if !ok {
		return
	}

	versionTable, ok := rtTable[version]
	if !ok {
		return
	}

	return versionTable
}
