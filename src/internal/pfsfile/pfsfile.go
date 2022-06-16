package pfsfile

import (
	"path"
	"strings"
)

// CleanPath converts paths to a canonical form used in the driver
// "" -> "/"
// "abc" -> "/abc"
// "/abc" -> "/abc"
// "abc/" -> "/abc"
// "/" -> "/"
func CleanPath(p string) string {
	p = path.Clean(p)
	if p == "." {
		return "/"
	}
	return "/" + strings.Trim(p, "/")
}
