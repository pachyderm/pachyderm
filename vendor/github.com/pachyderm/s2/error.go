package s2

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// Error is an XML marshalable error response
type Error struct {
	// HTTPStatus is the HTTP status that will be set in the response
	HTTPStatus int `xml:"-"`

	Code      string `xml:"Code"`
	Message   string `xml:"Message"`
	Resource  string `xml:"Resource"`
	RequestID string `xml:"RequestId"`
}

// NewError creates a new S3 error, to be serialized in a response
func NewError(r *http.Request, httpStatus int, code string, message string) *Error {
	vars := mux.Vars(r)
	requestID := vars["requestID"]

	return &Error{
		HTTPStatus: httpStatus,
		Code:       code,
		Message:    message,
		Resource:   r.URL.Path,
		RequestID:  requestID,
	}
}

func (e *Error) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// BadDigestError creates a new S3 error with a standard BadDigest S3 code.
func BadDigestError(r *http.Request) *Error {
	return NewError(r, http.StatusBadRequest, "BadDigest", "The Content-MD5 you specified did not match what we received.")
}

// BucketNotEmptyError creates a new S3 error with a standard BucketNotEmpty
// S3 code.
func BucketNotEmptyError(r *http.Request) *Error {
	return NewError(r, http.StatusConflict, "BucketNotEmpty", "The bucket you tried to delete is not empty.")
}

// BucketAlreadyOwnedByYouError creates a new S3 error with a standard
// BucketAlreadyOwnedByYou S3 code.
func BucketAlreadyOwnedByYouError(r *http.Request) *Error {
	return NewError(r, http.StatusConflict, "BucketAlreadyOwnedByYou", "The bucket you tried to create already exists, and you own it.")
}

// EntityTooSmallError creates a new S3 error with a standard EntityTooSmall
// S3 code.
func EntityTooSmallError(r *http.Request) *Error {
	return NewError(r, http.StatusBadRequest, "EntityTooSmall", "Your proposed upload is smaller than the minimum allowed object size. Each part must be at least 5 MB in size, except the last part.")
}

// InternalError creates a new S3 error with a standard InternalError S3 code.
func InternalError(r *http.Request, err error) *Error {
	return NewError(r, http.StatusInternalServerError, "InternalError", err.Error())
}

// InvalidBucketNameError creates a new S3 error with a standard
// InvalidBucketName S3 code.
func InvalidBucketNameError(r *http.Request) *Error {
	return NewError(r, http.StatusBadRequest, "InvalidBucketName", "The specified bucket is not valid.")
}

// InvalidArgument creates a new S3 error with a standard InvalidArgument S3
// code.
func InvalidArgument(r *http.Request) *Error {
	return NewError(r, http.StatusBadRequest, "InvalidArgument", "Invalid Argument")
}

// InvalidDigestError creates a new S3 error with a standard InvalidDigest S3
// code.
func InvalidDigestError(r *http.Request) *Error {
	return NewError(r, http.StatusBadRequest, "InvalidDigest", "The Content-MD5 you specified is not valid.")
}

// InvalidPartError creates a new S3 error with a standard InvalidPart S3
// code.
func InvalidPartError(r *http.Request) *Error {
	return NewError(r, http.StatusBadRequest, "InvalidPart", "One or more of the specified parts could not be found. The part might not have been uploaded, or the specified entity tag might not have matched the part's entity tag.")
}

// InvalidPartOrderError creates a new S3 error with a standard
// InvalidPartOrder S3 code.
func InvalidPartOrderError(w http.ResponseWriter, r *http.Request) *Error {
	return NewError(r, http.StatusBadRequest, "InvalidPartOrder", "The list of parts was not in ascending order. Parts list must be specified in order by part number.")
}

// MalformedXMLError creates a new S3 error with a standard MalformedXML S3
// code.
func MalformedXMLError(r *http.Request) *Error {
	return NewError(r, http.StatusBadRequest, "MalformedXML", "The XML you provided was not well-formed or would not validate against S3's published schema.")
}

// MethodNotAllowedError creates a new S3 error with a standard
// MethodNotAllowed S3 code.
func MethodNotAllowedError(r *http.Request) *Error {
	return NewError(r, http.StatusMethodNotAllowed, "MethodNotAllowed", "The specified method is not allowed against this resource.")
}

// NoSuchBucketError creates a new S3 error with a standard NoSuchBucket S3
// code.
func NoSuchBucketError(r *http.Request) *Error {
	return NewError(r, http.StatusNotFound, "NoSuchBucket", "The specified bucket does not exist.")
}

// NoSuchKeyError creates a new S3 error with a standard NoSuchKey S3 code.
func NoSuchKeyError(r *http.Request) *Error {
	return NewError(r, http.StatusNotFound, "NoSuchKey", "The specified key does not exist.")
}

// NoSuchUploadError creates a new S3 error with a standard NoSuchUpload S3
// code.
func NoSuchUploadError(r *http.Request) *Error {
	return NewError(r, http.StatusNotFound, "NoSuchUpload", "The specified multipart upload does not exist. The upload ID might be invalid, or the multipart upload might have been aborted or completed.")
}

// NotImplementedError creates a new S3 error with a standard NotImplemented
// S3 code.
func NotImplementedError(r *http.Request) *Error {
	return NewError(r, http.StatusNotImplemented, "NotImplemented", "This functionality is not implemented.")
}
