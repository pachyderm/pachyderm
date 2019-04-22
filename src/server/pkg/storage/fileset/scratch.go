package fileset

import (
	"io"

	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
)

const (
	Scratch = "scratch"
)

// Scratch is a scratch space that stores serialized
// fileset parts on disk, and will merge and write
// fileset parts to object storage if they use up too much
// space on disk.
// This is NOT safe for concurrent operations.
type Scratch struct {
	client        obj.Client
	dir           string
	diskAvailable int64
	// Parts is a simple incrementing value that is used as the file/object base name to keep track of the
	// relative order of the fileset parts as they are added to disk/object storage.
	// Data operations in higher number parts are applied after lower number parts during the merge step.
	parts int64
}

func NewScratch(diskAvailable int64) *Scratch {
	// dir is scratch prefix + uuid
	return nil
}

// Notes:
// - Shuffle step could simply pass in the reader for the GetChunk request here.
func (s *Scratch) Put(r io.Reader) error {
	// Write to disk.
	//   If the write goes past the disk threshold, merge everything on disk
	//   into one fileset part in object storage. Merging everything makes
	//   it easier to keep track of the ordering of operations during
	//   the final merge step.
	//   Will need to be careful about handling logic for the partially
	//   written file that pushed us over the disk threshold.
	//   We may want to write an index and deduplicate here.
}

func (s *Scratch) Merge(name string) error {
	// Merge everything under dir on disk and in object storage into object storage path name.
}
