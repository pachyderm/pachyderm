package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
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
			Issuer:       "http://dex:32000",
			ClientID:     "pachyderm",
			ClientSecret: "notsecret",
			RedirectURI:  "http://pachd:657/authorization-code/callback",
		},
	}},
}

// TestOIDCAuthCodeFlow tests that we can configure an OIDC provider and do the
// auth code flow
func TestOIDCAuthCodeFlow(t *testing.T) {
	if os.Getenv("RUN_BAD_TESTS") == "" {
		t.Skip("Skipping because RUN_BAD_TESTS was empty")
	}
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)
	adminClient, testClient := getPachClient(t, admin), getPachClient(t, "")

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

	deleteAll(t)
}

// TestOIDCTrustedApp tests using an ID token issued to another OIDC app to authenticate.
func TestOIDCTrustedApp(t *testing.T) {
	if os.Getenv("RUN_BAD_TESTS") == "" {
		t.Skip("Skipping because RUN_BAD_TESTS was empty")
	}
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}
	deleteAll(t)
	adminClient, testClient := getPachClient(t, admin), getPachClient(t, "")

	_, err := adminClient.SetConfiguration(adminClient.Ctx(),
		&auth.SetConfigurationRequest{Configuration: OIDCAuthConfig})
	require.NoError(t, err)

	vals := url.Values{
		"client_id":     []string{"testapp"},
		"client_secret": []string{"test"},
		"grant_type":    []string{"password"},
		"scope":         []string{"openid email profile audience:server:client_id:pachyderm"},
		"username":      []string{"admin"},
		"password":      []string{"password"},
	}

	c := &http.Client{}
	resp, err := c.PostForm(fmt.Sprintf("http://%v/token", dexHost(testClient)), vals)
	require.NoError(t, err)

	tokenResp := make(map[string]interface{})
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&tokenResp))

	// Authenticate using an OIDC token with pachyderm in the audience claim
	authResp, err := testClient.Authenticate(testClient.Ctx(),
		&auth.AuthenticateRequest{IdToken: tokenResp["id_token"].(string)})
	require.NoError(t, err)
	testClient.SetAuthToken(authResp.PachToken)

	// Check that testClient authenticated as the right user
	whoAmIResp, err := testClient.WhoAmI(testClient.Ctx(), &auth.WhoAmIRequest{})
	require.NoError(t, err)
	require.Equal(t, "idp:"+dexMockConnectorEmail, whoAmIResp.Username)
	// idp:admin is an admin of the IDP but not Pachyderm
	require.False(t, whoAmIResp.IsAdmin)

	deleteAll(t)
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
	return parts[0] + ":32000"
}

func pachHost(c *client.APIClient) string {
	parts := strings.Split(c.GetAddress(), ":")
	if parts[1] == "650" {
		return parts[0] + ":657"
	}
	return parts[0] + ":30657"
}
