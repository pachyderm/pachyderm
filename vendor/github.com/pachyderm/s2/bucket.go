package s2

import (
	"encoding/xml"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

const (
	defaultMaxKeys int = 1000

	VersioningDisabled  string = ""
	VersioningSuspended string = "Suspended"
	VersioningEnabled   string = "Enabled"
)

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

// DeleteMarker specifies an object that has been deleted from a
// versioning-enabled bucket.
type DeleteMarker struct {
	Key          string    `xml:"Key"`
	Version      string    `xml:"VersionId"`
	IsLatest     bool      `xml:"IsLatest"`
	LastModified time.Time `xml:"LastModified"`
	Owner        User      `xml:"Owner"`
}

// Version specifies a specific version of an object in a
// versioning-enabled bucket.
type Version struct {
	Key          string    `xml:"Key"`
	Version      string    `xml:"VersionId"`
	IsLatest     bool      `xml:"IsLatest"`
	LastModified time.Time `xml:"LastModified"`
	ETag         string    `xml:"ETag"`
	Size         uint64    `xml:"Size"`
	StorageClass string    `xml:"StorageClass"`
	Owner        User      `xml:"Owner"`
}

// BucketController is an interface that specifies bucket-level functionality.
type BucketController interface {
	// GetLocation gets the location of a bucket
	GetLocation(r *http.Request, bucket string) (location string, err error)

	// ListObjects lists objects within a bucket
	ListObjects(r *http.Request, bucket, prefix, marker, delimiter string, maxKeys int) (contents []Contents, commonPrefixes []CommonPrefixes, isTruncated bool, err error)

	// ListOVersionedbjects lists objects' versions within a bucket
	ListVersionedObjects(r *http.Request, bucket, prefix, keyMarker, versionMarker string, delimiter string, maxKeys int) (versions []Version, deleteMarkers []DeleteMarker, isTruncated bool, err error)

	// CreateBucket creates a bucket
	CreateBucket(r *http.Request, bucket string) error

	// DeleteBucket deletes a bucket
	DeleteBucket(r *http.Request, bucket string) error

	// GetBucketVersioning gets the state of versioning on the given bucket
	GetBucketVersioning(r *http.Request, bucket string) (status string, err error)

	// SetBucketVersioning sets the state of versioning on the given bucket
	SetBucketVersioning(r *http.Request, bucket, status string) error
}

// unimplementedBucketController defines a controller that returns
// `NotImplementedError` for all functionality
type unimplementedBucketController struct{}

func (c unimplementedBucketController) GetLocation(r *http.Request, bucket string) (location string, err error) {
	return "", NotImplementedError(r)
}

func (c unimplementedBucketController) ListObjects(r *http.Request, bucket, prefix, marker, delimiter string, maxKeys int) (contents []Contents, commonPrefixes []CommonPrefixes, isTruncated bool, err error) {
	err = NotImplementedError(r)
	return
}

func (c unimplementedBucketController) ListVersionedObjects(r *http.Request, bucket, prefix, keyMarker, versionMarker string, delimiter string, maxKeys int) (versions []Version, deleteMarkers []DeleteMarker, isTruncated bool, err error) {
	err = NotImplementedError(r)
	return
}

func (c unimplementedBucketController) CreateBucket(r *http.Request, bucket string) error {
	return NotImplementedError(r)
}

func (c unimplementedBucketController) DeleteBucket(r *http.Request, bucket string) error {
	return NotImplementedError(r)
}

func (c unimplementedBucketController) GetBucketVersioning(r *http.Request, bucket string) (status string, err error) {
	return "", NotImplementedError(r)
}

func (c unimplementedBucketController) SetBucketVersioning(r *http.Request, bucket, status string) error {
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

func (h bucketHandler) versioning(w http.ResponseWriter, r *http.Request) {
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

func (h bucketHandler) setVersioning(w http.ResponseWriter, r *http.Request) {
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

	err := h.controller.SetBucketVersioning(r, bucket, payload.Status)
	if err != nil {
		WriteError(h.logger, w, r, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h bucketHandler) listVersions(w http.ResponseWriter, r *http.Request) {
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

	versions, deleteMarkers, isTruncated, err := h.controller.ListVersionedObjects(r, bucket, prefix, keyMarker, versionIDMarker, delimiter, maxKeys)
	if err != nil {
		WriteError(h.logger, w, r, err)
		return
	}

	for _, v := range versions {
		v.ETag = addETagQuotes(v.ETag)
	}

	result := struct {
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
		IsTruncated:     isTruncated,
		KeyMarker:       keyMarker,
		MaxKeys:         maxKeys,
		Name:            bucket,
		VersionIDMarker: versionIDMarker,
		Prefix:          prefix,
		Versions:        versions,
		DeleteMarkers:   deleteMarkers,
	}

	if result.IsTruncated {
		if len(result.Versions) > 0 && len(result.DeleteMarkers) == 0 {
			last := result.Versions[len(result.Versions)-1]
			result.NextKeyMarker = last.Key
			result.NextVersionIDMarker = last.Version
		} else if len(result.Versions) == 0 && len(result.DeleteMarkers) > 0 {
			last := result.DeleteMarkers[len(result.DeleteMarkers)-1]
			result.NextKeyMarker = last.Key
			result.NextVersionIDMarker = last.Version
		} else if len(result.Versions) > 0 && len(result.DeleteMarkers) > 0 {
			lastVersion := result.Versions[len(result.Versions)-1]
			lastDeleteMarker := result.DeleteMarkers[len(result.DeleteMarkers)-1]
			if lastVersion.Key == lastDeleteMarker.Key {
				result.NextKeyMarker = lastVersion.Key
				if lastVersion.Version > lastDeleteMarker.Version {
					result.NextVersionIDMarker = lastVersion.Version
				} else {
					result.NextVersionIDMarker = lastDeleteMarker.Version
				}
			} else if lastVersion.Key > lastDeleteMarker.Key {
				result.NextKeyMarker = lastVersion.Key
				result.NextVersionIDMarker = lastVersion.Version
			} else {
				result.NextKeyMarker = lastDeleteMarker.Key
				result.NextVersionIDMarker = lastDeleteMarker.Version
			}
		}
	}

	writeXML(h.logger, w, r, http.StatusOK, result)
}
