package server

import (
	"strings"

	globlib "github.com/pachyderm/ohmyglob"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage"
)

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

func ValidateFilename(p string) error {
	return storage.ValidateFilename(p)
}
