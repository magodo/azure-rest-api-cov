package analyzer

import (
	"encoding/json"
	"fmt"
	"github.com/magodo/azure-rest-api-cov/analyzer/coverage"
	"github.com/magodo/azure-rest-api-cov/analyzer/grant"
	"github.com/magodo/azure-rest-api-cov/analyzer/workspace"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"strings"
)

type GrantContext struct {
	WorkspaceOpt     workspace.InitOption
	TryMigrate       bool
	MigrateGrantHint bool
}

// grant leverage the grant repo to grant the schemas (both active and predicted schemas).
// In case there is any grant exist but is for some other API version, a hint message will be returned.
// If tryMigrate is specified, these API mismatched grant will take effect.
func Grant(coverageIn io.Reader, ctx GrantContext) (string, error) {
	log.Info("Processing grant...")
	wsp, err := workspace.Init(ctx.WorkspaceOpt)
	if err != nil {
		return "", err
	}

	grants, err := grant.NewSchemaGrantsFromFiles(wsp.GrantRepo.RootDir, wsp.AzureRepo.SpecRootDir, nil)
	if err != nil {
		return "", err
	}

	// Parse the coverage
	b, err := ioutil.ReadAll(coverageIn)
	if err != nil {
		return "", err
	}
	var cov coverage.Coverage
	if err := json.Unmarshal(b, &cov); err != nil {
		return "", err
	}

	flatCov := cov.Flatten()

	var hintMessages []string
	for k := range flatCov {
		schema := flatCov[k]
		g, matchRef, err := grants.GetGrantForSchema(schema.Ref())
		if err != nil {
			return "", err
		}
		if g != nil {
			log.Debugf("Granting for schema %v", k)
			// The grant file is exactly for this schema.
			if matchRef == schema.Ref() {
				if err := g.Grant(&schema, true); err != nil {
					return "", fmt.Errorf("granting for schema %v: %v", k, err)
				}
			} else {
				hintMessages = append(hintMessages, fmt.Sprintf("there is a grant with different API version %q, that is able to be migrated for %q", matchRef, schema.Ref()))
				if ctx.TryMigrate {
					if err := g.Grant(&schema, false); err != nil {
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

	b, err = json.MarshalIndent(cov, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}
