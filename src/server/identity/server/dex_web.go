package server

import (
	"context"
	"net/http"
	"sync"

	"github.com/pachyderm/pachyderm/src/client/identity"

	dex_server "github.com/dexidp/dex/server"
	logrus "github.com/sirupsen/logrus"
)

// webDir is the path to find the static assets for the web server.
// This is always /web in the docker image, but it can be overriden for testing
var webDir = "/web"

// dexWeb wraps a Dex web server and hot reloads it when the
// issuer is reconfigured.
type dexWeb struct {
	sync.RWMutex

	issuer       string
	server       *dex_server.Server
	serverCancel context.CancelFunc

	logger          *logrus.Entry
	storageProvider StorageProvider
}

func newDexWeb(sp StorageProvider, logger *logrus.Entry) *dexWeb {
	return &dexWeb{
		logger:          logger,
		storageProvider: sp,
	}
}

func (w *dexWeb) updateConfig(conf identity.IdentityConfig) {
	w.Lock()
	defer w.Unlock()

	w.issuer = conf.Issuer
	w.stopWebServer()
	w.startWebServer()
}

// stopWebServer must be called while holding the write mutex
func (w *dexWeb) stopWebServer() {
	// Stop the background jobs for the existing server
	if w.serverCancel != nil {
		w.serverCancel()
	}

	w.server = nil
}

// startWebServer must called while holding the write mutex
func (w *dexWeb) startWebServer() {
	storage, err := w.storageProvider.GetStorage(w.logger)
	if err != nil {
		return
	}

	serverConfig := dex_server.Config{
		Storage:            storage,
		Issuer:             w.issuer,
		SkipApprovalScreen: true,
		Web: dex_server.WebConfig{
			Dir: webDir,
		},
		Logger: w.logger,
	}

	var ctx context.Context
	ctx, w.serverCancel = context.WithCancel(context.Background())
	dexServer, err := dex_server.NewServer(ctx, serverConfig)
	if err != nil {
		w.logger.WithError(err).Error("dex web server failed to start")
		return
	}

	w.server = dexServer
}

func (w *dexWeb) getServer() *dex_server.Server {
	var server *dex_server.Server
	// Get a read lock to check if the server is nil (this is rare)
	w.RLock()
	if w.server == nil {
		// If the server is nil, unlock and acquire the write lock
		w.RUnlock()
		w.Lock()

		// Once we have the write lock, check that the server hasn't already been started
		if w.server == nil {
			w.startWebServer()
		}
		server = w.server
		w.Unlock()
		return server
	}

	server = w.server
	w.RUnlock()
	return server
}

// ServeHTTP proxies requests to the Dex server, if it's configured.
//
func (w *dexWeb) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	server := w.getServer()
	if server == nil {
		http.Error(rw, "unable to start Dex server, check logs", http.StatusInternalServerError)
		return
	}

	server.ServeHTTP(rw, r)
}
