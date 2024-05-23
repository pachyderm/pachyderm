package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/identity"
	"github.com/pachyderm/pachyderm/v2/src/internal/clusterstate"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/testetcd"

	dex_storage "github.com/dexidp/dex/storage"
	dex_memory "github.com/dexidp/dex/storage/memory"
)

func getTestEnv(t *testing.T) serviceenv.ServiceEnv {
	ctx := pctx.TestContext(t)
	env := &serviceenv.TestServiceEnv{
		DBClient:      dockertestenv.NewTestDB(t),
		DexDB:         dex_memory.New(log.NewLogrus(ctx)),
		EtcdClient:    testetcd.NewEnv(ctx, t).EtcdClient,
		Configuration: pachconfig.NewConfiguration(&pachconfig.PachdFullConfiguration{}),
		Ctx:           ctx,
	}
	require.NoError(t, migrations.ApplyMigrations(ctx, env.GetDBClient(), migrations.MakeEnv(nil, env.GetEtcdClient()), clusterstate.DesiredClusterState))
	require.NoError(t, migrations.BlockUntil(ctx, env.GetDBClient(), clusterstate.DesiredClusterState))
	return env
}

// TestLazyStartWebServer tests that the web server redirects to a static page when no connectors are configured,
// and redirects to a real connector when one is configured
func TestLazyStartWebServer(t *testing.T) {
	webDir = "../../../../dex-assets"
	env := getTestEnv(t)
	api := NewIdentityServer(EnvFromServiceEnv(env), false)

	// server is instantiated but hasn't started
	server := newDexWeb(EnvFromServiceEnv(env), api)
	defer server.stopWebServer()

	// attempt to start the server, no connectors are available so we should get a redirect to a static page
	require.NoError(t, env.GetDexDB().CreateClient(dex_storage.Client{
		ID:           "test",
		RedirectURIs: []string{"http://example.com/callback"},
	}))

	// check if the dex server returns 200 on '/'
	req := httptest.NewRequest("GET", "/", nil)
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusOK, recorder.Result().StatusCode)

	req = httptest.NewRequest("GET", "/auth?client_id=test&nonce=abc&redirect_uri=http%3A%2F%2Fexample.com%2Fcallback&response_type=code&scope=openid+profile+email&state=abcd", nil)

	recorder = httptest.NewRecorder()
	server.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusFound, recorder.Result().StatusCode)
	require.Matches(t, "/placeholder", recorder.Result().Header.Get("Location"))

	// configure a connector so the server should redirect to github automatically - the placeholder provider shouldn't be enabled anymore
	err := env.GetDexDB().CreateConnector(dex_storage.Connector{ID: "conn", Type: "github"})
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
	env := getTestEnv(t)
	api := NewIdentityServer(EnvFromServiceEnv(env), false)

	server := newDexWeb(EnvFromServiceEnv(env), api)
	defer server.stopWebServer()

	err := env.GetDexDB().CreateConnector(dex_storage.Connector{ID: "conn", Type: "github"})
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
	env := getTestEnv(t)
	api := NewIdentityServer(EnvFromServiceEnv(env), false)

	server := newDexWeb(EnvFromServiceEnv(env), api)
	defer server.stopWebServer()

	// Configure a connector with a given ID
	err := env.GetDexDB().CreateConnector(dex_storage.Connector{ID: "conn", Type: "github", Config: []byte(`{"clientID": "test1", "redirectURI": "/callback"}`)})
	require.NoError(t, err)

	require.NoError(t, env.GetDexDB().CreateClient(dex_storage.Client{
		ID:           "testclient",
		RedirectURIs: []string{"http://example.com/callback"},
	}))

	// make a request to authenticate
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/auth/conn?req=testreq&client_id=testclient&redirect_uri=http%3A%2F%2Fexample.com%2Fcallback&scope=openid&response_type=code", nil)
	server.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusFound, recorder.Result().StatusCode)
	redirect, err := url.Parse(recorder.Result().Header.Get("Location"))
	require.NoError(t, err)
	require.Equal(t, "test1", redirect.Query()["client_id"][0])

	// update the connector config
	_, err = api.UpdateIDPConnector(context.Background(), &identity.UpdateIDPConnectorRequest{
		Connector: &identity.IDPConnector{
			Id:   "conn",
			Type: "github",
			Config: mustStruct(
				map[string]any{
					"clientID":    "test2",
					"redirectURI": "/callback",
				},
			),
			ConfigVersion: 1,
		},
	})
	require.NoError(t, err)

	// make a second request with the updated client id
	recorder = httptest.NewRecorder()
	server.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusFound, recorder.Result().StatusCode)
	redirect, err = url.Parse(recorder.Result().Header.Get("Location"))
	require.NoError(t, err)
	require.Equal(t, "test2", redirect.Query()["client_id"][0])
}

// TestLogApprovedUsers tests that the web server intercepts approvals and records the email
// in the database.
func TestLogApprovedUsers(t *testing.T) {
	webDir = "../../../../dex-assets"
	env := getTestEnv(t)
	api := NewIdentityServer(EnvFromServiceEnv(env), false)

	server := newDexWeb(EnvFromServiceEnv(env), api)
	defer server.stopWebServer()

	err := env.GetDexDB().CreateConnector(dex_storage.Connector{ID: "conn", Type: "github"})
	require.NoError(t, err)

	// Create an auth request that has already done the Github flow
	require.NoError(t, env.GetDexDB().CreateAuthRequest(dex_storage.AuthRequest{
		ID:       "testreq",
		HMACKey:  []byte{1},
		ClientID: "testclient",
		Expiry:   time.Now().Add(time.Hour),
		LoggedIn: true,
		Claims:   dex_storage.Claims{Email: "test@example.com"},
	}))

	// Hit the approval endpoint to be redirected
	req := httptest.NewRequest("GET", "/approval?req=testreq&hmac=hEdjDhdQtKNWQg5sc8Kk1oFhT93hhs87STU-ysYvToI", nil)
	req = req.WithContext(env.Context())
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusSeeOther, recorder.Result().StatusCode)

	users, err := listUsers(context.Background(), env.GetDBClient())
	require.NoError(t, err)
	require.Equal(t, 1, len(users))
	require.Equal(t, "test@example.com", users[0].Email)

	// Create a second request and confirm the last-authenticated date is updated
	require.NoError(t, env.GetDexDB().CreateAuthRequest(dex_storage.AuthRequest{
		ID:       "testreq2",
		HMACKey:  []byte{1},
		ClientID: "testclient",
		Expiry:   time.Now().Add(time.Hour),
		LoggedIn: true,
		Claims:   dex_storage.Claims{Email: "test@example.com"},
	}))

	req = httptest.NewRequest("GET", "/approval?req=testreq2&hmac=JdYXChZhwtxW_pkp3pIbNmY0Kj9HNJbh58bwpXwsoww", nil)
	req = req.WithContext(env.Context())
	recorder = httptest.NewRecorder()
	server.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusSeeOther, recorder.Result().StatusCode)

	newUsers, err := listUsers(context.Background(), env.GetDBClient())
	require.NoError(t, err)
	require.Equal(t, 1, len(newUsers))
	require.Equal(t, "test@example.com", newUsers[0].Email)
	require.NotEqual(t, users[0].LastAuthenticated, newUsers[0].LastAuthenticated)
}
