package s2

import (
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sort"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

const (
	defaultMaxUploads     = 1000
	defaultMaxParts       = 1000
	maxPartsAllowed       = 10000
	completeMultipartPing = 10 * time.Second
)

// Upload is an XML marshalable representation of an in-progress multipart
// upload
type Upload struct {
	Key          string    `xml:"Key"`
	UploadID     string    `xml:"UploadId"`
	Initiator    User      `xml:"Initiator"`
	Owner        User      `xml:"Owner"`
	StorageClass string    `xml:"StorageClass"`
	Initiated    time.Time `xml:"Initiated"`
}

// Part is an XML marshalable representation of a chunk of an in-progress
// multipart upload
type Part struct {
	PartNumber int    `xml:"PartNumber"`
	ETag       string `xml:"ETag"`
}

// MultipartController is an interface that specifies multipart-related
// functionality
type MultipartController interface {
	// ListMultipart lists in-progress multipart uploads in a bucket
	ListMultipart(r *http.Request, bucket, keyMarker, uploadIDMarker string, maxUploads int) (isTruncated bool, uploads []Upload, err error)
	// InitMultipart initializes a new multipart upload
	InitMultipart(r *http.Request, bucket, key string) (uploadID string, err error)
	// AbortMultipart aborts an in-progress multipart upload
	AbortMultipart(r *http.Request, bucket, key, uploadID string) error
	// CompleteMultipart finishes a multipart upload
	CompleteMultipart(r *http.Request, bucket, key, uploadID string, parts []Part) (location, etag string, err error)
	// ListMultipartChunks lists the constituent chunks of an in-progress
	// multipart upload
	ListMultipartChunks(r *http.Request, bucket, key, uploadID string, partNumberMarker, maxParts int) (initiator, owner *User, storageClass string, isTruncated bool, parts []Part, err error)
	// UploadMultipartChunk uploads a chunk of an in-progress multipart upload
	UploadMultipartChunk(r *http.Request, bucket, key, uploadID string, partNumber int, reader io.Reader) (etag string, err error)
	// DeleteMultipartChunk deletes a chunk of an in-progress multipart
	// upload. This is not exposed directly through the API, but rather to
	// remove an uploaded multipart chunk if an issue occurred after it
	// finished uploading.
	DeleteMultipartChunk(r *http.Request, bucket, key, uploadID string, partNumber int) error
}

// unimplementedMultipartController defines a controller that returns
// `NotImplementedError` for all functionality
type unimplementedMultipartController struct{}

func (c unimplementedMultipartController) ListMultipart(r *http.Request, bucket, keyMarker, uploadIDMarker string, maxUploads int) (isTruncated bool, uploads []Upload, err error) {
	return false, nil, NotImplementedError(r)
}

func (c unimplementedMultipartController) InitMultipart(r *http.Request, bucket, key string) (string, error) {
	return "", NotImplementedError(r)
}

func (c unimplementedMultipartController) AbortMultipart(r *http.Request, bucket, key, uploadID string) error {
	return NotImplementedError(r)
}

func (c unimplementedMultipartController) CompleteMultipart(r *http.Request, bucket, key, uploadID string, parts []Part) (location, etag string, err error) {
	return "", "", NotImplementedError(r)
}

func (c unimplementedMultipartController) ListMultipartChunks(r *http.Request, bucket, key, uploadID string, partNumberMarker, maxcParts int) (initiator, owner *User, storageClass string, isTruncated bool, parts []Part, err error) {
	return nil, nil, "", false, nil, NotImplementedError(r)
}

func (c unimplementedMultipartController) UploadMultipartChunk(r *http.Request, bucket, key, uploadID string, partNumber int, reader io.Reader) (etag string, err error) {
	return "", NotImplementedError(r)
}

func (c unimplementedMultipartController) DeleteMultipartChunk(r *http.Request, bucket, key, uploadID string, partNumber int) error {
	return NotImplementedError(r)
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

	isTruncated, uploads, err := h.controller.ListMultipart(r, bucket, keyMarker, uploadIDMarker, maxUploads)
	if err != nil {
		WriteError(h.logger, w, r, err)
		return
	}

	result := struct {
		XMLName            xml.Name `xml:"ListMultipartUploadsResult"`
		Bucket             string   `xml:"Bucket"`
		KeyMarker          string   `xml:"KeyMarker"`
		UploadIDMarker     string   `xml:"UploadIdMarker"`
		NextKeyMarker      string   `xml:"NextKeyMarker"`
		NextUploadIDMarker string   `xml:"NextUploadIdMarker"`
		MaxUploads         int      `xml:"MaxUploads"`
		IsTruncated        bool     `xml:"IsTruncated"`
		Uploads            []Upload `xml:"Upload"`
	}{
		Bucket:         bucket,
		KeyMarker:      keyMarker,
		UploadIDMarker: uploadIDMarker,
		MaxUploads:     maxUploads,
		IsTruncated:    isTruncated,
		Uploads:        uploads,
	}

	if result.IsTruncated && len(result.Uploads) > 0 {
		result.NextKeyMarker = result.Uploads[len(result.Uploads)-1].Key
		result.NextUploadIDMarker = result.Uploads[len(result.Uploads)-1].UploadID
	}

	writeXML(h.logger, w, r, http.StatusOK, result)
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

	initiator, owner, storageClass, isTruncated, parts, err := h.controller.ListMultipartChunks(r, bucket, key, uploadID, partNumberMarker, maxParts)
	if err != nil {
		WriteError(h.logger, w, r, err)
		return
	}

	result := struct {
		XMLName              xml.Name `xml:"ListPartsResult"`
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
		Parts                []Part   `xml:"Part"`
	}{
		Bucket:           bucket,
		Key:              key,
		UploadID:         uploadID,
		PartNumberMarker: partNumberMarker,
		MaxParts:         maxParts,
		Initiator:        initiator,
		Owner:            owner,
		StorageClass:     storageClass,
		IsTruncated:      isTruncated,
		Parts:            parts,
	}

	if result.IsTruncated && len(result.Parts) > 0 {
		result.NextPartNumberMarker = result.Parts[len(result.Parts)-1].PartNumber
	}

	writeXML(h.logger, w, r, http.StatusOK, result)
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

	result := struct {
		XMLName  xml.Name `xml:"InitiateMultipartUploadResult"`
		Bucket   string   `xml:"Bucket"`
		Key      string   `xml:"Key"`
		UploadID string   `xml:"UploadId"`
	}{
		Bucket:   bucket,
		Key:      key,
		UploadID: uploadID,
	}

	writeXML(h.logger, w, r, http.StatusOK, result)
}

func (h *multipartHandler) complete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	bucket := vars["bucket"]
	key := vars["key"]

	uploadID := r.FormValue("uploadId")

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		err = InternalError(r, fmt.Errorf("could not read request body: %v", err))
		WriteError(h.logger, w, r, err)
		return
	}
	payload := struct {
		XMLName xml.Name `xml:"CompleteMultipartUpload"`
		Parts   []Part   `xml:"Part"`
	}{}
	err = xml.Unmarshal(bodyBytes, &payload)
	if err != nil {
		WriteError(h.logger, w, r, MalformedXMLError(r))
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
		location string
		etag     string
		err      error
	})

	go func() {
		location, etag, err := h.controller.CompleteMultipart(r, bucket, key, uploadID, payload.Parts)
		ch <- struct {
			location string
			etag     string
			err      error
		}{
			location: location,
			etag:     etag,
			err:      err,
		}
	}()

	streaming := false

	for {
		select {
		case value := <-ch:
			if value.err != nil {
				var s3Error *Error

				switch e := value.err.(type) {
				case *Error:
					s3Error = e
				default:
					s3Error = InternalError(r, e)
				}

				if streaming {
					writeXMLBody(h.logger, w, s3Error)
				} else {
					WriteError(h.logger, w, r, s3Error)
				}
			} else {
				result := struct {
					XMLName  xml.Name `xml:"CompleteMultipartUploadResult"`
					Location string   `xml:"Location"`
					Bucket   string   `xml:"Bucket"`
					Key      string   `xml:"Key"`
					ETag     string   `xml:"ETag"`
				}{
					Bucket:   bucket,
					Key:      key,
					Location: value.location,
					ETag:     addETagQuotes(value.etag),
				}

				if streaming {
					writeXMLBody(h.logger, w, result)
				} else {
					writeXML(h.logger, w, r, http.StatusOK, result)
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

	etag := ""
	uploadID := r.FormValue("uploadId")
	partNumber, err := intFormValue(r, "partNumber", 0, maxPartsAllowed, 0)
	if err != nil {
		WriteError(h.logger, w, r, err)
		return
	}

	shouldCleanup, err := withBodyReader(r, func(reader io.Reader) error {
		fetchedETag, err := h.controller.UploadMultipartChunk(r, bucket, key, uploadID, partNumber, reader)
		etag = fetchedETag
		return err
	})

	if shouldCleanup {
		// try to clean up the chunk
		if err := h.controller.DeleteMultipartChunk(r, bucket, key, uploadID, partNumber); err != nil {
			h.logger.Errorf("could not clean up multipart chunk after an error: %+v", err)
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
