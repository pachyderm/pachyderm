package s2

// This is largely lifted from go's stdlib net/http, but modified to be
// simpler and to work with amazon's proprietary `x-amz-`-prefixed
// conditional matching headers, rather than the standard HTTP ones

import (
	"net/http"
	"net/textproto"
	"strings"
	"time"
)

func checkIfMatch(im string, etag string) bool {
	if im == "" {
		return true
	}

	for {
		im = textproto.TrimString(im)
		if len(im) == 0 {
			break
		}
		if im[0] == ',' {
			im = im[1:]
			continue
		}
		if im[0] == '*' {
			return true
		}
		checkEtag, remain := scanETag(im)
		if checkEtag == "" {
			break
		}
		if etagStrongMatch(checkEtag, etag) {
			return true
		}
		im = remain
	}

	return false
}

func checkIfNoneMatch(inm string, etag string) bool {
	if inm == "" {
		return true
	}

	buf := inm

	for {
		buf = textproto.TrimString(buf)
		if len(buf) == 0 {
			break
		}
		if buf[0] == ',' {
			buf = buf[1:]
		}
		if buf[0] == '*' {
			return false
		}
		checkEtag, remain := scanETag(buf)
		if checkEtag == "" {
			break
		}
		if etagWeakMatch(checkEtag, etag) {
			return false
		}
		buf = remain
	}
	return true
}

func checkIfUnmodifiedSince(ius string, modtime time.Time) bool {
	if ius == "" || isZeroTime(modtime) {
		return true
	}
	t, err := http.ParseTime(ius)
	if err != nil {
		return true
	}

	// The Last-Modified header truncates sub-second precision so
	// the modtime needs to be truncated too.
	modtime = modtime.Truncate(time.Second)
	if modtime.Before(t) || modtime.Equal(t) {
		return true
	}
	return false
}

func checkIfModifiedSince(ims string, modtime time.Time) bool {
	if ims == "" || isZeroTime(modtime) {
		return true
	}
	t, err := http.ParseTime(ims)
	if err != nil {
		return true
	}
	// The Last-Modified header truncates sub-second precision so
	// the modtime needs to be truncated too.
	modtime = modtime.Truncate(time.Second)
	if modtime.Before(t) || modtime.Equal(t) {
		return false
	}
	return true
}

// scanETag determines if a syntactically valid ETag is present at s. If so,
// the ETag and remaining text after consuming ETag is returned. Otherwise,
// it returns "", "".
func scanETag(s string) (etag string, remain string) {
	s = textproto.TrimString(s)
	start := 0
	if strings.HasPrefix(s, "W/") {
		start = 2
	}
	if len(s[start:]) < 2 || s[start] != '"' {
		return "", ""
	}
	// ETag is either W/"text" or "text".
	// See RFC 7232 2.3.
	for i := start + 1; i < len(s); i++ {
		c := s[i]
		switch {
		// Character values allowed in ETags.
		case c == 0x21 || c >= 0x23 && c <= 0x7E || c >= 0x80:
		case c == '"':
			return s[:i+1], s[i+1:]
		default:
			return "", ""
		}
	}
	return "", ""
}

// etagStrongMatch reports whether a and b match using strong ETag comparison.
// Assumes a and b are valid ETags.
func etagStrongMatch(a, b string) bool {
	return a == b && a != "" && a[0] == '"'
}

// etagWeakMatch reports whether a and b match using weak ETag comparison.
// Assumes a and b are valid ETags.
func etagWeakMatch(a, b string) bool {
	return strings.TrimPrefix(a, "W/") == strings.TrimPrefix(b, "W/")
}

// isZeroTime reports whether t is obviously unspecified (either zero or Unix()=0).
func isZeroTime(t time.Time) bool {
	return t.IsZero() || t.Equal(unixEpoch)
}
