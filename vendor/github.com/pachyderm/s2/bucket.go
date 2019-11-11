package s2

import (
	"encoding/xml"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

const (
	// defaultMaxKeys specifies the maximum number of keys returned in object
	// listings by default
	defaultMaxKeys int = 1000
	// VersioningDisabled specifies that versioning is not enabled on a bucket
	VersioningDisabled string = ""
	// VersioningDisabled specifies that versioning is suspended on a bucket
	VersioningSuspended string = "Suspended"
	// VersioningDisabled specifies that versioning is enabled on a bucket
	VersioningEnabled string = "Enabled"
)

// Contents is an individual file/object
type Contents struct {
	// Key specifies the object key
	Key string `xml:"Key"`
	// LastModified specifies when the object was last modified
	LastModified time.Time `xml:"LastModified"`
	// ETag is a hex encoding of the hash of the object contents, with or
	// without surrounding quotes.
	ETag string `xml:"ETag"`
	// Size specifies the size of the object
	Size uint64 `xml:"Size"`
	// StorageClass specifies the storage class used for the object
	StorageClass string `xml:"StorageClass"`
	// Owner specifies the owner of the object
	Owner User `xml:"Owner"`
}

// CommonPrefixes specifies a common prefix of S3 keys. This is akin to a
// directory.
type CommonPrefixes struct {
	// Prefix specifies the common prefix value.
	Prefix string `xml:"Prefix"`
	// Owner specifies the owner of the object
	Owner User `xml:"Owner"`
}

// DeleteMarker specifies an object that has been deleted from a
// versioning-enabled bucket.
type DeleteMarker struct {
	// Key specifies the object key
	Key string `xml:"Key"`
	// Version is the version of the object, or an empty string if versioning
	// is not enabled or supported.
	Version string `xml:"VersionId"`
	// IsLatest specifies whether this is the latest version of the object.
	IsLatest bool `xml:"IsLatest"`
	// LastModified specifies when the object was last modified
	LastModified time.Time `xml:"LastModified"`
	// Owner specifies the owner of the object
	Owner User `xml:"Owner"`
}

// Version specifies a specific version of an object in a
// versioning-enabled bucket.
type Version struct {
	// Key specifies the object key
	Key string `xml:"Key"`
	// Version is the version of the object, or an empty string if versioning
	// is not enabled or supported.
	Version string `xml:"VersionId"`
	// IsLatest specifies whether this is the latest version of the object.
	IsLatest bool `xml:"IsLatest"`
	// LastModified specifies when the object was last modified
	LastModified time.Time `xml:"LastModified"`
	// ETag is a hex encoding of the hash of the object contents, with or
	// without surrounding quotes.
	ETag string `xml:"ETag"`
	// Size specifies the size of the object
	Size uint64 `xml:"Size"`
	// StorageClass specifies the storage class used for the object
	StorageClass string `xml:"StorageClass"`
	// Owner specifies the owner of the object
	Owner User `xml:"Owner"`
}

// ListObjectsResult is a response from a ListObjects call
type ListObjectsResult struct {
	// Contents are the list of objects returned
	Contents []Contents
	// CommonPrefixes are the list of common prefixes returned
	CommonPrefixes []CommonPrefixes
	// IsTruncated specifies whether this is the end of the list or not
	IsTruncated bool
}

// ListObjectVersionsResult is a response from a ListObjectVersions call
type ListObjectVersionsResult struct {
	// Versions are the list of versions returned
	Versions []Version
	// DeleteMarkers are the list of delete markers returned
	DeleteMarkers []DeleteMarker
	// IsTruncated specifies whether this is the end of the list or not
	IsTruncated bool
}

// BucketController is an interface that specifies bucket-level functionality.
type BucketController interface {
	// GetLocation gets the location of a bucket
	GetLocation(r *http.Request, bucket string) (string, error)

	// ListObjects lists objects within a bucket
	ListObjects(r *http.Request, bucket, prefix, marker, delimiter string, maxKeys int) (*ListObjectsResult, error)

	// ListObjectVersions lists objects' versions within a bucket
	ListObjectVersions(r *http.Request, bucket, prefix, keyMarker, versionMarker string, delimiter string, maxKeys int) (*ListObjectVersionsResult, error)

	// CreateBucket creates a bucket
	CreateBucket(r *http.Request, bucket string) error

	// DeleteBucket deletes a bucket
	DeleteBucket(r *http.Request, bucket string) error

	// GetBucketVersioning gets the state of versioning on the given bucket
	GetBucketVersioning(r *http.Request, bucket string) (string, error)

	// SetBucketVersioning sets the state of versioning on the given bucket
	SetBucketVersioning(r *http.Request, bucket, status string) error
}

// unimplementedBucketController defines a controller that returns
// `NotImplementedError` for all functionality
type unimplementedBucketController struct{}

func (c unimplementedBucketController) GetLocation(r *http.Request, bucket string) (string, error) {
	return "", NotImplementedError(r)
}

func (c unimplementedBucketController) ListObjects(r *http.Request, bucket, prefix, marker, delimiter string, maxKeys int) (*ListObjectsResult, error) {
	return nil, NotImplementedError(r)
}

func (c unimplementedBucketController) ListObjectVersions(r *http.Request, bucket, prefix, keyMarker, versionMarker string, delimiter string, maxKeys int) (*ListObjectVersionsResult, error) {
	return nil, NotImplementedError(r)
}

func (c unimplementedBucketController) CreateBucket(r *http.Request, bucket string) error {
	return NotImplementedError(r)
}

func (c unimplementedBucketController) DeleteBucket(r *http.Request, bucket string) error {
	return NotImplementedError(r)
}

func (c unimplementedBucketController) GetBucketVersioning(r *http.Request, bucket string) (string, error) {
	return "", NotImplementedError(r)
}

func (c unimplementedBucketController) SetBucketVersioning(r *http.Request, bucket, status string) error {
	return NotImplementedError(r)
}

type bucketHandler struct {
	controller BucketController
	logger     *logrus.Entry
}

func (h *bucketHandler) location(w http.ResponseWriter, r *http.Request) {
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

func (h *bucketHandler) get(w http.ResponseWriter, r *http.Request) {
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

	result, err := h.controller.ListObjects(r, bucket, prefix, marker, delimiter, maxKeys)
	if err != nil {
		WriteError(h.logger, w, r, err)
		return
	}

	for _, c := range result.Contents {
		c.ETag = addETagQuotes(c.ETag)
	}

	marshallable := struct {
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
		IsTruncated:    result.IsTruncated,
		Contents:       result.Contents,
		CommonPrefixes: result.CommonPrefixes,
	}

	if marshallable.IsTruncated {
		high := ""

		for _, contents := range marshallable.Contents {
			if contents.Key > high {
				high = contents.Key
			}
		}
		for _, commonPrefix := range marshallable.CommonPrefixes {
			if commonPrefix.Prefix > high {
				high = commonPrefix.Prefix
			}
		}

		marshallable.NextMarker = high
	}

	writeXML(h.logger, w, r, http.StatusOK, marshallable)
}

func (h *bucketHandler) put(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]

	if err := h.controller.CreateBucket(r, bucket); err != nil {
		WriteError(h.logger, w, r, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *bucketHandler) del(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]

	if err := h.controller.DeleteBucket(r, bucket); err != nil {
		WriteError(h.logger, w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *bucketHandler) versioning(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]

	status, err := h.controller.GetBucketVersioning(r, bucket)
	if err != nil {
		WriteError(h.logger, w, r, err)
		return
	}

	result := struct {
		XMLName xml.Name `xml:"VersioningConfiguration"`
		Status  string   `xml:"Status,omitempty"`
	}{
		Status: status,
	}

	writeXML(h.logger, w, r, http.StatusOK, result)
}

func (h *bucketHandler) setVersioning(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]

	payload := struct {
		XMLName xml.Name `xml:"VersioningConfiguration"`
		Status  string   `xml:"Status"`
	}{}
	if err := readXMLBody(r, &payload); err != nil {
		WriteError(h.logger, w, r, err)
		return
	}

	if payload.Status != VersioningDisabled && payload.Status != VersioningSuspended && payload.Status != VersioningEnabled {
		WriteError(h.logger, w, r, IllegalVersioningConfigurationError(r))
		return
	}

	err := h.controller.SetBucketVersioning(r, bucket, payload.Status)
	if err != nil {
		WriteError(h.logger, w, r, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *bucketHandler) listVersions(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]

	maxKeys, err := intFormValue(r, "max-keys", 0, defaultMaxKeys, defaultMaxKeys)
	if err != nil {
		WriteError(h.logger, w, r, err)
		return
	}

	prefix := r.FormValue("prefix")
	keyMarker := r.FormValue("key-marker")
	versionIDMarker := r.FormValue("version-id-marker")
	delimiter := r.FormValue("delimiter")

	result, err := h.controller.ListObjectVersions(r, bucket, prefix, keyMarker, versionIDMarker, delimiter, maxKeys)
	if err != nil {
		WriteError(h.logger, w, r, err)
		return
	}

	for _, v := range result.Versions {
		v.ETag = addETagQuotes(v.ETag)
	}

	marshallable := struct {
		XMLName             xml.Name       `xml:"ListVersionsResult"`
		Delimiter           string         `xml:"Delimiter,omitempty"`
		IsTruncated         bool           `xml:"IsTruncated"`
		KeyMarker           string         `xml:"KeyMarker"`
		NextKeyMarker       string         `xml:"NextKeyMarker,omitempty"`
		MaxKeys             int            `xml:"MaxKeys"`
		Name                string         `xml:"Name"`
		VersionIDMarker     string         `xml:"VersionIdKeyMarker"`
		NextVersionIDMarker string         `xml:"NextVersionIdKeyMarker,omitempty"`
		Prefix              string         `xml:"Prefix"`
		Versions            []Version      `xml:"Version"`
		DeleteMarkers       []DeleteMarker `xml:"DeleteMarker"`
	}{
		IsTruncated:     result.IsTruncated,
		KeyMarker:       keyMarker,
		MaxKeys:         maxKeys,
		Name:            bucket,
		VersionIDMarker: versionIDMarker,
		Prefix:          prefix,
		Versions:        result.Versions,
		DeleteMarkers:   result.DeleteMarkers,
	}

	if marshallable.IsTruncated {
		highKey := ""
		highVersion := ""

		for _, version := range marshallable.Versions {
			if version.Key > highKey {
				highKey = version.Key
			}
			if version.Version > highVersion {
				highVersion = version.Version
			}
		}
		for _, deleteMarker := range marshallable.DeleteMarkers {
			if deleteMarker.Key > highKey {
				highKey = deleteMarker.Key
			}
			if deleteMarker.Version > highVersion {
				highVersion = deleteMarker.Version
			}
		}

		marshallable.NextKeyMarker = highKey
		marshallable.NextVersionIDMarker = highVersion
	}

	writeXML(h.logger, w, r, http.StatusOK, marshallable)
}
