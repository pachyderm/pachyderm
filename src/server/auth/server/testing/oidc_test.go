package server

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
)

const (
	k8sHost = "192.168.64.35"
)

// TestOIDCAuthCodeFlow tests that we can configure an OIDC provider and do the
// auth code flow
func TestOIDCAuthCodeFlow(t *testing.T) {
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

	// Do a GET to Dex and get the redirect to the login page.
	// This gets us the `req` param, which ties together the session.
	loginUrl, err := url.Parse(loginInfo.LoginURL)
	require.NoError(t, err)
	loginUrl.Host = k8sHost + ":32000"

	resp, err := http.Get(loginUrl.String())
	require.NoError(t, err)

	// POST our hard-coded credentials to the login page with the
	// appropriate req params.
	c := &http.Client{}
	vals := make(url.Values)
	vals.Add("login", "admin@example.com")
	vals.Add("password", "password")

	redirectUrl, err := url.Parse(resp.Header.Get("Location"))
	require.NoError(t, err)
	redirectUrl.Scheme = "http"
	redirectUrl.Host = k8sHost + ":32000"

	resp, err = c.PostForm(redirectUrl.String(), vals)
	require.NoError(t, err)

	// Follow the resulting redirect back to pachd to complete the flow
	callbackUrl, err := url.Parse(resp.Header.Get("Location"))
	require.NoError(t, err)
	callbackUrl.
		callbackUrl.Host = k8sHost + ":30657"

	_, err = http.Get(callbackUrl.String())
	require.NoError(t, err)

	// Check that pachd recorded the response from the redirect
	adminClient.Authenticate(adminClient.Ctx(), &auth.AuthenticateRequest{OIDCState: loginInfo.State})
	require.NoError(t, err)

	deleteAll(t)
}
