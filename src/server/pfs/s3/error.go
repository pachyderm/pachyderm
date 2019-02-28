package s3

import (
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
	"net/http"
)

// Error is an XML-encodable error response
type Error struct {
	httpStatus int    `xml"-"`
	Code       string `xml:"Code"`
	Message    string `xml:"Message"`
	Resource   string `xml:"Resource"`
	RequestID  string `xml:"RequestId"`
}

func (e *Error) write(w http.ResponseWriter) {
	writeXML(w, e.httpStatus, e)
}

func newError(r *http.Request, httpStatus int, code string, message string) *Error {
	return &Error{
		Code:      code,
		Message:   message,
		Resource:  r.URL.Path, // fmt.Sprintf("/%s", r.URL.Path),
		RequestID: uuid.NewWithoutDashes(),
	}
}

func newBadDigestError(r *http.Request) *Error {
	return newError(r, http.StatusBadRequest, "BadDigest", "The Content-MD5 you specified did not match what we received.")
}

func newBucketAlreadyExistsError(r *http.Request) *Error {
	return newError(r, http.StatusBadRequest, "BucketAlreadyExists", "There is already a repo with that name.")
}

func newInternalError(r *http.Request, err error) *Error {
	return newError(r, http.StatusInternalServerError, "InternalError", err.Error())
}

func newInvalidBucketNameError(r *http.Request) *Error {
	return newError(r, http.StatusBadRequest, "InvalidBucketName", "The specified repo has either an invalid name, or is not serviceable.")
}

func newInvalidDigestError(r *http.Request) *Error {
	return newError(r, http.StatusBadRequest, "InvalidDigest", "The Content-MD5 you specified is not valid.")
}

func newInvalidPartError(r *http.Request) *Error {
	return newError(r, http.StatusBadRequest, "InvalidPart", "One or more of the specified parts could not be found. The part might not have been uploaded, or the specified entity tag might not have matched the part's entity tag.")
}

func newInvalidPartOrderError(r *http.Request) *Error {
	return newError(r, http.StatusBadRequest, "InvalidPartOrder", "The list of parts was not in ascending order. Parts list must be specified in order by part number.")
}

func newMalformedXMLError(r *http.Request) *Error {
	return newError(r, http.StatusBadRequest, "MalformedXML", "The XML you provided was not well-formed or would not validate against S3's published schema.")
}

func newMethodNotAllowedError(r *http.Request) *Error {
	return newError(r, http.StatusMethodNotAllowed, "MethodNotAllowed", "The specified method is not allowed against this resource.")
}

// note: this is not a typical s3 error
func newMissingBranchError(r *http.Request) *Error {
	return newError(r, http.StatusBadRequest, "MissingBranch", "The key should include a branch.")
}

func newNoSuchBucketError(r *http.Request) *Error {
	return newError(r, http.StatusNotFound, "NoSuchBucket", "The specified bucket does not exist.")
}

func newNoSuchKeyError(r *http.Request) *Error {
	return newError(r, http.StatusNotFound, "NoSuchKey", "The specified key does not exist.")
}
