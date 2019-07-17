package s2

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// Error is an XML-encodable error response
type Error struct {
	HttpStatus int    `xml:"-"`
	Code       string `xml:"Code"`
	Message    string `xml:"Message"`
	Resource   string `xml:"Resource"`
	RequestID  string `xml:"RequestId"`
}

func NewError(r *http.Request, httpStatus int, code string, message string) *Error {
	vars := mux.Vars(r)
	requestID := vars["requestID"]

	return &Error{
		HttpStatus: httpStatus,
		Code:       code,
		Message:    message,
		Resource:   r.URL.Path,
		RequestID:  requestID,
	}
}

func (e *Error) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func BadDigestError(r *http.Request) *Error {
	return NewError(r, http.StatusBadRequest, "BadDigest", "The Content-MD5 you specified did not match what we received.")
}

func BucketNotEmptyError(r *http.Request) *Error {
	return NewError(r, http.StatusConflict, "BucketNotEmpty", "The bucket you tried to delete is not empty.")
}

func BucketAlreadyOwnedByYouError(r *http.Request) *Error {
	return NewError(r, http.StatusConflict, "BucketAlreadyOwnedByYou", "The bucket you tried to create already exists, and you own it.")
}

func EntityTooSmallError(r *http.Request) *Error {
	return NewError(r, http.StatusBadRequest, "EntityTooSmall", "Your proposed upload is smaller than the minimum allowed object size. Each part must be at least 5 MB in size, except the last part.")
}

func InternalError(r *http.Request, err error) *Error {
	return NewError(r, http.StatusInternalServerError, "InternalError", err.Error())
}

func InvalidBucketNameError(r *http.Request) *Error {
	return NewError(r, http.StatusBadRequest, "InvalidBucketName", "The specified bucket is not valid.")
}

func InvalidArgument(r *http.Request) *Error {
	return NewError(r, http.StatusBadRequest, "InvalidArgument", "Invalid Argument")
}

func InvalidDigestError(r *http.Request) *Error {
	return NewError(r, http.StatusBadRequest, "InvalidDigest", "The Content-MD5 you specified is not valid.")
}

func InvalidPartError(r *http.Request) *Error {
	return NewError(r, http.StatusBadRequest, "InvalidPart", "One or more of the specified parts could not be found. The part might not have been uploaded, or the specified entity tag might not have matched the part's entity tag.")
}

func InvalidPartOrderError(w http.ResponseWriter, r *http.Request) *Error {
	return NewError(r, http.StatusBadRequest, "InvalidPartOrder", "The list of parts was not in ascending order. Parts list must be specified in order by part number.")
}

func MalformedXMLError(r *http.Request) *Error {
	return NewError(r, http.StatusBadRequest, "MalformedXML", "The XML you provided was not well-formed or would not validate against S3's published schema.")
}

func MethodNotAllowedError(r *http.Request) *Error {
	return NewError(r, http.StatusMethodNotAllowed, "MethodNotAllowed", "The specified method is not allowed against this resource.")
}

func NoSuchBucketError(r *http.Request) *Error {
	return NewError(r, http.StatusNotFound, "NoSuchBucket", "The specified bucket does not exist.")
}

func NoSuchKeyError(r *http.Request) *Error {
	return NewError(r, http.StatusNotFound, "NoSuchKey", "The specified key does not exist.")
}

func NoSuchUploadError(r *http.Request) *Error {
	return NewError(r, http.StatusNotFound, "NoSuchUpload", "The specified multipart upload does not exist. The upload ID might be invalid, or the multipart upload might have been aborted or completed.")
}

func NotImplementedError(r *http.Request) *Error {
	return NewError(r, http.StatusNotImplemented, "NotImplemented", "This functionality is not implemented.")
}
