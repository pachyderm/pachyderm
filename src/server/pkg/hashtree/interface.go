package hashtree

import (
	"errors"

	"github.com/pachyderm/pachyderm/src/client/pfs"
)

var (
	PathNotFoundErr  = errors.New("path not found")
	MalformedGlobErr = errors.New("glob pattern malformed")

	// this error is returned when of the following occurs:
	// 1. PutFile is called with a path that points to a directory.
	// 2. PutFile is called with a path that contains a prefix that
	// points to a file.
	PathConflictErr = errors.New("path conflict")
)

type Interface interface {
	// Read, Update, and Delete files
	PutFile(path string, blockRefs []*pfs.BlockRef) error
	DeleteFile(path string) error
	Get(path string) (*Node, error)

	// Read, Update, and Delete directories
	PutDir(path string) error
	DeleteDir(path string) error
	List(path string) ([]*Node, error)

	// Returns a list of files and directories that match 'pattern'
	Glob(pattern string) ([]*Node, error)
	// Merges another hash tree into this tree
	Merge(tree Interface) error
}
