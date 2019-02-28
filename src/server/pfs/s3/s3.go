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
// Bucket names correspond to repo names, and files are accessible via the s3
// key pattern "<branch>/<filepath>". For example, to get the file "a/b/c.txt"
// on the "foo" repo's "master" branch, you'd making an s3 get request with
// bucket = "foo", key = "master/a/b/c.txt".
//
// Note: in s3, bucket names are constrained by IETF RFC 1123, (and its
// predecessor RFC 952) but pachyderm's repo naming constraints are slightly
// more liberal. This server only supports the subset of pachyderm repos whose
// names are RFC 1123 compatible.
//
// Note: In `s3cmd`, you must set the access key and secret key, even though
// this API will ignore them - otherwise, you'll get an opaque config error:
// https://github.com/s3tools/s3cmd/issues/845#issuecomment-464885959
func Server(pc *client.APIClient, port uint16, errLogWriter io.Writer, multipartDir string) *http.Server {
	// strict slash is enabled because we don't want to serve back redirects
	// for typical s3 requests, which usually include a trailing /
	router := mux.NewRouter().StrictSlash(true)
	router.Handle(`/`, newRootHandler(pc)).Methods("GET", "HEAD")

	// bucket-related routes
	// repo validation regex is the same as minio
	bucketHandler := newBucketHandler(pc)
	bucketRouter := router.Path(`/{repo:[a-z0-9][a-z0-9\.\-]{1,61}[a-z0-9]}/`).Subrouter()
	bucketRouter.Methods("GET", "HEAD").Queries("location", "").HandlerFunc(bucketHandler.location)
	bucketRouter.Methods("GET", "HEAD").HandlerFunc(bucketHandler.get)
	bucketRouter.Methods("PUT").HandlerFunc(bucketHandler.put)
	bucketRouter.Methods("DELETE").HandlerFunc(bucketHandler.delete)

	// object-related routes
	objectHandler := newObjectHandler(pc, multipartDir)
	objectRouter := router.Path(`/{repo:[a-z0-9][a-z0-9\.\-]{1,61}[a-z0-9]}/{branch}/{file:.+}`).Subrouter()
	objectRouter.Methods("GET", "HEAD").Queries("uploadId", "").HandlerFunc(objectHandler.listMultipart)
	objectRouter.Methods("GET", "HEAD").HandlerFunc(objectHandler.get)
	objectRouter.Methods("POST").Queries("uploads", "").HandlerFunc(objectHandler.initMultipart)
	objectRouter.Methods("POST").Queries("uploadId", "").HandlerFunc(objectHandler.completeMultipart)
	objectRouter.Methods("PUT").Queries("uploadId", "").HandlerFunc(objectHandler.uploadMultipart)
	objectRouter.Methods("PUT").HandlerFunc(objectHandler.put)
	objectRouter.Methods("DELETE").Queries("uploadId", "").HandlerFunc(objectHandler.abortMultipart)
	objectRouter.Methods("DELETE").HandlerFunc(objectHandler.delete)

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
