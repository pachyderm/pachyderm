package s3

import (
	"fmt"
	stdlog "log"
	"net/http"
	"regexp"
	"time"

	"golang.org/x/net/context"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/pachyderm/pachyderm/src/client"
	enterpriseclient "github.com/pachyderm/pachyderm/src/client/enterprise"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
)

var enterpriseTimeout = 24 * time.Hour
var bucketNameValidator = regexp.MustCompile(`^/[a-zA-Z0-9\-_]{1,255}\.[a-zA-Z0-9\-_]{1,255}/`)

func attachBucketRoutes(router *mux.Router, handler bucketHandler) {
	router.Methods("GET", "PUT").Queries("accelerate", "").HandlerFunc(notImplementedError)
	router.Methods("GET", "PUT").Queries("acl", "").HandlerFunc(notImplementedError)
	router.Methods("GET", "PUT", "DELETE").Queries("analytics", "").HandlerFunc(notImplementedError)
	router.Methods("GET", "PUT", "DELETE").Queries("cors", "").HandlerFunc(notImplementedError)
	router.Methods("GET", "PUT", "DELETE").Queries("encryption", "").HandlerFunc(notImplementedError)
	router.Methods("GET", "PUT", "DELETE").Queries("inventory", "").HandlerFunc(notImplementedError)
	router.Methods("GET", "PUT", "DELETE").Queries("lifecycle", "").HandlerFunc(notImplementedError)
	router.Methods("GET", "PUT").Queries("logging", "").HandlerFunc(notImplementedError)
	router.Methods("GET", "PUT", "DELETE").Queries("metrics", "").HandlerFunc(notImplementedError)
	router.Methods("GET", "PUT").Queries("notification", "").HandlerFunc(notImplementedError)
	router.Methods("GET", "PUT").Queries("object-lock", "").HandlerFunc(notImplementedError)
	router.Methods("GET", "PUT", "DELETE").Queries("policy", "").HandlerFunc(notImplementedError)
	router.Methods("GET").Queries("policyStatus", "").HandlerFunc(notImplementedError)
	router.Methods("GET", "PUT", "DELETE").Queries("publicAccessBlock", "").HandlerFunc(notImplementedError)
	router.Methods("PUT", "DELETE").Queries("replication", "").HandlerFunc(notImplementedError)
	router.Methods("GET", "PUT").Queries("requestPayment", "").HandlerFunc(notImplementedError)
	router.Methods("GET", "PUT", "DELETE").Queries("tagging", "").HandlerFunc(notImplementedError)
	router.Methods("GET").Queries("uploads", "").HandlerFunc(notImplementedError)
	router.Methods("GET", "PUT").Queries("versioning", "").HandlerFunc(notImplementedError)
	router.Methods("GET").Queries("versions", "").HandlerFunc(notImplementedError)
	router.Methods("GET", "PUT", "DELETE").Queries("website", "").HandlerFunc(notImplementedError)
	router.Methods("POST").HandlerFunc(notImplementedError)

	router.Methods("GET", "HEAD").Queries("location", "").HandlerFunc(handler.location)
	router.Methods("GET", "HEAD").HandlerFunc(handler.get)
	router.Methods("PUT").HandlerFunc(handler.put)
	router.Methods("DELETE").HandlerFunc(handler.del)
}

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
	router := mux.NewRouter()
	router.Handle(`/`, newRootHandler(pc)).Methods("GET", "HEAD")

	// Bucket-related routes. Repo validation regex is the same that the aws
	// cli uses. There's two routers - one with a trailing a slash and one
	// without. Both route to the same handlers, i.e. a request to `/foo` is
	// the same as `/foo/`. This is used instead of mux's builtin "strict
	// slash" functionality, because that uses redirects which doesn't always
	// play nice with s3 clients.
	bucketHandler := newBucketHandler(pc)
	trailingSlashBucketRouter := router.Path(`/{branch:[a-zA-Z0-9\-_]{1,255}}.{repo:[a-zA-Z0-9\-_]{1,255}}/`).Subrouter()
	attachBucketRoutes(trailingSlashBucketRouter, bucketHandler)
	bucketRouter := router.Path(`/{branch:[a-zA-Z0-9\-_]{1,255}}.{repo:[a-zA-Z0-9\-_]{1,255}}`).Subrouter()
	attachBucketRoutes(bucketRouter, bucketHandler)

	// object-related routes
	objectRouter := router.Path(`/{branch:[a-zA-Z0-9\-_]{1,255}}.{repo:[a-zA-Z0-9\-_]{1,255}}/{file:.+}`).Subrouter()

	objectRouter.Methods("GET", "PUT").Queries("acl", "").HandlerFunc(notImplementedError)
	objectRouter.Methods("GET", "PUT").Queries("legal-hold", "").HandlerFunc(notImplementedError)
	objectRouter.Methods("GET", "PUT").Queries("retention", "").HandlerFunc(notImplementedError)
	objectRouter.Methods("GET", "PUT", "DELETE").Queries("tagging", "").HandlerFunc(notImplementedError)
	objectRouter.Methods("GET").Queries("torrent", "").HandlerFunc(notImplementedError)
	objectRouter.Methods("POST").Queries("restore", "").HandlerFunc(notImplementedError)
	objectRouter.Methods("POST").Queries("select", "").HandlerFunc(notImplementedError)
	objectRouter.Methods("PUT").Headers("x-amz-copy-source", "").HandlerFunc(notImplementedError) // maybe worth implementing at some point
	objectRouter.Methods("GET", "HEAD").Queries("uploadId", "").HandlerFunc(notImplementedError)
	objectRouter.Methods("POST").Queries("uploads", "").HandlerFunc(notImplementedError)
	objectRouter.Methods("POST").Queries("uploadId", "").HandlerFunc(notImplementedError)
	objectRouter.Methods("PUT").Queries("uploadId", "").HandlerFunc(notImplementedError)
	objectRouter.Methods("DELETE").Queries("uploadId", "").HandlerFunc(notImplementedError)

	objectHandler := newObjectHandler(pc)
	objectRouter.Methods("GET", "HEAD").HandlerFunc(objectHandler.get)
	objectRouter.Methods("PUT").HandlerFunc(objectHandler.put)
	objectRouter.Methods("DELETE").HandlerFunc(objectHandler.del)

	router.MethodNotAllowedHandler = http.HandlerFunc(methodNotAllowedError)
	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestLogger(r).Infof("not found: %+v", r.URL.Path)
		if bucketNameValidator.MatchString(r.URL.Path) {
			noSuchKeyError(w, r)
		} else {
			invalidBucketNameError(w, r)
		}
	})

	// NOTE: this is not closed. If the standard logger gets customized, this will need to be fixed
	serverErrorLog := logrus.WithFields(logrus.Fields{
		"source": "s3gateway",
	}).Writer()

	var lastEnterpriseCheck time.Time
	isEnterprise := false

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
			requestLogger(r).Debugf("http request: %s %s", r.Method, r.RequestURI)

			// Ensure enterprise is enabled
			now := time.Now()
			if !isEnterprise || now.Sub(lastEnterpriseCheck) > enterpriseTimeout {
				resp, err := pc.Enterprise.GetState(context.Background(), &enterpriseclient.GetStateRequest{})
				if err != nil {
					err = fmt.Errorf("could not get Enterprise status: %v", grpcutil.ScrubGRPC(err))
					internalError(w, r, err)
					return
				}

				isEnterprise = resp.State == enterpriseclient.State_ACTIVE
			}
			if !isEnterprise {
				enterpriseDisabledError(w, r)
				return
			}

			router.ServeHTTP(w, r)
		}),
		ErrorLog: stdlog.New(serverErrorLog, "", 0),
	}
}
