package server

import (
	"path"
	"regexp"
	"strings"

	globlib "github.com/pachyderm/ohmyglob"
)

var globRegex = regexp.MustCompile(`[*?[\]{}!()@+^]`)

func globLiteralPrefix(glob string) string {
	idx := globRegex.FindStringIndex(glob)
	if idx == nil {
		return glob
	}
	return glob[:idx[0]]
}

func globMatchFunction(glob string) (func(string) bool, error) {
	g, err := globlib.Compile(glob, '/')
	if err != nil {
		return nil, err
	}
	return func(path string) bool {
		// TODO: This does not seem like a good approach for this edge case.
		if path == "/" && glob == "/" {
			return true
		}
		path = strings.TrimRight(path, "/")
		return g.Match(path)
	}, nil
}

// pathIsChild determines if the path child is an immediate child of the path parent
// it assumes cleaned paths
func pathIsChild(parent, child string) bool {
	if !strings.HasPrefix(child, parent) {
		return false
	}
	rel := child[len(parent):]
	rel = strings.Trim(rel, "/")
	return !strings.Contains(rel, "/")
}

// cleanPath converts paths to a canonical form used in the driver
// "" -> "/"
// "abc" -> "/abc"
// "/abc" -> "/abc"
// "abc/" -> "/abc"
// "/" -> "/"
func cleanPath(p string) string {
	p = path.Clean(p)
	if p == "." {
		return "/"
	}
	return "/" + strings.Trim(p, "/")
}
