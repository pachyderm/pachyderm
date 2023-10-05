package fuse

import (
	"path/filepath"
	"strings"
)

var Separator = string(filepath.Separator)

// Adds a leading slash and a trailing slash to the path if they don't already exist
func standardizeSlashes(path string) string {
	addTrailingSlash := func(p string) string {
		if len(p) == 0 {
			return Separator
		}
		if p[len(p)-1] != filepath.Separator {
			return p + Separator
		}
		return p
	}
	addLeadingSlash := func(p string) string {
		if len(p) == 0 {
			return Separator
		}
		if p[0] != filepath.Separator {
			return Separator + p
		}
		return p
	}
	return addLeadingSlash(addTrailingSlash(path))
}

// Returns true if path1 is at least a grandparent of path2
// along with the intermediate path between path1 and path2
// (excludes last path component of path2)
func isGrandparentOf(path1, path2 string) (bool, string) {
	path1 = standardizeSlashes(path1)
	path2 = standardizeSlashes(path2)
	relative, err := filepath.Rel(path1, path2)
	if err != nil {
		return false, ""
	}
	parts := strings.Split(relative, string(filepath.Separator))
	if len(parts) < 2 || parts[0] == ".." {
		return false, ""
	}
	intermediate := filepath.Join(parts[:len(parts)-1]...)
	return true, standardizeSlashes(intermediate)
}

// Returns true if path1 is parent of path2
func isParentOf(path1, path2 string) bool {
	path1 = standardizeSlashes(path1)
	path2 = strings.TrimSuffix(standardizeSlashes(path2), Separator) // filepath.Dir doesn't work as expected if path2 ends with a slash
	parentDir := filepath.Dir(path2)
	return path1 == standardizeSlashes(parentDir)
}

// Returns true if path1 is a descendant of or equal to path2
func isDescendantOf(path1, path2 string) bool {
	path1 = standardizeSlashes(path1)
	path2 = standardizeSlashes(path2)
	return strings.HasPrefix(path1, path2)
}
