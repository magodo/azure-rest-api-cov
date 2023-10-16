package report

import (
	"fmt"

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
		Commit:            index.Commit(),
		AdditionalMessage: additionalMessage,
		ResourceProviders: ResourceProviders{},
	}

	for rp, versions := range index.ResourceProviders {
		paths := &APIPaths{}
		coverage.ResourceProviders[ResourceProvider(rp)] = paths
		for version, specs := range versions {
			for _, apiPaths := range specs {
				for apiPath, ops := range apiPaths {

					path, ok := (*paths)[apiPath]
					if !ok {
						path = &APIPath{}
						(*paths)[apiPath] = path
					}

					if len(ops) != 0 && path.Operations == nil {
						path.Operations = Operations{}
					}

					for op := range ops {

						covVersions, ok := path.Operations[string(op)]
						if !ok {
							covVersions = &APIVersions{}
							path.Operations[string(op)] = covVersions
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
