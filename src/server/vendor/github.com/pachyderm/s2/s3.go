package s2

import (
	"net/http"
	"regexp"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

var bucketNameValidator = regexp.MustCompile(`^/[a-zA-Z0-9\-_\.]{1,255}/`)

func attachBucketRoutes(logger *logrus.Entry, router *mux.Router, handler *bucketHandler) {
	router.Methods("GET", "PUT").Queries("accelerate", "").HandlerFunc(NotImplementedEndpoint(logger))
	router.Methods("GET", "PUT").Queries("acl", "").HandlerFunc(NotImplementedEndpoint(logger))
	router.Methods("GET", "PUT", "DELETE").Queries("analytics", "").HandlerFunc(NotImplementedEndpoint(logger))
	router.Methods("GET", "PUT", "DELETE").Queries("cors", "").HandlerFunc(NotImplementedEndpoint(logger))
	router.Methods("GET", "PUT", "DELETE").Queries("encryption", "").HandlerFunc(NotImplementedEndpoint(logger))
	router.Methods("GET", "PUT", "DELETE").Queries("inventory", "").HandlerFunc(NotImplementedEndpoint(logger))
	router.Methods("GET", "PUT", "DELETE").Queries("lifecycle", "").HandlerFunc(NotImplementedEndpoint(logger))
	router.Methods("GET", "PUT").Queries("logging", "").HandlerFunc(NotImplementedEndpoint(logger))
	router.Methods("GET", "PUT", "DELETE").Queries("metrics", "").HandlerFunc(NotImplementedEndpoint(logger))
	router.Methods("GET", "PUT").Queries("notification", "").HandlerFunc(NotImplementedEndpoint(logger))
	router.Methods("GET", "PUT").Queries("object-lock", "").HandlerFunc(NotImplementedEndpoint(logger))
	router.Methods("GET", "PUT", "DELETE").Queries("policy", "").HandlerFunc(NotImplementedEndpoint(logger))
	router.Methods("GET").Queries("policyStatus", "").HandlerFunc(NotImplementedEndpoint(logger))
	router.Methods("GET", "PUT", "DELETE").Queries("publicAccessBlock", "").HandlerFunc(NotImplementedEndpoint(logger))
	router.Methods("PUT", "DELETE").Queries("replication", "").HandlerFunc(NotImplementedEndpoint(logger))
	router.Methods("GET", "PUT").Queries("requestPayment", "").HandlerFunc(NotImplementedEndpoint(logger))
	router.Methods("GET", "PUT", "DELETE").Queries("tagging", "").HandlerFunc(NotImplementedEndpoint(logger))
	router.Methods("GET").Queries("uploads", "").HandlerFunc(NotImplementedEndpoint(logger))
	router.Methods("GET", "PUT").Queries("versioning", "").HandlerFunc(NotImplementedEndpoint(logger))
	router.Methods("GET").Queries("versions", "").HandlerFunc(NotImplementedEndpoint(logger))
	router.Methods("GET", "PUT", "DELETE").Queries("website", "").HandlerFunc(NotImplementedEndpoint(logger))
	router.Methods("POST").HandlerFunc(NotImplementedEndpoint(logger))

	router.Methods("GET", "HEAD").Queries("location", "").HandlerFunc(handler.location)
	router.Methods("GET", "HEAD").HandlerFunc(handler.get)
	router.Methods("PUT").HandlerFunc(handler.put)
	router.Methods("DELETE").HandlerFunc(handler.del)
}

func attachObjectRoutes(logger *logrus.Entry, router *mux.Router, handler *objectHandler) {
	router.Methods("GET", "PUT").Queries("acl", "").HandlerFunc(NotImplementedEndpoint(logger))
	router.Methods("GET", "PUT").Queries("legal-hold", "").HandlerFunc(NotImplementedEndpoint(logger))
	router.Methods("GET", "PUT").Queries("retention", "").HandlerFunc(NotImplementedEndpoint(logger))
	router.Methods("GET", "PUT", "DELETE").Queries("tagging", "").HandlerFunc(NotImplementedEndpoint(logger))
	router.Methods("GET").Queries("torrent", "").HandlerFunc(NotImplementedEndpoint(logger))
	router.Methods("POST").Queries("restore", "").HandlerFunc(NotImplementedEndpoint(logger))
	router.Methods("POST").Queries("select", "").HandlerFunc(NotImplementedEndpoint(logger))
	router.Methods("PUT").Headers("x-amz-copy-source", "").HandlerFunc(NotImplementedEndpoint(logger))
	router.Methods("GET", "HEAD").Queries("uploadId", "").HandlerFunc(NotImplementedEndpoint(logger))
	router.Methods("POST").Queries("uploads", "").HandlerFunc(NotImplementedEndpoint(logger))
	router.Methods("POST").Queries("uploadId", "").HandlerFunc(NotImplementedEndpoint(logger))
	router.Methods("PUT").Queries("uploadId", "").HandlerFunc(NotImplementedEndpoint(logger))
	router.Methods("DELETE").Queries("uploadId", "").HandlerFunc(NotImplementedEndpoint(logger))

	router.Methods("GET", "HEAD").HandlerFunc(handler.get)
	router.Methods("PUT").HandlerFunc(handler.put)
	router.Methods("DELETE").HandlerFunc(handler.del)
}

type S3 struct {
	Root   RootController
	Bucket BucketController
	Object ObjectController
}

func NewS3() *S3 {
	return &S3{
		Root:   UnimplementedRootController{},
		Bucket: UnimplementedBucketController{},
		Object: UnimplementedObjectController{},
	}
}

func (h *S3) Router(logger *logrus.Entry) *mux.Router {
	rootHandler := &rootHandler{
		controller: h.Root,
		logger:     logger,
	}
	bucketHandler := &bucketHandler{
		controller: h.Bucket,
		logger:     logger,
	}
	objectHandler := &objectHandler{
		controller: h.Object,
		logger:     logger,
	}

	router := mux.NewRouter()
	router.Path(`/`).Methods("GET", "HEAD").HandlerFunc(rootHandler.get)

	// Bucket-related routes. Repo validation regex is the same that the aws
	// cli uses. There's two routers - one with a trailing a slash and one
	// without. Both route to the same handlers, i.e. a request to `/foo` is
	// the same as `/foo/`. This is used instead of mux's builtin "strict
	// slash" functionality, because that uses redirects which doesn't always
	// play nice with s3 clients.
	trailingSlashBucketRouter := router.Path(`/{bucket:[a-zA-Z0-9\-_\.]{1,255}}/`).Subrouter()
	attachBucketRoutes(logger, trailingSlashBucketRouter, bucketHandler)
	bucketRouter := router.Path(`/{bucket:[a-zA-Z0-9\-_\.]{1,255}}`).Subrouter()
	attachBucketRoutes(logger, bucketRouter, bucketHandler)

	// Object-related routes
	objectRouter := router.Path(`/{bucket:[a-zA-Z0-9\-_\.]{1,255}}/{key:.+}`).Subrouter()
	attachObjectRoutes(logger, objectRouter, objectHandler)

	router.MethodNotAllowedHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Infof("method not allowed: %s %s", r.Method, r.URL.Path)
		MethodNotAllowedError(r).Write(logger, w)
	})

	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Infof("not found: %s", r.URL.Path)
		if bucketNameValidator.MatchString(r.URL.Path) {
			NoSuchKeyError(r).Write(logger, w)
		} else {
			InvalidBucketNameError(r).Write(logger, w)
		}
	})

	return router
}
