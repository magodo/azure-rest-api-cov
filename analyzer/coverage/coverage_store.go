package coverage

import (
	"fmt"

	"github.com/magodo/azure-rest-api-cov/analyzer/swagger"

	"github.com/magodo/azure-rest-api-cov/analyzer/workspace"
)

type RepoCoverageStore struct {
	Commit            string            `json:"commit"`
	AdditionalMessage string            `json:"additional_message,omitempty"`
	ResourceProviders ResourceProviders `json:"resource_providers"`
}

type ResourceProviders map[ResourceProvider]APIPaths

type ResourceProvider string

type APIPaths map[string]Operations

type Operations map[string]APIVersions

type APIVersions map[string]*SchemaCoverageStore

func NewEmptyRepoCoverageStore(index workspace.AzureRepoIndex, additionalMessage string) RepoCoverageStore {
	coverage := RepoCoverageStore{
		Commit:            index.Commit,
		AdditionalMessage: additionalMessage,
		ResourceProviders: ResourceProviders{},
	}

	for rp, versions := range index.ResourceProviders {
		paths := APIPaths{}
		coverage.ResourceProviders[ResourceProvider(rp)] = paths
		for version, specs := range versions {
			for _, apiPaths := range specs {
				for apiPath, ops := range apiPaths {

					covOps, ok := paths[apiPath]
					if !ok {
						covOps = Operations{}
						paths[apiPath] = covOps
					}

					for op := range ops {

						covVersions, ok := covOps[string(op)]
						if !ok {
							covVersions = APIVersions{}
							covOps[string(op)] = covVersions
						}

						covVersions[string(version)] = nil
					}
				}
			}
		}
	}

	return coverage
}

func (repoCov *RepoCoverageStore) AddSchema(repoURLPrefix string, s swagger.Schema, locator workspace.SchemaLocator, active bool) error {
	schemaCov, err := NewSchemaCoverageStore(repoURLPrefix, s, active)
	if err != nil {
		return err
	}

	p := s.SwaggerRelPath
	pathInfo, err := p.PathInfo()
	if err != nil {
		return err
	}

	rp := ResourceProvider(pathInfo.ResourceProvider)
	paths, ok := repoCov.ResourceProviders[rp]
	if !ok {
		return fmt.Errorf("no resource provider %q found", rp)
	}

	ops, ok := paths[locator.ApiPath]
	if !ok {
		return fmt.Errorf("no API path %q found for resource provider %q", locator.ApiPath, rp)
	}

	versions, ok := ops[string(locator.Operation)]
	if !ok {
		return fmt.Errorf("no operation %q found for API path %q, resource provider %q", locator.Operation, locator.ApiPath, rp)
	}

	versions[locator.ApiVersion] = schemaCov
	return nil
}
