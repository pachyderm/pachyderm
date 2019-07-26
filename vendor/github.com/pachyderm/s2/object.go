package s2

import (
	"io"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// ObjectController is an interface that specifies object-level functionality.
type ObjectController interface {
	// GetObject gets an object
	GetObject(r *http.Request, bucket, key, version string) (etag, fetchedVersion string, modTime time.Time, content io.ReadSeeker, err error)
	// PutObject sets an object
	PutObject(r *http.Request, bucket, key string, reader io.Reader) (etag, createdVersion string, err error)
	// DeleteObject deletes an object
	DeleteObject(r *http.Request, bucket, key, version string) (removedVersion string, err error)
}

// unimplementedObjectController defines a controller that returns
// `NotImplementedError` for all functionality
type unimplementedObjectController struct{}

func (c unimplementedObjectController) GetObject(r *http.Request, bucket, key, version string) (etag, fetchedVersion string, modTime time.Time, content io.ReadSeeker, err error) {
	err = NotImplementedError(r)
	return
}

func (c unimplementedObjectController) PutObject(r *http.Request, bucket, key string, reader io.Reader) (etag, createdVersion string, err error) {
	err = NotImplementedError(r)
	return
}

func (c unimplementedObjectController) DeleteObject(r *http.Request, bucket, key, version string) (removedVersion string, err error) {
	return "", NotImplementedError(r)
}

type objectHandler struct {
	controller ObjectController
	logger     *logrus.Entry
}

func (h *objectHandler) get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]
	key := vars["key"]
	versionId := r.FormValue("versionId")

	etag, version, modTime, content, err := h.controller.GetObject(r, bucket, key, versionId)
	if err != nil {
		WriteError(h.logger, w, r, err)
		return
	}

	if etag != "" {
		w.Header().Set("ETag", addETagQuotes(etag))
	}
	if version != "" {
		w.Header().Set("x-amz-version-id", version)
	}
	http.ServeContent(w, r, key, modTime, content)
}

func (h *objectHandler) put(w http.ResponseWriter, r *http.Request) {
	if err := requireContentLength(r); err != nil {
		WriteError(h.logger, w, r, err)
		return
	}

	vars := mux.Vars(r)
	bucket := vars["bucket"]
	key := vars["key"]

	etag, version, err := h.controller.PutObject(r, bucket, key, r.Body)
	if err != nil {
		WriteError(h.logger, w, r, err)
		return
	}

	if etag != "" {
		w.Header().Set("ETag", addETagQuotes(etag))
	}
	if version != "" {
		w.Header().Set("x-amz-version-id", version)
	}
	w.WriteHeader(http.StatusOK)
}

func (h *objectHandler) del(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]
	key := vars["key"]
	versionId := r.FormValue("versionId")

	version, err := h.controller.DeleteObject(r, bucket, key, versionId)
	if err != nil {
		WriteError(h.logger, w, r, err)
		return
	}

	if version != "" {
		w.Header().Set("x-amz-version-id", version)
	}
	w.WriteHeader(http.StatusNoContent)
}
