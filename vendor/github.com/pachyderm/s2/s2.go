package s2

import (
	"fmt"
	"net/http"
	"regexp"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

var bucketNameValidator = regexp.MustCompile(`^/[a-zA-Z0-9\-_\.]{1,255}/`)

func attachBucketRoutes(logger *logrus.Entry, router *mux.Router, handler *bucketHandler, multipartHandler *multipartHandler) {
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
	router.Methods("GET", "PUT").Queries("versioning", "").HandlerFunc(NotImplementedEndpoint(logger))
	router.Methods("GET").Queries("versions", "").HandlerFunc(NotImplementedEndpoint(logger))
	router.Methods("GET", "PUT", "DELETE").Queries("website", "").HandlerFunc(NotImplementedEndpoint(logger))
	router.Methods("POST").HandlerFunc(NotImplementedEndpoint(logger))

	router.Methods("GET").Queries("uploads", "").HandlerFunc(multipartHandler.list)
	router.Methods("GET", "HEAD").Queries("location", "").HandlerFunc(handler.location)
	router.Methods("GET", "HEAD").HandlerFunc(handler.get)
	router.Methods("PUT").HandlerFunc(handler.put)
	router.Methods("DELETE").HandlerFunc(handler.del)
}

func attachObjectRoutes(logger *logrus.Entry, router *mux.Router, handler *objectHandler, multipartHandler *multipartHandler) {
	router.Methods("GET", "PUT").Queries("acl", "").HandlerFunc(NotImplementedEndpoint(logger))
	router.Methods("GET", "PUT").Queries("legal-hold", "").HandlerFunc(NotImplementedEndpoint(logger))
	router.Methods("GET", "PUT").Queries("retention", "").HandlerFunc(NotImplementedEndpoint(logger))
	router.Methods("GET", "PUT", "DELETE").Queries("tagging", "").HandlerFunc(NotImplementedEndpoint(logger))
	router.Methods("GET").Queries("torrent", "").HandlerFunc(NotImplementedEndpoint(logger))
	router.Methods("POST").Queries("restore", "").HandlerFunc(NotImplementedEndpoint(logger))
	router.Methods("POST").Queries("select", "").HandlerFunc(NotImplementedEndpoint(logger))
	router.Methods("PUT").Headers("x-amz-copy-source", "").HandlerFunc(NotImplementedEndpoint(logger))

	router.Methods("GET", "HEAD").Queries("uploadId", "").HandlerFunc(multipartHandler.listChunks)
	router.Methods("POST").Queries("uploads", "").HandlerFunc(multipartHandler.init)
	router.Methods("POST").Queries("uploadId", "").HandlerFunc(multipartHandler.complete)
	router.Methods("PUT").Queries("uploadId", "").HandlerFunc(multipartHandler.put)
	router.Methods("DELETE").Queries("uploadId", "").HandlerFunc(multipartHandler.del)
	router.Methods("GET", "HEAD").HandlerFunc(handler.get)
	router.Methods("PUT").HandlerFunc(handler.put)
	router.Methods("DELETE").HandlerFunc(handler.del)
}

type S2 struct {
	Root      RootController
	Bucket    BucketController
	Object    ObjectController
	Multipart MultipartController
	logger    *logrus.Entry
}

func NewS2(logger *logrus.Entry) *S2 {
	return &S2{
		Root:      UnimplementedRootController{},
		Bucket:    UnimplementedBucketController{},
		Object:    UnimplementedObjectController{},
		Multipart: UnimplementedMultipartController{},
		logger:    logger,
	}
}

func (h *S2) requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		id, err := uuid.NewV4()
		if err != nil {
			baseErr := fmt.Errorf("could not generate request ID: %v", err)
			writeError(h.logger, w, r, InternalError(r, baseErr))
			return
		}

		vars["requestID"] = id.String()
		next.ServeHTTP(w, r)
	})
}

func (h *S2) Router() *mux.Router {
	rootHandler := &rootHandler{
		controller: h.Root,
		logger:     h.logger,
	}
	bucketHandler := &bucketHandler{
		controller: h.Bucket,
		logger:     h.logger,
	}
	objectHandler := &objectHandler{
		controller: h.Object,
		logger:     h.logger,
	}
	multipartHandler := &multipartHandler{
		controller: h.Multipart,
		logger:     h.logger,
	}

	router := mux.NewRouter()
	router.Use(h.requestIDMiddleware)

	router.Path(`/`).Methods("GET", "HEAD").HandlerFunc(rootHandler.get)

	// Bucket-related routes. Repo validation regex is the same that the aws
	// cli uses. There's two routers - one with a trailing a slash and one
	// without. Both route to the same handlers, i.e. a request to `/foo` is
	// the same as `/foo/`. This is used instead of mux's builtin "strict
	// slash" functionality, because that uses redirects which doesn't always
	// play nice with s3 clients.
	trailingSlashBucketRouter := router.Path(`/{bucket:[a-zA-Z0-9\-_\.]{1,255}}/`).Subrouter()
	attachBucketRoutes(h.logger, trailingSlashBucketRouter, bucketHandler, multipartHandler)
	bucketRouter := router.Path(`/{bucket:[a-zA-Z0-9\-_\.]{1,255}}`).Subrouter()
	attachBucketRoutes(h.logger, bucketRouter, bucketHandler, multipartHandler)

	// Object-related routes
	objectRouter := router.Path(`/{bucket:[a-zA-Z0-9\-_\.]{1,255}}/{key:.+}`).Subrouter()
	attachObjectRoutes(h.logger, objectRouter, objectHandler, multipartHandler)

	router.MethodNotAllowedHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.logger.Infof("method not allowed: %s %s", r.Method, r.URL.Path)
		writeError(h.logger, w, r, MethodNotAllowedError(r))
	})

	router.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.logger.Infof("not found: %s", r.URL.Path)
		if bucketNameValidator.MatchString(r.URL.Path) {
			writeError(h.logger, w, r, NoSuchKeyError(r))
		} else {
			writeError(h.logger, w, r, InvalidBucketNameError(r))
		}
	})

	return router
}
