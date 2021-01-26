package swagger

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/magodo/azure-rest-api-cov/analyzer/specpath"

	"github.com/magodo/azure-rest-api-cov/analyzer/addr"
)

type SchemaGrants map[string]SchemaGrant // key is the Schema.Ref()

type SchemaGrant struct {
	Variant string `json:"variant,omitempty"`

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
func NewSchemaGrantsFromFiles(grantBaseDir string, specBaseDir string) (SchemaGrants, error) {
	var swgGrant = SchemaGrants{}
	return swgGrant, filepath.Walk(grantBaseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(info.Name(), ".json") {
			return err
		}

		infileGrants := map[string]SchemaGrant{}

		b, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		if len(b) == 0 {
			return nil
		}

		if err := json.Unmarshal(b, &infileGrants); err != nil {
			return err
		}

		specRelPath, err := filepath.Rel(grantBaseDir, path)
		if err != nil {
			return err
		}

		for schemaName, schemaGrant := range infileGrants {

			// Sanity check: ensure the properties are mutual exclusive.
			for prop1 := range schemaGrant.Properties {
				addr1 := addr.MustParseProperty(prop1)
				for prop2 := range schemaGrant.Properties {
					if prop1 == prop2 {
						continue
					}
					addr2 := addr.MustParseProperty(prop2)
					if addr1.Contains(addr2) || addr2.Contains(addr1) {
						return fmt.Errorf("sanity check failed: non-mutual inclusive properties %q and %q detected in schema %q (%q)",
							prop1, prop2, schemaName, path)
					}
				}
			}

			// Sanity check: ensure the schema and its properties exist.
			coveredProps := make([]string, 0, len(schemaGrant.Properties))
			for k := range schemaGrant.Properties {
				coveredProps = append(coveredProps, k)
			}

			_, err := ProcessCoverageForSchema(specBaseDir, specRelPath, schemaName, schemaGrant.Variant, coveredProps, false, true)
			if err != nil {
				return err
			}

			swgGrant[SchemaMetaInfo{
				Name:           schemaName,
				SwaggerRelPath: specpath.Path(specRelPath),
			}.Ref()] = schemaGrant
		}
		return nil
	})
}
