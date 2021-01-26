package server

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
)

var globRegex = regexp.MustCompile(`[*?[\]{}!()@+^]`)

func globLiteralPrefix(glob string) string {
	idx := globRegex.FindStringIndex(glob)
	if idx == nil {
		return glob
	}
	return glob[:idx[0]]
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

func checkFilePath(path string) error {
	path = filepath.Clean(path)
	if strings.HasPrefix(path, "../") {
		return errors.Errorf("path (%s) invalid: traverses above root", path)
	}
	return nil
}
