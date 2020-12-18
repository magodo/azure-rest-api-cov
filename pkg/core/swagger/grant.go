package swagger

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type SchemaGrants map[string]SchemaGrant // key is the Schema.Ref()

type SchemaGrant struct {
	// The grant comment why the schema is granted.
	Comment string `json:"comment,omitempty"`

	// Property grant map, key is the RelProperty, and value is the grant comment.
	Properties map[string]string `json:"properties,omitempty"`
}

func (g SchemaGrant) IsSchemaGranted() bool {
	return len(g.Properties) == 0
}

// NewGrantsFromFiles construct a SchemaGrants from a grant folder which contains the
// folder layout as defined by the SchemaAddr.
func NewSchemaGrantsFromFiles(grantBaseDir string) (SchemaGrants, error) {
	var swgGrant = SchemaGrants{}
	return swgGrant, filepath.Walk(grantBaseDir, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() ||
			(!strings.HasSuffix(info.Name(), ".json") &&
				!strings.HasSuffix(info.Name(), ".yaml") &&
				!strings.HasSuffix(info.Name(), ".yml")) {
			return err
		}

		infileGrants := map[string]SchemaGrant{}

		b, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		if err := json.Unmarshal(b, &infileGrants); err != nil {
			return err
		}

		relPath, err := filepath.Rel(grantBaseDir, path)
		if err != nil {
			return err
		}

		for schemaName, schemaGrant := range infileGrants {
			swgGrant[SchemaMetaInfo{
				Name:           schemaName,
				SwaggerRelPath: relPath,
			}.Ref()] = schemaGrant
		}
		return nil
	})
}
