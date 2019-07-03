package s3server

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

type GetObjectResult struct {
	Name    string
	Hash    []byte
	ModTime time.Time
	Content io.ReadSeeker
}

type ObjectController interface {
	Get(r *http.Request, bucket, key string, result *GetObjectResult) *Error
	Put(r *http.Request, bucket, key string, reader io.Reader) *Error
	Del(r *http.Request, bucket, key string) *Error
}

type UnimplementedObjectController struct{}

func (c UnimplementedObjectController) Get(r *http.Request, bucket, key string, result *GetObjectResult) *Error {
	return NotImplementedError(r)
}

func (c UnimplementedObjectController) Put(r *http.Request, bucket, key string, reader io.Reader) *Error {
	return NotImplementedError(r)
}

func (c UnimplementedObjectController) Del(r *http.Request, bucket, key string) *Error {
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

	if err := h.controller.Get(r, bucket, key, result); err != nil {
		err.Write(h.logger, w)
		return
	}

	if result.Hash != nil {
		w.Header().Set("ETag", fmt.Sprintf("\"%x\"", result.Hash))
	}
	http.ServeContent(w, r, result.Name, result.ModTime, result.Content)
}

func (h *objectHandler) put(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]
	key := vars["key"]

	expectedHash, ok := r.Header["Content-Md5"]
	var expectedHashBytes []uint8
	var err error
	if ok && len(expectedHash) == 1 {
		expectedHashBytes, err = base64.StdEncoding.DecodeString(expectedHash[0])
		if err != nil || len(expectedHashBytes) != 16 {
			InvalidDigestError(r).Write(h.logger, w)
			return
		}
	}

	hasher := md5.New()
	reader := io.TeeReader(r.Body, hasher)
	if err := h.controller.Put(r, bucket, key, reader); err != nil {
		err.Write(h.logger, w)
		return
	}

	actualHashBytes := hasher.Sum(nil)
	if expectedHashBytes != nil && !bytes.Equal(expectedHashBytes, actualHashBytes) {
		BadDigestError(r).Write(h.logger, w)

		// try to clean up the file
		if err := h.controller.Del(r, bucket, key); err != nil {
			h.logger.Errorf("could not clean up file after an error: %+v", err)
		}

		return
	}

	w.Header().Set("ETag", fmt.Sprintf("\"%x\"", actualHashBytes))
	w.WriteHeader(http.StatusOK)
}

func (h *objectHandler) del(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]
	key := vars["key"]

	if err := h.controller.Del(r, bucket, key); err != nil {
		err.Write(h.logger, w)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
