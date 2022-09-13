package chunk

import (
	"context"
	"encoding/hex"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachhash"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
)

// ID uniquely identifies a chunk. It is the hash of its content
type ID []byte

// Hash produces an ID by hashing data
func Hash(data []byte) ID {
	sum := pachhash.Sum(data)
	return sum[:]
}

// ParseTrackerID parses a trackerID into a chunk
func ParseTrackerID(trackerID string) (ID, error) {
	if !strings.HasPrefix(trackerID, TrackerPrefix) {
		return nil, errors.Errorf("tracker ID is not for chunk: %q", trackerID)
	}
	return IDFromHex(trackerID[len(TrackerPrefix):])
}

// IDFromHex parses a hex string into an ID
func IDFromHex(h string) (ID, error) {
	res, err := hex.DecodeString(h)
	return res, errors.EnsureStack(err)
}

func (id ID) String() string {
	return id.HexString()
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

// Entry is an chunk object mapping
type Entry struct {
	ChunkID   ID     `db:"chunk_id"`
	Gen       uint64 `db:"gen"`
	Uploaded  bool   `db:"uploaded"`
	Tombstone bool   `db:"tombstone"`
}

// SetupPostgresStoreV0 sets up tables in db
// DO NOT MODIFY THIS FUNCTION
// IT HAS BEEN USED IN A RELEASED MIGRATION
func SetupPostgresStoreV0(tx *pachsql.Tx) error {
	_, err := tx.Exec(`
	CREATE TABLE storage.chunk_objects (
		chunk_id BYTEA NOT NULL,
		gen BIGSERIAL NOT NULL,
		uploaded BOOLEAN NOT NULL DEFAULT FALSE,
		tombstone BOOLEAN NOT NULL DEFAULT FALSE,
		size INT8 NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

		PRIMARY KEY(chunk_id, gen)
	);

	CREATE TABLE storage.keys (
		name VARCHAR(128) NOT NULL,
		data BYTEA NOT NULL,

		PRIMARY KEY(name)
	)
	`)
	return errors.EnsureStack(err)
}

// KeyStore is a store for named secret keys
type KeyStore interface {
	Create(ctx context.Context, name string, data []byte) error
	Get(ctx context.Context, name string) ([]byte, error)
}

type postgresKeyStore struct {
	db *pachsql.DB
}

func NewPostgresKeyStore(db *pachsql.DB) *postgresKeyStore {
	return &postgresKeyStore{
		db: db,
	}
}

func (s *postgresKeyStore) Create(ctx context.Context, name string, data []byte) error {
	_, err := s.db.ExecContext(ctx, `
	INSERT INTO storage.keys (name, data) VALUES ($1, $2)
	`, name, data)
	return errors.EnsureStack(err)
}

func (s *postgresKeyStore) Get(ctx context.Context, name string) ([]byte, error) {
	var data []byte
	if err := s.db.GetContext(ctx, &data, `SELECT data FROM storage.keys WHERE name = $1 LIMIT 1`, name); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return data, nil
}
