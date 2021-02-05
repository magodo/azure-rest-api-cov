package swagger

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strconv"
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

type GrantFileFilter struct {
	RP        string
	Version   string
	IsPreview *bool
	SpecName  string
}

func (g SchemaGrant) IsSchemaGranted() bool {
	return len(g.Properties) == 0
}

// NewGrantsFromFiles construct a SchemaGrants from a grant folder which contains the
// folder layout as defined by the SchemaAddr.
func NewSchemaGrantsFromFiles(grantBaseDir string, specBaseDir string, filter *GrantFileFilter) (SchemaGrants, error) {
	var swgGrant = SchemaGrants{}
	return swgGrant, filepath.Walk(grantBaseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(info.Name(), ".json") {
			return err
		}

		specRelPath, err := filepath.Rel(grantBaseDir, path)
		if err != nil {
			return err
		}

		if filter != nil {
			pathInfo, err := specpath.Path(specRelPath).PathInfo()
			if err != nil {
				return err
			}

			if filter.RP != "" {
				if pathInfo.ResourceProvider != filter.RP {
					return nil
				}
			}

			if filter.IsPreview != nil {
				if pathInfo.IsPreview != *filter.IsPreview {
					return nil
				}
			}

			if filter.Version != "" {
				if pathInfo.Version != filter.Version {
					return nil
				}
			}

			if filter.SpecName != "" {
				if pathInfo.SpecName != filter.SpecName {
					return nil
				}
			}
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

			_, err := ProcessCoverageForSchema(specBaseDir, specRelPath, schemaName, schemaGrant.Variant, coveredProps, true, true)
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

// SchemaGrantFor get the schema grant for a certain schemaRef.
// If the schemaRef is included in the schemaGrants, the grant is returned, and the "exact" is set to true.
// If the schemaRef is not included, but another version of that schema ref (same RP, same spec name and same model name) is included,
// the best match (with the least version gap) will be returned, and the "exact" is set to false.
// Otherwise, a nil SchemaGrant is returned.
func (schemaGrants SchemaGrants) SchemaGrantFor(schemaRef string) (grant *SchemaGrant, ref string, err error) {
	if grant, ok := schemaGrants[schemaRef]; ok {
		return &grant, schemaRef, nil
	}

	metaInfo, err := ParseSchemaMetaInfoRef(schemaRef)
	if err != nil {
		return nil, "", err
	}

	pathInfo, err := metaInfo.SwaggerRelPath.PathInfo()
	if err != nil {
		return nil, "", err
	}

	versionToNum := func(ver string) (int, error) {
		version := strings.TrimSuffix(ver, "-preview")
		parts := strings.Split(version, "-")
		if len(parts) != 3 {
			return 0, fmt.Errorf("version %q doesn't conform to format yyyy-mm-dd", version)
		}
		y, m, d := parts[0], parts[1], parts[2]

		yearNum, err := strconv.Atoi(y)
		if err != nil {
			return 0, err
		}
		monthNum, err := strconv.Atoi(m)
		if err != nil {
			return 0, err
		}

		dayNum, err := strconv.Atoi(d)
		if err != nil {
			return 0, err
		}

		return yearNum*365 + monthNum*30 + dayNum, nil
	}

	var (
		bestMatchRef    string
		leastVersionGap = math.MaxFloat64
	)
	for ref := range schemaGrants {
		thisMetaInfo, err := ParseSchemaMetaInfoRef(ref)
		if err != nil {
			return nil, "", err
		}
		thisPathInfo, err := thisMetaInfo.SwaggerRelPath.PathInfo()
		if err != nil {
			return nil, "", err
		}

		if thisMetaInfo.Name != metaInfo.Name {
			continue
		}

		if thisPathInfo.ResourceProvider != pathInfo.ResourceProvider {
			continue
		}

		if thisPathInfo.ResourceProviderMS != pathInfo.ResourceProviderMS {
			continue
		}

		if thisPathInfo.SpecName != pathInfo.SpecName {
			continue
		}

		// Compare the version with "-preview" stripped (if any). Pick the best match.
		thisVersion, err := versionToNum(thisPathInfo.Version)
		if err != nil {
			return nil, "", err
		}
		version, err := versionToNum(pathInfo.Version)
		if err != nil {
			return nil, "", err
		}

		versionGap := math.Abs(float64(version) - float64(thisVersion))
		if versionGap < leastVersionGap {
			leastVersionGap = versionGap
			bestMatchRef = ref
		}
		// Always pick the lower API version to make the result consistent
		if versionGap == leastVersionGap {
			if float64(version)-float64(thisVersion) > 0 {
				leastVersionGap = versionGap
				bestMatchRef = ref
			}
		}
	}

	if bestMatchRef != "" {
		grant := schemaGrants[bestMatchRef]
		return &grant, bestMatchRef, nil
	}
	return nil, "", nil
}

func (schemaGrant SchemaGrant) ToSchema(schemaRef string, specRootDir string, strict bool) (*Schema, error) {
	metaInfo, err := ParseSchemaMetaInfoRef(schemaRef)
	if err != nil {
		return nil, err
	}
	properties := make([]string, 0, len(schemaGrant.Properties))
	for prop := range schemaGrant.Properties {
		properties = append(properties, prop)
	}

	schema, err := ProcessCoverageForSchema(specRootDir, string(metaInfo.SwaggerRelPath), metaInfo.Name, schemaGrant.Variant, properties, true, strict)
	if err != nil {
		return nil, err
	}
	return schema, nil
}
