package testutil

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

// OIDCOIDCConfig is an auth config which can be used to connect to the identity service in tests
func OIDCOIDCConfig() *auth.OIDCConfig {
	return &auth.OIDCConfig{
		Issuer:          "http://pachd:1658/dex",
		ClientID:        "pachyderm",
		ClientSecret:    "notsecret",
		RedirectURI:     "http://pachd:1657/authorization-code/callback",
		LocalhostIssuer: true,
		Scopes:          auth.DefaultOIDCScopes,
	}
}

// RewriteRedirect rewrites the Location header to point to the returned path at `host`
func RewriteRedirect(t testing.TB, resp *http.Response, host string) string {
	return RewriteURL(t, resp.Header.Get("Location"), host)
}

// RewriteURL rewrites the host and scheme in urlStr
func RewriteURL(t testing.TB, urlStr, host string) string {
	redirectURL, err := url.Parse(urlStr)
	require.NoError(t, err)
	redirectURL.Scheme = "http"
	redirectURL.Host = host
	return redirectURL.String()
}

// DexHost returns the address to access the identity server during tests
func DexHost(c *client.APIClient) string {
	if c.GetAddress().Port == 1650 {
		return c.GetAddress().Host + ":1658"
	}
	if c.GetAddress().Port == 31650 {
		return c.GetAddress().Host + ":31658"
	}
	// TODO(acohen4): revisit the way we are doing rewrites here
	// NOTE: the identity port is dynamically allocated in
	// src/internal/minikubetestenv/deploy.go
	return fmt.Sprintf("%v:%v", c.GetAddress().Host, c.GetAddress().Port+8)
}

func pachHost(c *client.APIClient) string {
	if c.GetAddress().Port == 1650 {
		return c.GetAddress().Host + ":1657"
	}
	if c.GetAddress().Port == 31650 {
		return c.GetAddress().Host + ":31657"
	}
	// NOTE: the identity port is dynamically allocated in
	// src/internal/minikubetestenv/deploy.go
	return fmt.Sprintf("%v:%v", c.GetAddress().Host, c.GetAddress().Port+7)
}
