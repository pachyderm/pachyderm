package chunk

import (
	"context"
	"encoding/hex"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
)

type ChunkID []byte

func ChunkIDFromHex(h string) (ChunkID, error) {
	return hex.DecodeString(h)
}

func (id ChunkID) HexString() string {
	return hex.EncodeToString(id)
}

type ChunkMetadata struct {
	Size     int
	PointsTo []ChunkID
}

var (
	ErrChunkLocked      = errors.Errorf("chunk is already locked in another set")
	ErrChunkSetNotExist = errors.Errorf("chunk set does not exist")
)

// Tracker tracks chunk metadata and keeps track of chunk sets being uploaded or deleted.
type Tracker interface {
	// ChunkSet methods. Keep track of active chunks
	NewChunkSet(ctx context.Context, chunkSet string, delete bool, ttl time.Duration) error
	// SetTTL sets the TTL on a chunkSet
	// If the ChunkSet has already expired, was never created, or has been dropped. SetTTL returns ErrChunkSetNotExist
	SetTTL(ctx context.Context, chunkSet string, ttl time.Duration) (expiresAt time.Time, err error)
	// LockChunk adds a chunk id to a chunk set.
	// LockChunk returns ErrChunkLocked if the chunk is already locked to a delete ChunkSet
	LockChunk(ctx context.Context, chunkSet string, chunkID ChunkID) error
	// Unlock chunk removes a chunk from a chunk set.
	UnlockChunk(ctx context.Context, chunkSet string, chunkID ChunkID) error
	// DropChunkSet causes all chunks in a chunk set to be deleted. It returns the number of chunks dropped.
	// ChunkSets are expired automatically by the tracker, but DropChunkSet will drop it immediately.
	DropChunkSet(ctx context.Context, chunkSet string) (int, error)

	// Chunk metadata
	// SetChunkInfo adds chunk metadata to the tracker
	SetChunkInfo(ctx context.Context, chunkID ChunkID, md ChunkMetadata) error
	// GetChunkInfo returns info about the chunk if it exists
	GetChunkInfo(ctx context.Context, chunkID ChunkID) (*ChunkMetadata, error)
	// DeleteChunkInfo removes chunk metadata from the tracker
	DeleteChunkInfo(ctx context.Context, chunkID ChunkID) error
}

func WithTestTracker(t testing.TB, cb func(tracker Tracker)) {
	cb(nil)
}

type PGTracker struct {
	db *sqlx.DB
}

func NewPGTracker(db *sqlx.DB) *PGTracker {
	return &PGTracker{db: db}
}

func (tr *PGTracker) NewChunkSet(ctx context.Context, chunkSet string, delete bool, ttl time.Duration) (time.Time, error) {
	panic("not implemented")
}

func (tr *PGTracker) SetTTL(ctx context.Context, chunkSet string, ttl time.Duration) (time.Time, error) {
	panic("not implemented")
}

const schema = `
	CREATE SCHEMA IF NOT EXISTS storage;

	CREATE TABLE storage.chunks (
		int_id BIGSERIAL PRIMARY KEY,
		hash_id BYTEA NOT NULL UNIQUE,
		size INT8 NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE storage.chunk_refs (
		from INT8 NOT NULL,
		to INT8 NOT NULL,
		PRIMARY KEY(from, to)
	);

	CREATE TABLE storage.chunk_locks (
		chunkset_id VARCHAR(64) NOT NULL PRIMARY KEY
		chunk_id INT8 NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		expires_at TIMESTAMP NOT NULL
	);
`
