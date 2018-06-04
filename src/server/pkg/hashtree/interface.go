package hashtree

import (
	"io"

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

	// CannotSerialize is returned when Serialize(io.Writer) fails, normally
	// due to it being called on an OpenHashTree
	CannotSerialize

	// CannotDeserialize is returned when Deserialize(bytes) fails, perhaps due to
	// 'bytes' being corrupted. Or it being called on an OpenHashTree.
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
	// Read methods
	// Get retrieves a file.
	Get(path string) (*NodeProto, error)

	// List retrieves the list of files and subdirectories of the directory at
	// 'path'.
	List(path string) ([]*NodeProto, error)

	// Glob returns a map of file/directory paths to nodes that match 'pattern'.
	Glob(pattern string) (map[string]*NodeProto, error)

	// FSSize gets the size of the file system that this tree represents.
	// It's essentially a helper around h.Get("/").SubtreeBytes
	FSSize() int64

	// Walk calls a given function against every node in the hash tree.
	// The order of traversal is not guaranteed.  If any invocation of the
	// function returns an error, the walk stops and returns the error.
	Walk(path string, f func(path string, node *NodeProto) error) error

	// Diff returns the diff of 2 HashTrees at particular Paths. It takes a
	// callback function f, which will be called with paths that are not
	// identical to the same path in the other HashTree.
	// Specify '-1' for fully recursive, or '1' for shallow diff
	Diff(oldHashTree HashTree, newPath string, oldPath string, recursiveDepth int64, f func(path string, node *NodeProto, new bool) error) error

	// Serialize serializes a binary version of the HashTree to w.
	Serialize(w io.Writer) error

	// Copy returns a copy of the HashTree
	Copy() (HashTree, error)

	// Write methods

	// PutFile appends data to a file (and creates the file if it doesn't exist).
	PutFile(path string, objects []*pfs.Object, size int64) error

	// PutFileOverwrite is the same as PutFile, except that instead of
	// appending the objects to the end of the given file, the objects
	// are inserted to the given index, and the existing objects starting
	// from the given index are removed.
	//
	// sizeDelta is the delta between the size of the objects added and
	// the size of the objects removed.
	PutFileOverwrite(path string, objects []*pfs.Object, overwriteIndex *pfs.OverwriteIndex, sizeDelta int64) error

	// PutDir creates a directory (or does nothing if one exists).
	PutDir(path string) error

	// DeleteFile deletes a regular file or directory (along with its children).
	DeleteFile(path string) error

	// Merge adds all of the files and directories in each tree in 'trees' into
	// this tree. If it errors this tree will be left in a undefined state and
	// should be discarded. If you'd like to be able to revert to the previous
	// state of the tree you should Finish and then Open the tree.
	Merge(trees ...HashTree) error

	// Hash updates all of the hashes and node size metadata, it also checks
	// for conflicts.
	Hash() error

	// Deserialize deserializes a HashTree from r, into the receiver of the function.
	Deserialize(r io.Reader) error
}
