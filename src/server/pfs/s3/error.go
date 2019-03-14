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
	repoNotFoundMatcher   = regexp.MustCompile(`repos/ ?[a-zA-Z0-9.\-_]{1,255} not found`)
	branchNotFoundMatcher = regexp.MustCompile(`branches/[a-zA-Z0-9.\-_]{1,255}/ [^ ]+ not found`)
	fileNotFoundMatcher   = regexp.MustCompile(`file .+ not found`)
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
		Resource:   r.URL.Path,
		RequestID:  uuid.NewWithoutDashes(),
	}
}

func badDigestError(w http.ResponseWriter, r *http.Request) {
	newError(r, http.StatusBadRequest, "BadDigest", "The Content-MD5 you specified did not match what we received.").write(w)
}

func bucketAlreadyExistsError(w http.ResponseWriter, r *http.Request) {
	newError(r, http.StatusBadRequest, "BucketAlreadyExists", "There is already a repo with that name.").write(w)
}

func internalError(w http.ResponseWriter, r *http.Request, err error) {
	newError(r, http.StatusInternalServerError, "InternalError", err.Error()).write(w)
}

func invalidBucketNameError(w http.ResponseWriter, r *http.Request) {
	newError(r, http.StatusBadRequest, "InvalidBucketName", "The specified repo or branch either has an invalid name, or is not serviceable.").write(w)
}

// note: this is not a standard s3 error
func invalidDelimiterError(w http.ResponseWriter, r *http.Request) {
	newError(r, http.StatusBadRequest, "InvalidDelimiter", "The delimiter you specified is invalid. It must be '' or '/'.").write(w)
}

func invalidDigestError(w http.ResponseWriter, r *http.Request) {
	newError(r, http.StatusBadRequest, "InvalidDigest", "The Content-MD5 you specified is not valid.").write(w)
}

// note: this is not a standard s3 error
func invalidFilePathError(w http.ResponseWriter, r *http.Request) {
	newError(r, http.StatusBadRequest, "InvalidFilePath", "You cannot operate on s3gateway metadata files (files with the extension '.s3g.json')").write(w)
}

func invalidPartError(w http.ResponseWriter, r *http.Request) {
	newError(r, http.StatusBadRequest, "InvalidPart", "One or more of the specified parts could not be found. The part might not have been uploaded, or it may have been deleted.").write(w)
}

func invalidPartOrderError(w http.ResponseWriter, r *http.Request) {
	newError(r, http.StatusBadRequest, "InvalidPartOrder", "The list of parts was not in ascending order. Parts list must be specified in order by part number.").write(w)
}

func malformedXMLError(w http.ResponseWriter, r *http.Request) {
	newError(r, http.StatusBadRequest, "MalformedXML", "The XML you provided was not well-formed or would not validate against S3's published schema.").write(w)
}

func methodNotAllowedError(w http.ResponseWriter, r *http.Request) {
	newError(r, http.StatusMethodNotAllowed, "MethodNotAllowed", "The specified method is not allowed against this resource.").write(w)
}
func noSuchBucketError(w http.ResponseWriter, r *http.Request) {
	newError(r, http.StatusNotFound, "NoSuchBucket", "The specified bucket does not exist.").write(w)
}

func noSuchKeyError(w http.ResponseWriter, r *http.Request) {
	newError(r, http.StatusNotFound, "NoSuchKey", "The specified key does not exist.").write(w)
}

func noSuchUploadError(w http.ResponseWriter, r *http.Request) {
	newError(r, http.StatusNotFound, "NoSuchUpload", "The specified multipart upload does not exist. The upload ID might be invalid, or the multipart upload might have been aborted or completed, or it may have otherwise been deleted.").write(w)
}

func notFoundError(w http.ResponseWriter, r *http.Request, err error) {
	s := err.Error()

	if repoNotFoundMatcher.MatchString(s) || branchNotFoundMatcher.MatchString(s) {
		noSuchBucketError(w, r)
	} else if fileNotFoundMatcher.MatchString(s) {
		noSuchKeyError(w, r)
	} else {
		internalError(w, r, err)
	}
}

func notImplementedError(w http.ResponseWriter, r *http.Request) {
	newError(r, http.StatusNotImplemented, "NotImplemented", "This functionality is not implemented in the pachyderm s3gateway.").write(w)
}
