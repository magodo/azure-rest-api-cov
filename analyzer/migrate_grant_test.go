package analyzer

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMigrateGrant(t *testing.T) {
	specRootDir := filepath.Join("testdata", "spec")
	grantRootDir := filepath.Join("testdata", "grant_migrate")

	cases := []struct {
		rp         string
		spec       string
		srcVersion string
		dstVersion string
		expect     map[string][]byte
	}{
		{
			rp:         "dummy",
			spec:       "foo.json",
			srcVersion: "2020-12-01",
			dstVersion: "2021-01-01",
			expect: map[string][]byte{
				filepath.Join("dummy", "resource-manager", "Microsoft.Dummy", "stable", "2021-01-01", "foo.yaml"): []byte(`
def_grant_update:
  property_grant_rules: |
    p1: grant`),
			},
		},
	}

	for idx, c := range cases {
		out, err := migrateGrant(specRootDir, grantRootDir, c.rp, c.spec, c.srcVersion, c.dstVersion)
		require.NoError(t, err, idx)
		require.Equal(t, len(c.expect), len(out), idx)
		for k, v := range out {
			require.YAMLEq(t, string(c.expect[k]), string(v), idx)
		}
	}
}
