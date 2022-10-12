package testutil

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"testing"

	"golang.org/x/oauth2"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/identity"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

// DexMockConnectorEmail is the identity returned for all requests to the mock Dex connector
// (see https://github.com/dexidp/dex/blob/c113df2730052e20881dd68561289f8ae121300b/connector/mock/connectortest.go#L21)
// Kilgore Trout is a recurring character of Kurt Vonnegut's
const DexMockConnectorEmail = `kilgore@kilgore.trout`

// OIDCOIDCConfig is an auth config which can be used to connect to the identity service in tests
func OIDCOIDCConfig(host, issuerPort, redirectPort string, local bool) *auth.OIDCConfig {

	return &auth.OIDCConfig{
		Issuer:          "http://" + host + ":" + issuerPort + "/dex",
		ClientID:        "pachyderm",
		ClientSecret:    "notsecret",
		RedirectURI:     "http://" + host + ":" + redirectPort + "/authorization-code/callback",
		LocalhostIssuer: local,
		Scopes:          auth.DefaultOIDCScopes,
	}
}

// ConfigureOIDCProvider configures the identity service and the auth service to
// use a mock connector.
func ConfigureOIDCProvider(t *testing.T, c *client.APIClient, unitTest bool) error {
	adminClient := AuthenticateClient(t, c, auth.RootUser)

	_, err := adminClient.IdentityAPIClient.DeleteAll(adminClient.Ctx(), &identity.DeleteAllRequest{})
	require.NoError(t, err)

	// In the case of integration tests, the following defaults are used to due to routing constraints
	// imposed by minikube in integration tests. The address won't be reachable because the service external IP
	// is not reachable without a tunnel on minikube's IP.
	local := true
	issuerHost := "pachd"
	issuerPort := "1658"
	redirectPort := "1657"

	if unitTest {
		// In unit tests, all grpc servers listen on the peerPort which is randomly generated.
		local = false
		issuerHost = c.GetAddress().Host
		issuerPort = strconv.Itoa(int(c.GetAddress().Port + 8))
		redirectPort = strconv.Itoa(int(c.GetAddress().Port + 7))
	}

	_, err = adminClient.SetIdentityServerConfig(adminClient.Ctx(), &identity.SetIdentityServerConfigRequest{
		Config: &identity.IdentityServerConfig{
			Issuer: "http://" + issuerHost + ":" + issuerPort + "/dex",
		},
	})
	require.NoError(t, err)

	// Block until the web server has restarted with the right config
	require.NoError(t, backoff.Retry(func() error {
		resp, err := adminClient.GetIdentityServerConfig(adminClient.Ctx(), &identity.GetIdentityServerConfigRequest{})
		require.NoError(t, err)
		return require.EqualOrErr(
			"http://"+issuerHost+":"+issuerPort+"/dex", resp.Config.Issuer,
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
			Name:         "testapp",
			RedirectUris: []string{"http://test.example.com:" + redirectPort + "/authorization-code/callback"},
			Secret:       "test",
		},
	})
	require.NoError(t, err)

	_, err = adminClient.CreateOIDCClient(adminClient.Ctx(), &identity.CreateOIDCClientRequest{
		Client: &identity.OIDCClient{
			Id:           "pachyderm",
			Name:         "pachyderm",
			RedirectUris: []string{"http://" + issuerHost + ":" + redirectPort + "/authorization-code/callback"},
			Secret:       "notsecret",
			TrustedPeers: []string{"testapp"},
		},
	})
	require.NoError(t, err)

	_, err = adminClient.SetConfiguration(adminClient.Ctx(),
		&auth.SetConfigurationRequest{Configuration: OIDCOIDCConfig(issuerHost, issuerPort, redirectPort, local)})
	require.NoError(t, err)

	return nil
}

// DoOAuthExchange does the OAuth dance to log in to the mock provider, given a login URL
func DoOAuthExchange(t testing.TB, pachClient, enterpriseClient *client.APIClient, loginURL string) {
	// Create an HTTP client that doesn't follow redirects.
	// We rewrite the host names for each redirect to avoid issues because
	// pachd is configured to reach dex with kube dns, but the tests might be
	// outside the cluster.
	c := &http.Client{}
	c.CheckRedirect = func(_ *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	// Get the initial URL from the grpc, which should point to the dex login page
	resp, err := c.Get(RewriteURL(t, loginURL, DexHost(enterpriseClient)))
	require.NoError(t, err)

	// Dex login redirects to the provider page, which will generate it's own state
	resp, err = c.Get(RewriteRedirect(t, resp, DexHost(enterpriseClient)))
	require.NoError(t, err)

	// Because we've only configured username/password login, there's a redirect
	// to the login page. The params have the session state. POST our hard-coded
	// credentials to the login page.
	vals := make(url.Values)
	vals.Add("login", "admin")
	vals.Add("password", "password")

	resp, err = c.PostForm(RewriteRedirect(t, resp, DexHost(enterpriseClient)), vals)
	require.NoError(t, err)

	// The username/password flow redirects back to the dex /approval endpoint
	resp, err = c.Get(RewriteRedirect(t, resp, DexHost(enterpriseClient)))
	require.NoError(t, err)

	// Follow the resulting redirect back to pachd to complete the flow
	_, err = c.Get(RewriteRedirect(t, resp, pachHost(pachClient)))
	require.NoError(t, err)
}

func GetOIDCTokenForTrustedApp(t testing.TB, testClient *client.APIClient, unitTest bool) string {
	// Create an HTTP client that doesn't follow redirects.
	// We rewrite the host names for each redirect to avoid issues because
	// pachd is configured to reach dex with kube dns, but the tests might be
	// outside the cluster.
	c := &http.Client{}
	c.CheckRedirect = func(_ *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	redirectPort := "1657"
	if unitTest {
		redirectPort = strconv.Itoa(int(testClient.GetAddress().Port + 7))
	}

	oauthConfig := oauth2.Config{
		ClientID:     "testapp",
		ClientSecret: "test",
		RedirectURL:  "http://test.example.com:" + redirectPort + "/authorization-code/callback",
		Endpoint: oauth2.Endpoint{
			AuthURL:  RewriteURL(t, "http://pachd:30658/dex/auth", DexHost(testClient)),
			TokenURL: RewriteURL(t, "http://pachd:30658/dex/token", DexHost(testClient)),
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

	// Dex login redirects to the provider page, which will generate it's own state
	resp, err = c.Get(RewriteRedirect(t, resp, DexHost(testClient)))
	require.NoError(t, err)

	// Because we've only configured username/password login, there's a redirect
	// to the login page. The params have the session state. POST our hard-coded
	// credentials to the login page.
	vals := make(url.Values)
	vals.Add("login", "admin")
	vals.Add("password", "password")

	resp, err = c.PostForm(RewriteRedirect(t, resp, DexHost(testClient)), vals)
	require.NoError(t, err)

	// The username/password flow redirects back to the dex /approval endpoint
	resp, err = c.Get(RewriteRedirect(t, resp, DexHost(testClient)))
	require.NoError(t, err)

	codeURL, err := url.Parse(resp.Header.Get("Location"))
	require.NoError(t, err)

	token, err := oauthConfig.Exchange(context.Background(), codeURL.Query().Get("code"))
	require.NoError(t, err)

	return token.Extra("id_token").(string)
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
