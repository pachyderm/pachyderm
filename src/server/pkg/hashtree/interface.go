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

	// ObjectNotFound is returned when GetObject() is called with an object
	// that doesn't exist.
	ObjectNotFound

	// HeaderFooterConflict is returned when PutFileHeaderFooter is called on a
	// path of the form parent/child, but the DirectoryNode at 'parent' doesn't
	// have a header or footer (headers and footers cannot be added to directories
	// retroactively, as that would require modifying all of the directory's
	// children to indicate that they include header data in their parent)
	HeaderFooterConflict

	// MixedObjectsAndBlockRefs is returned when PutFile attempts to append
	// Objects to a file that already has BlockRefs or BlockRefs to a file that
	// already has Objects. This generally happens when you user tries to copy
	// from an output file to an input file that already has content, or tries
	// to write to an input file that was created by copying from an output
	// file.
	MixedObjectsAndBlockRefs
)

// HashTree is the signature of a hash tree provided by this library. To get a
// new HashTree, create an OpenHashTree with NewHashTree(), modify it, and then
// call Finish() on it.
type HashTree interface {
	// Read methods
	// Get retrieves a file.
	Get(path string) (*NodeProto, error)

	// List calls f with the files and subdirectories of the directory at 'path'.
	List(path string, f func(node *NodeProto) error) error

	// ListAll is like List but aggregates its results into a slice.
	ListAll(path string) ([]*NodeProto, error)

	// Glob calls f with the file/directory paths and nodes that match 'pattern'.
	Glob(pattern string, f func(path string, node *NodeProto) error) error

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

	// PutDirHeaderFooter creates a directory at 'path' with the given header
	// and/or footer, or updates the header used by the directory at 'path' if
	// one is already set (if a directory was created without header/footer
	// metadata, PutDirHeaderFooter cannot convert it to a header/footer
	// directory, and will simply return an error.
	//
	// If a directory is a header/footer directory, files that it directly
	// contains will inherit the header/footer that the directory stores without
	// incurring any extra storage themselves. Also, this header metadata will be
	// exposed to PFS, allowing for nice semantics there (for example,
	// pfs.GetFile with a glob argument will only append the header data once to
	// the concatenated contents of all matching files, so that uploading
	// header-containing data with PutFileSplit and then downloading it with
	// GetFile+glob yields the original data. Also, pfs.InspectFile will make
	// sure file hashes reflect the header content, even if it changes after the
	// file is initially created and the file's non-header content stays
	// unchanged, etc)
	//
	// Note: If 'header' or 'footer' is empty, then header data will no longer be
	// appended to children of 'file's parent directory, but because
	// HasHeaderFooter will still be set on all of those children, Pachyderm will
	// still check the parent directory for header data in the future).
	PutDirHeaderFooter(path string, header, footer *pfs.Object, headerSize, footerSize int64) error

	// PutFile appends data to a file (and creates the file if it doesn't exist).
	PutFile(path string, objects []*pfs.Object, size int64) error

	// PutFileBlockRefs is like PutFile, but it uses block refs instead of objects.
	PutFileBlockRefs(path string, brs []*pfs.BlockRef, size int64) error

	// PutFileHeaderFooter is the same as PutFile, except that it marks the
	// FileNode at 'path' as having a header stored in its parent directory and
	// validates that the parent directory has the right header/footer data
	// structures (see PutDirHeaderFooter). Note that before calling
	// PutFileHeaderFooter(path, ...), you must call
	// PutDirHeaderFooter(dir(path)) to create the parent directory correctly
	PutFileHeaderFooter(path string, objects []*pfs.Object, size int64) error

	// PutFileOverwrite is the same as PutFile, except that instead of
	// appending the objects to the end of the given file, the objects
	// are inserted to the given index, and the existing objects starting
	// from the given index are removed.
	//
	// sizeDelta is the delta between the size of the objects added and
	// the size of the objects removed.
	PutFileOverwrite(path string, objects []*pfs.Object, overwriteIndex *pfs.OverwriteIndex, sizeDelta int64) error

	// PutFileOverwriteBlockRefs is the same as PutFileOverwrite except that it
	// uses Block Refs instead of objects.
	PutFileOverwriteBlockRefs(path string, brs []*pfs.BlockRef, overwriteIndex *pfs.OverwriteIndex, sizeDelta int64) error

	// PutDir creates a directory (or does nothing if one exists).
	PutDir(path string) error

	// DeleteFile deletes a regular file or directory (along with its children).
	DeleteFile(path string) error

	// Hash updates all of the hashes and node size metadata, it also checks
	// for conflicts.
	Hash() error

	// Deserialize deserializes a HashTree from r, into the receiver of the function.
	Deserialize(r io.Reader) error

	// Destroy cleans up the on disk structures for the hashtree. Further
	// operations on the database will error. Blocks for pending txns.
	Destroy() error
}
