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
	GetObject(r *http.Request, bucket, key string) (etag string, modTime time.Time, content io.ReadSeeker, err error)
	// PutObject sets an object
	PutObject(r *http.Request, bucket, key string, reader io.Reader) (etag string, err error)
	// DeleteObject deletes an object
	DeleteObject(r *http.Request, bucket, key string) error
}

// unimplementedObjectController defines a controller that returns
// `NotImplementedError` for all functionality
type unimplementedObjectController struct{}

func (c unimplementedObjectController) GetObject(r *http.Request, bucket, key string) (etag string, modTime time.Time, content io.ReadSeeker, err error) {
	err = NotImplementedError(r)
	return
}

func (c unimplementedObjectController) PutObject(r *http.Request, bucket, key string, reader io.Reader) (etag string, err error) {
	err = NotImplementedError(r)
	return
}

func (c unimplementedObjectController) DeleteObject(r *http.Request, bucket, key string) error {
	return NotImplementedError(r)
}

type objectHandler struct {
	controller ObjectController
	logger     *logrus.Entry
}

func (h *objectHandler) get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]
	key := vars["key"]

	etag, modTime, content, err := h.controller.GetObject(r, bucket, key)
	if err != nil {
		WriteError(h.logger, w, r, err)
		return
	}

	if etag != "" {
		w.Header().Set("ETag", addETagQuotes(etag))
	}

	http.ServeContent(w, r, key, modTime, content)
}

func (h *objectHandler) put(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]
	key := vars["key"]
	etag := ""

	shouldCleanup, err := withBodyReader(r, func(reader io.Reader) error {
		fetchedETag, err := h.controller.PutObject(r, bucket, key, reader)
		etag = fetchedETag
		return err
	})

	if shouldCleanup {
		// try to clean up the file
		if err := h.controller.DeleteObject(r, bucket, key); err != nil {
			h.logger.Errorf("could not clean up file after an error: %+v", err)
		}
	}

	if err != nil {
		WriteError(h.logger, w, r, err)
		return
	}

	if etag != "" {
		w.Header().Set("ETag", addETagQuotes(etag))
	}
	w.WriteHeader(http.StatusOK)
}

func (h *objectHandler) del(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]
	key := vars["key"]

	if err := h.controller.DeleteObject(r, bucket, key); err != nil {
		WriteError(h.logger, w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
