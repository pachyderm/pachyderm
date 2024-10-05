package snapshot

import (
	"github.com/google/uuid"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
)

type SnapshotID int64

type Snapshot struct {
	ID               SnapshotID
	ChunksetID       fileset.ChunkSetID
	SQLDumpFilesetID *uuid.UUID
	Metadata         map[string]string
	PachydermVersion string
	CreatedAt        time.Time
}
