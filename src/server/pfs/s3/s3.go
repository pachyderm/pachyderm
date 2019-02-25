package s3

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"github.com/pachyderm/pachyderm/src/client"
)


// Server runs an HTTP server with an S3-like API for PFS. This allows you to
// use s3 clients to acccess PFS contents.
//
// This returns an `http.Server` instance. It is the responsibility of the
// caller to start the server. This also makes it possible for the caller to
// enable graceful shutdown if desired; see the `http` package for details.
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
func Server(pc *client.APIClient, port uint16) *http.Server {
	// repo validation regex is the same as minio
	router := mux.NewRouter()
	router.Handle(`/`, newRootHandler(pc)).Methods("GET", "HEAD")
	router.Handle(`/{repo:[a-z0-9][a-z0-9\.\-]{1,61}[a-z0-9]}/`, newBucketHandler(pc)).Methods("GET", "HEAD", "PUT", "DELETE")
	router.Handle(`/{repo:[a-z0-9][a-z0-9\.\-]{1,61}[a-z0-9]}/{branch}/{file:.+}`, newObjectHandler(pc)).Methods("GET", "HEAD", "PUT", "DELETE")

	// Note: error log is not customized on this `http.Server`, which means
	// it'll default to using the stdlib logger and produce log messages that
	// don't look like the ones produced elsewhere, and aren't configured
	// properly. In testing, this didn't seem to be a big deal because it's
	// rather hard to trigger it anyways, but if we find a reliable way to
	// create error logs, it might be worthwhile to fix.
	return &http.Server{
		Addr: fmt.Sprintf(":%d", port),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO: reduce log level
			log.Infof("s3 gateway request: %s %s", r.Method, r.RequestURI)
			router.ServeHTTP(w, r)
		}),
	}
}
