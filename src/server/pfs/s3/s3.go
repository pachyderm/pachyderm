package s3

import (
	"fmt"
	stdlog "log"
	"net/http"
	"regexp"
	"time"

	"golang.org/x/net/context"

	"github.com/sirupsen/logrus"

	"github.com/pachyderm/pachyderm/src/client"
	enterpriseclient "github.com/pachyderm/pachyderm/src/client/enterprise"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
	"github.com/pachyderm/s3server"
)

var enterpriseTimeout = 24 * time.Hour
var bucketNameValidator = regexp.MustCompile(`^/[a-zA-Z0-9\-_]{1,255}\.[a-zA-Z0-9\-_]{1,255}/`)

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
func Server(pc *client.APIClient, port uint16) *http.Server {
	logger := logrus.WithFields(logrus.Fields{
		"source": "s3gateway",
	})

	var lastEnterpriseCheck time.Time
	isEnterprise := false

	s3 := s3server.NewS3()
	s3.Root = rootController{
		pc:     pc,
		logger: logger,
	}
	s3.Bucket = bucketController{
		pc:     pc,
		logger: logger,
	}
	s3.Object = objectController{
		pc:     pc,
		logger: logger,
	}
	router := s3.Router(logger)

	return &http.Server{
		Addr: fmt.Sprintf(":%d", port),
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
				resp, err := pc.Enterprise.GetState(context.Background(), &enterpriseclient.GetStateRequest{})
				if err != nil {
					err = fmt.Errorf("could not get Enterprise status: %v", grpcutil.ScrubGRPC(err))
					s3server.InternalError(r, err).Write(logger, w)
					return
				}

				isEnterprise = resp.State == enterpriseclient.State_ACTIVE
			}
			if !isEnterprise {
				enterpriseDisabledError(r).Write(logger, w)
				return
			}

			router.ServeHTTP(w, r)
		}),
		// NOTE: this is not closed. If the standard logger gets customized, this will need to be fixed
		ErrorLog: stdlog.New(logger.Writer(), "", 0),
	}
}
