package feeddb

import (
	"encoding/json"
	"strings"

	"github.com/magodo/aztfq/aztfq"
	"github.com/magodo/azure-rest-api-cov/analyzer/addr"
	"github.com/magodo/azure-rest-api-cov/analyzer/swagger"
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
	TFMapping  *TFMap     `json:"tfMapping,omitempty"`
}

type TFMap struct {
	TFResource []string `json:"tfResource,omitempty"`
	TFAddr     []string `json:"tfAddr,omitempty"`
}

var AztfqCnt int = 0

func NewAPIModel(s swagger.Schema) (*APIModel, error) {
	// In case this is a schemaless operation
	if s.Name == "" && len(s.Properties) == 0 {
		return nil, nil
	}

	model := APIModel{
		Name: s.Name,
	}

	var property Property
	for propRAddr, prop := range s.Properties {
		property.Add(addr.MustParseProperty(propRAddr), prop.PropertyState)
	}
	model.Properties = property.Properties
	return &model, nil
}

func (model *APIModel) MapTFAddr(tfPropTable map[string][]aztfq.TFResult) {
	for propAddr, prop := range model.Properties {
		if prop != nil {
			prop.AddTFMap("", propAddr, tfPropTable)
		}
	}
}

func (prop *Property) Add(propAddr addr.Property, state swagger.PropertyState) {
	if state.IsIgnored() {
		return
	}

	boolPtr := func(b bool) *bool {
		return &b
	}
	if len(propAddr) == 0 {
		prop.IsCovered = boolPtr(state.IsCovered())
		return
	}

	if prop.Properties == nil {
		prop.Properties = Properties{}
	}

	if _, ok := prop.Properties[propAddr[0]]; !ok {
		prop.Properties[propAddr[0]] = &Property{}
	}
	prop.Properties[propAddr[0]].Add(propAddr[1:], state)

	return
}

func (prop *Property) AddTFMap(prefix string, propAddr addr.PropertySegment, tfPropTable map[string][]aztfq.TFResult) {
	tfIndex := propAddr.StringInAztfqFormat()
	if prefix != "" {
		tfIndex = strings.Join([]string{prefix, tfIndex}, ".")
	}

	tfIndex = strings.ReplaceAll(tfIndex, `{}`, "*")
	tfMapping := TFMap{
		TFResource: make([]string, 0),
		TFAddr:     make([]string, 0),
	}

	if tfAddr, ok := tfPropTable[tfIndex]; ok {
		AztfqCnt++
		for _, tf := range tfAddr {
			tfMapping.TFAddr = append(tfMapping.TFAddr, tf.PropertyAddr)
			tfMapping.TFResource = append(tfMapping.TFResource, tf.ResourceType)
		}
		prop.TFMapping = &tfMapping
	}

	if prop.Properties != nil && len(prop.Properties) > 0 {
		for propAddr, prop := range prop.Properties {
			prop.AddTFMap(tfIndex, propAddr, tfPropTable)
		}
	}

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
