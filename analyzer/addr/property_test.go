package addr_test

import (
	"testing"

	"github.com/magodo/azure-rest-api-cov/analyzer/addr"

	"github.com/stretchr/testify/require"
)

func TestProperty_Contains(t *testing.T) {
	cases := []struct {
		addr     addr.Property
		oaddr    addr.Property
		contains bool
	}{
		// 0
		{
			addr: nil,
			oaddr: addr.Property{
				{Name: "p1"},
			},
			contains: true,
		},
		// 1
		{
			addr: addr.Property{
				{Name: "p1"},
			},
			oaddr: addr.Property{
				{Name: "p1"},
				{Name: "p2"},
			},
			contains: true,
		},
		// 2
		{
			addr: addr.Property{
				{Name: "p1"},
				{Name: "p2"},
			},
			oaddr: addr.Property{
				{Name: "p1"},
			},
			contains: false,
		},
		// 3
		{
			addr: addr.Property{
				{Name: "p1"},
			},
			oaddr: addr.Property{
				{Name: "p1", DiscriminatorValue: "v1"},
			},
			contains: true,
		},
		// 4
		{
			addr: addr.Property{
				{Name: "p1", DiscriminatorValue: "v1"},
			},
			oaddr: addr.Property{
				{Name: "p1", DiscriminatorValue: "v1"},
			},
			contains: false,
		},
		// 5
		{
			addr: addr.Property{
				{Name: "p1", DiscriminatorValue: "v1"},
			},
			oaddr: addr.Property{
				{Name: "p1", DiscriminatorValue: "v2"},
			},
			contains: false,
		},
		// 6
		{
			addr: addr.Property{
				{Name: "p1"},
			},
			oaddr: addr.Property{
				{Name: "p1", DiscriminatorValue: "v2"},
				{Name: "p2"},
			},
			contains: true,
		},
		// 7
		{
			addr: addr.Property{
				{Name: "p1"},
			},
			oaddr: addr.Property{
				{Name: "p1"},
			},
			contains: false,
		},
	}
	for idx, c := range cases {
		require.Equal(t, c.contains, c.addr.Contains(c.oaddr), idx)
	}
}

func TestProperty_Equals(t *testing.T) {
	cases := []struct {
		addr  addr.Property
		oaddr addr.Property
		equal bool
	}{
		{
			addr:  addr.Property{},
			oaddr: addr.Property{},
			equal: true,
		},
		{
			addr: addr.Property{
				{Name: "p1"},
			},
			oaddr: addr.Property{
				{Name: "p1"},
			},
			equal: true,
		},
		{
			addr: addr.Property{
				{Name: "p1"},
				{Name: "p2"},
			},
			oaddr: addr.Property{
				{Name: "p1"},
			},
			equal: false,
		},
	}
	for idx, c := range cases {
		require.Equal(t, c.equal, c.addr.Equals(c.oaddr), idx)
	}
}

func TestProperty_String(t *testing.T) {
	cases := []struct {
		input  addr.Property
		expect string
	}{
		// 0
		{
			input:  addr.Property{},
			expect: "",
		},
		// 1
		{
			input:  addr.Property{{Name: "p1"}},
			expect: "p1",
		},
		// 2
		{
			input:  addr.Property{{Name: "p1"}, {Name: "p2"}},
			expect: "p1/p2",
		},
		// 3
		{
			input:  addr.Property{{Name: "p1"}, {Name: "p2", DiscriminatorValue: "v1"}},
			expect: "p1/p2{v1}",
		},
		// 4
		{
			input:  addr.Property{{Name: "p1/"}, {Name: "p2"}},
			expect: `p1\//p2`,
		},
		// 5
		{
			input:  addr.Property{{Name: `p1\`}, {}, {Name: "p2"}},
			expect: `p1\\//p2`,
		},
		// 6
		{
			input:  addr.Property{{Name: `p1\`}, {Name: "p2"}},
			expect: `p1\\/p2`,
		},
		// 7
		{
			input:  addr.Property{{Name: `p1{`}, {Name: "p2"}},
			expect: `p1\{/p2`,
		},
		// 8
		{
			input:  addr.Property{{Name: `p1}`}, {Name: "p2"}},
			expect: `p1}/p2`,
		},
		// 9
		{
			input:  addr.Property{{Name: "p1"}, {Name: "p2", DiscriminatorValue: `v1}`}},
			expect: `p1/p2{v1\}}`,
		},
		// 10
		{
			input:  addr.Property{{Name: "p1"}, {Name: "p2", DiscriminatorValue: `v1{`}},
			expect: `p1/p2{v1{}`,
		},
		// 11
		{
			input:  addr.Property{{Name: "p1"}, {Name: "p2", DiscriminatorValue: `v1/`}},
			expect: `p1/p2{v1/}`,
		},
		// 12
		{
			expect: `p1/p2{v1\\}`,
			input:  addr.Property{{Name: "p1"}, {Name: "p2", DiscriminatorValue: `v1\`}},
		},
	}
	for idx, c := range cases {
		require.Equal(t, c.expect, c.input.String(), idx)
	}
}

func TestProperty_MarshalJSON(t *testing.T) {
	cases := []struct {
		addr   addr.Property
		expect string
	}{
		{
			addr:   addr.Property{},
			expect: `""`,
		},
		{
			addr:   addr.Property{{Name: "p1"}, {Name: "p2"}},
			expect: `"p1/p2"`,
		},
		{
			addr:   addr.Property{{Name: "p1"}, {Name: "p2", DiscriminatorValue: "v1"}},
			expect: `"p1/p2{v1}"`,
		},
	}

	for idx, c := range cases {
		actual, err := c.addr.MarshalJSON()
		require.NoError(t, err)
		if c.expect == "" {
			require.Equal(t, c.expect, string(actual), idx)
		} else {
			require.JSONEq(t, c.expect, string(actual), idx)
		}
	}
}

func TestProperty_UnmarshalJSON(t *testing.T) {
	cases := []struct {
		input  string
		expect addr.Property
	}{
		{
			input:  `""`,
			expect: nil,
		},
		{
			input:  `"p1/p2"`,
			expect: addr.Property{{Name: "p1"}, {Name: "p2"}},
		},
		{
			input:  `"p1/p2{v1}"`,
			expect: addr.Property{{Name: "p1"}, {Name: "p2", DiscriminatorValue: "v1"}},
		},
	}
	for idx, c := range cases {
		var addr addr.Property
		err := addr.UnmarshalJSON([]byte(c.input))
		require.NoError(t, err, idx)
		require.Equal(t, c.expect, addr, idx)
	}
}
