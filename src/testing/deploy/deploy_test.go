package main

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/minikubetestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

func TestDeployEnterprise(t *testing.T) {
	k := testutil.GetKubeClient(t)
	minikubetestenv.DeleteRelease(t, context.Background(), k) // cleanup cluster
	c := minikubetestenv.InstallReleaseEnterprise(t, context.Background(), k, auth.RootUser)
	whoami, err := c.AuthAPIClient.WhoAmI(c.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, auth.RootUser, whoami.Username)
	mockIDPLogin(t)
}

func mockIDPLogin(t *testing.T) {
	// login using mock IDP admin
	hc := &http.Client{}
	var loginURL string
	hc.Get(loginURL)
	// Get the initial URL from the grpc, which should point to the dex login page
	resp, err := hc.Get(loginURL)
	require.NoError(t, err)

	vals := make(url.Values)
	vals.Add("login", "admin")
	vals.Add("password", "password")

	resp, err = hc.PostForm(resp.Header.Get("Location"), vals)
	require.NoError(t, err)

	// The username/password flow redirects back to the dex /approval endpoint
	resp, err = hc.Get(resp.Header.Get("Location"))
	require.NoError(t, err)

	// Follow the resulting redirect back to pachd to complete the flow
	_, err = hc.Get(resp.Header.Get("Location"))
	require.NoError(t, err)
}
