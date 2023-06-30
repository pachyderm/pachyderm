package server

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/pachyderm/pachyderm/v2/src/identity"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"

	dex_server "github.com/dexidp/dex/server"
	dex_storage "github.com/dexidp/dex/storage"
)

var (
	dexRequestCountMetric = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "pachyderm",
		Subsystem: "auth_dex",
		Name:      "http_requests_total",
		Help:      "Count of http requests handled by Dex, by response status code and HTTP method",
	}, []string{"code", "method"})
	dexRequestsInFlightMetric = promauto.NewGauge(prometheus.GaugeOpts{
		Namespace: "pachyderm",
		Subsystem: "auth_dex",
		Name:      "http_requests_in_flight",
		Help:      "Number of requests currently being handled by Dex",
	})
	dexRequestsDurationMetric = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "pachyderm",
		Subsystem: "auth_dex",
		Name:      "http_requests_duration_seconds",
		Help:      "Histogram of time spent processing Dex requests, by response status code and HTTP method",
		Buckets:   []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1, 5, 10, 30, 60, 300, 600},
	}, []string{"code", "method"})
	dexApprovalErrorCountMetric = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "pachyderm",
		Subsystem: "auth_dex",
		Name:      "approval_errors_total",
		Help:      "Count of HTTP requests to /approval that ended in error",
	})
	dexStartupErrorCountMetric = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "pachyderm",
		Subsystem: "auth_dex",
		Name:      "startup_errors_total",
		Help:      "Count of HTTP requests that were rejected because the server can't start",
	})
)

// webDir is the path to find the static assets for the web server.
// This is always /dex-assets in the docker image, but it can be overridden for testing
var webDir = "/dex-assets"

// dexWeb wraps a Dex web server and hot reloads it when the
// issuer is reconfigured.
type dexWeb struct {
	sync.RWMutex

	env Env

	// Rather than restart the server on every request, we cache it
	// along with the config and set of connectors. If either of these
	// change we restart the server because Dex doesn't support
	// reconfiguring them on the fly.
	currentConfig     *identity.IdentityServerConfig
	currentConnectors *identity.ListIDPConnectorsResponse
	server            *dex_server.Server
	serverCancel      context.CancelFunc

	storageProvider dex_storage.Storage
	apiServer       identity.APIServer
	assetsDir       string
	addr            string
}

func newDexWeb(env Env, apiServer identity.APIServer, options ...IdentityServerOption) *dexWeb {
	web := &dexWeb{
		env:             env,
		storageProvider: env.DexStorage,
		apiServer:       apiServer,
		addr:            dexHTTPPort,
		assetsDir:       webDir,
	}
	for _, opt := range options {
		opt(web)
	}
	return web
}

// stopWebServer must be called while holding the write mutex
func (w *dexWeb) stopWebServer() {
	ctx := pctx.Child(w.env.BackgroundContext, "dex")
	log.Info(ctx, "stopping identity web server")
	// Stop the background jobs for the existing server
	if w.serverCancel != nil {
		w.serverCancel()
	}

	w.server = nil
}

// serverNeedsRestart returns true if the server hasn't started yet, or if the config or set of
// connectors has changed. Must be called while holding a read lock on `w`.
func (w *dexWeb) serverNeedsRestart(config *identity.IdentityServerConfig, connectors *identity.ListIDPConnectorsResponse) bool {
	return w.server == nil ||
		w.currentConfig == nil ||
		w.currentConnectors == nil ||
		!proto.Equal(config, w.currentConfig) ||
		!proto.Equal(connectors, w.currentConnectors)
}

// startWebServer starts a new web server with the appropriate configuration and connectors.
func (w *dexWeb) startWebServer(config *identity.IdentityServerConfig, connectors *identity.ListIDPConnectorsResponse) (*dex_server.Server, error) {
	ctx := pctx.Child(w.env.BackgroundContext, "dex")
	w.Lock()
	defer w.Unlock()

	// If the config and connectors have already been updated while we were blocked,
	// don't restart the server again.
	if !w.serverNeedsRestart(config, connectors) {
		return w.server, nil
	}

	w.stopWebServer()
	log.Info(ctx, "starting identity web server")

	storage := w.storageProvider

	// If no connectors are configured, add a static placeholder which directs the user
	// to configure a connector
	if len(connectors.Connectors) == 0 {
		log.Info(ctx, "no idp connectors configured, using placeholder")
		dex_server.ConnectorsConfig["placeholder"] = func() dex_server.ConnectorConfig { return new(placeholderConfig) }
		storage = dex_storage.WithStaticConnectors(storage, []dex_storage.Connector{
			{
				ID:     "placeholder",
				Type:   "placeholder",
				Name:   "No IDPs Configured",
				Config: []byte(""),
			},
		})
	}

	var err error
	idTokenExpiry := 24 * time.Hour

	if config.IdTokenExpiry != "" {
		idTokenExpiry, err = time.ParseDuration(config.IdTokenExpiry)
		if err != nil {
			return nil, errors.EnsureStack(err)
		}
	}

	var refreshTokenPolicy *dex_server.RefreshTokenPolicy
	if config.RotationTokenExpiry != "" {
		refreshTokenPolicy, err = dex_server.NewRefreshTokenPolicy(log.NewLogrus(pctx.Child(ctx, "NewRefreshTokenPolicy")), false, "", config.RotationTokenExpiry, "")
		if err != nil {
			return nil, errors.EnsureStack(err)
		}
	}
	serverConfig := dex_server.Config{
		Storage:            storage,
		Issuer:             config.Issuer,
		IDTokensValidFor:   idTokenExpiry,
		SkipApprovalScreen: true,
		Web: dex_server.WebConfig{
			Issuer:  "Pachyderm",
			LogoURL: "/theme/logo.svg",
			Theme:   "pachyderm",
			Dir:     w.assetsDir,
		},
		RefreshTokenPolicy: refreshTokenPolicy,
		Logger:             log.NewLogrus(ctx),
	}

	ctx, w.serverCancel = pctx.WithCancel(ctx)
	w.server, err = dex_server.NewServer(ctx, serverConfig)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	w.currentConfig = config
	w.currentConnectors = connectors

	return w.server, nil
}

// getServer either returns a cached web server, or starts a new one
// if the config or set of connectors has changed
func (w *dexWeb) getServer(ctx context.Context) (*dex_server.Server, error) {
	var server *dex_server.Server
	config, err := w.apiServer.GetIdentityServerConfig(ctx, &identity.GetIdentityServerConfigRequest{})
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	connectors, err := w.apiServer.ListIDPConnectors(ctx, &identity.ListIDPConnectorsRequest{})
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	// Get a read lock to check if the server needs a restart
	w.RLock()
	if w.serverNeedsRestart(config.Config, connectors) {
		// If the server needs to restart, unlock and acquire the write lock
		w.RUnlock()
		return w.startWebServer(config.Config, connectors)
	}

	server = w.server
	w.RUnlock()
	return server, nil
}

// interceptApproval handles the `/approval` route which is called after a user has
// authenticated to the IDP but before they're redirected back to the OIDC server
func (w *dexWeb) interceptApproval(server *dex_server.Server) func(http.ResponseWriter, *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		authReq, err := w.storageProvider.GetAuthRequest(r.FormValue("req"))
		if err != nil {
			dexApprovalErrorCountMetric.Inc()
			log.Error(ctx, "failed to get auth request", zap.Error(err))
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		if !authReq.LoggedIn {
			dexApprovalErrorCountMetric.Inc()
			log.Error(ctx, "auth request does not have an identity for approval")
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		if err := dbutil.WithTx(r.Context(), w.env.DB, func(ctx context.Context, tx *pachsql.Tx) error {
			err := addUserInTx(r.Context(), tx, authReq.Claims.Email)
			return errors.Wrapf(err, "unable to record user identity for login")
		}); err != nil {
			dexApprovalErrorCountMetric.Inc()
			rw.WriteHeader(http.StatusInternalServerError)
			log.Error(ctx, "error while adding user in tx", zap.Error(err))
			return
		}
		server.ServeHTTP(rw, r)
	}
}

// ServeHTTP proxies requests to the Dex server, if it's configured.
func (w *dexWeb) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	server, err := w.getServer(r.Context())
	if server == nil {
		dexStartupErrorCountMetric.Inc()
		log.Error(r.Context(), "unable to start Dex server", zap.Error(err))
		http.Error(rw, "unable to start Dex server, check logs", http.StatusInternalServerError)
		return
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/approval", w.interceptApproval(server))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			fmt.Fprintf(w, "200 OK")
			return
		}
		server.ServeHTTP(w, r)
	})

	instrumented := promhttp.InstrumentHandlerInFlight(dexRequestsInFlightMetric,
		promhttp.InstrumentHandlerDuration(dexRequestsDurationMetric,
			promhttp.InstrumentHandlerCounter(dexRequestCountMetric, mux)))
	instrumented.ServeHTTP(rw, r)
}
