package report

import (
	"fmt"
	"strings"

	"slices"

	"github.com/magodo/azure-rest-api-cov/analyzer/operation"
	"github.com/magodo/azure-rest-api-cov/analyzer/workspace"
)

type Coverage struct {
	Commit            string            `json:"commit"`
	AdditionalMessage string            `json:"additional_message,omitempty"`
	ResourceProviders ResourceProviders `json:"resource_providers"`
}

type ResourceProviders map[ResourceProvider]*APIPaths

type ResourceProvider string

type APIPaths map[string]*APIPath

type APIPath struct {
	GrantComment string     `json:"grant_comment,omitempty"`
	Operations   Operations `json:"operations,omitempty"`
}

type Operations map[string]*APIVersions

type APIVersions map[string]*SchemaCoverage

func NewEmptyCoverage(index *workspace.AzureRepoIndex, additionalMessage string) Coverage {
	coverage := Coverage{
		Commit:            index.Index.Commit,
		AdditionalMessage: additionalMessage,
		ResourceProviders: ResourceProviders{},
	}

	for rp, versions := range index.Index.ResourceProviders {
		paths := &APIPaths{}
		coverage.ResourceProviders[ResourceProvider(rp)] = paths
		for version, methods := range versions {
			for opKind, rts := range methods {
				for _, opInfo := range rts {
					if len(opInfo.OperationRefs) == 0 {
						continue
					}

					//  Enumerate all the original api path patterns defined in the spec
					pm := map[string]bool{}
					for _, ref := range opInfo.OperationRefs {
						p := strings.ToUpper(ref.GetPointer().DecodedTokens()[1])
						pm[p] = true
					}
					var pl []string
					for p := range pm {
						pl = append(pl, p)
					}
					slices.Sort(pl)

					for _, p := range pl {
						path, ok := (*paths)[p]
						if !ok {
							path = &APIPath{Operations: Operations{}}
							(*paths)[p] = path
						}

						covVersions, ok := path.Operations[string(opKind)]
						if !ok {
							covVersions = &APIVersions{}
							path.Operations[string(opKind)] = covVersions
						}
						(*covVersions)[string(version)] = nil
					}
				}
			}
		}
	}

	return coverage
}

func (cov *Coverage) AddOperation(repoURLPrefix string, opInfo *operation.Info) error {
	schemaCov, err := NewSchemaCoverage(repoURLPrefix, opInfo)
	if err != nil {
		return err
	}

	rp := ResourceProvider(opInfo.RP)
	paths, ok := cov.ResourceProviders[rp]
	if !ok {
		return fmt.Errorf("no resource provider %q found", rp)
	}

	path, ok := (*paths)[opInfo.APIPath]
	if !ok {
		return fmt.Errorf("no API path %q found for resource provider %q", opInfo.APIPath, rp)
	}

	versions, ok := path.Operations[string(opInfo.Operation)]
	if !ok {
		return fmt.Errorf("no operation %q found for API path %q, resource provider %q", opInfo.Operation, opInfo.APIPath, rp)
	}

	(*versions)[opInfo.APIVersion] = schemaCov
	return nil
}
