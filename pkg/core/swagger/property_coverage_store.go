package swagger

import (
	"fmt"

	"github.com/magodo/azure-rest-api-cov/pkg/core/addr"
)

type PropertyCoverageStore struct {
	node propertyCoverageNode
}

func NewPropertyCoverageStore() PropertyCoverageStore {
	return PropertyCoverageStore{
		node: propertyCoverageNode{
			Children: map[addr.PropertySegment]*propertyCoverageNode{},
		},
	}
}

// Add adds a Property and record the coverage PropertyState in each property level.
// If the property is granted, then nothing will happen.
// If a property is added more than once, an error will be returned.
// NOTE: The granted property will be ignored
func (store *PropertyCoverageStore) Add(propAddr addr.Property, prop Property) error {
	if prop.PropertyState.IsGranted() {
		return nil
	}

	if _, _, ok := store.FindCoverage(propAddr); ok {
		return fmt.Errorf("%q has already been added to the coverage store", propAddr)
	}

	if len(propAddr) == 0 {
		return nil
	}

	store.node.add(propAddr, prop.PropertyState.IsCovered())
	return nil
}

func (store *PropertyCoverageStore) SchemaCoverage() (covered, total int) {
	return store.node.CoveredAmount, store.node.TotalAmount
}

func (store *PropertyCoverageStore) FindCoverage(prop addr.Property) (covered, total int, ok bool) {
	node := store.node
	for _, addrSeg := range prop {
		tmpNode, ok := node.Children[addrSeg]
		if !ok {
			return 0, 0, false
		}
		node = *tmpNode
	}
	return node.CoveredAmount, node.TotalAmount, true
}

type propertyCoverageNode struct {
	TotalAmount   int
	CoveredAmount int
	Children      map[addr.PropertySegment]*propertyCoverageNode
}

func (node *propertyCoverageNode) add(propAddr addr.Property, isCovered bool) {
	node.TotalAmount++
	if isCovered {
		node.CoveredAmount++
	}

	if len(propAddr) == 0 {
		return
	}

	child, ok := node.Children[propAddr[0]]
	if !ok {
		node.Children[propAddr[0]] = &propertyCoverageNode{
			Children: map[addr.PropertySegment]*propertyCoverageNode{},
		}
		child = node.Children[propAddr[0]]
	}
	child.add(propAddr[1:], isCovered)

	return
}
