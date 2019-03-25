package files

import (
	"bytes"
	"io"

	"github.com/pachyderm/pachyderm/src/server/pkg/storage/obj"
)

// A set of files (this may be a subset of a full commit, e.g. a datum/shard)
type Files struct {
	root         string
	memAvailable int64
	id           string
	// Parts represents the parts of the files written to scratch space (inserts went past the in-memory threshold).
	// The ids for the parts are [0, parts), with no parts when parts == 0.
	// Operations in higher number parts happen after lower number parts during the merge.
	parts   int64
	ops     map[string][]*DataOp
	data    map[string]*bytes.Buffer
	scratch obj.Storage
}

func (f *Files) New(root string, memAvailable int64, scratch obj.Storage) *Files {
	// Sets up files with mem available and scratch space.
	// Supports arbitrary rooting (puts are rooted at root).
	return nil
}

func (f *Files) Put(path string, r io.Reader) error {
	// Append put operation to file ops.
	// If past in-memory threshold, serialize and store in scratch space (also increment parts).
	return nil
}

func (f *Files) Overwrite(path string, r io.Reader) error {
	// Append overwrite operation to file ops.
	// If past in-memory threshold, serialize and store in scratch space (also increment parts).
	return nil
}

func (f *Files) Delete(path string) error {
	// Append delete operation to file ops.
	return nil
}

func (f *Files) serialize() error {
	// Sort paths, apply operations, then serialize files to on-disk format.
	return nil
}

func (f *Files) Finish(tag ...string) error {
	// Merge memory, disk, object storage trees into object storage.
	// Create file/chunk index.
	// should we re-think tagging? don't want to go to object storage for each datum.
	return nil
}
