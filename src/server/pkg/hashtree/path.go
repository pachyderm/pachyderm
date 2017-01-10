// This is a library that provides some path-cleaning and path manipulation
// functions for hashtree.go. The functions it defines are very similar to the
// function in go's "path" library.

// In both, a canonicalized path has a leading slash and no trailing slash in
// general. The difference is:
//
// - In this library, the canonical version of "/" is "" (i.e. preserve "no
//   trailing slash" invariant, at the cost of the "no leading slash"
//   invariant), whereas
//
// - in go's "path" library, the canonical version of "/" is "/" and "" becomes
//   ".".
//
// We prefer our canonicalizion because it gives us globbing behavior we like:
// "/" and ""   (canonicalize to: "")   match "/" ("") but not "/foo"
// "*" and "/*" (canonicalize to: "/*") match "/foo"   but not "/" ("")

package hashtree

import (
	"path"
	"strings"
)

// Overrides the internal defaults of many functions in the "path" library.
// Specifically, the top-level dir "/" and the special string "." (which is what
// most "path" functions return for the empty string) both map to the empty
// string here, so that we get the globbing behavior we want (see top).
func internalDefault(s string) string {
	if s == "/" || s == "." {
		return ""
	}
	return s
}

// Canonicalizes 'path' for internal use: leading slash and no trailing slash.
// Also,clean the result with internalDefault.
func clean(p string) string {
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return internalDefault(path.Clean(p))
}

// Like path.Base, but uses this library's defaults for canonical paths
func base(p string) string {
	return internalDefault(path.Base(p))
}

// Like path.Dir, but uses this library's defaults for canonical paths
func dir(p string) string {
	return clean(path.Dir(p))
}

// Like path.Split, but uses this library's defaults for canonical paths
func split(p string) (string, string) {
	return dir(p), base(p)
}

// Joints the path components 'ps', using our version of 'clean()' instead of
// path.Clean()
func join(ps ...string) string {
	return clean(path.Join(ps...))
}
