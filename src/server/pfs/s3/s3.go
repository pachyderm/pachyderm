package s3

import (
	"fmt"
	"io"
	stdlog "log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"

	"github.com/pachyderm/pachyderm/src/client"
)

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
	router.Methods("GET").Queries("uploads", "").HandlerFunc(notImplementedError) // maybe worth implementing at some point
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
// caller to:
// 1) start the returned server
// 2) close `errLogWriter`
// 3) remove `multipartDir`, unless you want to persist in-flight multipart
//    contents between server runs
// Furthermore, it's possible for the caller to gracefully shutdown the server
// if desired; see the `http` package for details.
//
// If `multipartDir` is an empty string, multipart uploads are disabled.
//
// Note: In `s3cmd`, you must set the access key and secret key, even though
// this API will ignore them - otherwise, you'll get an opaque config error:
// https://github.com/s3tools/s3cmd/issues/845#issuecomment-464885959
func Server(pc *client.APIClient, port uint16, errLogWriter io.Writer, multipartDir string) *http.Server {
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

	if multipartDir != "" {
		// Multipart handlers are only registered if a root dir is specified
		multipartHandler := newMultipartHandler(pc, multipartDir)
		objectRouter.Methods("GET", "HEAD").Queries("uploadId", "").HandlerFunc(multipartHandler.list)
		objectRouter.Methods("POST").Queries("uploads", "").HandlerFunc(multipartHandler.init)
		objectRouter.Methods("POST").Queries("uploadId", "").HandlerFunc(multipartHandler.complete)
		objectRouter.Methods("PUT").Queries("uploadId", "").HandlerFunc(multipartHandler.put)
		objectRouter.Methods("DELETE").Queries("uploadId", "").HandlerFunc(multipartHandler.del)
	}

	objectRouter.Methods("GET", "PUT").Queries("acl", "").HandlerFunc(notImplementedError)
	objectRouter.Methods("GET", "PUT").Queries("legal-hold", "").HandlerFunc(notImplementedError)
	objectRouter.Methods("GET", "PUT").Queries("retention", "").HandlerFunc(notImplementedError)
	objectRouter.Methods("GET", "PUT", "DELETE").Queries("tagging", "").HandlerFunc(notImplementedError)
	objectRouter.Methods("GET").Queries("torrent", "").HandlerFunc(notImplementedError)
	objectRouter.Methods("POST").Queries("restore", "").HandlerFunc(notImplementedError)
	objectRouter.Methods("POST").Queries("select", "").HandlerFunc(notImplementedError)
	objectRouter.Methods("PUT").Headers("x-amz-copy-source", "").HandlerFunc(notImplementedError) // maybe worth implementing at some point

	objectHandler := newObjectHandler(pc)
	objectRouter.Methods("GET", "HEAD").HandlerFunc(objectHandler.get)
	objectRouter.Methods("PUT").HandlerFunc(objectHandler.put)
	objectRouter.Methods("DELETE").HandlerFunc(objectHandler.del)

	// TODO: this will trigger for paths that are not valid utf-8 strings, giving the incorrect error message. See:
	// ./etc/testing/s3gateway/conformance.py --nose-args 's3tests.functional.test_s3:test_object_create_unreadable' --no-persist
	router.NotFoundHandler = http.HandlerFunc(invalidBucketNameError)
	router.MethodNotAllowedHandler = http.HandlerFunc(methodNotAllowedError)

	return &http.Server{
		Addr: fmt.Sprintf(":%d", port),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO: reduce log level
			logrus.Infof("s3gateway: http request: %s %s", r.Method, r.RequestURI)
			router.ServeHTTP(w, r)
		}),
		ErrorLog: stdlog.New(errLogWriter, "s3gateway: ", 0),
	}
}
