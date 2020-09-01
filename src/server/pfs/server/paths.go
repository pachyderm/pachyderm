package server

import (
	"path"
	"path/filepath"
	"regexp"
	"strings"

	globlib "github.com/pachyderm/ohmyglob"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset"
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

func commitPath(commit *pfs.Commit) string {
	return commitKey(commit)
}

func compactedCommitPath(commit *pfs.Commit) string {
	return path.Join(commitPath(commit), fileset.Compacted)
}

func checkFilePath(path string) error {
	path = filepath.Clean(path)
	if strings.HasPrefix(path, "../") {
		return errors.Errorf("path (%s) invalid: traverses above root", path)
	}
	return nil
}
