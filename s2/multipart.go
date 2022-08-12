package s2

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

const (
	// defaultMaxUploads specifies the maximum number of uploads returned in
	// multipart upload listings by default
	defaultMaxUploads = 1000
	// defaultMaxParts specifies the maximum number of parts returned in
	// multipart upload part listings by default
	defaultMaxParts = 1000
	// maxPartsAllowed specifies the maximum number of parts that can be
	// uploaded in a multipart upload
	maxPartsAllowed = 10000
	// completeMultipartPing is how long to wait before sending whitespace in
	// a complete multipart response (to ensure the connection doesn't close.)
	completeMultipartPing = 10 * time.Second
)

// Upload is an XML marshallable representation of an in-progress multipart
// upload
type Upload struct {
	// Key specifies the object key
	Key string `xml:"Key"`
	// UploadID is an ID identifying the multipart upload
	UploadID string `xml:"UploadId"`
	// Initiator is the user that initiated the multipart upload
	Initiator User `xml:"Initiator"`
	// Owner specifies the owner of the object
	Owner User `xml:"Owner"`
	// StorageClass specifies the storage class used for the object
	StorageClass string `xml:"StorageClass"`
	// Initiated is a timestamp specifying when the multipart upload was
	// started
	Initiated time.Time `xml:"Initiated"`
}

// Part is an XML marshallable representation of a chunk of an in-progress
// multipart upload
type Part struct {
	// PartNumber is the index of the part
	PartNumber int `xml:"PartNumber"`
	// ETag is a hex encoding of the hash of the object contents, with or
	// without surrounding quotes.
	ETag string `xml:"ETag"`
}

// ListMultipartResult is a response from a ListMultipart call
type ListMultipartResult struct {
	// IsTruncated specifies whether this is the end of the list or not
	IsTruncated bool
	// Uploads are the list of uploads returned
	Uploads []*Upload
}

// CompleteMultipartResult is a response from a CompleteMultipart call
type CompleteMultipartResult struct {
	// Location is the location of the newly uploaded object
	Location string
	// ETag is a hex encoding of the hash of the object contents, with or
	// without surrounding quotes.
	ETag string
	// Version is the version of the object, or an empty string if versioning
	// is not enabled or supported.
	Version string
}

// ListMultipartChunksResult is a response from a ListMultipartChunks call
type ListMultipartChunksResult struct {
	// Initiator is the user that initiated the multipart upload
	Initiator *User
	// Owner specifies the owner of the object
	Owner *User
	// StorageClass specifies the storage class used for the object
	StorageClass string
	// IsTruncated specifies whether this is the end of the list or not
	IsTruncated bool
	// Parts are the list of parts returned
	Parts []*Part
}

// MultipartController is an interface that specifies multipart-related
// functionality
type MultipartController interface {
	// ListMultipart lists in-progress multipart uploads in a bucket
	ListMultipart(r *http.Request, bucket, keyMarker, uploadIDMarker string, maxUploads int) (*ListMultipartResult, error)
	// InitMultipart initializes a new multipart upload
	InitMultipart(r *http.Request, bucket, key string) (string, error)
	// AbortMultipart aborts an in-progress multipart upload
	AbortMultipart(r *http.Request, bucket, key, uploadID string) error
	// CompleteMultipart finishes a multipart upload
	CompleteMultipart(r *http.Request, bucket, key, uploadID string, parts []*Part) (*CompleteMultipartResult, error)
	// ListMultipartChunks lists the constituent chunks of an in-progress
	// multipart upload
	ListMultipartChunks(r *http.Request, bucket, key, uploadID string, partNumberMarker, maxParts int) (*ListMultipartChunksResult, error)
	// UploadMultipartChunk uploads a chunk of an in-progress multipart upload
	UploadMultipartChunk(r *http.Request, bucket, key, uploadID string, partNumber int, reader io.Reader) (string, error)
}

// unimplementedMultipartController defines a controller that returns
// `NotImplementedError` for all functionality
type unimplementedMultipartController struct{}

func (c unimplementedMultipartController) ListMultipart(r *http.Request, bucket, keyMarker, uploadIDMarker string, maxUploads int) (*ListMultipartResult, error) {
	return nil, NotImplementedError(r)
}

func (c unimplementedMultipartController) InitMultipart(r *http.Request, bucket, key string) (string, error) {
	return "", NotImplementedError(r)
}

func (c unimplementedMultipartController) AbortMultipart(r *http.Request, bucket, key, uploadID string) error {
	return NotImplementedError(r)
}

func (c unimplementedMultipartController) CompleteMultipart(r *http.Request, bucket, key, uploadID string, parts []*Part) (*CompleteMultipartResult, error) {
	return nil, NotImplementedError(r)
}

func (c unimplementedMultipartController) ListMultipartChunks(r *http.Request, bucket, key, uploadID string, partNumberMarker, maxcParts int) (*ListMultipartChunksResult, error) {
	return nil, NotImplementedError(r)
}

func (c unimplementedMultipartController) UploadMultipartChunk(r *http.Request, bucket, key, uploadID string, partNumber int, reader io.Reader) (string, error) {
	return "", NotImplementedError(r)
}

type multipartHandler struct {
	controller MultipartController
	logger     *logrus.Entry
}

func (h *multipartHandler) list(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]

	keyMarker := r.FormValue("key-marker")
	uploadIDMarker := r.FormValue("upload-id-marker")
	if keyMarker == "" {
		uploadIDMarker = ""
	}

	maxUploads, err := intFormValue(r, "max-uploads", 0, defaultMaxUploads, defaultMaxUploads)
	if err != nil {
		WriteError(h.logger, w, r, err)
		return
	}

	result, err := h.controller.ListMultipart(r, bucket, keyMarker, uploadIDMarker, maxUploads)
	if err != nil {
		WriteError(h.logger, w, r, err)
		return
	}

	// some clients (e.g. minio-python) can't handle sub-seconds in datetime
	// output
	for _, upload := range result.Uploads {
		upload.Initiated = upload.Initiated.UTC().Round(time.Second)
	}

	marshallable := struct {
		XMLName            xml.Name  `xml:"http://s3.amazonaws.com/doc/2006-03-01/ ListMultipartUploadsResult"`
		Bucket             string    `xml:"Bucket"`
		KeyMarker          string    `xml:"KeyMarker"`
		UploadIDMarker     string    `xml:"UploadIdMarker"`
		NextKeyMarker      string    `xml:"NextKeyMarker"`
		NextUploadIDMarker string    `xml:"NextUploadIdMarker"`
		MaxUploads         int       `xml:"MaxUploads"`
		IsTruncated        bool      `xml:"IsTruncated"`
		Uploads            []*Upload `xml:"Upload"`
	}{
		Bucket:         bucket,
		KeyMarker:      keyMarker,
		UploadIDMarker: uploadIDMarker,
		MaxUploads:     maxUploads,
		IsTruncated:    result.IsTruncated,
		Uploads:        result.Uploads,
	}

	if marshallable.IsTruncated {
		highKey := ""
		highUploadID := ""

		for _, upload := range marshallable.Uploads {
			if upload.Key > highKey {
				highKey = upload.Key
			}
			if upload.UploadID > highUploadID {
				highUploadID = upload.UploadID
			}
		}

		marshallable.NextKeyMarker = highKey
		marshallable.NextUploadIDMarker = highUploadID
	}

	writeXML(h.logger, w, r, http.StatusOK, marshallable)
}

func (h *multipartHandler) listChunks(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]
	key := vars["key"]

	maxParts, err := intFormValue(r, "max-parts", 0, defaultMaxParts, defaultMaxParts)
	if err != nil {
		WriteError(h.logger, w, r, err)
		return
	}

	partNumberMarker, err := intFormValue(r, "part-number-marker", 0, maxPartsAllowed, 0)
	if err != nil {
		WriteError(h.logger, w, r, err)
		return
	}

	uploadID := r.FormValue("uploadId")

	result, err := h.controller.ListMultipartChunks(r, bucket, key, uploadID, partNumberMarker, maxParts)
	if err != nil {
		WriteError(h.logger, w, r, err)
		return
	}

	marshallable := struct {
		XMLName              xml.Name `xml:"http://s3.amazonaws.com/doc/2006-03-01/ ListPartsResult"`
		Bucket               string   `xml:"Bucket"`
		Key                  string   `xml:"Key"`
		UploadID             string   `xml:"UploadId"`
		Initiator            *User    `xml:"Initiator"`
		Owner                *User    `xml:"Owner"`
		StorageClass         string   `xml:"StorageClass"`
		PartNumberMarker     int      `xml:"PartNumberMarker"`
		NextPartNumberMarker int      `xml:"NextPartNumberMarker"`
		MaxParts             int      `xml:"MaxParts"`
		IsTruncated          bool     `xml:"IsTruncated"`
		Parts                []*Part  `xml:"Part"`
	}{
		Bucket:           bucket,
		Key:              key,
		UploadID:         uploadID,
		PartNumberMarker: partNumberMarker,
		MaxParts:         maxParts,
		Initiator:        result.Initiator,
		Owner:            result.Owner,
		StorageClass:     result.StorageClass,
		IsTruncated:      result.IsTruncated,
		Parts:            result.Parts,
	}

	if marshallable.IsTruncated {
		high := 0

		for _, part := range marshallable.Parts {
			if part.PartNumber > high {
				high = part.PartNumber
			}
		}

		marshallable.NextPartNumberMarker = high
	}

	writeXML(h.logger, w, r, http.StatusOK, marshallable)
}

func (h *multipartHandler) init(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]
	key := vars["key"]

	uploadID, err := h.controller.InitMultipart(r, bucket, key)
	if err != nil {
		WriteError(h.logger, w, r, err)
		return
	}

	marshallable := struct {
		XMLName  xml.Name `xml:"http://s3.amazonaws.com/doc/2006-03-01/ InitiateMultipartUploadResult"`
		Bucket   string   `xml:"Bucket"`
		Key      string   `xml:"Key"`
		UploadID string   `xml:"UploadId"`
	}{
		Bucket:   bucket,
		Key:      key,
		UploadID: uploadID,
	}

	writeXML(h.logger, w, r, http.StatusOK, marshallable)
}

func (h *multipartHandler) complete(w http.ResponseWriter, r *http.Request) {
	if err := requireContentLength(r); err != nil {
		WriteError(h.logger, w, r, err)
		return
	}

	vars := mux.Vars(r)
	bucket := vars["bucket"]
	key := vars["key"]

	uploadID := r.FormValue("uploadId")

	payload := struct {
		XMLName xml.Name `xml:"CompleteMultipartUpload"`
		Parts   []*Part  `xml:"Part"`
	}{}
	if err := readXMLBody(r, &payload); err != nil {
		WriteError(h.logger, w, r, err)
		return
	}

	// verify that there's at least part, and all parts are in ascending order
	isSorted := sort.SliceIsSorted(payload.Parts, func(i, j int) bool {
		return payload.Parts[i].PartNumber < payload.Parts[j].PartNumber
	})
	if len(payload.Parts) == 0 || !isSorted {
		WriteError(h.logger, w, r, InvalidPartOrderError(w, r))
		return
	}

	for _, part := range payload.Parts {
		part.ETag = addETagQuotes(part.ETag)
	}

	ch := make(chan struct {
		result *CompleteMultipartResult
		err    error
	})

	go func() {
		result, err := h.controller.CompleteMultipart(r, bucket, key, uploadID, payload.Parts)
		ch <- struct {
			result *CompleteMultipartResult
			err    error
		}{
			result: result,
			err:    err,
		}
	}()

	streaming := false

	for {
		select {
		case value := <-ch:
			if value.err != nil {
				s3Error := newGenericError(r, value.err)

				if streaming {
					writeXMLBody(h.logger, w, s3Error)
				} else {
					WriteError(h.logger, w, r, s3Error)
				}
			} else {
				marshallable := struct {
					XMLName  xml.Name `xml:"http://s3.amazonaws.com/doc/2006-03-01/ CompleteMultipartUploadResult"`
					Location string   `xml:"Location"`
					Bucket   string   `xml:"Bucket"`
					Key      string   `xml:"Key"`
					ETag     string   `xml:"ETag"`
				}{
					Bucket:   bucket,
					Key:      key,
					Location: value.result.Location,
					ETag:     addETagQuotes(value.result.ETag),
				}

				if value.result.Version != "" {
					w.Header().Set("x-amz-version-id", value.result.Version)
				}

				if streaming {
					writeXMLBody(h.logger, w, marshallable)
				} else {
					writeXML(h.logger, w, r, http.StatusOK, marshallable)
				}
			}
			return
		case <-time.After(completeMultipartPing):
			if !streaming {
				streaming = true
				writeXMLPrelude(w, r, http.StatusOK)
			} else {
				fmt.Fprint(w, " ")
			}
		}
	}
}

func (h *multipartHandler) put(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]
	key := vars["key"]

	uploadID := r.FormValue("uploadId")
	partNumber, err := intFormValue(r, "partNumber", 0, maxPartsAllowed, 0)
	if err != nil {
		WriteError(h.logger, w, r, err)
		return
	}

	etag, err := h.controller.UploadMultipartChunk(r, bucket, key, uploadID, partNumber, r.Body)
	if err != nil {
		WriteError(h.logger, w, r, err)
		return
	}

	if etag != "" {
		w.Header().Set("ETag", addETagQuotes(etag))
	}

	w.WriteHeader(http.StatusOK)
}

func (h *multipartHandler) del(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]
	key := vars["key"]

	uploadID := r.FormValue("uploadId")

	if err := h.controller.AbortMultipart(r, bucket, key, uploadID); err != nil {
		WriteError(h.logger, w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
