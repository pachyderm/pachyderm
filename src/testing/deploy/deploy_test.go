package main

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/minikubetestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"golang.org/x/sync/errgroup"
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

func TestUpgradeEnterpriseWithEnv(t *testing.T) {
	k := testutil.GetKubeClient(t)
	opts := &minikubetestenv.DeployOpts{
		AuthUser:     auth.RootUser,
		Enterprise:   true,
		CleanupAfter: true,
	}
	c := minikubetestenv.InstallRelease(t, context.Background(), "default", k, opts)
	whoami, err := c.AuthAPIClient.WhoAmI(c.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, auth.RootUser, whoami.Username)
	// set new root token via env
	opts.AuthUser = ""
	token := "new-root-token"
	opts.ValueOverrides = map[string]string{"pachd.rootToken": token}
	c = minikubetestenv.UpgradeRelease(t, context.Background(), "default", k, opts)
	c.SetAuthToken(token)
	whoami, err = c.AuthAPIClient.WhoAmI(c.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, auth.RootUser, whoami.Username)
	// old token should no longer work
	c.SetAuthToken(testutil.RootToken)
	_, err = c.AuthAPIClient.WhoAmI(c.Ctx(), &auth.WhoAmIRequest{})
	require.YesError(t, err)
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

func TestParallelDeployments(t *testing.T) {
	eg, _ := errgroup.WithContext(context.Background())
	var c1 *client.APIClient
	var c2 *client.APIClient
	eg.Go(func() error {
		c1, _ = minikubetestenv.AcquireCluster(t)
		_, err := c1.PfsAPIClient.CreateRepo(context.Background(), &pfs.CreateRepoRequest{Repo: client.NewRepo("c1")})
		return errors.Wrap(err, "CreateRepo error")
	})
	eg.Go(func() error {
		c2, _ = minikubetestenv.AcquireCluster(t)
		_, err := c2.PfsAPIClient.CreateRepo(context.Background(), &pfs.CreateRepoRequest{Repo: client.NewRepo("c2")})
		return errors.Wrap(err, "CreateRepo error")
	})
	require.NoError(t, eg.Wait())

	c1List, err := c1.ListRepo()
	require.NoError(t, err)
	require.Equal(t, 1, len(c1List))
	require.Equal(t, c1List[0].Repo.Name, "c1")

	c2List, err := c2.ListRepo()
	require.NoError(t, err)
	require.Equal(t, 1, len(c2List))
	require.Equal(t, c2List[0].Repo.Name, "c2")
}
