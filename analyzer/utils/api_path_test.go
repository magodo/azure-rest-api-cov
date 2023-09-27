package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeAPIPath(t *testing.T) {
	cases := []struct {
		input  string
		expect string
	}{
		{
			input:  "",
			expect: "",
		},
		{
			input:  "/",
			expect: "/",
		},
		{
			input:  "/foos",
			expect: "/FOOS",
		},
		{
			input:  "/foos/{fooname}/bars/{barname}",
			expect: "/FOOS/{}/BARS/{}",
		},
	}
	for _, tt := range cases {
		t.Run(tt.input, func(t *testing.T) {
			require.Equal(t, tt.expect, NormalizeAPIPath(tt.input))
		})
	}
}

func TestSpecializeAPIPath(t *testing.T) {
	cases := []struct {
		input  string
		em     map[string]string
		expect string
	}{
		{
			input:  "",
			expect: "",
		},
		{
			input:  "/",
			expect: "/",
		},
		{
			input:  "/foos",
			expect: "/FOOS",
		},
		{
			input:  "/foos/{fooname}/bars/{barname}",
			em:     map[string]string{"fooname": "default"},
			expect: "/FOOS/DEFAULT/BARS/{}",
		},
	}
	for _, tt := range cases {
		t.Run(tt.input, func(t *testing.T) {
			require.Equal(t, tt.expect, SpecializeAPIPath(tt.input, tt.em))
		})
	}
}
