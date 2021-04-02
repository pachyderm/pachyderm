package server

import (
	"regexp"
	"strings"

	globlib "github.com/pachyderm/ohmyglob"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset/index"
)

var globRegex = regexp.MustCompile(`[*?[\]{}!()@+^]`)

func globLiteralPrefix(glob string) string {
	idx := globRegex.FindStringIndex(glob)
	if idx == nil {
		return glob
	}
	return glob[:idx[0]]
}

func parseGlob(glob string) (index.Option, func(string) bool, error) {
	glob = cleanPath(glob)
	opt := index.WithPrefix(globLiteralPrefix(glob))
	g, err := globlib.Compile(glob, '/')
	if err != nil {
		return nil, nil, err
	}
	mf := func(path string) bool {
		// TODO: This does not seem like a good approach for this edge case.
		if path == "/" && glob == "/" {
			return true
		}
		path = strings.TrimRight(path, "/")
		return g.Match(path)
	}
	return opt, mf, nil
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
func cleanPath(x string) string {
	return "/" + strings.Trim(x, "/")
}
