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
// new HashTree, create an OpenHashTree with NewHashTree(), modify it, and then
// call Finish() on it.
type HashTree interface {
	// Open makes a deep copy of the HashTree and returns the copy
	Open() OpenHashTree

	// Get retrieves a file.
	Get(path string) (*NodeProto, error)

	// List retrieves the list of files and subdirectories of the directory at
	// 'path'.
	List(path string) ([]*NodeProto, error)

	// Glob returns a list of files and directories that match 'pattern'.
	Glob(pattern string) ([]*NodeProto, error)

	// FSSize gets the size of the file system that this tree represents.
	// It's essentially a helper around h.Get("/").SubtreeBytes
	FSSize() int64

	// Walk calls a given function against every node in the hash tree.
	// The order of traversal is not guaranteed.  If any invocation of the
	// function returns an error, the walk stops and returns the error.
	Walk(func(path string, node *NodeProto) error) error

	// Diff returns a the diff of 2 HashTrees at particular Paths. It takes a
	// callback function f, which will be called with paths that are not
	// identical to the same path in the other HashTree.
	Diff(oldHashTree HashTree, newPath string, oldPath string, f func(path string, node *NodeProto, new bool) error) error
}

// OpenNode is similar to NodeProto, except that it doesn't include the Hash
// field (which is not generally meaningful in an OpenHashTree)
type OpenNode struct {
	Name string
	Size int64

	FileNode *FileNodeProto
	DirNode  *DirectoryNodeProto
}

// OpenHashTree is like HashTree, except that it can be modified. Once an
// OpenHashTree is Finish()ed, the hash and size stored with each node will be
// updated (until then, the hashes and sizes stored in an OpenHashTree will be
// stale).
type OpenHashTree interface {
	HashTree
	// GetOpen retrieves a file.
	GetOpen(path string) (*OpenNode, error)

	// PutFile appends data to a file (and creates the file if it doesn't exist).
	PutFile(path string, objects []*pfs.Object, size int64) error

	// PutDir creates a directory (or does nothing if one exists).
	PutDir(path string) error

	// DeleteFile deletes a regular file or directory (along with its children).
	DeleteFile(path string) error

	// Merge adds all of the files and directories in each tree in 'trees' into
	// this tree. If it errors this tree will be left in a undefined state and
	// should be discarded. If you'd like to be able to revert to the previous
	// state of the tree you should Finish and then Open the tree.
	Merge(trees ...HashTree) error

	// Finish makes a deep copy of the OpenHashTree, updates all of the hashes and
	// node size metadata in the copy, and returns the copy
	Finish() (HashTree, error)
}
