package propmatch

import (
	"testing"

	"github.com/magodo/azure-rest-api-cov/analyzer/addr"
	"github.com/stretchr/testify/require"
)

func TestPropMatch(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		pattern string
		expect  bool
	}{
		{
			name:    "empty matches empty",
			input:   "",
			pattern: "",
			expect:  true,
		},
		{
			name:    "empty matches any wildcard",
			input:   "",
			pattern: "**",
			expect:  true,
		},
		{
			name:    "empty doesn't matches wildcard",
			input:   "",
			pattern: "*",
		},
		{
			name:    "literally matches",
			input:   "a/b",
			pattern: "a/b",
			expect:  true,
		},
		{
			name:    "literally not matches",
			input:   "a/b",
			pattern: "a/c",
		},
		{
			name:    "wildcard match",
			input:   "a/b",
			pattern: "a/*",
			expect:  true,
		},
		{
			name:    "wildcard not match match (too short)",
			input:   "a",
			pattern: "a/*",
		},
		{
			name:    "wildcard not match match (too long)",
			input:   "a/b/c",
			pattern: "a/*",
		},
		{
			name:    "wildcard not match match (different prefix)",
			input:   "b/b",
			pattern: "a/*",
		},
		{
			name:    "wildcard not match match (different suffix)",
			input:   "b/b",
			pattern: "*/a",
		},
		{
			name:    "any wildcard match (prefix)",
			input:   "a",
			pattern: "**/a",
			expect:  true,
		},
		{
			name:    "any wildcard match (suffix)",
			input:   "a",
			pattern: "a/**",
			expect:  true,
		},
		{
			name:    "any wildcard match (middle nothing)",
			input:   "a/b",
			pattern: "a/**/b",
			expect:  true,
		},
		{
			name:    "any wildcard match (middle more)",
			input:   "a/x/y/z/b",
			pattern: "a/**/b",
			expect:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expect,
				propMatch(addr.MustParseProperty(tt.input), addr.MustParseProperty(tt.pattern)))
		})
	}
}
