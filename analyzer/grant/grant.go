package grant

import (
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/magodo/azure-rest-api-cov/analyzer/swagger"
	"gopkg.in/yaml.v2"

	"github.com/magodo/azure-rest-api-cov/analyzer/specpath"

	"github.com/magodo/azure-rest-api-cov/analyzer/addr"
)

type SchemaGrants map[string]SchemaGrant // key is the Schema.Ref()

type SchemaGrant struct {
	Variant string `yaml:"variant,omitempty"`

	// The grant comment why the schema is granted. When this is set, it means the schema is to be granted at a whole.
	Comment string `yaml:"comment,omitempty"`

	PropertyGrantRules PropertyGrantRules `yaml:"property_grant_rules,omitempty"`

	// A set of expanded properties that is to be granted, where the value is the property addr, and the value is the grant comment.
	// This is evaluated from "Properties".
	grantedProperties map[string]string
}

type PropertyGrantRules []PropertyGrantRule

type PropertyGrantRule struct {
	Prop    addr.Property
	Negate  bool
	Comment string
}

func (prop PropertyGrantRule) String() string {
	if prop.Negate {
		return fmt.Sprintf("!%s", prop.Prop)
	}
	return fmt.Sprintf("%s: %s", prop.Prop, prop.Comment)
}

func (props PropertyGrantRules) MarshalYAML() (interface{}, error) {
	propertyLines := make([]string, 0, len(props))
	for _, prop := range props {
		propertyLines = append(propertyLines, prop.String())
	}
	return strings.Join(propertyLines, "\n"), nil
}

func (props *PropertyGrantRules) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var propertyLines string
	if err := unmarshal(&propertyLines); err != nil {
		return fmt.Errorf("unmarshaling properties: %v", err)
	}

	for _, propertyLine := range strings.Split(strings.TrimSpace(propertyLines), "\n") {
		var prop PropertyGrantRule
		columnIndex := strings.Index(propertyLine, ":")
		if strings.HasPrefix(propertyLine, "!") {
			prop.Negate = true
			if columnIndex != -1 {
				return fmt.Errorf("negate properrty grant (%q) shouldn't contain comment", propertyLine)
			}
			prop.Prop = addr.MustParseProperty(propertyLine[1:])
		} else {
			if columnIndex == -1 {
				return fmt.Errorf("properrty grant (%q) should contain comment", propertyLine)
			}
			prop.Prop = addr.MustParseProperty(propertyLine[:columnIndex])
			prop.Comment = strings.TrimSpace(propertyLine[columnIndex+1:])
		}
		*props = append(*props, prop)
	}
	return nil
}

type GrantFileFilter struct {
	RP        string
	Version   string
	IsPreview *bool
	SpecName  string
}

// NewGrantsFromFiles construct a SchemaGrants from a grant folder which contains the
// folder layout as defined by the SchemaAddr.
func NewSchemaGrantsFromFiles(grantBaseDir string, specBaseDir string, filter *GrantFileFilter) (SchemaGrants, error) {
	var swgGrant = SchemaGrants{}
	return swgGrant, filepath.Walk(grantBaseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if info.Name() == ".git" || info.Name() == ".github" {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(info.Name(), ".yaml") && !strings.HasSuffix(info.Name(), ".yml") {
			return nil
		}

		specRelPath, err := filepath.Rel(grantBaseDir, path[:strings.LastIndex(path, ".")]+".json")
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

		if err := yaml.Unmarshal(b, &infileGrants); err != nil {
			return fmt.Errorf("unmarshalling %s: %v", path, err)
		}

		// Evaluate the property grant rules reflect to the expanded properties.
		for schemaName, schemaGrant := range infileGrants {
			schema, err := swagger.NewEmptySchema(specBaseDir, specRelPath, schemaName, true, false)
			if err != nil {
				return err
			}

			if err := schema.ExpandProperties(); err != nil {
				return err
			}

			schemaGrant.grantedProperties = map[string]string{}
			for _, grant := range schemaGrant.PropertyGrantRules {
				for propName := range schema.Properties {
					propAddr := addr.MustParseProperty(propName)

					if grant.Prop.Contains(propAddr) || grant.Prop.Equals(propAddr) {
						if grant.Negate {
							delete(schemaGrant.grantedProperties, propName)
						} else {
							schemaGrant.grantedProperties[propName] = grant.Comment
						}
					}
				}
			}

			swgGrant[swagger.SchemaMetaInfo{
				Name:           schemaName,
				SwaggerRelPath: specpath.Path(specRelPath),
			}.Ref()] = schemaGrant
		}
		return nil
	})
}

// GetGrantForSchema get the schema grant for a certain schemaRef.
// If the schemaRef is included in the schemaGrants, the grant is returned.
// If the schemaRef is not included, but another version of that schema ref (same RP, same spec name and same model name) is included,
// the best match (with the least version gap) will be returned.
// Otherwise, a nil SchemaGrant is returned.
func (schemaGrants SchemaGrants) GetGrantForSchema(schemaRef string) (grant *SchemaGrant, ref string, err error) {
	if grant, ok := schemaGrants[schemaRef]; ok {
		return &grant, schemaRef, nil
	}

	metaInfo, err := swagger.ParseSchemaMetaInfoRef(schemaRef)
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
		if len(parts) < 3 {
			return 0, fmt.Errorf("version %q doesn't conform to format yyyy-mm-dd[-xxx]", version)
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
		thisMetaInfo, err := swagger.ParseSchemaMetaInfoRef(ref)
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

func (g SchemaGrant) IsSchemaGranted() bool {
	return g.Comment != "" && len(g.PropertyGrantRules) == 0
}

// Grant inquiries the SchemaGrants to Add the granting information onto the Schema
// If strict is on, this function will raise error if any properties appear in grant doesn't exist in the schema.
// In the strict mode, it will also ensure the to be granted property is in state "uncovered".
// Non-strict mode is useful for using grant from other API versions to try to grant current API version, in the best effort way.
func (g SchemaGrant) Grant(c *swagger.Schema, strict bool) error {
	if g.IsSchemaGranted() {
		c.IsGranted = true
		c.GrantComment = g.Comment
		c.Properties = nil
		return nil
	}

	for propAddr, grantComment := range g.grantedProperties {
		property, ok := c.Properties[propAddr]
		if !ok {
			if strict {
				return fmt.Errorf(`property %q that is to be granted doesn't exist in Swagger schema %q'`, propAddr, c.Ref())
			}
			continue
		}
		if property.PropertyState != nil {
			if property.IsIgnored() {
				continue
			}
			if property.PropertyState.IsCovered() && strict {
				return fmt.Errorf(`property %q that is to be granted is a covered property`, propAddr)
			}
		}
		property.PropertyState = swagger.NewPropertyState(swagger.PropStateGranted, grantComment)
	}

	return nil
}
