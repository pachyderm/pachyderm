package chunk

import (
	"context"
	"database/sql"
	"encoding/hex"
	fmt "fmt"
	"regexp"

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

// MetadataStore stores metadata about chunks
type MetadataStore interface {
	DB() *sqlx.DB
	// Set adds chunk metadata to the tracker
	SetTx(tx *sqlx.Tx, chunkID ID, md Metadata) error
	// Get returns info about the chunk if it exists
	Get(ctx context.Context, chunkID ID) (*Metadata, error)
	// Delete removes chunk metadata from the tracker
	DeleteTx(tx *sqlx.Tx, chunkID ID) error
}

var _ MetadataStore = &postgresStore{}

type postgresStore struct {
	db *sqlx.DB
}

// NewPostgresStore returns a Metadata backed by db
func NewPostgresStore(db *sqlx.DB) MetadataStore {
	return &postgresStore{db: db}
}

func (s *postgresStore) DB() *sqlx.DB {
	return s.db
}

func (s *postgresStore) SetTx(tx *sqlx.Tx, chunkID ID, md Metadata) error {
	_, err := tx.Exec(
		`INSERT INTO storage.chunks (hash_id, size) VALUES ($1, $2)
		ON CONFLICT (hash_id) DO NOTHING
		`, chunkID, md.Size)
	return err
}

func (s *postgresStore) Get(ctx context.Context, chunkID ID) (*Metadata, error) {
	type chunkRow struct {
		size int `db:"size"`
	}
	var x chunkRow
	if err := s.db.GetContext(ctx, &x, `SELECT size FROM storage.chunks WHERE hash_id = $1`, chunkID); err != nil {
		if err == sql.ErrNoRows {
			err = ErrChunkNotExists
		}
		return nil, err
	}
	return &Metadata{
		Size: x.size,
	}, nil
}

func (s *postgresStore) DeleteTx(tx *sqlx.Tx, chunkID ID) error {
	_, err := tx.Exec(`DELETE FROM storage.chunks WHERE hash_id = $1`, chunkID)
	return err
}

// SetupPostgresStoreV0 sets up tables in db
func SetupPostgresStoreV0(ctx context.Context, tableName string, tx *sqlx.Tx) error {
	ok, err := regexp.MatchString("[A-z_]+", tableName)
	if err != nil {
		panic(err)
	}
	if !ok {
		panic("invalid table name: " + tableName)
	}
	query := fmt.Sprintf(`
	CREATE TABLE %s (
		hash_id BYTEA NOT NULL PRIMARY KEY,
		size INT8 NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`, tableName)
	_, err = tx.ExecContext(ctx, query)
	return err
}
