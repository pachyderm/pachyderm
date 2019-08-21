package s2

import (
	"encoding/xml"
	"io"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// GetObjectResult is a response from a GetObject call
type GetObjectResult struct {
	// ETag is a hex encoding of the hash of the object contents, with or
	// without surrounding quotes.
	ETag string
	// Version is the version of the object, or an empty string if versioning
	// is not enabled or supported.
	Version string
	// DeleteMarker specifies whether there's a delete marker in place of the
	// object.
	DeleteMarker bool
	// ModTime specifies when the object was modified.
	ModTime time.Time
	// Content is the contents of the object.
	Content io.ReadSeeker
}

// PutObjectResult is a response from a PutObject call
type PutObjectResult struct {
	// ETag is a hex encoding of the hash of the object contents, with or
	// without surrounding quotes.
	ETag string
	// Version is the version of the object, or an empty string if versioning
	// is not enabled or supported.
	Version string
}

// DeleteObjectResult is a response from a DeleteObject call
type DeleteObjectResult struct {
	// Version is the version of the object, or an empty string if versioning
	// is not enabled or supported.
	Version string
	// DeleteMarker specifies whether there's a delete marker in place of the
	// object.
	DeleteMarker bool
}

// ObjectController is an interface that specifies object-level functionality.
type ObjectController interface {
	// GetObject gets an object
	GetObject(r *http.Request, bucket, key, version string) (*GetObjectResult, error)
	// PutObject sets an object
	PutObject(r *http.Request, bucket, key string, reader io.Reader) (*PutObjectResult, error)
	// DeleteObject deletes an object
	DeleteObject(r *http.Request, bucket, key, version string) (*DeleteObjectResult, error)
}

// unimplementedObjectController defines a controller that returns
// `NotImplementedError` for all functionality
type unimplementedObjectController struct{}

func (c unimplementedObjectController) GetObject(r *http.Request, bucket, key, version string) (*GetObjectResult, error) {
	return nil, NotImplementedError(r)
}

func (c unimplementedObjectController) PutObject(r *http.Request, bucket, key string, reader io.Reader) (*PutObjectResult, error) {
	return nil, NotImplementedError(r)
}

func (c unimplementedObjectController) DeleteObject(r *http.Request, bucket, key, version string) (*DeleteObjectResult, error) {
	return nil, NotImplementedError(r)
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

	result, err := h.controller.GetObject(r, bucket, key, versionId)
	if err != nil {
		WriteError(h.logger, w, r, err)
		return
	}

	if result.ETag != "" {
		w.Header().Set("ETag", addETagQuotes(result.ETag))
	}
	if result.Version != "" {
		w.Header().Set("x-amz-version-id", result.Version)
	}

	if result.DeleteMarker {
		w.Header().Set("x-amz-delete-marker", "true")
		WriteError(h.logger, w, r, NoSuchKeyError(r))
		return
	}

	http.ServeContent(w, r, key, result.ModTime, result.Content)
}

func (h *objectHandler) put(w http.ResponseWriter, r *http.Request) {
	if err := requireContentLength(r); err != nil {
		WriteError(h.logger, w, r, err)
		return
	}

	vars := mux.Vars(r)
	bucket := vars["bucket"]
	key := vars["key"]

	result, err := h.controller.PutObject(r, bucket, key, r.Body)
	if err != nil {
		WriteError(h.logger, w, r, err)
		return
	}

	if result.ETag != "" {
		w.Header().Set("ETag", addETagQuotes(result.ETag))
	}
	if result.Version != "" {
		w.Header().Set("x-amz-version-id", result.Version)
	}
	w.WriteHeader(http.StatusOK)
}

func (h *objectHandler) del(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]
	key := vars["key"]
	versionId := r.FormValue("versionId")

	result, err := h.controller.DeleteObject(r, bucket, key, versionId)
	if err != nil {
		WriteError(h.logger, w, r, err)
		return
	}

	if result.Version != "" {
		w.Header().Set("x-amz-version-id", result.Version)
	}
	if result.DeleteMarker {
		w.Header().Set("x-amz-delete-marker", "true")
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *objectHandler) post(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]

	payload := struct {
		XMLName xml.Name `xml:"Delete"`
		Quiet   bool     `xml:"Quiet"`
		Objects []struct {
			Key     string `xml:"Key"`
			Version string `xml:"VersionId"`
		} `xml:"Object"`
	}{}
	if err := readXMLBody(r, &payload); err != nil {
		WriteError(h.logger, w, r, err)
		return
	}

	marshallable := struct {
		XMLName xml.Name `xml:"DeleteResult"`
		Deleted []struct {
			Key                 string `xml:"Key"`
			Version             string `xml:"Version,omitempty"`
			DeleteMarker        bool   `xml:"Code,omitempty"`
			DeleteMarkerVersion string `xml:"DeleteMarkerVersionId,omitempty"`
		} `xml:"Deleted"`
		Errors []struct {
			Key     string `xml:"Key"`
			Code    string `xml:"Code"`
			Message string `xml:"Message"`
		} `xml:"Error"`
	}{
		Deleted: []struct {
			Key                 string `xml:"Key"`
			Version             string `xml:"Version,omitempty"`
			DeleteMarker        bool   `xml:"Code,omitempty"`
			DeleteMarkerVersion string `xml:"DeleteMarkerVersionId,omitempty"`
		}{},
		Errors: []struct {
			Key     string `xml:"Key"`
			Code    string `xml:"Code"`
			Message string `xml:"Message"`
		}{},
	}

	for _, object := range payload.Objects {
		result, err := h.controller.DeleteObject(r, bucket, object.Key, object.Version)
		if err != nil {
			s3Err := NewFromGenericError(r, err)

			marshallable.Errors = append(marshallable.Errors, struct {
				Key     string `xml:"Key"`
				Code    string `xml:"Code"`
				Message string `xml:"Message"`
			}{
				Key:     object.Key,
				Code:    s3Err.Code,
				Message: s3Err.Message,
			})
		} else {
			deleteMarkerVersion := ""
			if result.DeleteMarker {
				deleteMarkerVersion = result.Version
			}

			if !payload.Quiet {
				marshallable.Deleted = append(marshallable.Deleted, struct {
					Key                 string `xml:"Key"`
					Version             string `xml:"Version,omitempty"`
					DeleteMarker        bool   `xml:"Code,omitempty"`
					DeleteMarkerVersion string `xml:"DeleteMarkerVersionId,omitempty"`
				}{
					Key:                 object.Key,
					Version:             object.Version,
					DeleteMarker:        result.DeleteMarker,
					DeleteMarkerVersion: deleteMarkerVersion,
				})
			}
		}
	}

	writeXML(h.logger, w, r, http.StatusOK, marshallable)
}
