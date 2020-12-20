package addr_test

import (
	"testing"

	"github.com/magodo/azure-rest-api-cov/pkg/core/addr"
	"github.com/stretchr/testify/require"
)

func TestParsePropertySegment(t *testing.T) {
	cases := []struct {
		input  string
		expect addr.PropertySegment
		error  bool
	}{
		// 0
		{
			input:  "",
			expect: addr.PropertySegment{},
		},
		// 1
		{
			input:  "a",
			expect: addr.PropertySegment{Name: "a"},
		},
		// 2
		{
			input:  "a{v}",
			expect: addr.PropertySegment{Name: "a", DiscriminatorValue: "v"},
		},
		// 3
		{
			input:  "{v}",
			expect: addr.PropertySegment{DiscriminatorValue: "v"},
		},
		// 4
		{
			input:  `a\{`,
			expect: addr.PropertySegment{Name: `a{`},
		},
		// 5
		{
			input:  `a}`,
			expect: addr.PropertySegment{Name: `a}`},
		},
		// 6
		{
			input:  `a\/`,
			expect: addr.PropertySegment{Name: `a/`},
		},
		// 7
		{
			input:  `{v{}`,
			expect: addr.PropertySegment{DiscriminatorValue: "v{"},
		},
		// 8
		{
			input:  `{v\}}`,
			expect: addr.PropertySegment{DiscriminatorValue: "v}"},
		},
		// 9
		{
			input:  `{v\\}`,
			expect: addr.PropertySegment{DiscriminatorValue: `v\`},
		},
		// 10
		{
			input:  `a\\\{}\/{v\\{\}/}`,
			expect: addr.PropertySegment{Name: `a\{}/`, DiscriminatorValue: `v\{}/`},
		},
	}

	for idx, c := range cases {
		actual, err := addr.ParsePropertySegment(c.input)
		if c.error {
			require.Error(t, err, idx)
			continue
		}
		require.NoError(t, err, idx, err)
		require.Equal(t, c.expect, actual, idx)
	}
}

func TestParseProperty(t *testing.T) {
	cases := []struct {
		input  string
		expect addr.Property
		error  bool
	}{
		// 0
		{
			input:  "",
			expect: nil,
		},
		// 1
		{
			input:  "a",
			expect: addr.Property{addr.PropertySegment{Name: "a"}},
		},
		// 2
		{
			input:  "a{v}",
			expect: addr.Property{addr.PropertySegment{Name: "a", DiscriminatorValue: "v"}},
		},
		// 3
		{
			input:  "{v}",
			expect: addr.Property{addr.PropertySegment{DiscriminatorValue: "v"}},
		},
		// 4
		{
			input:  `a\{`,
			expect: addr.Property{addr.PropertySegment{Name: `a{`}},
		},
		// 5
		{
			input:  `a}`,
			expect: addr.Property{addr.PropertySegment{Name: `a}`}},
		},
		// 6
		{
			input:  `a\/`,
			expect: addr.Property{addr.PropertySegment{Name: `a/`}},
		},
		// 7
		{
			input:  `{v{}`,
			expect: addr.Property{addr.PropertySegment{DiscriminatorValue: "v{"}},
		},
		// 8
		{
			input:  `{v\}}`,
			expect: addr.Property{addr.PropertySegment{DiscriminatorValue: "v}"}},
		},
		// 9
		{
			input:  `{v\\}`,
			expect: addr.Property{addr.PropertySegment{DiscriminatorValue: `v\`}},
		},
		// 10
		{
			input:  `a\\\{}\/{v\\{\}/}`,
			expect: addr.Property{addr.PropertySegment{Name: `a\{}/`, DiscriminatorValue: `v\{}/`}},
		},
		// 11
		{
			input: `a\\\{}\/{v1\\{\}/}/b\\\{}\/{v2\\{\}/}/c`,
			expect: addr.Property{
				addr.PropertySegment{Name: `a\{}/`, DiscriminatorValue: `v1\{}/`},
				addr.PropertySegment{Name: `b\{}/`, DiscriminatorValue: `v2\{}/`},
				addr.PropertySegment{Name: `c`},
			},
		},
	}

	for idx, c := range cases {
		actual, err := addr.ParseProperty(c.input)
		if c.error {
			require.Error(t, err, idx)
			continue
		}
		require.NoError(t, err, idx, err)
		require.Equal(t, c.expect, actual, idx)
	}
}
