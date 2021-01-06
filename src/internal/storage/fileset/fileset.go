package fileset

import (
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/index"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
)

// ID is the unique identifier for a fileset
// TODO: change this to [16]byte, add ParseID, and HexString methods.
type ID = string

func newID() ID {
	return uuid.NewWithoutDashes()
}

func (p *Primitive) PointsTo() []chunk.ID {
	var ids []chunk.ID
	for _, id := range index.PointsTo(p.Additive) {
		ids = append(ids, id)
	}
	for _, id := range index.PointsTo(p.Deletive) {
		ids = append(ids, id)
	}
	return ids
}

func (c *Composite) PointsTo() []ID {
	return c.Layers
}
