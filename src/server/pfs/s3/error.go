package s3

import (
	"net/http"
	"regexp"

	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
)

// note: the repo matcher is stricter than what PFS enforces, because this is
// based off of minio's bucket matching regexp. At the point where we're
// matching errors, we shouldn't be attempting to operate on buckets that
// don't match the minio regexp anyways.
var (
	repoNotFoundMatcher   = regexp.MustCompile(`repos/ [a-z0-9][a-z0-9\.\-]{1,61}[a-z0-9] not found`)
	branchNotFoundMatcher = regexp.MustCompile(`branches/[a-z0-9][a-z0-9\.\-]{1,61}[a-z0-9]/ [^ ]+ not found`)
	fileNotFoundMatcher   = regexp.MustCompile(`file .+ not found in repo`)
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
		httpStatus: httpStatus,
		Code:       code,
		Message:    message,
		Resource:   r.URL.Path, // fmt.Sprintf("/%s", r.URL.Path),
		RequestID:  uuid.NewWithoutDashes(),
	}
}

func newBadDigestError(r *http.Request) *Error {
	return newError(r, http.StatusBadRequest, "BadDigest", "The Content-MD5 you specified did not match what we received.")
}

func newBucketAlreadyExistsError(r *http.Request) *Error {
	return newError(r, http.StatusBadRequest, "BucketAlreadyExists", "There is already a repo with that name.")
}

// note: this is not a standard s3 error
func newGlobbyPrefixError(r *http.Request) *Error {
	return newError(r, http.StatusBadRequest, "GlobbyPrefix", "The prefix cannot contain special glob characters.")
}

func newInternalError(r *http.Request, err error) *Error {
	return newError(r, http.StatusInternalServerError, "InternalError", err.Error())
}

func newInvalidBucketNameError(r *http.Request) *Error {
	return newError(r, http.StatusBadRequest, "InvalidBucketName", "The specified repo has either an invalid name, or is not serviceable.")
}

// note: this is not a standard s3 error
func newInvalidDelimiterError(r *http.Request) *Error {
	return newError(r, http.StatusBadRequest, "InvalidDelimiter", "The delimiter you specified is invalid. It must be '' or '/'.")
}

func newInvalidDigestError(r *http.Request) *Error {
	return newError(r, http.StatusBadRequest, "InvalidDigest", "The Content-MD5 you specified is not valid.")
}

func newInvalidPartError(r *http.Request) *Error {
	return newError(r, http.StatusBadRequest, "InvalidPart", "One or more of the specified parts could not be found. The part might not have been uploaded, or it may have been deleted.")
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

// note: this is not a standard s3 error
func newMissingBranchError(r *http.Request) *Error {
	return newError(r, http.StatusBadRequest, "MissingBranch", "The key should include a branch.")
}

func newNoSuchBucketError(r *http.Request) *Error {
	return newError(r, http.StatusNotFound, "NoSuchBucket", "The specified bucket does not exist.")
}

func newNoSuchKeyError(r *http.Request) *Error {
	return newError(r, http.StatusNotFound, "NoSuchKey", "The specified key does not exist.")
}

func newNoSuchUploadError(r *http.Request) *Error {
	return newError(r, http.StatusNotFound, "NoSuchUpload", "The specified multipart upload does not exist. The upload ID might be invalid, or the multipart upload might have been aborted or completed, or it may have otherwise been deleted.")
}

func newNotFoundError(r *http.Request, err error) *Error {
	s := err.Error()

	if repoNotFoundMatcher.MatchString(s) {
		return newNoSuchBucketError(r)
	} else if branchNotFoundMatcher.MatchString(s) {
		return newNoSuchKeyError(r)
	} else if fileNotFoundMatcher.MatchString(s) {
		return newNoSuchKeyError(r)
	} else {
		return newInternalError(r, err)
	}
}
