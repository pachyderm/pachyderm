// Package conditionalrequest handles HTTP conditional requests based on modification time and
// etags.
package conditionalrequest

import (
	"net/http"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

var (
	ErrPreconditionFailed = errors.New("precondition failed")
	ErrNotModified        = errors.New("not modified")
)

// ResourceInfo is information about a resource to be tested with conditionals.
type ResourceInfo struct {
	LastModified time.Time
	ETag         string // Includes the quotes.
}

// https://www.rfc-editor.org/rfc/rfc9110#field.if-modified-since
func (info *ResourceInfo) ifModifiedSince(httpDate string) error {
	t, err := http.ParseTime(httpDate)
	if err != nil {
		// A recipient MUST ignore the If-Modified-Since header field if the received field
		// value is not a valid HTTP-date.
		return nil
	}
	if info.LastModified.IsZero() {
		// A recipient MUST ignore the If-Modified-Since header field if the resource does
		// not have a modification date available.
		return nil
	}
	// If the selected representation's last modification date is earlier or equal to the
	// date provided in the field value, the condition is false.
	if info.LastModified.Before(t) || info.LastModified.Equal(t) {
		return ErrNotModified
	}
	// Otherwise, the condition is true.
	return nil
}

// https://www.rfc-editor.org/rfc/rfc9110#field.if-unmodified-since
func (info *ResourceInfo) ifUnmodifiedSince(httpDate string) error {
	t, err := http.ParseTime(httpDate)
	if err != nil {
		return nil
	}
	if info.LastModified.IsZero() {
		return nil
	}
	// If the selected representation's last modification date is earlier or equal to the
	// date provided in the field value, the condition is true.
	if info.LastModified.Before(t) || info.LastModified.Equal(t) {
		return nil
	}
	// Otherwise, the condition is false.
	return ErrPreconditionFailed
}

// https://www.rfc-editor.org/rfc/rfc9110#field.if-match
func (info *ResourceInfo) ifMatch(etags string) error {
	if etags == "*" {
		// If the field value is "*", the condition is true if the origin server has a
		// current representation for the target resource.
		return nil
	}
	// If the field value is a list of entity tags, the condition is true if any of the listed
	// tags match the entity tag of the selected representation.
	for _, part := range strings.Split(etags, ",") {
		part = strings.TrimSpace(part)
		if part == info.ETag {
			return nil
		}
	}
	// Otherwise, the condition is false.
	return ErrPreconditionFailed
}

// https://www.rfc-editor.org/rfc/rfc9110#field.if-none-match
func (info *ResourceInfo) ifNoneMatch(etags string) error {
	if etags == "*" {
		// If the field value is "*", the condition is false if the origin server has a
		// current representation for the target resource.
		return ErrPreconditionFailed
	}
	// If the field value is a list of entity tags, the condition is false if one of the listed
	// tags matches the entity tag of the selected representation.
	for _, part := range strings.Split(etags, ",") {
		part = strings.TrimSpace(part)
		if part == info.ETag {
			return ErrPreconditionFailed
		}
	}
	// Otherwise, the condition is true.
	return nil
}

// https://www.rfc-editor.org/rfc/rfc9110#name-if-range
func (info *ResourceInfo) ifRange(value string) error {
	// A valid entity-tag can be distinguished from a valid HTTP-date by examining the first
	// three characters for a DQUOTE.
	if len(value) > 3 && strings.ContainsAny(value, `"`) {
		// To evaluate a received If-Range header field containing an entity-tag:
		if value == info.ETag {
			// 1. If the entity-tag validator provided exactly matches the ETag field
			// value for the selected representation using the strong comparison
			// function (Section 8.8.3.2), the condition is true.
			return nil
		} else {
			// 2. Otherwise, the condition is false.
			return ErrPreconditionFailed
		}
	}

	// To evaluate a received If-Range header field containing an HTTP-date:

	// 1. If the HTTP-date validator provided is not a strong validator in the sense
	// defined by Section 8.8.2.2, the condition is false.
	// 2. If the HTTP-date validator provided exactly matches the Last-Modified field
	// value for the selected representation, the condition is true.
	if value == info.LastModified.UTC().Format(http.TimeFormat) {
		return nil
	} else {
		// 3. Otherwise, the condition is false.
		return ErrPreconditionFailed
	}
}

// Evaluate evaluates an HTTP conditional request, returning 0 if the request should be handled
// normally, or a status code if it should be aborted.
func Evaluate(req *http.Request, info *ResourceInfo) int {
	// https://www.rfc-editor.org/rfc/rfc9110#name-precedence-of-preconditions
	if ifMatch := req.Header.Get("if-match"); ifMatch != "" {
		// Step 1: When recipient is the origin server and If-Match is present, evaluate the
		// If-Match precondition:
		if err := info.ifMatch(ifMatch); err != nil {
			// if false, respond 412 (Precondition Failed)
			return http.StatusPreconditionFailed
		}
		// if true, continue to step 3
	} else if ifUnmodifiedSince := req.Header.Get("if-unmodified-since"); ifUnmodifiedSince != "" {
		// Step 2: When recipient is the origin server, If-Match is not present, and
		// If-Unmodified-Since is present, evaluate the If-Unmodified-Since precondition
		if err := info.ifUnmodifiedSince(ifUnmodifiedSince); err != nil {
			// if false, respond 412 (Precondition Failed)
			return http.StatusPreconditionFailed
		}
		// if true, continue to step 3
	}
	if ifNoneMatch := req.Header.Get("if-none-match"); ifNoneMatch != "" {
		// Step 3: When If-None-Match is present, evaluate the If-None-Match precondition
		if err := info.ifNoneMatch(ifNoneMatch); err != nil {
			switch req.Method {
			case http.MethodGet, http.MethodHead:
				// if false for GET/HEAD, respond 304 (Not Modified)
				return http.StatusNotModified
			default:
				// if false for other methods, respond 412 (Precondition Failed)
				return http.StatusPreconditionFailed
			}
		}
		// if true, continue to step 5
	} else if ifModifiedSince := req.Header.Get("if-modified-since"); (req.Method == http.MethodGet || req.Method == http.MethodHead) && ifModifiedSince != "" {
		// Step 4: When the method is GET or HEAD, If-None-Match is not present, and
		// If-Modified-Since is present, evaluate the If-Modified-Since precondition.
		if err := info.ifModifiedSince(ifModifiedSince); err != nil {
			// if false, respond 304 (Not Modified)
			return http.StatusNotModified
		}
		// if true, continue to step 5
	}
	if ifRange := req.Header.Get("if-range"); req.Method == http.MethodGet && req.Header.Get("range") != "" && ifRange != "" {
		// Step 5: When the method is GET and both Range and If-Range are present, evaluate the
		// If-Range precondition.
		if err := info.ifRange(ifRange); err == nil {
			// if true and the Range is applicable to the selected representation,
			// respond 206 (Partial Content)
			return http.StatusPartialContent
		}
		// otherwise, ignore the Range header field and respond 200 (OK)
		req.Header.Del("range")
		return 0
	}
	// Step 6: Otherwise, perform the requested method and respond according to its success or
	// failure.
	return 0
}
