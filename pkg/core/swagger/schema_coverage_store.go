package swagger

import (
	"encoding/json"
	"fmt"

	"github.com/magodo/azure-rest-api-cov/pkg/core/addr"
)

type SchemaCoverageStore struct {
	SchemaAddr SchemaAddr           `json:"schema_addr"`
	Node       PropertyCoverageNode `json:"node"`
}

func NewSchemaCoverageStore() SchemaCoverageStore {
	return SchemaCoverageStore{
		Node: PropertyCoverageNode{
			Children: map[addr.PropertySegment]*PropertyCoverageNode{},
		},
	}
}

// Add adds a Property and record the coverage PropertyState in each property level.
// If the property is granted, then nothing will happen.
// If a property is added more than once, an error will be returned.
// NOTE: The granted property will be ignored
func (store *SchemaCoverageStore) Add(propAddr addr.Property, prop Property) error {
	if prop.PropertyState.IsGranted() {
		return nil
	}

	if _, _, ok := store.FindCoverage(propAddr); ok {
		return fmt.Errorf("%q has already been added to the coverage store", propAddr)
	}

	if len(propAddr) == 0 {
		return nil
	}

	store.Node.Add(propAddr, prop.PropertyState.IsCovered())
	return nil
}

func (store *SchemaCoverageStore) SchemaCoverage() (covered, total int) {
	return store.Node.CoveredAmount, store.Node.TotalAmount
}

func (store *SchemaCoverageStore) FindCoverage(prop addr.Property) (covered, total int, ok bool) {
	node := store.Node
	for _, addrSeg := range prop {
		tmpNode, ok := node.Children[addrSeg]
		if !ok {
			return 0, 0, false
		}
		node = *tmpNode
	}
	return node.CoveredAmount, node.TotalAmount, true
}

type PropertyCoverageNodes map[addr.PropertySegment]*PropertyCoverageNode

type PropertyCoverageNode struct {
	TotalAmount   int                   `json:"total_amount"`
	CoveredAmount int                   `json:"covered_amount"`
	Children      PropertyCoverageNodes `json:"children"`
}

func (nodes PropertyCoverageNodes) MarshalJSON() ([]byte, error) {
	m := map[string]*PropertyCoverageNode{}
	for k, v := range nodes {
		m[k.String()] = v
	}
	return json.Marshal(m)
}

func (nodes *PropertyCoverageNodes) UnmarshalJSON(b []byte) error {
	var m map[string]*PropertyCoverageNode
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	tmpNodes := PropertyCoverageNodes{}
	for k, v := range m {
		tmpNodes[addr.MustParsePropertySegment(k)] = v
	}
	*nodes = tmpNodes
	return nil
}

func (node *PropertyCoverageNode) Add(propAddr addr.Property, isCovered bool) {
	node.TotalAmount++
	if isCovered {
		node.CoveredAmount++
	}

	if len(propAddr) == 0 {
		return
	}

	child, ok := node.Children[propAddr[0]]
	if !ok {
		node.Children[propAddr[0]] = &PropertyCoverageNode{
			Children: map[addr.PropertySegment]*PropertyCoverageNode{},
		}
		child = node.Children[propAddr[0]]
	}
	child.Add(propAddr[1:], isCovered)

	return
}
