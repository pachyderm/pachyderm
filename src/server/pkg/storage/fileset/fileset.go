package fileset

import (
	"bytes"
	"io"

	"github.com/pachyderm/pachyderm/src/server/pkg/storage/obj"
)

// Notes:
// - I am drawing a distinction between parts of a fileset and the merge log because I think it makes sense to require
// the full merging of data within a commit. The merge log will apply across commit boundaries. This also makes the finishing
// step a bit easier because we simply merge everything in memory and in the scratch space, then maybe merge with the parent commits
// based on some heuristics.
// - This should be used primarily for a fileset that requires Put/Overwrite/Delete operations (input repos, uploading datum/chunk data).
// For merging multiple serialized filesets, a Scratch and/or Merge operation should be used.
// - Example structure in object storage:
// /pfs/repo/commit/shard-0/fileset
// /pfs/repo/commit/shard-0/index
// /pfs/repo/commit/shard-0/mergelog
// /pfs/repo/commit/shard-1/fileset
// /pfs/repo/commit/shard-1/index
// /pfs/repo/commit/shard-1/mergelog
// /pfs/chunks/8cfd
// /pfs/chunks/ef7d
// /pfs/chunks/823a
// /pfs/chunks/b54d
// - Handling of shards will happen outside of the storage layer (does not make sense for storage layer to understand
// workers and how to filter chunk data from them.

const (
	FileSet = "fileset"
)

// FileSet is a set of files.
// This may be a full filesystem or a subfilesystem (e.g. datum/chunk/shard).
type FileSet struct {
	client           obj.Client
	root             string
	memThreshold     int64
	name, parentName string
	// A sequence of data operations represent the state of a file.
	// Data operations stored in this map reference the data map.
	ops     map[string][]*DataOp
	data    map[string]*bytes.Buffer
	scratch Scratch
}

func (f *FileSet) New(name string, opts ...Option) *FileSet {
	// Apply opts, no scratch means everything happens in memory.
	// Supports arbitrary rooting (puts are rooted at root).
	return nil
}

func (f *FileSet) Put(path string, r io.Reader) error {
	// Append put operation to file ops.
	// If past in-memory threshold, serialize and store in scratch space.
	return nil
}

func (f *FileSet) Overwrite(path string, r io.Reader) error {
	// Append overwrite operation to file ops.
	// If past in-memory threshold, serialize and store in scratch space.
	return nil
}

func (f *FileSet) Delete(path string) error {
	// Append delete operation to file ops.
	return nil
}

// serialize will be called whenever the in-memory fileset is past the memory threshold (if it is set).
// A new in-memory fileset will be created for the following operations.
func (f *FileSet) serialize() error {
	// Sort paths, apply operations, then serialize files.
	// Put in the scratch space.
	return nil
}

func (f *FileSet) Finish(tag ...string) error {
	// Serialize in-memory tree to scratch space, then merge scratch space.
	// Call CompactMergeLog with parent name, if it exists.
	// Create file index and deduplicate.
	// should we re-think tagging? don't want to go to object storage for each datum.
	// maybe it makes sense for the tag to be associated with a chunk.
	return nil
}
