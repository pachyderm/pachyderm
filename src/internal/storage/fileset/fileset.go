package fileset

import (
	"context"
	"io"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset/index"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
)

// ID is the unique identifier for a fileset
// TODO: change this to [16]byte, add ParseID, and HexString methods.
type ID string

func newID() ID {
	return ID(uuid.NewWithoutDashes())
}

// ParseID parses a string into an ID or returns an error
func ParseID(x string) (*ID, error) {
	if len(x) < 32 {
		return nil, errors.Errorf("string (%v) too short to be ID", x)
	}
	id := ID(x)
	return &id, nil
}

// HexString returns the ID encoded with the hex alphabet.
func (id ID) HexString() string {
	return string(id)
}

// PointsTo returns a slice of the chunk.IDs which this fileset immediately points to.
// Transitively reachable chunks are not included in the slice.
func (p *Primitive) PointsTo() []chunk.ID {
	var ids []chunk.ID
	ids = append(ids, index.PointsTo(p.Additive)...)
	ids = append(ids, index.PointsTo(p.Deletive)...)
	return ids
}

// PointsTo returns the IDs of the filesets which this composite fileset points to
func (c *Composite) PointsTo() []ID {
	ids := make([]ID, len(c.Layers))
	for i := range c.Layers {
		ids[i] = ID(c.Layers[i])
	}
	return ids
}

// File represents a file.
type File interface {
	// Index returns the index for the file.
	Index() *index.Index
	// Content writes the content of the file.
	Content(w io.Writer) error
}

var _ File = &MergeFileReader{}
var _ File = &FileReader{}

// FileSet represents a set of files.
type FileSet interface {
	// Iterate iterates over the files in the file set.
	Iterate(ctx context.Context, cb func(File) error, deletive ...bool) error
	// TODO: Implement IterateDeletes or pull deletion information out of the fileset API.
}

var _ FileSet = &MergeReader{}
var _ FileSet = &Reader{}

type emptyFileSet struct{}

func (efs emptyFileSet) Iterate(ctx context.Context, cb func(File) error, deletive ...bool) error {
	return nil
}

func stringsToIDs(xs []string) []ID {
	ids := make([]ID, len(xs))
	for i := range xs {
		ids[i] = ID(xs[i])
	}
	return ids
}

func idsToHex(xs []ID) []string {
	ys := make([]string, len(xs))
	for i := range xs {
		ys[i] = xs[i].HexString()
	}
	return ys
}
