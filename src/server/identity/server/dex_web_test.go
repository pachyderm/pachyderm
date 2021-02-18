package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/identity"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	logrus "github.com/sirupsen/logrus"

	dex_storage "github.com/dexidp/dex/storage"
	dex_memory "github.com/dexidp/dex/storage/memory"
)

type InMemoryStorageProvider struct {
	provider dex_storage.Storage
	err      error
}

func (p *InMemoryStorageProvider) GetStorage(logger *logrus.Entry) (dex_storage.Storage, error) {
	return p.provider, p.err
}

// TestLazyStartWebServer tests that the web server returns a 500 when the database isn't available,
// redirects to a static page when no connectors are configured, and redirects to a real connector when
// one is configured
func TestLazyStartWebServer(t *testing.T) {
	webDir = "../../../../dex-assets"
	logger := logrus.NewEntry(logrus.New())
	sp := &InMemoryStorageProvider{err: errors.New("unable to connect to database")}

	// server is instantiated but hasn't started
	server := newDexWeb(sp, logger, dbutil.NewTestDB(t))
	defer server.stopWebServer()

	// request the well-known endpoint, this should return a 500 because the database is failing
	req := httptest.NewRequest("GET", "/.well-known/openid-configuration", nil)

	// attempt to start the server but the database is unavailable
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusInternalServerError, recorder.Result().StatusCode)

	// attempt to start the server again, this time the database is available but no connectors are available
	// so we should get a redirect to a static page
	sp.provider = dex_memory.New(logger)
	sp.err = nil

	require.NoError(t, sp.provider.CreateClient(dex_storage.Client{
		ID:           "test",
		RedirectURIs: []string{"http://example.com/callback"},
	}))

	req = httptest.NewRequest("GET", "/auth?client_id=test&nonce=abc&redirect_uri=http%3A%2F%2Fexample.com%2Fcallback&response_type=code&scope=openid+profile+email&state=abcd", nil)
	recorder = httptest.NewRecorder()
	server.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusFound, recorder.Result().StatusCode)
	require.Matches(t, "/placeholder", recorder.Result().Header.Get("Location"))

	// configure a connector so the server should redirect to github automatically - the placeholder
	// provider shouldn't be enabled anymore
	err := sp.provider.CreateConnector(dex_storage.Connector{ID: "conn", Type: "github"})
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
	sp := &InMemoryStorageProvider{provider: dex_memory.New(logger)}

	server := newDexWeb(sp, logger, dbutil.NewTestDB(t))
	defer server.stopWebServer()

	err := sp.provider.CreateConnector(dex_storage.Connector{ID: "conn", Type: "github"})
	require.NoError(t, err)

	// request the OIDC configuration endpoint - the issuer is an empty string
	req := httptest.NewRequest("GET", "/.well-known/openid-configuration", nil)
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusOK, recorder.Result().StatusCode)

	var oidcConfig map[string]interface{}
	require.NoError(t, json.NewDecoder(recorder.Result().Body).Decode(&oidcConfig))
	require.Equal(t, "", oidcConfig["issuer"].(string))

	// reconfigure the issuer, the server should reload and serve the new issuer value
	server.updateConfig(identity.IdentityServerConfig{Issuer: "http://example.com:1234"})

	recorder = httptest.NewRecorder()
	server.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusOK, recorder.Result().StatusCode)

	require.NoError(t, json.NewDecoder(recorder.Result().Body).Decode(&oidcConfig))
	require.Equal(t, "http://example.com:1234", oidcConfig["issuer"].(string))
}
