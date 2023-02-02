//go:build k8s

package main

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/identity"
	"github.com/pachyderm/pachyderm/v2/src/internal/minikubetestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

var valueOverrides map[string]string

func TestInstallAndUpgradeEnterpriseWithEnv(t *testing.T) {
	t.Parallel()
	ns, portOffset := minikubetestenv.ClaimCluster(t)
	k := testutil.GetKubeClient(t)
	opts := &minikubetestenv.DeployOpts{
		AuthUser:   auth.RootUser,
		Enterprise: true,
		PortOffset: portOffset,
	}
	opts.ValueOverrides = valueOverrides
	// Test Install
	minikubetestenv.PutNamespace(t, ns)
	c := minikubetestenv.InstallRelease(t, context.Background(), ns, k, opts)
	whoami, err := c.AuthAPIClient.WhoAmI(c.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, auth.RootUser, whoami.Username)
	c.SetAuthToken("")
	mockIDPLogin(t, c)
	// Test Upgrade
	opts.CleanupAfter = true
	// set new root token via env
	opts.AuthUser = ""
	token := "new-root-token"
	opts.ValueOverrides = map[string]string{"pachd.rootToken": token}
	// add config file with trusted peers & new clients
	opts.ValuesFiles = []string{createAdditionalClientsFile(t), createTrustedPeersFile(t)}
	// apply upgrade
	c = minikubetestenv.UpgradeRelease(t, context.Background(), ns, k, opts)
	c.SetAuthToken(token)
	whoami, err = c.AuthAPIClient.WhoAmI(c.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, auth.RootUser, whoami.Username)
	// old token should no longer work
	c.SetAuthToken(testutil.RootToken)
	_, err = c.AuthAPIClient.WhoAmI(c.Ctx(), &auth.WhoAmIRequest{})
	require.YesError(t, err)
	c.SetAuthToken("")
	mockIDPLogin(t, c)
	// assert new trusted peer and client
	resp, err := c.IdentityAPIClient.GetOIDCClient(c.Ctx(), &identity.GetOIDCClientRequest{Id: "pachd"})
	require.NoError(t, err)
	require.EqualOneOf(t, resp.Client.TrustedPeers, "example-app")
	_, err = c.IdentityAPIClient.GetOIDCClient(c.Ctx(), &identity.GetOIDCClientRequest{Id: "example-app"})
	require.NoError(t, err)
}

func TestEnterpriseServerMember(t *testing.T) {
	t.Parallel()
	ns, portOffset := minikubetestenv.ClaimCluster(t)
	k := testutil.GetKubeClient(t)
	minikubetestenv.PutNamespace(t, "enterprise")
	ec := minikubetestenv.InstallRelease(t, context.Background(), "enterprise", k, &minikubetestenv.DeployOpts{
		AuthUser:         auth.RootUser,
		EnterpriseServer: true,
		CleanupAfter:     true,
		ValueOverrides:   valueOverrides,
	})
	whoami, err := ec.AuthAPIClient.WhoAmI(ec.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, auth.RootUser, whoami.Username)
	mockIDPLogin(t, ec)
	minikubetestenv.PutNamespace(t, ns)
	c := minikubetestenv.InstallRelease(t, context.Background(), ns, k, &minikubetestenv.DeployOpts{
		AuthUser:         auth.RootUser,
		EnterpriseMember: true,
		Enterprise:       true,
		PortOffset:       portOffset,
		CleanupAfter:     true,
		ValueOverrides:   valueOverrides,
	})
	whoami, err = c.AuthAPIClient.WhoAmI(c.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, auth.RootUser, whoami.Username)
	c.SetAuthToken("")
	loginInfo, err := c.GetOIDCLogin(c.Ctx(), &auth.GetOIDCLoginRequest{})
	require.NoError(t, err)
	require.True(t, strings.Contains(loginInfo.LoginURL, ":31658"))
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
	getResp, err := hc.Get(loginInfo.LoginURL)
	require.NoError(t, err)
	require.Equal(t, 200, getResp.StatusCode, "Mock login could not retrieve the login page.")

	vals := make(url.Values)
	vals.Add("login", "admin")
	vals.Add("password", "password")

	postResp, err := hc.PostForm(getResp.Request.URL.String(), vals)
	require.NoError(t, err)
	require.Equal(t, 200, postResp.StatusCode, "POST to perform mock login failed.")
	postBody, err := io.ReadAll(postResp.Body)
	require.NoError(t, err, "Could not read login form POST response.")
	// There is a login failure case in which the response returned is a redirect back to the login page that returns 200
	require.Matches(t, "^(You are now logged in).*$", string(postBody), "Recieved an unexpected response from mock IDP login form POST.")

	authResp, err := c.AuthAPIClient.Authenticate(c.Ctx(), &auth.AuthenticateRequest{OIDCState: state})
	require.NoError(t, err)
	c.SetAuthToken(authResp.PachToken)
	whoami, err := c.AuthAPIClient.WhoAmI(c.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, "user:"+testutil.DexMockConnectorEmail, whoami.Username)
}

func createTrustedPeersFile(t testing.TB) string {
	data := []byte(`pachd:
  additionalTrustedPeers:
    - example-app
`)
	tf, err := os.CreateTemp("", "pachyderm-trusted-peers-*.yaml")
	require.NoError(t, err)
	_, err = tf.Write(data)
	require.NoError(t, err)
	return tf.Name()
}

func createAdditionalClientsFile(t testing.TB) string {
	data := []byte(`oidc:
  additionalClients:
    - id: example-app
      secret: example-app-secret
      name: 'Example App'
      redirectURIs:
      - 'http://127.0.0.1:5555/callback'
`)
	tf, err := os.CreateTemp("", "pachyderm-additional-clients-*.yaml")
	require.NoError(t, err)
	_, err = tf.Write(data)
	require.NoError(t, err)
	return tf.Name()
}
