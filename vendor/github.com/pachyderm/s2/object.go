package s2

import (
	"io"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type GetObjectResult struct {
	Name    string        `xml:"Name"`
	ETag    string        `xml:"ETag"`
	ModTime time.Time     `xml:"ModTime"`
	Content io.ReadSeeker `xml:"Content"`
}

type ObjectController interface {
	GetObject(r *http.Request, bucket, key string, result *GetObjectResult) error
	PutObject(r *http.Request, bucket, key string, reader io.Reader) (string, error)
	DeleteObject(r *http.Request, bucket, key string) error
}

type UnimplementedObjectController struct{}

func (c UnimplementedObjectController) GetObject(r *http.Request, bucket, key string, result *GetObjectResult) error {
	return NotImplementedError(r)
}

func (c UnimplementedObjectController) PutObject(r *http.Request, bucket, key string, reader io.Reader) (string, error) {
	return "", NotImplementedError(r)
}

func (c UnimplementedObjectController) DeleteObject(r *http.Request, bucket, key string) error {
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

	result := &GetObjectResult{}

	if err := h.controller.GetObject(r, bucket, key, result); err != nil {
		writeError(h.logger, w, r, err)
		return
	}

	if result.ETag != "" {
		w.Header().Set("ETag", addETagQuotes(result.ETag))
	}

	http.ServeContent(w, r, result.Name, result.ModTime, result.Content)
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
		writeError(h.logger, w, r, err)
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
		writeError(h.logger, w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
