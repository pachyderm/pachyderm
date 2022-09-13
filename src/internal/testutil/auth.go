package testutil

import (
	"context"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/config"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
)

const (
	// RootToken is the hard-coded admin token used on all activated test clusters
	RootToken = "iamroot"
)

func activateAuthHelper(tb testing.TB, client *client.APIClient) {
	client.SetAuthToken(RootToken)

	ActivateEnterprise(tb, client)

	_, err := client.Activate(client.Ctx(),
		&auth.ActivateRequest{RootToken: RootToken},
	)
	if err != nil && !strings.HasSuffix(err.Error(), "already activated") {
		tb.Fatalf("could not activate auth service: %v", err.Error())
	}
	require.NoError(tb, config.WritePachTokenToConfig(RootToken, false))

	// Activate auth for PPS
	client = client.WithCtx(context.Background())
	client.SetAuthToken(RootToken)
	_, err = client.PfsAPIClient.ActivateAuth(client.Ctx(), &pfs.ActivateAuthRequest{})
	require.NoError(tb, err)
	_, err = client.PpsAPIClient.ActivateAuth(client.Ctx(), &pps.ActivateAuthRequest{})
	require.NoError(tb, err)
}

// creates a new authenticated pach client, without re-activating
func AuthenticateClient(tb testing.TB, c *client.APIClient, subject string) *client.APIClient {
	tb.Helper()
	rootClient := UnauthenticatedPachClient(tb, c)
	rootClient.SetAuthToken(RootToken)
	if subject == auth.RootUser {
		return rootClient
	}
	token, err := rootClient.GetRobotToken(rootClient.Ctx(), &auth.GetRobotTokenRequest{Robot: subject})
	require.NoError(tb, err)
	client := UnauthenticatedPachClient(tb, c)
	client.SetAuthToken(token.Token)
	return client
}

// GetUnauthenticatedPachClient returns a copy of the testing pach client with no auth token
func UnauthenticatedPachClient(tb testing.TB, c *client.APIClient) *client.APIClient {
	tb.Helper()
	client := c.WithCtx(context.Background())
	client.SetAuthToken("")
	return client
}
