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
	GetFile(path string) (*FileNode, error)
	PutFile(path string, blockRefs []*pfs.BlockRef) error
	ListFile(path string) ([]*FileNode, error)
	GlobFile(pattern string) ([]*Node, error)
}
