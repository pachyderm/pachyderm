package fileset

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset/index"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
)

// ID is the unique identifier for a fileset
// TODO: change this to [16]byte, add ParseID, and HexString methods.
type ID = string

func newID() ID {
	return uuid.NewWithoutDashes()
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
	return c.Layers
}
