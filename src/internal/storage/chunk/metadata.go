package chunk

import (
	"context"
	"encoding/hex"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachhash"
)

// ID uniquely identifies a chunk. It is the hash of its content
type ID []byte

// Hash produces an ID by hashing data
func Hash(data []byte) ID {
	sum := pachhash.Sum(data)
	return sum[:]
}

// IDFromHex parses a hex string into an ID
func IDFromHex(h string) (ID, error) {
	return hex.DecodeString(h)
}

// HexString hex encodes the ID
func (id ID) HexString() string {
	return hex.EncodeToString(id)
}

// TrackerID returns an ID for use with the tracker.
func (id ID) TrackerID() string {
	return TrackerPrefix + id.HexString()
}

// Metadata holds metadata about a chunk
type Metadata struct {
	Size     int
	PointsTo []ID
}

var (
	// ErrMetadataExists metadata exists
	ErrMetadataExists = errors.Errorf("metadata exists")
	// ErrChunkNotExists chunk does not exist
	ErrChunkNotExists = errors.Errorf("chunk does not exist")
)

// Entry is an chunk object mapping
type Entry struct {
	ChunkID   ID     `db:"chunk_id"`
	Gen       uint64 `db:"gen"`
	Uploaded  uint64 `db:"uploaded"`
	Tombstone uint64 `db:"tombstone"`
}

// SetupPostgresStoreV0 sets up tables in db
func SetupPostgresStoreV0(ctx context.Context, tx *sqlx.Tx) error {
	_, err := tx.ExecContext(ctx, `
	CREATE TABLE storage.chunk_objects (
		chunk_id BYTEA NOT NULL,
		gen BIGSERIAL NOT NULL,
		uploaded BOOLEAN NOT NULL DEFAULT FALSE,
		tombstone BOOLEAN NOT NULL DEFAULT FALSE,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

		PRIMARY KEY(chunk_id, gen)
	);
	`)
	return err
}
