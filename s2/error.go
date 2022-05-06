package s2

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

// Error is an XML marshallable error response
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

// newGenericError takes in a generic error, and returns an s2 `Error`. If
// the input error is not already an s2 `Error`, it is turned into an
// `InternalError`.
func newGenericError(r *http.Request, err error) *Error {
	switch e := err.(type) {
	case *Error:
		return e
	default:
		return InternalError(r, e)
	}
}

func (e *Error) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// AccessDeniedError creates a new S3 error with a standard AccessDenied S3
// code.
func AccessDeniedError(r *http.Request) *Error {
	return NewError(r, http.StatusForbidden, "AccessDenied", "Access Denied")
}

// AuthorizationHeaderMalformedError creates a new S3 error with a standard
// AuthorizationHeaderMalformed S3 code.
func AuthorizationHeaderMalformedError(r *http.Request) *Error {
	return NewError(r, http.StatusBadRequest, "AuthorizationHeaderMalformed", "The authorization header you provided is invalid.")
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

// EntityTooLargeError creates a new S3 error with a standard EntityTooLarge
// S3 code.
func EntityTooLargeError(r *http.Request) *Error {
	return NewError(r, http.StatusBadRequest, "EntityTooLarge", "Your proposed upload exceeds the maximum allowed object size.")
}

// EntityTooSmallError creates a new S3 error with a standard EntityTooSmall
// S3 code.
func EntityTooSmallError(r *http.Request) *Error {
	return NewError(r, http.StatusBadRequest, "EntityTooSmall", "Your proposed upload is smaller than the minimum allowed object size. Each part must be at least 5 MB in size, except the last part.")
}

// IllegalVersioningConfigurationError creates a new S3 error with a standard
// IllegalVersioningConfigurationException S3 code.
func IllegalVersioningConfigurationError(r *http.Request) *Error {
	return NewError(r, http.StatusBadRequest, "IllegalVersioningConfigurationException", "The versioning configuration specified in the request is invalid.")
}

// IncompleteBodyError creates a new S3 error with a standard IncompleteBody S3 code.
func IncompleteBodyError(r *http.Request) *Error {
	return NewError(r, http.StatusBadRequest, "IncompleteBody", "You did not provide the number of bytes specified by the Content-Length HTTP header.")
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

// InvalidAccessKeyIDError creates a new S3 error with a standard
// InvalidAccessKeyId S3 code.
func InvalidAccessKeyIDError(r *http.Request) *Error {
	return NewError(r, http.StatusForbidden, "InvalidAccessKeyId", "The AWS access key ID you provided does not exist in our records.")
}

// InvalidArgumentError creates a new S3 error with a standard InvalidArgument S3
// code.
func InvalidArgumentError(r *http.Request) *Error {
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

// InvalidRequestError creates a new S3 error with a standard
// InvalidRequest S3 code.
func InvalidRequestError(r *http.Request, message string) *Error {
	return NewError(r, http.StatusBadRequest, "InvalidRequest", message)
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

// MissingContentLengthError creates a new S3 error with a standard
// MissingContentLength S3 code.
func MissingContentLengthError(r *http.Request) *Error {
	return NewError(r, http.StatusLengthRequired, "MissingContentLength", "You must provide the Content-Length HTTP header.")
}

// MissingRequestBodyError creates a new S3 error with a standard
// MissingRequestBodyError S3 code.
func MissingRequestBodyError(r *http.Request) *Error {
	return NewError(r, http.StatusBadRequest, "MissingRequestBodyError", "Request body is empty.")
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

// NoSuchVersionError creates a new S3 error with a standard NoSuchVersion S3
// code.
func NoSuchVersionError(r *http.Request) *Error {
	return NewError(r, http.StatusNotFound, "NoSuchVersion", "The version ID specified in the request does not match an existing version.")
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

// PreconditionFailedError creates a new S3 error with a standard
// PreconditionFailed S3 code.
func PreconditionFailedError(r *http.Request) *Error {
	return NewError(r, http.StatusPreconditionFailed, "PreconditionFailed", "At least one of the preconditions you specified did not hold.")
}

// RequestTimeoutError creates a new S3 error with a standard RequestTimeout
// S3 code.
func RequestTimeoutError(r *http.Request) *Error {
	return NewError(r, http.StatusBadRequest, "RequestTimeout", "Your socket connection to the server was not read from or written to within the timeout period.")
}

// RequestTimeTooSkewedError creates a new S3 error with a standard
// RequestTimeTooSkewed S3 code.
func RequestTimeTooSkewedError(r *http.Request) *Error {
	return NewError(r, http.StatusForbidden, "RequestTimeTooSkewed", "The difference between the request time and the server's time is too large.")
}

// SignatureDoesNotMatchError creates a new S3 error with a standard
// SignatureDoesNotMatch S3 code.
func SignatureDoesNotMatchError(r *http.Request) *Error {
	return NewError(r, http.StatusForbidden, "SignatureDoesNotMatch", "The request signature we calculated does not match the signature you provided. Check your auth credentials and signing method.")
}
