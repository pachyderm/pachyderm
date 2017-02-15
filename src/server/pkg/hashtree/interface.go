package hashtree

import (
	"github.com/pachyderm/pachyderm/src/client/pfs"
)

// ErrCode identifies different kinds of errors returned by methods in
// HashTree below. The ErrCode of any such error can be retrieved with Code().
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

	// CannotDeserialize is returned when Deserialize(bytes) fails, perhaps due to
	// 'bytes' being corrupted.
	CannotDeserialize

	// Unsupported is returned when Deserialize(bytes) encounters an unsupported
	// (likely old) serialized HashTree.
	Unsupported

	// PathNotFound is returned when Get() or DeleteFile() is called with a path
	// that doesn't lead to a node.
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

// HashTree is the signature of a hash tree provided by this library. To get a
// new HashTree, create one with NewHashTree().
type HashTree interface {
	// Get retrieves a file.
	Get(path string) (*NodeProto, error)

	// List retrieves the list of files and subdirectories of the directory at
	// 'path'.
	List(path string) ([]*NodeProto, error)

	// Glob returns a list of files and directories that match 'pattern'.
	Glob(pattern string) ([]*NodeProto, error)

	// Serialize serializes a HashTree so that it can be persisted. Also see
	// Deserialize(bytes).
	Serialize() ([]byte, error)

	// PutFile appends data to a file (and creates the file if it doesn't exist).
	PutFile(path string, blockRefs []*pfs.BlockRef) error

	// PutDir creates a directory (or does nothing if one exists).
	PutDir(path string) error

	// DeleteFile deletes a regular file or directory (along with its children).
	DeleteFile(path string) error

	// Merge adds all of the files and directories in each tree in 'trees' into
	// this tree. The effect is equivalent to calling this.PutFile with every
	// file in every tree in 'tree', though the performance may be slightly
	// better.
	Merge(trees []HashTree) error
}
