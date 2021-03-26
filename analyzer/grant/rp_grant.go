package grant

import (
	"errors"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type RPGrants map[string]RPGrant // key is the RP name

type RPGrant map[string]string // key is the api path, value is the grant comment

// NewRPGrantsFromFiles construct a RPGrants from a grant folder which contains the
// folder layout as defined by the SchemaAddr.
func NewRPGrantsFromFiles(grantBaseDir string) (RPGrants, error) {
	dirEntries, err := os.ReadDir(grantBaseDir)
	if err != nil {
		return nil, err
	}

	gs := RPGrants{}
	// Try to read the grant.yaml file under the top level folders, consume the file into RPGrants if exists.
	for _, rpDir := range dirEntries {
		if !rpDir.IsDir() {
			continue
		}
		b, err := os.ReadFile(filepath.Join(grantBaseDir, rpDir.Name(), "grant.yaml"))
		if errors.Is(err, os.ErrNotExist) {
			continue
		}

		g := RPGrant{}
		if err := yaml.Unmarshal(b, &g); err != nil {
			return nil, err
		}

		gs[rpDir.Name()] = g
	}
	return gs, nil
}
