package s2

import (
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

// Error is an XML-encodable error response
type Error struct {
	r          *http.Request `xml:"-"`
	httpStatus int           `xml:"-"`
	Code       string        `xml:"Code"`
	Message    string        `xml:"Message"`
	Resource   string        `xml:"Resource"`
	RequestID  string        `xml:"RequestId"`
}

func NewError(r *http.Request, httpStatus int, code string, message string) *Error {
	return &Error{
		r:          r,
		httpStatus: httpStatus,
		Code:       code,
		Message:    message,
		Resource:   r.URL.Path,
		RequestID:  r.Header.Get("X-Request-ID"),
	}
}

func (e *Error) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *Error) Write(logger *logrus.Entry, w http.ResponseWriter) {
	writeXML(logger, w, e.r, e.httpStatus, e)
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

func NotImplementedError(r *http.Request) *Error {
	return NewError(r, http.StatusNotImplemented, "NotImplemented", "This functionality is not implemented.")
}
