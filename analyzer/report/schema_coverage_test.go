package report_test

import (
	"fmt"
	"github.com/magodo/azure-rest-api-cov/analyzer/operation"
	"testing"

	"github.com/magodo/azure-rest-api-cov/analyzer/report"

	"github.com/magodo/azure-rest-api-cov/analyzer/swagger"

	"github.com/magodo/azure-rest-api-cov/analyzer/addr"
	"github.com/stretchr/testify/require"
)

func TestSchemaCoverageStore_Add(t *testing.T) {
	type swgproperty struct {
		addr addr.Property
		prop swagger.Property
	}
	cases := []struct {
		propertiesToAdd []swgproperty
		expectStore     report.SchemaCoverage
	}{
		{
			propertiesToAdd: []swgproperty{
				{
					addr: addr.MustParseProperty("prop1/not_covered"),
					prop: swagger.Property{
						PropertyState: swagger.PropertyStateUncovered,
					},
				},
				{
					addr: addr.MustParseProperty("prop1/covered"),
					prop: swagger.Property{
						PropertyState: swagger.PropertyStateCovered,
					},
				},
				{
					addr: addr.MustParseProperty("prop2/covered"),
					prop: swagger.Property{
						PropertyState: swagger.PropertyStateCovered,
					},
				},
				{
					addr: addr.MustParseProperty("prop3"),
					prop: swagger.Property{
						PropertyState: swagger.NewPropertyState(swagger.PropStateGranted, "granted"),
					},
				},
				{
					addr: addr.MustParseProperty("prop_ignored"),
					prop: swagger.Property{
						PropertyState: swagger.PropertyStateIgnoredForReqROProp,
					},
				},
			},
			expectStore: report.SchemaCoverage{
				Node: report.PropertyCoverageNode{
					TotalAmount:   4,
					CoveredAmount: 3,
					Children: map[addr.PropertySegment]*report.PropertyCoverageNode{
						addr.MustParsePropertySegment("prop1"): {
							TotalAmount:   2,
							CoveredAmount: 1,
							Children: map[addr.PropertySegment]*report.PropertyCoverageNode{
								addr.MustParsePropertySegment("not_covered"): {
									TotalAmount:   1,
									CoveredAmount: 0,
									Children:      map[addr.PropertySegment]*report.PropertyCoverageNode{},
								},
								addr.MustParsePropertySegment("covered"): {
									TotalAmount:   1,
									CoveredAmount: 1,
									Children:      map[addr.PropertySegment]*report.PropertyCoverageNode{},
								},
							},
						},
						addr.MustParsePropertySegment("prop2"): {
							TotalAmount:   1,
							CoveredAmount: 1,
							Children: map[addr.PropertySegment]*report.PropertyCoverageNode{
								addr.MustParsePropertySegment("covered"): {
									TotalAmount:   1,
									CoveredAmount: 1,
									Children:      map[addr.PropertySegment]*report.PropertyCoverageNode{},
								},
							},
						},
						addr.MustParsePropertySegment("prop3"): {
							TotalAmount:   1,
							CoveredAmount: 1,
							Granted:       true,
							GrantComment:  "granted",
							Children:      map[addr.PropertySegment]*report.PropertyCoverageNode{},
						},
					},
				},
			},
		},
	}

	for idx, c := range cases {
		store := report.SchemaCoverage{
			Node: report.PropertyCoverageNode{
				Children: report.PropertyCoverageNodes{},
			},
		}
		for idx2, prop := range c.propertiesToAdd {
			require.NoError(t, store.Add(prop.addr, prop.prop), fmt.Sprintf("%d.%d", idx, idx2))
		}
		require.Equal(t, c.expectStore, store, idx)
	}
}

func TestSchemaCoverageStore_FindCoverage(t *testing.T) {
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
		store               report.SchemaCoverage
		subtest             []subtest
		expectTotalCoverage result
	}{
		{
			store: report.SchemaCoverage{
				Node: report.PropertyCoverageNode{
					TotalAmount:   3,
					CoveredAmount: 1,
					Children: map[addr.PropertySegment]*report.PropertyCoverageNode{
						addr.MustParsePropertySegment("prop1"): {
							TotalAmount:   2,
							CoveredAmount: 1,
							Children: map[addr.PropertySegment]*report.PropertyCoverageNode{
								addr.MustParsePropertySegment("nest"): {
									TotalAmount:   2,
									CoveredAmount: 1,
									Children: map[addr.PropertySegment]*report.PropertyCoverageNode{
										addr.MustParsePropertySegment("covered"): {
											TotalAmount:   1,
											CoveredAmount: 1,
											Children:      map[addr.PropertySegment]*report.PropertyCoverageNode{},
										},
										addr.MustParsePropertySegment("uncovered"): {
											TotalAmount:   1,
											CoveredAmount: 0,
											Children:      map[addr.PropertySegment]*report.PropertyCoverageNode{},
										},
									},
								},
							},
						},
						addr.MustParsePropertySegment("uncovered"): {
							TotalAmount:   1,
							CoveredAmount: 0,
							Children:      map[addr.PropertySegment]*report.PropertyCoverageNode{},
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

func TestNewSchemaCoverageStore(t *testing.T) {
	cases := []struct {
		opInfo operation.Info
		expect *report.SchemaCoverage
	}{
		{
			opInfo: operation.Info{
				IsPreview: true,
				Model: &swagger.Schema{
					SchemaMetaInfo: swagger.SchemaMetaInfo{
						Name:           "foo",
						SwaggerRelPath: "foo/resource-manager/Microsoft.Foo/preview/2021-02-01/foo.json",
					},
					Properties: swagger.Properties{
						"prop1/covered": &swagger.Property{
							PropertyState: swagger.PropertyStateCovered,
						},
						"prop1/not_covered": {
							PropertyState: swagger.PropertyStateUncovered,
						},
						"prop2/covered": {
							PropertyState: swagger.PropertyStateCovered,
						},
						"prop2/granted": {
							PropertyState: swagger.NewPropertyState(swagger.PropStateGranted, "granted"),
						},
						"prop_granted": {
							PropertyState: swagger.NewPropertyState(swagger.PropStateGranted, "granted"),
						},
					},
				},
			},
			expect: &report.SchemaCoverage{
				Name:      "foo",
				URL:       "foo/resource-manager/Microsoft.Foo/preview/2021-02-01/foo.json",
				IsPreview: true,
				Node: report.PropertyCoverageNode{
					TotalAmount:   5,
					CoveredAmount: 4,
					Children: map[addr.PropertySegment]*report.PropertyCoverageNode{
						addr.MustParsePropertySegment("prop1"): {
							TotalAmount:   2,
							CoveredAmount: 1,
							Children: map[addr.PropertySegment]*report.PropertyCoverageNode{
								addr.MustParsePropertySegment("not_covered"): {
									TotalAmount:   1,
									CoveredAmount: 0,
									Children:      map[addr.PropertySegment]*report.PropertyCoverageNode{},
								},
								addr.MustParsePropertySegment("covered"): {
									TotalAmount:   1,
									CoveredAmount: 1,
									Children:      map[addr.PropertySegment]*report.PropertyCoverageNode{},
								},
							},
						},
						addr.MustParsePropertySegment("prop2"): {
							TotalAmount:   2,
							CoveredAmount: 2,
							Children: map[addr.PropertySegment]*report.PropertyCoverageNode{
								addr.MustParsePropertySegment("covered"): {
									TotalAmount:   1,
									CoveredAmount: 1,
									Children:      map[addr.PropertySegment]*report.PropertyCoverageNode{},
								},
								addr.MustParsePropertySegment("granted"): {
									TotalAmount:   1,
									CoveredAmount: 1,
									Granted:       true,
									GrantComment:  "granted",
									Children:      map[addr.PropertySegment]*report.PropertyCoverageNode{},
								},
							},
						},

						addr.MustParsePropertySegment("prop_granted"): {
							TotalAmount:   1,
							CoveredAmount: 1,
							Granted:       true,
							GrantComment:  "granted",
							Children:      map[addr.PropertySegment]*report.PropertyCoverageNode{},
						},
					},
				},
			},
		},
	}

	for idx, c := range cases {
		store, err := report.NewSchemaCoverage("", &c.opInfo)
		require.NoError(t, err, idx)
		require.Equal(t, c.expect, store, idx)
	}
}
