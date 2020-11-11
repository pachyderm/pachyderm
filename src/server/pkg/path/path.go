// This is a library that provides some path-cleaning and path manipulation
// functions for hashtree.go and worker/datum/iterator.go. The functions it
// defines are very similar to the functions in go's "path" library.

// In both, a canonicalized path has a leading slash and no trailing slash in
// general. The difference is:
//
// - In this library, the canonical version of "/" is "" (i.e. preserve "no
//   trailing slash" invariant, at the cost of the "always have a leading slash"
//   invariant), whereas
//
// - in go's "path" library, the canonical version of "/" is "/" and "" becomes
//   ".".
//
// We prefer our canonicalizion because it gives us globbing behavior we like:
// "/" and ""   (canonicalize to: "")   match "/" ("") but not "/foo"
// "*" and "/*" (canonicalize to: "/*") match "/foo"   but not "/" ("")

package path

import (
	"path"
	"regexp"
	"strings"

	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
)

var globRegex = regexp.MustCompile(`[*?[\]{}!()@+^]`)

// internalDefault overrides the internal defaults of many functions in the
// "path" library.  Specifically, the top-level dir "/" and the special string
// "." (which is what most "path" functions return for the empty string) both
// map to the empty string here, so that we get the globbing behavior we want
// (see top).
func internalDefault(s string) string {
	if s == "/" || s == "." {
		return ""
	}
	return s
}

// Clean canonicalizes 'path' (which may be either a file path or a glob
// pattern) for internal use: leading slash and no trailing slash (see top).
// Also, clean the result with internalDefault.
func Clean(p string) string {
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return internalDefault(path.Clean(p))
}

// Base is like path.Base, but uses this library's defaults for canonical paths
func Base(p string) string {
	// remove leading '/' added by Clean() (to all paths except "", which is
	// unchanged)
	return strings.TrimPrefix(Clean(path.Base(Clean(p))), "/")
}

// Dir is like path.Dir, but uses this library's defaults for canonical paths
func Dir(p string) string {
	return Clean(path.Dir(Clean(p)))
}

// Split is like path.Split, but uses this library's defaults for canonical
// paths
func Split(p string) (string, string) {
	d, b := path.Split(Clean(p))
	return Clean(d), strings.TrimPrefix(Clean(b), "/")
}

// Join is like path.Join, but uses our version of 'clean()' instead of
// path.Clean()
func Join(ps ...string) string {
	return Clean(path.Join(ps...))
}

// ValidatePath checks if a file path is legal
func ValidatePath(path string) error {
	path = Clean(path)
	match, _ := regexp.MatchString("^[ -~]+$", path)

	if !match {
		return errors.Errorf("path (%v) invalid: only printable ASCII characters allowed", path)
	}

	if IsGlob(path) {
		return errors.Errorf("path (%v) invalid: globbing character (%v) not allowed in path", path, globRegex.FindString(path))
	}
	return nil
}

// IsGlob checks if the pattern contains a glob character
func IsGlob(pattern string) bool {
	pattern = Clean(pattern)
	return globRegex.Match([]byte(pattern))
}

// GlobLiteralPrefix returns the prefix before the first glob character
func GlobLiteralPrefix(pattern string) string {
	pattern = Clean(pattern)
	idx := globRegex.FindStringIndex(pattern)
	if idx == nil {
		return pattern
	}
	return pattern[:idx[0]]
}
