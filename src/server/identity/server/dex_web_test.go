package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/identity"
	"github.com/pachyderm/pachyderm/v2/src/internal/clusterstate"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	logrus "github.com/sirupsen/logrus"

	dex_storage "github.com/dexidp/dex/storage"
	dex_memory "github.com/dexidp/dex/storage/memory"
)

func getTestEnv(t *testing.T) serviceenv.ServiceEnv {
	env := &serviceenv.TestServiceEnv{DBClient: testutil.NewTestDB(t)}
	require.NoError(t, migrations.ApplyMigrations(context.Background(), env.GetDBClient(), migrations.Env{}, clusterstate.DesiredClusterState))
	require.NoError(t, migrations.BlockUntil(context.Background(), env.GetDBClient(), clusterstate.DesiredClusterState))
	return env
}

// TestLazyStartWebServer tests that the web server redirects to a static page when no connectors are configured,
// and redirects to a real connector when one is configured
func TestLazyStartWebServer(t *testing.T) {
	webDir = "../../../../dex-assets"
	logger := logrus.NewEntry(logrus.New())
	sp := dex_memory.New(logger)
	env := getTestEnv(t)
	api := NewIdentityServer(env, sp, false)

	// server is instantiated but hasn't started
	server := newDexWeb(env, sp, api)
	defer server.stopWebServer()

	// attempt to start the server, no connectors are available so we should get a redirect to a static page
	require.NoError(t, sp.CreateClient(dex_storage.Client{
		ID:           "test",
		RedirectURIs: []string{"http://example.com/callback"},
	}))

	req := httptest.NewRequest("GET", "/auth?client_id=test&nonce=abc&redirect_uri=http%3A%2F%2Fexample.com%2Fcallback&response_type=code&scope=openid+profile+email&state=abcd", nil)

	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusFound, recorder.Result().StatusCode)
	require.Matches(t, "/placeholder", recorder.Result().Header.Get("Location"))

	// configure a connector so the server should redirect to github automatically - the placeholder provider shouldn't be enabled anymore
	err := sp.CreateConnector(dex_storage.Connector{ID: "conn", Type: "github"})
	require.NoError(t, err)
	recorder = httptest.NewRecorder()
	server.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusFound, recorder.Result().StatusCode)
	require.Matches(t, "/auth/conn", recorder.Result().Header.Get("Location"))

	// make a second request to the running web server
	recorder = httptest.NewRecorder()
	server.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusFound, recorder.Result().StatusCode)
	require.Matches(t, "/auth/conn", recorder.Result().Header.Get("Location"))
}

// TestConfigureIssuer tests that the web server is restarted when the issuer is changed.
func TestConfigureIssuer(t *testing.T) {
	webDir = "../../../../dex-assets"
	logger := logrus.NewEntry(logrus.New())
	sp := dex_memory.New(logger)
	env := getTestEnv(t)
	api := NewIdentityServer(env, sp, false)

	server := newDexWeb(env, sp, api)
	defer server.stopWebServer()

	err := sp.CreateConnector(dex_storage.Connector{ID: "conn", Type: "github"})
	require.NoError(t, err)

	// request the OIDC configuration endpoint - the issuer is an empty string
	req := httptest.NewRequest("GET", "/.well-known/openid-configuration", nil)
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusOK, recorder.Result().StatusCode)

	var oidcConfig map[string]interface{}
	require.NoError(t, json.NewDecoder(recorder.Result().Body).Decode(&oidcConfig))
	require.Equal(t, "", oidcConfig["issuer"].(string))

	//reconfigure the issuer, the server should reload and serve the new issuer value
	_, err = api.SetIdentityServerConfig(context.Background(), &identity.SetIdentityServerConfigRequest{
		Config: &identity.IdentityServerConfig{Issuer: "http://example.com:1234"},
	})
	require.NoError(t, err)

	recorder = httptest.NewRecorder()
	server.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusOK, recorder.Result().StatusCode)

	require.NoError(t, json.NewDecoder(recorder.Result().Body).Decode(&oidcConfig))
	require.Equal(t, "http://example.com:1234", oidcConfig["issuer"].(string))
}

// TestUpdateIDP tests that the web server is restarted when an IDP is reconfigured
func TestUpdateIDP(t *testing.T) {
	webDir = "../../../../dex-assets"
	logger := logrus.NewEntry(logrus.New())
	sp := dex_memory.New(logger)
	env := getTestEnv(t)
	api := NewIdentityServer(env, sp, false)

	server := newDexWeb(env, sp, api)
	defer server.stopWebServer()

	// Configure a connector with a given ID
	err := sp.CreateConnector(dex_storage.Connector{ID: "conn", Type: "github", Config: []byte(`{"clientID": "test1", "redirectURI": "/callback"}`)})
	require.NoError(t, err)

	// Create an auth request that hasn't done the flow
	require.NoError(t, sp.CreateAuthRequest(dex_storage.AuthRequest{
		ID:       "testreq",
		ClientID: "testclient",
	}))

	// make a request to authenticate
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/auth/conn?req=testreq", nil)
	server.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusFound, recorder.Result().StatusCode)
	require.Equal(t, "https://github.com/login/oauth/authorize?client_id=test1&redirect_uri=%2Fcallback&response_type=code&scope=user%3Aemail&state=testreq", recorder.Result().Header.Get("Location"))

	// update the connector config
	_, err = api.UpdateIDPConnector(context.Background(), &identity.UpdateIDPConnectorRequest{
		Connector: &identity.IDPConnector{Id: "conn", Type: "github", JsonConfig: `{"clientID": "test2", "redirectURI": "/callback"}`, ConfigVersion: 1},
	})
	require.NoError(t, err)

	// make a second request with the updated client id
	recorder = httptest.NewRecorder()
	server.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusFound, recorder.Result().StatusCode)
	require.Equal(t, "https://github.com/login/oauth/authorize?client_id=test2&redirect_uri=%2Fcallback&response_type=code&scope=user%3Aemail&state=testreq", recorder.Result().Header.Get("Location"))
}

// TestLogApprovedUsers tests that the web server intercepts approvals and records the email
// in the database.
func TestLogApprovedUsers(t *testing.T) {
	webDir = "../../../../dex-assets"
	logger := logrus.NewEntry(logrus.New())
	sp := dex_memory.New(logger)
	env := getTestEnv(t)
	api := NewIdentityServer(env, sp, false)

	server := newDexWeb(env, sp, api)
	defer server.stopWebServer()

	err := sp.CreateConnector(dex_storage.Connector{ID: "conn", Type: "github"})
	require.NoError(t, err)

	// Create an auth request that has already done the Github flow
	require.NoError(t, sp.CreateAuthRequest(dex_storage.AuthRequest{
		ID:       "testreq",
		ClientID: "testclient",
		Expiry:   time.Now().Add(time.Hour),
		LoggedIn: true,
		Claims:   dex_storage.Claims{Email: "test@example.com"},
	}))

	// Hit the approval endpoint to be redirected
	req := httptest.NewRequest("GET", "/approval?req=testreq", nil)
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusSeeOther, recorder.Result().StatusCode)

	users, err := listUsers(context.Background(), env.GetDBClient())
	require.NoError(t, err)
	require.Equal(t, 1, len(users))
	require.Equal(t, "test@example.com", users[0].Email)

	// Create a second request and confirm the last-authenticated date is updated
	require.NoError(t, sp.CreateAuthRequest(dex_storage.AuthRequest{
		ID:       "testreq2",
		ClientID: "testclient",
		Expiry:   time.Now().Add(time.Hour),
		LoggedIn: true,
		Claims:   dex_storage.Claims{Email: "test@example.com"},
	}))

	req = httptest.NewRequest("GET", "/approval?req=testreq2", nil)
	recorder = httptest.NewRecorder()
	server.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusSeeOther, recorder.Result().StatusCode)

	newUsers, err := listUsers(context.Background(), env.GetDBClient())
	require.NoError(t, err)
	require.Equal(t, 1, len(newUsers))
	require.Equal(t, "test@example.com", newUsers[0].Email)
	require.NotEqual(t, users[0].LastAuthenticated, newUsers[0].LastAuthenticated)
}
