package swagger

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLookupSpecPath(t *testing.T) {
	pwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	specBasePath := filepath.Join(filepath.Dir(pwd), "testdata", "spec")
	cases := []struct {
		locator        SpecLocator
		expectSpecPath string
		expectError    bool
	}{
		{
			locator: SpecLocator{
				ApiVersion: "2020-12-01",
				ApiPath:    "/{host}/{bar}",
			},
			expectSpecPath: "dummy/resource-manager/Microsoft.Dummy/preview/2020-12-01-preview/bar.json",
		},
		{
			locator: SpecLocator{
				ApiVersion: "2020-12-01",
				ApiPath:    "/{host}/{foo}",
			},
			expectSpecPath: "dummy/resource-manager/Microsoft.Dummy/stable/2020-12-01/foo.json",
		},
		{
			locator: SpecLocator{
				ApiVersion: "non-exist",
				ApiPath:    "/{host}/{foo}",
			},
			expectError: true,
		},
	}

	for idx, c := range cases {
		path, err := LookupSpecPath(specBasePath, c.locator)
		if c.expectError {
			require.Error(t, err, idx)
			continue
		}
		require.NoError(t, err, idx)
		require.Equal(t, c.expectSpecPath, path, idx)
	}
}
