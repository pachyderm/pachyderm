package s3

import (
	"fmt"
	stdlog "log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	enterpriseclient "github.com/pachyderm/pachyderm/src/client/enterprise"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
	"github.com/pachyderm/s2"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

const (
	multipartRepo        = "_s3gateway_multipart_"
	maxAllowedParts      = 10000
	maxRequestBodyLength = 128 * 1024 * 1024 //128mb
	requestTimeout       = 10 * time.Second
	readBodyTimeout      = 5 * time.Second
)

var enterpriseTimeout = 24 * time.Hour

// Server runs an HTTP server with an S3-like API for PFS. This allows you to
// use s3 clients to acccess PFS contents.
//
// This returns an `http.Server` instance. It is the responsibility of the
// caller to start the returned server. It's possible for the caller to
// gracefully shutdown the server if desired; see the `http` package for details.
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
func Server(port, pachdPort uint16) (*http.Server, error) {
	logger := logrus.WithFields(logrus.Fields{
		"source": "s3gateway",
	})

	var lastEnterpriseCheck time.Time
	isEnterprise := false

	c := &controller{
		pachdPort:       pachdPort,
		logger:          logger,
		repo:            multipartRepo,
		maxAllowedParts: maxAllowedParts,
	}

	s3Server := s2.NewS2(logger, maxRequestBodyLength, readBodyTimeout)
	s3Server.Auth = c
	s3Server.Service = c
	s3Server.Bucket = c
	s3Server.Object = c
	s3Server.Multipart = c
	router := s3Server.Router()

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		ReadTimeout:  requestTimeout,
		WriteTimeout: requestTimeout,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set a request ID, if it hasn't been set by the client already.
			// This can be used for tracing, and is included in error
			// responses.
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = uuid.NewWithoutDashes()
				r.Header.Set("X-Request-ID", requestID)
			}
			w.Header().Set("x-amz-request-id", requestID)

			// Log that a request was made
			logger.Infof("http request: %s %s", r.Method, r.RequestURI)

			// Ensure enterprise is enabled
			now := time.Now()
			if !isEnterprise || now.Sub(lastEnterpriseCheck) > enterpriseTimeout {
				vars := mux.Vars(r)
				pc, err := c.pachClient(vars["authAccessKey"])
				if err != nil {
					err = fmt.Errorf("gRPC client client: %v", err)
					s2.WriteError(logger, w, r, s2.InternalError(r, err))
					return
				}

				defer func() {
					if err := pc.Close(); err != nil {
						logger.Errorf("could not close enterprise check pach client: %v", err.Error())
					}
				}()

				resp, err := pc.Enterprise.GetState(context.Background(), &enterpriseclient.GetStateRequest{})
				if err != nil {
					err = fmt.Errorf("enterprise status check: %v", grpcutil.ScrubGRPC(err))
					s2.WriteError(logger, w, r, s2.InternalError(r, err))
					return
				}

				isEnterprise = resp.State == enterpriseclient.State_ACTIVE
			}
			if !isEnterprise {
				s2.WriteError(logger, w, r, enterpriseDisabledError(r))
				return
			}

			router.ServeHTTP(w, r)
		}),
		// NOTE: this is not closed. If the standard logger gets customized, this will need to be fixed
		ErrorLog: stdlog.New(logger.Writer(), "", 0),
	}

	return server, nil
}
