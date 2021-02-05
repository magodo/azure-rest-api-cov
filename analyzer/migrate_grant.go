package analyzer

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/magodo/azure-rest-api-cov/analyzer/swaggerloader"

	"github.com/magodo/azure-rest-api-cov/analyzer/swagger"
	"github.com/magodo/azure-rest-api-cov/analyzer/workspace"
)

func MigrateGrant(workspaceOpt workspace.InitOption, grantRepoDir, rp, spec, srcVersion, dstVersion, outDir string) error {
	// Initialize the workspace, which will return the repo and repo index.
	wsp, err := workspace.Init(workspaceOpt)
	if err != nil {
		return err
	}

	migrateGrantFiles, err := migrateGrant(wsp.AzureRepoIndex.SpecRootDir, grantRepoDir, rp, spec, srcVersion, dstVersion)
	if err != nil {
		return err
	}

	for path, content := range migrateGrantFiles {
		ofile := filepath.Join(outDir, path)
		err := os.MkdirAll(filepath.Dir(ofile), 0755)
		if err != nil && !errors.Is(err, os.ErrExist) {
			return err
		}
		if err := ioutil.WriteFile(ofile, content, 0644); err != nil {
			return err
		}
	}
	return nil
}

func migrateGrant(specRootDir, grantRootDir, rp, spec, srcVersion, dstVersion string) (map[string][]byte, error) {
	srcGrants, err := swagger.NewSchemaGrantsFromFiles(grantRootDir, specRootDir,
		&swagger.GrantFileFilter{
			RP:       rp,
			SpecName: spec,
			Version:  srcVersion,
		})
	if err != nil {
		return nil, err
	}

	migrateGrantFiles := map[string]swagger.SchemaGrants{}
	for schemaRef, srcGrant := range srcGrants {
		// Ensure the target spec exists, which might be in either preview or stable folder.
		metaInfo, err := swagger.ParseSchemaMetaInfoRef(schemaRef)
		if err != nil {
			return nil, err
		}
		pathInfo, err := metaInfo.SwaggerRelPath.PathInfo()
		if err != nil {
			return nil, err
		}

		pathInfo.Version = dstVersion
		dstPath := pathInfo.ToPath()
		absPath := filepath.Join(specRootDir, string(dstPath))
		_, err = os.Stat(absPath)
		if err == os.ErrNotExist {
			pathInfo.IsPreview = !pathInfo.IsPreview
			dstPath = pathInfo.ToPath()
			absPath = filepath.Join(specRootDir, string(dstPath))
			if _, err := os.Stat(absPath); err != nil {
				return nil, fmt.Errorf("dest spec %s with version %s (either preview or stable) in RP %s doesn't exist: %v",
					pathInfo.SpecName, pathInfo.Version, pathInfo.ResourceProvider, err)
			}
		}

		dstMetaInfo := swagger.SchemaMetaInfo{
			Name:           metaInfo.Name,
			SwaggerRelPath: dstPath,
		}

		// Check whether the model is defined in the target API version spec.
		swg, err := swaggerloader.LoadSwagger(absPath)
		if err != nil {
			return nil, err
		}
		if _, ok := swg.Definitions[dstMetaInfo.Name]; !ok {
			continue
		}

		// Here we convert the grant to a schema.Schema and try to expand it to the properties defined in the grant in non-strict mode.
		// The resulted "covered" properties mean those properties still exist in the dst API version, and will be kept in the dst grant.
		// The non-existed properties in the new API version will be skipped due to non-strict mode.
		schema, err := srcGrant.ToSchema(dstMetaInfo.Ref(), specRootDir, false)
		if err != nil {
			return nil, err
		}

		var dstGrant swagger.SchemaGrant

		// In case there are no properties left, it might because
		// - All the granted properties are not existed in the newer API version: in this case we shall skip this model at all.
		// - The model is to be granted: in this case we shall go on granting this model as a whole.
		if len(schema.Properties) == 0 {
			if srcGrant.IsSchemaGranted() {
				dstGrant = srcGrant
			} else {
				continue
			}
		} else {
			dstGrant = srcGrant
			dstGrant.Properties = map[string]string{}
			for prop, state := range schema.Properties {
				if state.IsCovered() {
					dstGrant.Properties[prop] = srcGrant.Properties[prop]
				}
			}
		}

		migrateGrantFile := migrateGrantFiles[string(dstPath)]
		if migrateGrantFile == nil {
			migrateGrantFile = swagger.SchemaGrants{}
			migrateGrantFiles[string(dstPath)] = migrateGrantFile
		}
		migrateGrantFile[dstMetaInfo.Name] = dstGrant
	}

	out := map[string][]byte{}
	for k, v := range migrateGrantFiles {
		out[k], err = json.MarshalIndent(v, "", "  ")
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}
