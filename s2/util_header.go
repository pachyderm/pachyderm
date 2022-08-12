package s2

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	// awsTimeFormat specifies the time format used in AWS requests
	awsTimeFormat = "20060102T150405Z"
	// skewTime specifies the maximum delta between the current time and the
	// time specified in the HTTP request
	skewTime = 15 * time.Minute
)

var (
	// unixEpoch represents the unix epoch time (Jan 1 1970)
	unixEpoch = time.Unix(0, 0)
)

// intFormValue extracts an int value from a request's form values, ensuring
// it's within specified bounds. If the value is unspecified, `def` is
// returned. If the value is not an int, or not with the specified bounds, an
// error is returned.
func intFormValue(r *http.Request, name string, min int, max int, def int) (int, error) {
	s := r.FormValue(name)
	if s == "" {
		return def, nil
	}

	i, err := strconv.Atoi(s)
	if err != nil || i < min || i > max {
		return 0, InvalidArgumentError(r)
	}

	return i, nil
}

//stripETagQuotes removes leading and trailing quotes in a string (if they
// exist.) This is used for ETags.
func stripETagQuotes(s string) string {
	if strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"") {
		return strings.Trim(s, "\"")
	}
	return s
}

// addETagQuotes ensures that a given string has leading and trailing quotes.
// This is used for ETags.
func addETagQuotes(s string) string {
	if !strings.HasPrefix(s, "\"") {
		return fmt.Sprintf("\"%s\"", s)
	}
	return s
}

// normURI normalizes a URI using AWS' technique
func normURI(uri string) string {
	parts := strings.Split(uri, "/")
	for i := range parts {
		parts[i] = encodePathFrag(parts[i])
	}
	return strings.Join(parts, "/")
}

// encodePathFrag encodes a fragment of a path in a URL using AWS' technique
func encodePathFrag(s string) string {
	hexCount := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if shouldEscape(c) {
			hexCount++
		}
	}
	t := make([]byte, len(s)+2*hexCount)
	j := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if shouldEscape(c) {
			t[j] = '%'
			t[j+1] = "0123456789ABCDEF"[c>>4]
			t[j+2] = "0123456789ABCDEF"[c&15]
			j += 3
		} else {
			t[j] = c
			j++
		}
	}
	return string(t)
}

// shouldEscape returns whether a character should be escaped under AWS' URL
// encoding
func shouldEscape(c byte) bool {
	if 'a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' {
		return false
	}
	if '0' <= c && c <= '9' {
		return false
	}
	if c == '-' || c == '_' || c == '.' || c == '~' {
		return false
	}
	return true
}

// normQuery normalizes query string values using AWS' technique
func normQuery(v url.Values) string {
	queryString := v.Encode()

	// Go encodes a space as '+' but Amazon requires '%20'. Luckily any '+' in the
	// original query string has been percent escaped so all '+' chars that are left
	// were originally spaces.

	return strings.Replace(queryString, "+", "%20", -1)
}

// hmacSHA1 computes HMAC with SHA1
func hmacSHA1(key []byte, content string) []byte {
	mac := hmac.New(sha1.New, key)
	mac.Write([]byte(content))
	return mac.Sum(nil)
}

// hmacSHA256 computes HMAC with SHA256
func hmacSHA256(key []byte, content string) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(content))
	return mac.Sum(nil)
}

// requireContentLength checks to ensure that an HTTP request includes a
// `Content-Length` header.
func requireContentLength(r *http.Request) error {
	if _, ok := singleHeader(r, "Content-Length"); !ok {
		return MissingContentLengthError(r)
	}
	return nil
}

// singleHeader gets a single header value. This is used in places instead of
// `r.Header.Get()` because it differentiates between missing headers versus
// empty header values.
func singleHeader(r *http.Request, name string) (string, bool) {
	values, ok := r.Header[name]
	if !ok {
		return "", false
	}
	if len(values) != 1 {
		return "", false
	}
	return values[0], true
}

func formatAWSTimestamp(t time.Time) string {
	return t.Format(awsTimeFormat)
}

//  parseTimestamp parses a timestamp value that is formatted in any of the
// following:
// 1) as AWS' custom format (e.g. 20060102T150405Z)
// 2) as RFC1123
// 3) as RFC1123Z
func parseAWSTimestamp(r *http.Request) (time.Time, error) {
	timestampStr := r.Header.Get("x-amz-date")
	if timestampStr == "" {
		timestampStr = r.Header.Get("date")
	}

	timestamp, err := time.Parse(time.RFC1123, timestampStr)
	if err != nil {
		timestamp, err = time.Parse(time.RFC1123Z, timestampStr)
		if err != nil {
			timestamp, err = time.Parse(awsTimeFormat, timestampStr)
			if err != nil {
				return time.Time{}, AccessDeniedError(r)
			}
		}
	}

	if !timestamp.After(unixEpoch) {
		return time.Time{}, AccessDeniedError(r)
	}

	now := time.Now()
	if !timestamp.After(now.Add(-skewTime)) || timestamp.After(now.Add(skewTime)) {
		return time.Time{}, RequestTimeTooSkewedError(r)
	}

	return timestamp, nil
}
