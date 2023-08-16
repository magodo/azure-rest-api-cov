package feeddb

import (
	"encoding/json"
	"strings"

	"github.com/magodo/aztfq/aztfq"
	"github.com/magodo/azure-rest-api-cov/analyzer/addr"
	"github.com/magodo/azure-rest-api-cov/analyzer/swagger"

	log "github.com/sirupsen/logrus"
)

type APIModel struct {
	Name       string     `json:"name"`
	Properties Properties `json:"properties"`
}

type Properties map[addr.PropertySegment]*Property

type Property struct {
	// Only exists for leaf properties
	IsCovered  *bool      `json:"isCovered,omitempty"`
	Properties Properties `json:"properties,omitempty"`
	TFMapping  TFMap      `json:"tfMapping,omitempty"`
}

type TFMap struct {
	TFResource []string `json:"tfResource,omitempty"`
	TFAddr     []string `json:"tfAddr,omitempty"`
}

var aztfqCnt int = 0
var mappedAztfqTable aztfq.LookupTable = make(map[string]map[string]map[string][]aztfq.TFResult, 0)

func GetAztfqCnt() int {
	return aztfqCnt
}

func GetMappedAztfqMap() aztfq.LookupTable {
	return mappedAztfqTable
}

func NewAPIModel(s swagger.Schema, tfPropTable map[string][]aztfq.TFResult, diag AztfqDiagCtx) (*APIModel, error) {
	// In case this is a schemaless operation
	if s.Name == "" && len(s.Properties) == 0 {
		return nil, nil
	}

	model := APIModel{
		Name: s.Name,
	}

	var property Property
	for propRAddr, prop := range s.Properties {
		// todo: maybe we can generate a map using `addr.Property` instead of `string` and the strings editing function.
		tfIndex := strings.ReplaceAll(propRAddr, "/", ".")
		tfIndex = strings.ReplaceAll(tfIndex, `properties{`, `properties.{`)
		tfIndex = strings.ReplaceAll(tfIndex, `{}`, "*")
		tfMapping := TFMap{
			TFResource: make([]string, 0),
			TFAddr:     make([]string, 0),
		}
		log.Infof("[aztfq map]resource type: %s, index: %s", diag.rtName, tfIndex)
		if tfAddr, ok := tfPropTable[tfIndex]; ok {
			aztfqCnt++
			for _, tf := range tfAddr {
				tfMapping.TFAddr = append(tfMapping.TFAddr, tf.PropertyAddr)
				tfMapping.TFResource = append(tfMapping.TFResource, tf.ResourceType)
			}
			mappedAztfqTable[diag.rtName][diag.apiVersion][tfIndex] = tfAddr
		}
		property.Add(addr.MustParseProperty(propRAddr), prop.PropertyState, tfMapping)
		addr.MustParseProperty(propRAddr)
	}
	model.Properties = property.Properties
	return &model, nil
}

func (prop *Property) Add(propAddr addr.Property, state swagger.PropertyState, tfMapping TFMap) {
	if state.IsIgnored() {
		return
	}

	boolPtr := func(b bool) *bool {
		return &b
	}
	if len(propAddr) == 0 {
		prop.IsCovered = boolPtr(state.IsCovered())
		prop.TFMapping = tfMapping
		return
	}

	if prop.Properties == nil {
		prop.Properties = Properties{}
	}

	if _, ok := prop.Properties[propAddr[0]]; !ok {
		prop.Properties[propAddr[0]] = &Property{}
	}
	prop.Properties[propAddr[0]].Add(propAddr[1:], state, tfMapping)

	return
}

func (props Properties) MarshalJSON() ([]byte, error) {
	m := map[string]*Property{}
	for k, v := range props {
		m[k.String()] = v
	}
	return json.Marshal(m)
}

func (props *Properties) UnmarshalJSON(b []byte) error {
	var m map[string]*Property
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	tmpNodes := Properties{}
	for k, v := range m {
		tmpNodes[addr.MustParsePropertySegment(k)] = v
	}
	*props = tmpNodes
	return nil
}
