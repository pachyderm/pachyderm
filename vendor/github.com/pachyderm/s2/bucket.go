package s2

import (
	"net/http"
	"strconv"
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
	List(r *http.Request, bucket string, result *ListBucketResult) error
	Create(r *http.Request, bucket string) error
	Delete(r *http.Request, bucket string) error
}

type UnimplementedBucketController struct{}

func (c UnimplementedBucketController) GetLocation(r *http.Request, bucket string, result *LocationConstraint) error {
	return NotImplementedError(r)
}

func (c UnimplementedBucketController) List(r *http.Request, bucket string, result *ListBucketResult) error {
	return NotImplementedError(r)
}

func (c UnimplementedBucketController) Create(r *http.Request, bucket string) error {
	return NotImplementedError(r)
}

func (c UnimplementedBucketController) Delete(r *http.Request, bucket string) error {
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
		writeError(h.logger, r, w, err)
		return
	}

	writeXML(h.logger, w, r, http.StatusOK, result)
}

func (h bucketHandler) get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]

	maxKeys := defaultMaxKeys
	maxKeysStr := r.FormValue("max-keys")
	if maxKeysStr != "" {
		i, err := strconv.Atoi(maxKeysStr)
		if err != nil || i < 0 || i > defaultMaxKeys {
			InvalidArgument(r).Write(h.logger, w)
			return
		}
		maxKeys = i
	}

	result := &ListBucketResult{
		Name:        bucket,
		Prefix:      r.FormValue("prefix"),
		Marker:      r.FormValue("marker"),
		Delimiter:   r.FormValue("delimiter"),
		MaxKeys:     maxKeys,
		IsTruncated: false,
	}

	if err := h.controller.List(r, bucket, result); err != nil {
		writeError(h.logger, r, w, err)
		return
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

	if err := h.controller.Create(r, bucket); err != nil {
		writeError(h.logger, r, w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h bucketHandler) del(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]

	if err := h.controller.Delete(r, bucket); err != nil {
		writeError(h.logger, r, w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
