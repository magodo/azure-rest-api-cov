package report

import (
	"fmt"

	"github.com/magodo/azure-rest-api-cov/analyzer/swagger"

	"github.com/magodo/azure-rest-api-cov/analyzer/workspace"
)

type Coverage struct {
	Commit            string            `json:"commit"`
	AdditionalMessage string            `json:"additional_message,omitempty"`
	ResourceProviders ResourceProviders `json:"resource_providers"`
}

type ResourceProviders map[ResourceProvider]APIPaths

type ResourceProvider string

type APIPaths map[string]APIPath

type APIPath struct {
	GrantComment string     `json:"grant_comment,omitempty"`
	Operations   Operations `json:"operations,omitempty"`
}

type Operations map[string]APIVersions

type APIVersions map[string]*SchemaCoverage

func NewEmptyCoverage(index workspace.AzureRepoIndex, additionalMessage string) Coverage {
	coverage := Coverage{
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

					path, ok := paths[apiPath]
					if !ok {
						path = APIPath{
							Operations: Operations{},
						}
						paths[apiPath] = path
					}

					for op := range ops {

						covVersions, ok := path.Operations[string(op)]
						if !ok {
							covVersions = APIVersions{}
							path.Operations[string(op)] = covVersions
						}

						covVersions[string(version)] = nil
					}
				}
			}
		}
	}

	return coverage
}

func (cov *Coverage) AddSchema(repoURLPrefix string, s swagger.Schema, locator workspace.SchemaLocator) error {
	schemaCov, err := NewSchemaCoverage(repoURLPrefix, s)
	if err != nil {
		return err
	}

	p := s.SwaggerRelPath
	pathInfo, err := p.PathInfo()
	if err != nil {
		return err
	}

	rp := ResourceProvider(pathInfo.ResourceProvider)
	paths, ok := cov.ResourceProviders[rp]
	if !ok {
		return fmt.Errorf("no resource provider %q found", rp)
	}

	path, ok := paths[locator.ApiPath]
	if !ok {
		return fmt.Errorf("no API path %q found for resource provider %q", locator.ApiPath, rp)
	}

	versions, ok := path.Operations[string(locator.Operation)]
	if !ok {
		return fmt.Errorf("no operation %q found for API path %q, resource provider %q", locator.Operation, locator.ApiPath, rp)
	}

	versions[locator.ApiVersion] = schemaCov
	return nil
}
