package addr

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseProperty(t *testing.T) {
	cases := []struct {
		input  string
		expect Property
		error  bool
	}{
		{
			input:  "",
			expect: nil,
		},
		{
			input:  "p1",
			expect: Property{{name: "p1"}},
		},
		{
			input:  "p1/p2",
			expect: Property{{name: "p1"}, {name: "p2"}},
		},
		{
			input:  "p1/p2{v1}",
			expect: Property{{name: "p1"}, {name: "p2", discriminatorValue: "v1"}},
		},
		{
			input: "p1/p2}",
			error: true,
		},
	}

	for idx, c := range cases {
		actual, err := ParseProperty(c.input)
		if c.error {
			require.Error(t, err, idx)
			continue
		}
		require.NoError(t, err, idx)
		require.Equal(t, c.expect, actual, idx)
	}
}

func TestProperty_Contains(t *testing.T) {
	cases := []struct {
		addr     Property
		oaddr    Property
		contains bool
	}{
		// 0
		{
			addr: nil,
			oaddr: Property{
				{name: "p1"},
			},
			contains: true,
		},
		// 1
		{
			addr: Property{
				{name: "p1"},
			},
			oaddr: Property{
				{name: "p1"},
				{name: "p2"},
			},
			contains: true,
		},
		// 2
		{
			addr: Property{
				{name: "p1"},
				{name: "p2"},
			},
			oaddr: Property{
				{name: "p1"},
			},
			contains: false,
		},
		// 3
		{
			addr: Property{
				{name: "p1"},
			},
			oaddr: Property{
				{name: "p1", discriminatorValue: "v1"},
			},
			contains: true,
		},
		// 4
		{
			addr: Property{
				{name: "p1", discriminatorValue: "v1"},
			},
			oaddr: Property{
				{name: "p1", discriminatorValue: "v1"},
			},
			contains: false,
		},
		// 5
		{
			addr: Property{
				{name: "p1", discriminatorValue: "v1"},
			},
			oaddr: Property{
				{name: "p1", discriminatorValue: "v2"},
			},
			contains: false,
		},
		// 6
		{
			addr: Property{
				{name: "p1"},
			},
			oaddr: Property{
				{name: "p1", discriminatorValue: "v2"},
				{name: "p2"},
			},
			contains: true,
		},
	}
	for idx, c := range cases {
		require.Equal(t, c.contains, c.addr.Contains(c.oaddr), idx)
	}
}

func TestProperty_Equals(t *testing.T) {
	cases := []struct {
		addr  Property
		oaddr Property
		equal bool
	}{
		{
			addr:  Property{},
			oaddr: Property{},
			equal: true,
		},
		{
			addr: Property{
				{name: "p1"},
			},
			oaddr: Property{
				{name: "p1"},
			},
			equal: true,
		},
		{
			addr: Property{
				{name: "p1"},
				{name: "p2"},
			},
			oaddr: Property{
				{name: "p1"},
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
		input  Property
		expect string
	}{
		{
			input:  Property{},
			expect: "",
		},
		{
			input:  Property{{name: "p1"}},
			expect: "p1",
		},
		{
			input:  Property{{name: "p1"}, {name: "p2"}},
			expect: "p1/p2",
		},
		{
			input:  Property{{name: "p1"}, {name: "p2", discriminatorValue: "v1"}},
			expect: "p1/p2{v1}",
		},
	}
	for idx, c := range cases {
		require.Equal(t, c.expect, c.input.String(), idx)
	}
}

func TestProperty_MarshalJSON(t *testing.T) {
	cases := []struct {
		addr   Property
		expect string
	}{
		{
			addr:   Property{},
			expect: `""`,
		},
		{
			addr:   Property{{name: "p1"}, {name: "p2"}},
			expect: `"p1/p2"`,
		},
		{
			addr:   Property{{name: "p1"}, {name: "p2", discriminatorValue: "v1"}},
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
		expect Property
	}{
		{
			input:  `""`,
			expect: nil,
		},
		{
			input:  `"p1/p2"`,
			expect: Property{{name: "p1"}, {name: "p2"}},
		},
		{
			input:  `"p1/p2{v1}"`,
			expect: Property{{name: "p1"}, {name: "p2", discriminatorValue: "v1"}},
		},
	}
	for idx, c := range cases {
		var addr Property
		err := addr.UnmarshalJSON([]byte(c.input))
		require.NoError(t, err, idx)
		require.Equal(t, c.expect, addr, idx)
	}
}
