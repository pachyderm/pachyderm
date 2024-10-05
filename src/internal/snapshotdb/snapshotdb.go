package snapshotdb

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/snapshot"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"time"
)

type snapshotRecord struct {
	ID               int64           `db:"id"`
	ChunksetID       int64           `db:"chunkset_id"`
	SQLDumpFilesetID *uuid.UUID      `db:"sql_dump_fileset_id"`
	Metadata         json.RawMessage `db:"metadata"`
	PachydermVersion string          `db:"pachyderm_version"`
	CreatedAt        time.Time       `db:"created_at"`
}

func (r snapshotRecord) toSnapshot() (snapshot.Snapshot, error) {
	var metadata map[string]string
	if err := json.Unmarshal(r.Metadata, &metadata); err != nil {
		return snapshot.Snapshot{}, errors.New("failed to unmarshal metadata: " + err.Error())
	}

	// Construct the Snapshot from the snapshotRecord
	s := snapshot.Snapshot{
		ID:               snapshot.SnapshotID(r.ID),
		ChunksetID:       fileset.ChunkSetID(r.ChunksetID),
		SQLDumpFilesetID: r.SQLDumpFilesetID,
		Metadata:         metadata,
		PachydermVersion: r.PachydermVersion,
		CreatedAt:        r.CreatedAt,
	}

	return s, nil
}
