// Package snapshotdb implements subsystem-independent disaster recovery database CRUD.
package snapshotdb

import (
	"database/sql"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/pgjsontypes"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	snapshotserver "github.com/pachyderm/pachyderm/v2/src/snapshot"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type snapshotRecord struct {
	ID               int64                 `db:"id"`
	Chunkset         int64                 `db:"chunkset"`
	SQLDumpPin       sql.NullInt64         `db:"sql_dump_pin"`
	Metadata         pgjsontypes.StringMap `db:"metadata"`
	PachydermVersion string                `db:"pachyderm_version"`
	CreatedAt        time.Time             `db:"created_at"`
}

func (r snapshotRecord) toSnapshot() snapshot {
	// Construct the Snapshot from the snapshotRecord
	s := snapshot{
		ID:               snapshotID(r.ID),
		ChunksetID:       fileset.ChunkSetID(r.Chunkset),
		SQLDumpPin:       fileset.Pin(r.SQLDumpPin.Int64),
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
	SQLDumpPin       fileset.Pin
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
	return &info
}

type InternalSnapshotInfo struct {
	SQLDumpPin fileset.Pin
}

func (s snapshot) toInteralSnapshotInfo() *InternalSnapshotInfo {
	return &InternalSnapshotInfo{SQLDumpPin: s.SQLDumpPin}
}
