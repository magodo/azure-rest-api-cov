package analyzer

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strings"

	"github.com/magodo/azure-rest-api-cov/analyzer/coverage"
	"github.com/magodo/azure-rest-api-cov/analyzer/grant"
	"github.com/magodo/azure-rest-api-cov/analyzer/workspace"
	log "github.com/sirupsen/logrus"
)

type GrantContext struct {
	WorkspaceOpt     workspace.InitOption
	TryMigrate       bool
	MigrateGrantHint bool
}

// grant leverage the grant repo to grant the schemas (both active and predicted schemas).
// In case there is any grant exist but is for some other API version, a hint message will be returned.
// If tryMigrate is specified, these API mismatched grant will take effect.
func Grant(in io.Reader, ctx GrantContext) (string, error) {
	log.Info("CMD: Grant")

	wsp, err := workspace.Init(ctx.WorkspaceOpt)
	if err != nil {
		return "", err
	}

	// Parse the coverage
	b, err := ioutil.ReadAll(in)
	if err != nil {
		return "", err
	}
	var cov coverage.Coverage
	if err := json.Unmarshal(b, &cov); err != nil {
		return "", err
	}

	// Grant schemas

	flatCov := cov.Flatten()

	sgs, err := grant.NewSchemaGrantsFromFiles(wsp.GrantRepo.RootDir, wsp.AzureRepo.SpecRootDir, nil)
	if err != nil {
		return "", err
	}

	var hintMessages []string
	for k := range flatCov {
		opInfo := flatCov[k]
		if opInfo.Model == nil {
			continue
		}
		schema := opInfo.Model
		g, matchRef, err := sgs.GetGrantForSchema(schema.Ref())
		if err != nil {
			return "", err
		}
		if g != nil {
			log.Infof("Granting for schema %v", k)
			// The grant file is exactly for this schema.
			if matchRef == schema.Ref() {
				if err := g.Grant(schema, true); err != nil {
					return "", fmt.Errorf("granting for schema %v: %v", k, err)
				}
			} else {
				hintMessages = append(hintMessages, fmt.Sprintf("there is a grant with different API version %q, that is able to be migrated for %q", matchRef, schema.Ref()))
				if ctx.TryMigrate {
					if err := g.Grant(schema, false); err != nil {
						return "", fmt.Errorf("(trying) granting for schema %v: %v", k, err)
					}
				}
			}
		}
	}

	if ctx.MigrateGrantHint {
		if err := ioutil.WriteFile("possible_grant_migration.txt", []byte(strings.Join(hintMessages, "\n")), 0644); err != nil {
			return "", err
		}
	}

	cov, err = flatCov.Structuralize()
	if err != nil {
		return "", err
	}

	// Grant rps
	rpgs, err := grant.NewRPGrantsFromFiles(wsp.GrantRepo.RootDir)
	if err != nil {
		return "", err
	}

	for rp, rpg := range rpgs {
		log.Infof("Granting for RP %v", rp)
		apiPaths, ok := cov[rp]
		if !ok {
			apiPaths = &coverage.APIPaths{}
			cov[rp] = apiPaths
		}

		for path, grantComment := range rpg {
			apiPath, ok := (*apiPaths)[path]
			if !ok {
				apiPath = &coverage.APIPath{}
				(*apiPaths)[path] = apiPath
			}
			apiPath.GrantComment = grantComment
		}
	}

	b, err = json.MarshalIndent(cov, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}
