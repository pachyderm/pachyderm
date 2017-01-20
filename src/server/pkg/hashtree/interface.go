package hashtree

import (
	"github.com/pachyderm/pachyderm/src/client/pfs"
)

// ErrCode identifies different kinds of errors returned by methods in
// Interface below. The ErrCode of any such error can be retrieved with Code().
type ErrCode uint8

const (
	// OK is returned on success
	OK ErrCode = iota

	// Unknown is returned by Code() when an error wasn't emitted by the HashTree
	// implementation.
	Unknown

	// Internal is returned when a HashTree encounters a bug (usually due to the
	// violation of an internal invariant).
	Internal

	// PathNotFound is returned when Get() is called with a path that doesn't
	// lead to a node.
	PathNotFound

	// MalformedGlob is returned when Glob() is called with an invalid glob
	// pattern.
	MalformedGlob

	// PathConflict is returned when a path that is expected to point to a
	// directory in fact points to a file, or the reverse. For example:
	// 1. PutFile is called with a path that points to a directory.
	// 2. PutFile is called with a path that contains a prefix that
	//    points to a file.
	// 3. Merge is forced to merge a directory into a file
	PathConflict
)

// Interface is the signature of a HashTree provided by this library
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
