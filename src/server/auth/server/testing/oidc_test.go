package server

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"golang.org/x/oauth2"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	tu "github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

// TestOIDCAuthCodeFlow tests that we can configure an OIDC provider and do the
// auth code flow
func TestOIDCAuthCodeFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.DeleteAll(t)
	tu.ConfigureOIDCProvider(t)
	defer tu.DeleteAll(t)

	testClient := tu.GetUnauthenticatedPachClient(t)
	loginInfo, err := testClient.GetOIDCLogin(testClient.Ctx(), &auth.GetOIDCLoginRequest{})
	require.NoError(t, err)

	tu.DoOAuthExchange(t, loginInfo.LoginURL)
	// Check that pachd recorded the response from the redirect
	authResp, err := testClient.Authenticate(testClient.Ctx(),
		&auth.AuthenticateRequest{OIDCState: loginInfo.State})
	require.NoError(t, err)
	testClient.SetAuthToken(authResp.PachToken)

	// Check that testClient authenticated as the right user
	whoAmIResp, err := testClient.WhoAmI(testClient.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, user(tu.DexMockConnectorEmail), whoAmIResp.Username)

	tu.DeleteAll(t)
}

// TestOIDCTrustedApp tests using an ID token issued to another OIDC app to authenticate.
func TestOIDCTrustedApp(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.DeleteAll(t)
	tu.ConfigureOIDCProvider(t)
	defer tu.DeleteAll(t)
	testClient := tu.GetUnauthenticatedPachClient(t)

	// Create an HTTP client that doesn't follow redirects.
	// We rewrite the host names for each redirect to avoid issues because
	// pachd is configured to reach dex with kube dns, but the tests might be
	// outside the cluster.
	c := &http.Client{}
	c.CheckRedirect = func(_ *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	oauthConfig := oauth2.Config{
		ClientID:     "testapp",
		ClientSecret: "test",
		RedirectURL:  "http://test.example.com:657/authorization-code/callback",
		Endpoint: oauth2.Endpoint{
			AuthURL:  tu.RewriteURL(t, "http://pachd:30658/auth", tu.DexHost(testClient)),
			TokenURL: tu.RewriteURL(t, "http://pachd:30658/token", tu.DexHost(testClient)),
		},
		Scopes: []string{
			"openid",
			"profile",
			"email",
			"audience:server:client_id:pachyderm",
		},
	}

	// Hit the dex login page for the test client with a fixed nonce
	resp, err := c.Get(oauthConfig.AuthCodeURL("state"))
	require.NoError(t, err)

	// Because we've only configured username/password login, there's a redirect
	// to the login page. The params have the session state. POST our hard-coded
	// credentials to the login page.
	vals := make(url.Values)
	vals.Add("login", "admin")
	vals.Add("password", "password")

	resp, err = c.PostForm(tu.RewriteRedirect(t, resp, tu.DexHost(testClient)), vals)
	require.NoError(t, err)

	// The username/password flow redirects back to the dex /approval endpoint
	resp, err = c.Get(tu.RewriteRedirect(t, resp, tu.DexHost(testClient)))
	require.NoError(t, err)

	codeURL, err := url.Parse(resp.Header.Get("Location"))
	require.NoError(t, err)

	token, err := oauthConfig.Exchange(context.Background(), codeURL.Query().Get("code"))
	require.NoError(t, err)

	// Use the id token from the previous OAuth flow with Pach
	authResp, err := testClient.Authenticate(testClient.Ctx(),
		&auth.AuthenticateRequest{IdToken: token.Extra("id_token").(string)})
	require.NoError(t, err)
	testClient.SetAuthToken(authResp.PachToken)

	// Check that testClient authenticated as the right user
	whoAmIResp, err := testClient.WhoAmI(testClient.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, user(tu.DexMockConnectorEmail), whoAmIResp.Username)

	tu.DeleteAll(t)
}
