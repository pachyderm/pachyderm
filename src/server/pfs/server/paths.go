package server

import (
	"regexp"
	"strings"

	globlib "github.com/pachyderm/ohmyglob"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
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
		return nil, errors.EnsureStack(err)
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

var validRangeRegex = regexp.MustCompile("^[ -~]+$")

func ValidateFilename(p string) error {
	switch {
	case !validRangeRegex.Match([]byte(p)):
		return errors.Errorf("path (%v) invalid: only printable ASCII characters allowed", p)
	case globRegex.Match([]byte(p)):
		return errors.Errorf("path (%v) invalid: globbing character (%v) not allowed in path", p, globRegex.FindString(p))
	case strings.HasSuffix(p, "/"):
		return errors.Errorf("path (%v) invalid: trailing slash not allowed in path", p)
	}
	for _, elem := range strings.Split(p, "/") {
		if elem == "." || elem == ".." {
			return errors.Errorf("path (%v) invalid: relative file paths are not allowed", p)
		}
	}
	return nil
}
