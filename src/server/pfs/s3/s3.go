//nolint:wrapcheck
// TODO: the s2 library checks the type of the error to decide how to handle it,
// which doesn't work properly with wrapped errors
package s3

import (
	"context"
	"fmt"
	stdlog "log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/pachyderm/pachyderm/v2/src/client"

	"github.com/pachyderm/s2"
	"github.com/sirupsen/logrus"
)

// ClientFactory is a function called by s3g to create request-scoped
// pachyderm clients
type ClientFactory = func(ctx context.Context) *client.APIClient

const (
	multipartRepo        = "_s3gateway_multipart_"
	maxAllowedParts      = 10000
	maxRequestBodyLength = 128 * 1024 * 1024 //128mb
	requestTimeout       = 10 * time.Second
	readBodyTimeout      = 5 * time.Second

	// The S3 storage class that all PFS content will be reported to be stored in
	globalStorageClass = "STANDARD"

	// The S3 location served back
	globalLocation = "PACHYDERM"
)

// The S3 user associated with all PFS content
var defaultUser = s2.User{ID: "00000000000000000000000000000000", DisplayName: "pachyderm"}

type controller struct {
	logger *logrus.Entry

	// Name of the PFS repo holding multipart content
	repo string

	// the maximum number of allowed parts that can be associated with any
	// given file
	maxAllowedParts int

	driver Driver

	clientFactory ClientFactory
}

// requestPachClient uses the clientFactory to construct a request-scoped
// pachyderm client
func (c *controller) requestClient(r *http.Request) *client.APIClient {
	pc := c.clientFactory(r.Context())

	vars := mux.Vars(r)
	if vars["s3gAuth"] != "disabled" {
		accessKey := vars["authAccessKey"]
		if accessKey != "" {
			pc.SetAuthToken(accessKey)
		}
	}

	return pc
}

// Router creates an http server like object that serves an S3-like API for PFS. This allows you to
// use s3 clients to access PFS contents.

// `inputBuckets` specifies which buckets should be served, referencing
// specific commit IDs. If nil, all PFS branches will be served as separate
// buckets, of the form `<branch name>.<bucket name>`. Some s3 features are
// enabled when all PFS branches are served as well; e.g. we add support for
// some s3 versioning functionality.
//
// This returns an `mux.Router` instance. It is the responsibility of the
// caller to configure a server to use this Router.
//
// Note: server errors are redirected to logrus' standard log writer. The log
// writer is never closed. This should not be a problem with logrus' default
// configuration, which just writes to stdio. But if the standard logger is
// overwritten (e.g. to write to a socket), it's possible for this to cause
// problems.
//
// Note: In `s3cmd`, you must set the access key and secret key, even though
// this API will ignore them - otherwise, you'll get an opaque config error:
// https://github.com/s3tools/s3cmd/issues/845#issuecomment-464885959
func Router(driver Driver, clientFactory ClientFactory) *mux.Router {
	logger := logrus.WithFields(logrus.Fields{
		"source": "s3gateway",
	})

	c := &controller{
		logger:          logger,
		repo:            multipartRepo,
		maxAllowedParts: maxAllowedParts,
		driver:          driver,
		clientFactory:   clientFactory,
	}

	s3Server := s2.NewS2(logger, maxRequestBodyLength, readBodyTimeout)
	s3Server.Auth = c
	s3Server.Service = c
	s3Server.Bucket = c
	s3Server.Object = c
	s3Server.Multipart = c
	return s3Server.Router()
}

// S3Server wraps an HTTP server with an S3-like API for PFS. This allows you to
// use s3 clients to access PFS contents.

// In addition to providing the http server itself, S3Server exposes methods
// to configure handlers (mux.Routers) corresponding to different request URI hostnames.
// This way one http Server can respond differently and accordingly to each specific job.
type S3Server struct {
	*http.Server
	routerMap   map[string]*mux.Router
	routersLock sync.RWMutex
}

func (s *S3Server) ContainsRouter(k string) bool {
	s.routersLock.RLock()
	defer s.routersLock.RUnlock()
	_, ok := s.routerMap[k]
	return ok
}

func (s *S3Server) AddRouter(k string, r *mux.Router) {
	s.routersLock.Lock()
	defer s.routersLock.Unlock()
	s.routerMap[k] = r
}

func (s *S3Server) RemoveRouter(k string) {
	s.routersLock.Lock()
	defer s.routersLock.Unlock()
	delete(s.routerMap, k)
}

// Server runs an HTTP server with an S3-like API for PFS. This allows you to
// use s3 clients to access PFS contents.
func Server(port uint16, defaultRouter *mux.Router) *S3Server {
	logger := logrus.WithFields(logrus.Fields{
		"source": "s3gateway",
	})
	s3Server := S3Server{routerMap: make(map[string]*mux.Router)}
	s3Server.Server = &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		ReadTimeout:  requestTimeout,
		WriteTimeout: requestTimeout,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Log that a request was made
			logger.Infof("http request: %s %s", r.Method, r.RequestURI)
			if strings.HasPrefix(r.Host, "s3-") {
				s3Server.routersLock.RLock()
				defer s3Server.routersLock.RUnlock()
				host := r.Host
				if sep := strings.Index(r.Host, "."); sep != -1 {
					host = host[:sep]
				}
				router, ok := s3Server.routerMap[host]
				if ok {
					router.ServeHTTP(w, r)
				} else {
					w.WriteHeader(http.StatusInternalServerError)
				}
			} else if defaultRouter != nil {
				defaultRouter.ServeHTTP(w, r)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
		}),
		// NOTE: this is not closed. If the standard logger gets customized, this will need to be fixed
		ErrorLog: stdlog.New(logger.Writer(), "", 0),
	}
	return &s3Server
}
