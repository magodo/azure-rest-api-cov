package coverage

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/magodo/azure-rest-api-cov/pkg/core/specpath"

	"github.com/magodo/azure-rest-api-cov/pkg/core/swagger"

	"github.com/magodo/azure-rest-api-cov/pkg/core/workspace"
)

type RepoCoverageStore struct {
	Commit            string            `json:"commit"`
	ResourceProviders ResourceProviders `json:"resource_providers"`
}

type ResourceProviders map[ResourceProvider]Operations

type ResourceProvider string

type Operations map[Operation]APIVersions

type Operation struct {
	APIPath       string                  `json:"api_path"`
	OperationKind workspace.OperationKind `json:"operation_kind"`
}

type APIVersions map[APIVersion]*SchemaCoverageStore

type APIVersion string

func NewEmptyRepoCoverageStore(index workspace.AzureRepoIndex) RepoCoverageStore {
	coverage := RepoCoverageStore{
		Commit:            index.Commit,
		ResourceProviders: ResourceProviders{},
	}

	for rp, versions := range index.ResourceProviders {
		covOps := Operations{}
		coverage.ResourceProviders[ResourceProvider(rp)] = covOps

		for version, specs := range versions {
			for _, apiPaths := range specs {
				for apiPath, ops := range apiPaths {
					for op := range ops {
						covOp := Operation{
							APIPath:       string(apiPath),
							OperationKind: op,
						}
						covVersions, ok := covOps[covOp]
						if !ok {
							covVersions = APIVersions{}
							covOps[covOp] = covVersions
						}
						covVersions[APIVersion(version)] = nil
					}
				}
			}
		}
	}

	return coverage
}

func (repoCov *RepoCoverageStore) AddSchema(s swagger.Schema, locator workspace.SchemaLocator) error {
	schemaCov, err := NewSchemaCoverageStore(s)
	if err != nil {
		return err
	}

	p := specpath.Path(s.SwaggerRelPath)
	pathInfo, err := p.PathInfo()
	if err != nil {
		return err
	}

	rp := ResourceProvider(pathInfo.ResourceProvider)
	ops, ok := repoCov.ResourceProviders[rp]
	if !ok {
		return fmt.Errorf("no resource provider %q found", rp)
	}

	op := Operation{
		APIPath:       locator.ApiPath,
		OperationKind: locator.Operation,
	}
	versions, ok := ops[op]
	if !ok {
		return fmt.Errorf("no operation %q found for resource provider %q", op, rp)
	}

	versions[APIVersion(locator.ApiVersion)] = schemaCov
	return nil
}

func (op Operation) String() string {
	return fmt.Sprintf("%s:%s", op.APIPath, op.OperationKind)
}

func ParseOperation(s string) Operation {
	l := strings.Split(s, ":")
	return Operation{
		APIPath:       l[0],
		OperationKind: workspace.OperationKind(l[1]),
	}
}

func (ops Operations) MarshalJSON() ([]byte, error) {
	m := map[string]interface{}{}
	for op, apis := range ops {
		m[op.String()] = apis
	}
	return json.Marshal(m)
}

func (ops *Operations) UnmarshalJSON(b []byte) error {
	var m map[string]APIVersions
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	tmpOperations := Operations{}
	for k, v := range m {
		op := ParseOperation(k)
		tmpOperations[op] = v
	}
	*ops = tmpOperations
	return nil
}
