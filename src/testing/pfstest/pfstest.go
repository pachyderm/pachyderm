package pfstest

import (
	"testing"

	pachclient "github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
)

// TestAPI is a test suite which ensure objects returned from newClient correctly implement the PFS API.
func TestAPI(t *testing.T, newClient func(t testing.TB) pfs.APIClient) {
	tests := []struct {
		Name string
		F    func(t *testing.T, client pfs.APIClient)
	}{
		// Add more tests here
		{"InvalidRepo", testInvalidRepo},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			t.Parallel()
			client := newClient(t)
			test.F(t, client)
		})
	}
}

func testInvalidRepo(t *testing.T, client pfs.APIClient) {
	pclient := pachclient.APIClient{
		PfsAPIClient: client,
	}
	require.YesError(t, pclient.CreateRepo("/repo"))

	require.NoError(t, pclient.CreateRepo("lenny"))
	require.NoError(t, pclient.CreateRepo("lenny123"))
	require.NoError(t, pclient.CreateRepo("lenny_123"))
	require.NoError(t, pclient.CreateRepo("lenny-123"))

	require.YesError(t, pclient.CreateRepo("lenny.123"))
	require.YesError(t, pclient.CreateRepo("lenny:"))
	require.YesError(t, pclient.CreateRepo("lenny,"))
	require.YesError(t, pclient.CreateRepo("lenny#"))
}
