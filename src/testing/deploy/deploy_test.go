package main

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/minikubetestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

func TestDeployEnterprise(t *testing.T) {
	k := testutil.GetKubeClient(t)
	c := minikubetestenv.InstallRelease(t,
		context.Background(),
		"default",
		k,
		&minikubetestenv.DeployOpts{
			AuthUser:     auth.RootUser,
			Enterprise:   true,
			CleanupAfter: true,
		})
	whoami, err := c.AuthAPIClient.WhoAmI(c.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, auth.RootUser, whoami.Username)
	c.SetAuthToken("")
	mockIDPLogin(t, c)
}

func mockIDPLogin(t testing.TB, c *client.APIClient) {
	// login using mock IDP admin
	hc := &http.Client{}
	c.SetAuthToken("")
	loginInfo, err := c.GetOIDCLogin(c.Ctx(), &auth.GetOIDCLoginRequest{})
	require.NoError(t, err)
	state := loginInfo.State

	// Get the initial URL from the grpc, which should point to the dex login page
	resp, err := hc.Get(loginInfo.LoginURL)
	require.NoError(t, err)

	vals := make(url.Values)
	vals.Add("login", "admin")
	vals.Add("password", "password")

	_, err = hc.PostForm(resp.Request.URL.String(), vals)
	require.NoError(t, err)

	authResp, err := c.AuthAPIClient.Authenticate(c.Ctx(), &auth.AuthenticateRequest{OIDCState: state})
	require.NoError(t, err)
	c.SetAuthToken(authResp.PachToken)
	whoami, err := c.AuthAPIClient.WhoAmI(c.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, "user:"+testutil.DexMockConnectorEmail, whoami.Username)
}
