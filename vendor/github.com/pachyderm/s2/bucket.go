package s2

import (
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

const defaultMaxKeys int = 1000

// LocationConstraint is an XML-encodable location specification of a bucket
type LocationConstraint struct {
	Location string `xml:",innerxml"`
}

// ListBucketResult is an XML-encodable listing of files/objects in a
// repo/bucket
type ListBucketResult struct {
	Contents       []Contents       `xml:"Contents"`
	CommonPrefixes []CommonPrefixes `xml:"CommonPrefixes"`
	Delimiter      string           `xml:"Delimiter,omitempty"`
	IsTruncated    bool             `xml:"IsTruncated"`
	Marker         string           `xml:"Marker"`
	MaxKeys        int              `xml:"MaxKeys"`
	Name           string           `xml:"Name"`
	NextMarker     string           `xml:"NextMarker,omitempty"`
	Prefix         string           `xml:"Prefix"`
}

func (r *ListBucketResult) IsFull() bool {
	return len(r.Contents)+len(r.CommonPrefixes) >= r.MaxKeys
}

// Contents is an individual file/object
type Contents struct {
	Key          string    `xml:"Key"`
	LastModified time.Time `xml:"LastModified"`
	ETag         string    `xml:"ETag"`
	Size         uint64    `xml:"Size"`
	StorageClass string    `xml:"StorageClass"`
	Owner        User      `xml:"Owner"`
}

// CommonPrefixes is an individual PFS directory
type CommonPrefixes struct {
	Prefix string `xml:"Prefix"`
	Owner  User   `xml:"Owner"`
}

type BucketController interface {
	GetLocation(r *http.Request, bucket string, result *LocationConstraint) error
	ListObjects(r *http.Request, bucket string, result *ListBucketResult) error
	CreateBucket(r *http.Request, bucket string) error
	DeleteBucket(r *http.Request, bucket string) error
}

type UnimplementedBucketController struct{}

func (c UnimplementedBucketController) GetLocation(r *http.Request, bucket string, result *LocationConstraint) error {
	return NotImplementedError(r)
}

func (c UnimplementedBucketController) ListObjects(r *http.Request, bucket string, result *ListBucketResult) error {
	return NotImplementedError(r)
}

func (c UnimplementedBucketController) CreateBucket(r *http.Request, bucket string) error {
	return NotImplementedError(r)
}

func (c UnimplementedBucketController) DeleteBucket(r *http.Request, bucket string) error {
	return NotImplementedError(r)
}

type bucketHandler struct {
	controller BucketController
	logger     *logrus.Entry
}

func (h bucketHandler) location(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]

	result := &LocationConstraint{}

	if err := h.controller.GetLocation(r, bucket, result); err != nil {
		writeError(h.logger, w, r, err)
		return
	}

	writeXML(h.logger, w, r, http.StatusOK, result)
}

func (h bucketHandler) get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]

	maxKeys, err := intFormValue(r, "max-keys", 0, defaultMaxKeys, defaultMaxKeys)
	if err != nil {
		writeError(h.logger, w, r, err)
		return
	}

	result := &ListBucketResult{
		Name:        bucket,
		Prefix:      r.FormValue("prefix"),
		Marker:      r.FormValue("marker"),
		Delimiter:   r.FormValue("delimiter"),
		MaxKeys:     maxKeys,
		IsTruncated: false,
	}

	if err := h.controller.ListObjects(r, bucket, result); err != nil {
		writeError(h.logger, w, r, err)
		return
	}

	for _, contents := range result.Contents {
		contents.ETag = addETagQuotes(contents.ETag)
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
		writeError(h.logger, w, r, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h bucketHandler) del(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]

	if err := h.controller.DeleteBucket(r, bucket); err != nil {
		writeError(h.logger, w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
