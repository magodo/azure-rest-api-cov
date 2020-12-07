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
				ApiVersion: "2020-05-01",
				ApiPath:    "/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Network/applicationGateways/{applicationGatewayName}",
			},
			expectSpecPath: "network/resource-manager/Microsoft.Network/stable/2020-05-01/applicationGateway.json",
		},
		{
			locator: SpecLocator{
				ApiVersion: "2015-05-01-preview",
				ApiPath:    "/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Network/applicationGateways/{applicationGatewayName}",
			},
			expectSpecPath: "network/resource-manager/Microsoft.Network/preview/2015-05-01-preview/network.json",
		},
		{
			locator: SpecLocator{
				ApiVersion: "non-exist",
				ApiPath:    "/subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/providers/Microsoft.Network/applicationGateways/{applicationGatewayName}",
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
