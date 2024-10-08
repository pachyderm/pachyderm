// Package snapshotdb implements subsystem-independent disaster recovery database CRUD.
package snapshotdb

import (
	"encoding/json"
	"github.com/google/uuid"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	snapshotserver "github.com/pachyderm/pachyderm/v2/src/snapshot"
	"google.golang.org/protobuf/types/known/timestamppb"
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

func (r snapshotRecord) toSnapshot() (Snapshot, error) {
	var metadata map[string]string
	if err := json.Unmarshal(r.Metadata, &metadata); err != nil {
		return Snapshot{}, errors.New("failed to unmarshal metadata: " + err.Error())
	}

	// Construct the Snapshot from the snapshotRecord
	s := Snapshot{
		ID:               SnapshotID(r.ID),
		ChunksetID:       fileset.ChunkSetID(r.ChunksetID),
		SQLDumpFilesetID: r.SQLDumpFilesetID,
		Metadata:         metadata,
		PachydermVersion: r.PachydermVersion,
		CreatedAt:        r.CreatedAt,
	}

	return s, nil
}

type SnapshotID int64

type Snapshot struct {
	ID               SnapshotID
	ChunksetID       fileset.ChunkSetID
	SQLDumpFilesetID *uuid.UUID
	Metadata         map[string]string
	PachydermVersion string
	CreatedAt        time.Time
}

func (s Snapshot) ToSnapshotInfo() *snapshotserver.SnapshotInfo {
	info := snapshotserver.SnapshotInfo{
		Id:               int64(s.ID),
		ChunksetId:       int64(s.ChunksetID),
		PachydermVersion: s.PachydermVersion,
		CreatedAt:        timestamppb.New(s.CreatedAt),
		Metadata:         s.Metadata,
	}
	if s.SQLDumpFilesetID != nil {
		info.SqlDumpFilesetId = s.SQLDumpFilesetID.String()
	}
	return &info
}
