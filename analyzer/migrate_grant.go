package analyzer

import (
	"errors"
	"fmt"
	"github.com/magodo/azure-rest-api-cov/analyzer/specpath"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/magodo/azure-rest-api-cov/analyzer/addr"
	"github.com/magodo/azure-rest-api-cov/analyzer/grant"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"

	"github.com/magodo/azure-rest-api-cov/analyzer/swaggerloader"

	"github.com/magodo/azure-rest-api-cov/analyzer/swagger"
	"github.com/magodo/azure-rest-api-cov/analyzer/workspace"
)

type MigrateGrantContext struct {
	WorkspaceOpt workspace.InitOption
	GrantRepoDir string
	RP           string
	Spec         string
	SrcVersion   string
	DstVersion   string
	OutDir       string
}

func MigrateGrant(ctx MigrateGrantContext) error {
	log.Info("CMD: Migrate Grant")

	// Initialize the workspace, which will return the repo and repo index.
	wsp, err := workspace.Init(ctx.WorkspaceOpt)
	if err != nil {
		return err
	}

	migrateGrantFiles, err := migrateGrant(wsp.AzureRepoIndex.SpecRootDir, ctx.GrantRepoDir, ctx.RP, ctx.Spec, ctx.SrcVersion, ctx.DstVersion)
	if err != nil {
		return err
	}

	for path, content := range migrateGrantFiles {
		ofile := filepath.Join(ctx.OutDir, path)
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
	srcGrants, err := grant.NewSchemaGrantsFromFiles(grantRootDir, specRootDir,
		&grant.GrantFileFilter{
			RP:       rp,
			SpecName: spec,
			Version:  srcVersion,
		})
	if err != nil {
		return nil, err
	}

	migrateGrantFiles := map[string]grant.SchemaGrants{}
	for schemaRef, srcGrant := range srcGrants {
		log.Infof("migrating grant of %q from %q to %q", schemaRef, srcVersion, dstVersion)

		// Ensure the target spec exists, which might be in either preview or stable folder.
		metaInfo, err := swagger.ParseSchemaMetaInfoRef(schemaRef)
		if err != nil {
			return nil, err
		}
		// TODO: the swagger path to the schema definition is not always conform to the specpath.Path pattern
		pathInfo, err := specpath.Path(metaInfo.SwaggerRelPath).PathInfo()
		if err != nil {
			return nil, err
		}

		pathInfo.Version = dstVersion
		dstPath := pathInfo.ToPath()
		absPath := filepath.Join(specRootDir, string(dstPath))

		// Check whether the target API version spec exists.
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
			SwaggerRelPath: string(dstPath),
		}

		// Check whether the model is defined in the target API version spec.
		swg, err := swaggerloader.LoadSwagger(absPath)
		if err != nil {
			return nil, err
		}

		// In case there is no such model in the target API version spec, there is nothing to migrate, simply skip it.
		if _, ok := swg.Definitions[dstMetaInfo.Name]; !ok {
			continue
		}

		var dstGrant grant.SchemaGrant

		dstSchema, err := swagger.NewEmptySchema(specRootDir, string(dstPath), dstMetaInfo.Name, true)
		if err != nil {
			return nil, err
		}

		if err := dstSchema.ExpandProperties(); err != nil {
			return nil, err
		}

		dstGrant = srcGrant
		// Validate each grant rule as defined in the source grant file to see whether it is take effect on the schema of the target API vesion.
		newGrantRules := grant.PropertyGrantRules{}
	ruleLoop:
		for _, grantRule := range dstGrant.PropertyGrantRules {
			for prop := range dstSchema.Properties {
				grantProp := grantRule.Prop
				existProp := addr.MustParseProperty(prop)
				if grantProp.Contains(existProp) || grantProp.Equals(existProp) {
					newGrantRules = append(newGrantRules, grantRule)
					continue ruleLoop
				}
			}
			log.Warnf("grant rule %q is thrown away as in the destination API version there is no property affected by this rule.", grantRule)
		}
		dstGrant.PropertyGrantRules = newGrantRules

		dstGrantPath := strings.ReplaceAll(string(dstPath), ".json", ".yaml")
		migrateGrantFile := migrateGrantFiles[dstGrantPath]
		if migrateGrantFile == nil {
			migrateGrantFile = grant.SchemaGrants{}
			migrateGrantFiles[dstGrantPath] = migrateGrantFile
		}
		migrateGrantFile[dstMetaInfo.Name] = dstGrant
	}

	out := map[string][]byte{}
	for k, v := range migrateGrantFiles {
		out[k], err = yaml.Marshal(v)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}
