package minipach

import (
	"context"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

const (
	// RootToken is the hard-coded admin token used on all activated test clusters
	RootToken = "iamroot"
)

// ActivateAuth activates the auth service in the test cluster, if it isn't already enabled
func ActivateAuth(tb testing.TB, client *client.APIClient) {
	tb.Helper()
	client = client.WithCtx(context.Background())
	client.SetAuthToken(RootToken)

	ActivateEnterprise(tb, client)

	_, err := client.Activate(client.Ctx(),
		&auth.ActivateRequest{RootToken: RootToken},
	)
	if err != nil && !strings.HasSuffix(err.Error(), "already activated") {
		tb.Fatalf("could not activate auth service: %v", err.Error())
	}

	// Activate auth for PPS
	_, err = client.PfsAPIClient.ActivateAuth(client.Ctx(), &pfs.ActivateAuthRequest{})
	require.NoError(tb, err)
	_, err = client.PpsAPIClient.ActivateAuth(client.Ctx(), &pps.ActivateAuthRequest{})
	require.NoError(tb, err)
}
