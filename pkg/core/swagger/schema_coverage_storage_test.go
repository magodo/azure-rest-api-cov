package swagger

import (
	"fmt"
	"testing"

	"github.com/magodo/azure-rest-api-cov/pkg/core/addr"
	"github.com/stretchr/testify/require"
)

func TestNewPropertyCoverageStore_Add(t *testing.T) {
	type swgproperty struct {
		addr addr.Property
		prop Property
	}
	cases := []struct {
		propertiesToAdd []swgproperty
		expectStore     SchemaCoverageStore
	}{
		{
			propertiesToAdd: []swgproperty{
				{
					addr: addr.MustParseProperty("prop1/not_covered"),
					prop: Property{
						PropertyState: NewPropertyState(PropStateUncovered, ""),
					},
				},
				{
					addr: addr.MustParseProperty("prop1/covered"),
					prop: Property{
						PropertyState: NewPropertyState(PropStateCovered, ""),
					},
				},
				{
					addr: addr.MustParseProperty("prop2/covered"),
					prop: Property{
						PropertyState: NewPropertyState(PropStateCovered, ""),
					},
				},
				{
					addr: addr.MustParseProperty("prop_granted"),
					prop: Property{
						PropertyState: NewPropertyState(PropStateGranted, ""),
					},
				},
			},
			expectStore: SchemaCoverageStore{
				Node: PropertyCoverageNode{
					TotalAmount:   3,
					CoveredAmount: 2,
					Children: map[addr.PropertySegment]*PropertyCoverageNode{
						addr.MustParsePropertySegment("prop1"): {
							TotalAmount:   2,
							CoveredAmount: 1,
							Children: map[addr.PropertySegment]*PropertyCoverageNode{
								addr.MustParsePropertySegment("not_covered"): {
									TotalAmount:   1,
									CoveredAmount: 0,
									Children:      map[addr.PropertySegment]*PropertyCoverageNode{},
								},
								addr.MustParsePropertySegment("covered"): {
									TotalAmount:   1,
									CoveredAmount: 1,
									Children:      map[addr.PropertySegment]*PropertyCoverageNode{},
								},
							},
						},
						addr.MustParsePropertySegment("prop2"): {
							TotalAmount:   1,
							CoveredAmount: 1,
							Children: map[addr.PropertySegment]*PropertyCoverageNode{
								addr.MustParsePropertySegment("covered"): {
									TotalAmount:   1,
									CoveredAmount: 1,
									Children:      map[addr.PropertySegment]*PropertyCoverageNode{},
								},
							},
						},
					},
				},
			},
		},
	}

	for idx, c := range cases {
		store := NewSchemaCoverageStore()
		for idx2, prop := range c.propertiesToAdd {
			require.NoError(t, store.Add(prop.addr, prop.prop), fmt.Sprintf("%d.%d", idx, idx2))
		}
		require.Equal(t, c.expectStore, store, idx)
	}
}

func TestNewPropertyCoverageStore_FindCoverage(t *testing.T) {
	type result struct {
		total   int
		covered int
		ok      bool
	}

	type subtest struct {
		propAddr string
		expect   result
	}

	cases := []struct {
		store               SchemaCoverageStore
		subtest             []subtest
		expectTotalCoverage result
	}{
		{
			store: SchemaCoverageStore{
				Node: PropertyCoverageNode{
					TotalAmount:   3,
					CoveredAmount: 1,
					Children: map[addr.PropertySegment]*PropertyCoverageNode{
						addr.MustParsePropertySegment("prop1"): {
							TotalAmount:   2,
							CoveredAmount: 1,
							Children: map[addr.PropertySegment]*PropertyCoverageNode{
								addr.MustParsePropertySegment("nest"): {
									TotalAmount:   2,
									CoveredAmount: 1,
									Children: map[addr.PropertySegment]*PropertyCoverageNode{
										addr.MustParsePropertySegment("covered"): {
											TotalAmount:   1,
											CoveredAmount: 1,
											Children:      map[addr.PropertySegment]*PropertyCoverageNode{},
										},
										addr.MustParsePropertySegment("uncovered"): {
											TotalAmount:   1,
											CoveredAmount: 0,
											Children:      map[addr.PropertySegment]*PropertyCoverageNode{},
										},
									},
								},
							},
						},
						addr.MustParsePropertySegment("uncovered"): {
							TotalAmount:   1,
							CoveredAmount: 0,
							Children:      map[addr.PropertySegment]*PropertyCoverageNode{},
						},
					},
				},
			},
			subtest: []subtest{
				{
					propAddr: "non_existed",
					expect: result{
						ok: false,
					},
				},
				{
					propAddr: "uncovered",
					expect: result{
						total: 1,
						ok:    true,
					},
				},
				{
					propAddr: "prop1/nest/covered",
					expect: result{
						total:   1,
						covered: 1,
						ok:      true,
					},
				},
				{
					propAddr: "prop1/nest/uncovered",
					expect: result{
						total:   1,
						covered: 0,
						ok:      true,
					},
				},
				{
					propAddr: "prop1/nest",
					expect: result{
						total:   2,
						covered: 1,
						ok:      true,
					},
				},
				{
					propAddr: "prop1",
					expect: result{
						total:   2,
						covered: 1,
						ok:      true,
					},
				},
			},
			expectTotalCoverage: result{
				total:   3,
				covered: 1,
				ok:      true,
			},
		},
	}

	for idx, c := range cases {
		for idx2, subtest := range c.subtest {
			covered, total, ok := c.store.FindCoverage(addr.MustParseProperty(subtest.propAddr))
			result := result{
				total:   total,
				covered: covered,
				ok:      ok,
			}
			require.Equal(t, subtest.expect, result, fmt.Sprintf("%d.%d", idx, idx2))
		}

		covered, total := c.store.SchemaCoverage()
		result := result{
			total:   total,
			covered: covered,
			ok:      true,
		}
		require.Equal(t, c.expectTotalCoverage, result, idx)
	}
}
