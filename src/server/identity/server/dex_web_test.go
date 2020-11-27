package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pachyderm/pachyderm/src/client/identity"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
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

// TestLazyStartWebServer tests that the web server returns a 500 and retries starting on each request
// until it is successful. This is necessary in case postgres hasn't started or no connectors are configured.
func TestLazyStartWebServer(t *testing.T) {
	webDir = "../../../../web"
	logger := logrus.NewEntry(logrus.New())
	sp := &InMemoryStorageProvider{err: errors.New("unable to connect to database")}

	// server is instantiated but hasn't started
	server := newDexWeb(sp, logger)
	defer server.stopWebServer()

	// request the OIDC configuration endpoint to check whether the server is up
	req := httptest.NewRequest("GET", "/.well-known/openid-configuration", nil)

	// attempt to start the server but the database is unavailable
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusInternalServerError, recorder.Result().StatusCode)

	// attempt to start the server again, this time the database is available but no connectors are available
	sp.provider = dex_memory.New(logger)
	sp.err = nil

	recorder = httptest.NewRecorder()
	server.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusInternalServerError, recorder.Result().StatusCode)

	// configure a connector so the server starts successfully
	err := sp.provider.CreateConnector(dex_storage.Connector{ID: "conn", Type: "github"})
	require.NoError(t, err)
	recorder = httptest.NewRecorder()
	server.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusOK, recorder.Result().StatusCode)
}

// TestConfigureIssuer tests that the web server is restarted when the issuer is changed.
func TestConfigureIssuer(t *testing.T) {
	webDir = "../../../../web"
	logger := logrus.NewEntry(logrus.New())
	sp := &InMemoryStorageProvider{provider: dex_memory.New(logger)}

	server := newDexWeb(sp, logger)
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
	server.updateConfig(identity.IdentityConfig{Issuer: "http://example.com:1234"})

	recorder = httptest.NewRecorder()
	server.ServeHTTP(recorder, req)
	require.Equal(t, http.StatusOK, recorder.Result().StatusCode)

	require.NoError(t, json.NewDecoder(recorder.Result().Body).Decode(&oidcConfig))
	require.Equal(t, "http://example.com:1234", oidcConfig["issuer"].(string))
}
