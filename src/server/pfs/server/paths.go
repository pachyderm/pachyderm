package server

import (
	"path"
	"regexp"
	"strings"

	globlib "github.com/pachyderm/ohmyglob"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/index"
)

var globRegex = regexp.MustCompile(`[*?[\]{}!()@+^]`)

func globLiteralPrefix(glob string) string {
	idx := globRegex.FindStringIndex(glob)
	if idx == nil {
		return glob
	}
	return glob[:idx[0]]
}

func matchFunc(glob string) (func(string) bool, error) {
	// TODO Need to do a review of the globbing functionality.
	var parentG *globlib.Glob
	parentGlob, baseGlob := path.Split(glob)
	if len(baseGlob) > 0 {
		var err error
		parentG, err = globlib.Compile(parentGlob, '/')
		if err != nil {
			return nil, err
		}
	}
	g, err := globlib.Compile(glob, '/')
	if err != nil {
		return nil, err
	}
	return func(s string) bool {
		s = path.Clean(s)
		return g.Match(s) && (parentG == nil || !parentG.Match(s))
	}, nil
}

func parseGlob(glob string) (index.Option, func(string) bool, error) {
	g, err := globlib.Compile(glob, '/')
	if err != nil {
		return nil, nil, err
	}
	prefix := globLiteralPrefix(glob)
	return index.WithPrefix(prefix), g.Match, nil
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
