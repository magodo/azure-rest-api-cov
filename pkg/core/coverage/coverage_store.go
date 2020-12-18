package coverage

import (
	"encoding/json"
	"fmt"
	"strings"

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
	APIPath       string `json:"api_path"`
	OperationKind string `json:"operation_kind"`
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
							OperationKind: string(op),
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

func (op Operation) String() string {
	return fmt.Sprintf("%s:%s", op.APIPath, op.OperationKind)
}

func ParseOperation(s string) Operation {
	l := strings.Split(s, ":")
	return Operation{
		APIPath:       l[0],
		OperationKind: l[1],
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
