package server

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"golang.org/x/oauth2"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/identity"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"
)

// Dex's mock connector always returns the same identity for all authentication
// requests, and this is that identity's email (see
// https://github.com/dexidp/dex/blob/c113df2730052e20881dd68561289f8ae121300b/connector/mock/connectortest.go#L21)
// Kilgore Trout is a recurring character of Kurt Vonnegut's
const dexMockConnectorEmail = `kilgore@kilgore.trout`

var OIDCAuthConfig = &auth.AuthConfig{
	LiveConfigVersion: 0,
	IDProviders: []*auth.IDProvider{&auth.IDProvider{
		Name:        "idp",
		Description: "fake IdP for testing",
		OIDC: &auth.IDProvider_OIDCOptions{
			Issuer:         "http://localhost:30658/",
			IssuerOverride: "http://localhost:658",
			ClientID:       "pachyderm",
			ClientSecret:   "notsecret",
			RedirectURI:    "http://pachd:657/authorization-code/callback",
		},
	}},
}

func setupIdentityServer(t *testing.T, adminClient *client.APIClient) error {
	_, err := adminClient.IdentityAPIClient.DeleteAll(adminClient.Ctx(), &identity.DeleteAllRequest{})
	require.NoError(t, err)

	_, err = adminClient.SetIdentityConfig(adminClient.Ctx(), &identity.SetIdentityConfigRequest{
		Config: &identity.IdentityConfig{
			Issuer: "http://localhost:30658/",
		},
	})
	require.NoError(t, err)

	// Block until the web server has restarted with the right config
	require.NoError(t, backoff.Retry(func() error {
		resp, err := adminClient.GetIdentityConfig(adminClient.Ctx(), &identity.GetIdentityConfigRequest{})
		require.NoError(t, err)
		return require.EqualOrErr(
			"http://localhost:30658/", resp.Config.Issuer,
		)
	}, backoff.NewTestingBackOff()))

	_, err = adminClient.CreateIDPConnector(adminClient.Ctx(), &identity.CreateIDPConnectorRequest{
		Connector: &identity.IDPConnector{
			Name:       "test",
			Id:         "test",
			Type:       "mockPassword",
			JsonConfig: `{"username": "admin", "password": "password"}`,
		},
	})
	require.NoError(t, err)

	_, err = adminClient.CreateOIDCClient(adminClient.Ctx(), &identity.CreateOIDCClientRequest{
		Client: &identity.OIDCClient{
			Id:           "testapp",
			RedirectUris: []string{"http://test.example.com:657/authorization-code/callback"},
			Secret:       "test",
		},
	})
	require.NoError(t, err)

	_, err = adminClient.CreateOIDCClient(adminClient.Ctx(), &identity.CreateOIDCClientRequest{
		Client: &identity.OIDCClient{
			Id:           "pachyderm",
			RedirectUris: []string{"http://pachd:657/authorization-code/callback"},
			Secret:       "notsecret",
			TrustedPeers: []string{"testapp"},
		},
	})
	require.NoError(t, err)

	return nil
}

// TestOIDCAuthCodeFlow tests that we can configure an OIDC provider and do the
// auth code flow
func TestOIDCAuthCodeFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.DeleteAll(t)
	adminClient, testClient := tu.GetAuthenticatedPachClient(t, tu.AdminUser), tu.GetAuthenticatedPachClient(t, "")

	setupIdentityServer(t, adminClient)

	_, err := adminClient.SetConfiguration(adminClient.Ctx(),
		&auth.SetConfigurationRequest{Configuration: OIDCAuthConfig})
	require.NoError(t, err)

	loginInfo, err := testClient.GetOIDCLogin(testClient.Ctx(), &auth.GetOIDCLoginRequest{})
	require.NoError(t, err)

	// Create an HTTP client that doesn't follow redirects.
	// We rewrite the host names for each redirect to avoid issues because
	// pachd is configured to reach dex with kube dns, but the tests might be
	// outside the cluster.
	c := &http.Client{}
	c.CheckRedirect = func(_ *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	// Get the initial URL from the grpc, which should point to the dex login page
	resp, err := c.Get(rewriteURL(t, loginInfo.LoginURL, dexHost(testClient)))
	require.NoError(t, err)

	// Because we've only configured username/password login, there's a redirect
	// to the login page. The params have the session state. POST our hard-coded
	// credentials to the login page.
	vals := make(url.Values)
	vals.Add("login", "admin")
	vals.Add("password", "password")

	resp, err = c.PostForm(rewriteRedirect(t, resp, dexHost(testClient)), vals)
	require.NoError(t, err)

	// The username/password flow redirects back to the dex /approval endpoint
	resp, err = c.Get(rewriteRedirect(t, resp, dexHost(testClient)))
	require.NoError(t, err)

	// Follow the resulting redirect back to pachd to complete the flow
	_, err = c.Get(rewriteRedirect(t, resp, pachHost(testClient)))
	require.NoError(t, err)

	// Check that pachd recorded the response from the redirect
	authResp, err := testClient.Authenticate(testClient.Ctx(),
		&auth.AuthenticateRequest{OIDCState: loginInfo.State})
	require.NoError(t, err)
	testClient.SetAuthToken(authResp.PachToken)

	// Check that testClient authenticated as the right user
	whoAmIResp, err := testClient.WhoAmI(testClient.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, "idp:"+dexMockConnectorEmail, whoAmIResp.Username)
	require.False(t, whoAmIResp.IsAdmin)

	tu.DeleteAll(t)
}

// TestOIDCTrustedApp tests using an ID token issued to another OIDC app to authenticate.
func TestOIDCTrustedApp(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	tu.DeleteAll(t)
	adminClient, testClient := tu.GetAuthenticatedPachClient(t, tu.AdminUser), tu.GetAuthenticatedPachClient(t, "")

	setupIdentityServer(t, adminClient)

	_, err := adminClient.SetConfiguration(adminClient.Ctx(),
		&auth.SetConfigurationRequest{Configuration: OIDCAuthConfig})
	require.NoError(t, err)

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
			AuthURL:  rewriteURL(t, "http://pachd:30658/auth", dexHost(testClient)),
			TokenURL: rewriteURL(t, "http://pachd:30658/token", dexHost(testClient)),
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

	resp, err = c.PostForm(rewriteRedirect(t, resp, dexHost(testClient)), vals)
	require.NoError(t, err)

	// The username/password flow redirects back to the dex /approval endpoint
	resp, err = c.Get(rewriteRedirect(t, resp, dexHost(testClient)))
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
	require.Equal(t, "idp:"+dexMockConnectorEmail, whoAmIResp.Username)
	// idp:admin is an admin of the IDP but not Pachyderm
	require.False(t, whoAmIResp.IsAdmin)

	tu.DeleteAll(t)
}

// Rewrite the Location header to point to the returned path at `host`
func rewriteRedirect(t *testing.T, resp *http.Response, host string) string {
	return rewriteURL(t, resp.Header.Get("Location"), host)
}

func rewriteURL(t *testing.T, urlStr, host string) string {
	redirectURL, err := url.Parse(urlStr)
	require.NoError(t, err)
	redirectURL.Scheme = "http"
	redirectURL.Host = host
	return redirectURL.String()
}

func dexHost(c *client.APIClient) string {
	parts := strings.Split(c.GetAddress(), ":")
	if parts[1] == "650" {
		return parts[0] + ":658"
	}
	return parts[0] + ":30658"
}

func pachHost(c *client.APIClient) string {
	parts := strings.Split(c.GetAddress(), ":")
	if parts[1] == "650" {
		return parts[0] + ":657"
	}
	return parts[0] + ":30657"
}
