package s2

import (
	"encoding/xml"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

const defaultMaxKeys int = 1000

// Contents is an individual file/object
type Contents struct {
	Key          string    `xml:"Key"`
	LastModified time.Time `xml:"LastModified"`
	ETag         string    `xml:"ETag"`
	Size         uint64    `xml:"Size"`
	StorageClass string    `xml:"StorageClass"`
	Owner        User      `xml:"Owner"`
}

// CommonPrefixes specifies a common prefix of S3 keys. This is akin to a
// directory.
type CommonPrefixes struct {
	Prefix string `xml:"Prefix"`
	Owner  User   `xml:"Owner"`
}

// BucketController is an interface that specifies bucket-level functionality.
type BucketController interface {
	// GetLocation gets the location of a bucket
	GetLocation(r *http.Request, bucket string) (location string, err error)

	// ListObjects lists objects within a bucket
	ListObjects(r *http.Request, bucket, prefix, marker, delimiter string, maxKeys int) (contents []Contents, commonPrefixes []CommonPrefixes, isTruncated bool, err error)

	// CreateBucket creates a bucket
	CreateBucket(r *http.Request, bucket string) error

	// DeleteBucket deletes a bucket
	DeleteBucket(r *http.Request, bucket string) error
}

// unimplementedBucketController defines a controller that returns
// `NotImplementedError` for all functionality
type unimplementedBucketController struct{}

func (c unimplementedBucketController) GetLocation(r *http.Request, bucket string) (location string, err error) {
	return "", NotImplementedError(r)
}

func (c unimplementedBucketController) ListObjects(r *http.Request, bucket, prefix, marker, delimiter string, maxKeys int) (contents []Contents, commonPrefixes []CommonPrefixes, isTruncated bool, err error) {
	return nil, nil, false, NotImplementedError(r)
}

func (c unimplementedBucketController) CreateBucket(r *http.Request, bucket string) error {
	return NotImplementedError(r)
}

func (c unimplementedBucketController) DeleteBucket(r *http.Request, bucket string) error {
	return NotImplementedError(r)
}

type bucketHandler struct {
	controller BucketController
	logger     *logrus.Entry
}

func (h bucketHandler) location(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]

	location, err := h.controller.GetLocation(r, bucket)
	if err != nil {
		WriteError(h.logger, w, r, err)
		return
	}

	writeXML(h.logger, w, r, http.StatusOK, struct {
		XMLName  xml.Name `xml:"LocationConstraint"`
		Location string   `xml:",innerxml"`
	}{
		Location: location,
	})
}

func (h bucketHandler) get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]

	maxKeys, err := intFormValue(r, "max-keys", 0, defaultMaxKeys, defaultMaxKeys)
	if err != nil {
		WriteError(h.logger, w, r, err)
		return
	}

	prefix := r.FormValue("prefix")
	marker := r.FormValue("marker")
	delimiter := r.FormValue("delimiter")

	contents, commonPrefixes, isTruncated, err := h.controller.ListObjects(r, bucket, prefix, marker, delimiter, maxKeys)
	if err != nil {
		WriteError(h.logger, w, r, err)
		return
	}

	for _, c := range contents {
		c.ETag = addETagQuotes(c.ETag)
	}

	result := struct {
		XMLName        xml.Name         `xml:"ListBucketResult"`
		Contents       []Contents       `xml:"Contents"`
		CommonPrefixes []CommonPrefixes `xml:"CommonPrefixes"`
		Delimiter      string           `xml:"Delimiter,omitempty"`
		IsTruncated    bool             `xml:"IsTruncated"`
		Marker         string           `xml:"Marker"`
		MaxKeys        int              `xml:"MaxKeys"`
		Name           string           `xml:"Name"`
		NextMarker     string           `xml:"NextMarker,omitempty"`
		Prefix         string           `xml:"Prefix"`
	}{
		Name:           bucket,
		Prefix:         prefix,
		Marker:         marker,
		Delimiter:      delimiter,
		MaxKeys:        maxKeys,
		IsTruncated:    isTruncated,
		Contents:       contents,
		CommonPrefixes: commonPrefixes,
	}

	if result.IsTruncated {
		if len(result.Contents) > 0 && len(result.CommonPrefixes) == 0 {
			result.NextMarker = result.Contents[len(result.Contents)-1].Key
		} else if len(result.Contents) == 0 && len(result.CommonPrefixes) > 0 {
			result.NextMarker = result.CommonPrefixes[len(result.CommonPrefixes)-1].Prefix
		} else if len(result.Contents) > 0 && len(result.CommonPrefixes) > 0 {
			lastContents := result.Contents[len(result.Contents)-1].Key
			lastCommonPrefixes := result.CommonPrefixes[len(result.CommonPrefixes)-1].Prefix

			if lastContents > lastCommonPrefixes {
				result.NextMarker = lastContents
			} else {
				result.NextMarker = lastCommonPrefixes
			}
		}
	}

	writeXML(h.logger, w, r, http.StatusOK, result)
}

func (h bucketHandler) put(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]

	if err := h.controller.CreateBucket(r, bucket); err != nil {
		WriteError(h.logger, w, r, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h bucketHandler) del(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]

	if err := h.controller.DeleteBucket(r, bucket); err != nil {
		WriteError(h.logger, w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
