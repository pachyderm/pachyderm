// Package snapshotdb implements subsystem-independent disaster recovery database CRUD.
package snapshotdb

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/pgjsontypes"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	snapshotserver "github.com/pachyderm/pachyderm/v2/src/snapshot"
	"google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

type snapshotRecord struct {
	ID               int64                 `db:"id"`
	ChunksetID       int64                 `db:"chunkset_id"`
	SQLDumpFilesetID *fileset.ID           `db:"sql_dump_fileset_id"`
	Metadata         pgjsontypes.StringMap `db:"metadata"`
	PachydermVersion string                `db:"pachyderm_version"`
	CreatedAt        time.Time             `db:"created_at"`
}

func (r snapshotRecord) toSnapshot() snapshot {
	// Construct the Snapshot from the snapshotRecord
	s := snapshot{
		ID:               snapshotID(r.ID),
		ChunksetID:       fileset.ChunkSetID(r.ChunksetID),
		SQLDumpFilesetID: r.SQLDumpFilesetID,
		Metadata:         r.Metadata.Data,
		PachydermVersion: r.PachydermVersion,
		CreatedAt:        r.CreatedAt,
	}

	return s
}

type snapshotID int64

type snapshot struct {
	ID               snapshotID
	ChunksetID       fileset.ChunkSetID
	SQLDumpFilesetID *fileset.ID
	Metadata         map[string]string
	PachydermVersion string
	CreatedAt        time.Time
}

func (s snapshot) toSnapshotInfo() *snapshotserver.SnapshotInfo {
	info := snapshotserver.SnapshotInfo{
		Id:               int64(s.ID),
		ChunksetId:       int64(s.ChunksetID),
		PachydermVersion: s.PachydermVersion,
		CreatedAt:        timestamppb.New(s.CreatedAt),
		Metadata:         s.Metadata,
	}
	if s.SQLDumpFilesetID != nil {
		info.SqlDumpFilesetId = s.SQLDumpFilesetID.HexString()
	}
	return &info
}
