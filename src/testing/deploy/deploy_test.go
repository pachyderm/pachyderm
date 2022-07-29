//go:build k8s

package main

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/identity"
	"github.com/pachyderm/pachyderm/v2/src/internal/minikubetestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestInstallAndUpgradeEnterpriseWithEnv(t *testing.T) {
	k := testutil.GetKubeClient(t)
	opts := &minikubetestenv.DeployOpts{
		AuthUser:   auth.RootUser,
		Enterprise: true,
	}
	// Test Install
	c := minikubetestenv.InstallRelease(t, context.Background(), "default", k, opts)
	whoami, err := c.AuthAPIClient.WhoAmI(c.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, auth.RootUser, whoami.Username)
	c.SetAuthToken("")
	mockIDPLogin(t, c)
	// Test Upgrade
	// set new root token via env
	opts.AuthUser = ""
	token := "new-root-token"
	opts.ValueOverrides = map[string]string{"pachd.rootToken": token}
	// add config file with trusted peers & new clients
	opts.ValuesFiles = []string{createAdditionalClientsFile(t), createTrustedPeersFile(t)}
	// apply upgrade
	c = minikubetestenv.UpgradeRelease(t, context.Background(), "default", k, opts)
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
	k := testutil.GetKubeClient(t)
	_, err := k.CoreV1().Namespaces().Create(context.Background(),
		&v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "enterprise",
			},
		},
		metav1.CreateOptions{})
	require.True(t, err == nil || strings.Contains(err.Error(), "already exists"), "Error '%v' does not contain 'already exists'", err)
	ec := minikubetestenv.InstallRelease(t, context.Background(), "enterprise", k, &minikubetestenv.DeployOpts{
		AuthUser:         auth.RootUser,
		EnterpriseServer: true,
	})
	whoami, err := ec.AuthAPIClient.WhoAmI(ec.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, auth.RootUser, whoami.Username)
	mockIDPLogin(t, ec)
	c := minikubetestenv.InstallRelease(t, context.Background(), "default", k, &minikubetestenv.DeployOpts{
		AuthUser:         auth.RootUser,
		EnterpriseMember: true,
		Enterprise:       true,
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

func createTrustedPeersFile(t testing.TB) string {
	data := []byte(`pachd:
  additionalTrustedPeers:
    - example-app
`)
	tf, err := ioutil.TempFile("", "pachyderm-trusted-peers-*.yaml")
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
	tf, err := ioutil.TempFile("", "pachyderm-additional-clients-*.yaml")
	require.NoError(t, err)
	_, err = tf.Write(data)
	require.NoError(t, err)
	return tf.Name()
}
