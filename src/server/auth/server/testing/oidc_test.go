package server

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
)

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
	adminClient := getPachClient(t, admin)

	conf := &auth.AuthConfig{
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
	_, err := adminClient.SetConfiguration(adminClient.Ctx(),
		&auth.SetConfigurationRequest{Configuration: conf})
	require.NoError(t, err)

	loginInfo, err := adminClient.GetOIDCLogin(adminClient.Ctx(), &auth.GetOIDCLoginRequest{})
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
	resp, err := c.Get(rewriteURL(t, loginInfo.LoginURL, dexHost(adminClient)))
	require.NoError(t, err)

	// Because we've only configured username/password login, there's a redirect
	// to the login page. The params have the session state. POST our hard-coded
	// credentials to the login page.
	vals := make(url.Values)
	vals.Add("login", "admin@example.com")
	vals.Add("password", "password")

	resp, err = c.PostForm(rewriteRedirect(t, resp, dexHost(adminClient)), vals)
	require.NoError(t, err)

	// The username/password flow redirects back to the dex /approval endpoint
	resp, err = c.Get(rewriteRedirect(t, resp, dexHost(adminClient)))
	require.NoError(t, err)

	// Follow the resulting redirect back to pachd to complete the flow
	_, err = c.Get(rewriteRedirect(t, resp, pachHost(adminClient)))
	require.NoError(t, err)

	// Check that pachd recorded the response from the redirect
	adminClient.Authenticate(adminClient.Ctx(), &auth.AuthenticateRequest{OIDCState: loginInfo.State})
	require.NoError(t, err)

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
