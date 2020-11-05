package chunk

import (
	"context"
	"crypto/sha512"
	"encoding/hex"

	"github.com/jmoiron/sqlx"
)

type ID []byte

func Hash(data []byte) ID {
	h := sha512.New()
	h.Write(data)
	return h.Sum(nil)[:32]
}

func ChunkIDFromHex(h string) (ID, error) {
	return hex.DecodeString(h)
}

func (id ID) HexString() string {
	return hex.EncodeToString(id)
}

type Metadata struct {
	Size     int
	PointsTo []ID
}

type MetadataStore interface {
	// SetChunkInfo adds chunk metadata to the tracker
	Set(ctx context.Context, chunkID ID, md Metadata) error
	// GetChunkInfo returns info about the chunk if it exists
	Get(ctx context.Context, chunkID ID) (*Metadata, error)
	// DeleteChunkInfo removes chunk metadata from the tracker
	Delete(ctx context.Context, chunkID ID) error
}

var _ MetadataStore = &PostgresStore{}

type PostgresStore struct {
	db *sqlx.DB
}

func NewPostgresStore(db *sqlx.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

func (s *PostgresStore) Set(ctx context.Context, chunkID ID, md Metadata) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO storage.chunks (hash_id, size) VALUES ($1, $2)
		ON CONFLICT (hash_id) DO UPDATE SET size = $2 WHERE storage.chunks.hash_id = $1
		`, chunkID, md.Size)
	return err
}

func (s *PostgresStore) Get(ctx context.Context, chunkID ID) (*Metadata, error) {
	type chunkRow struct {
		size int `db:"size"`
	}
	var x chunkRow
	if err := s.db.GetContext(ctx, &x, `SELECT size FROM storage.chunks WHERE hash_id = $1`, chunkID); err != nil {
		return nil, err
	}
	return &Metadata{
		Size: x.size,
	}, nil
}

func (s *PostgresStore) Delete(ctx context.Context, chunkID ID) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM storage.chunks WHERE hash_id = $1`, chunkID)
	return err
}

func SetupPostgresStore(db *sqlx.DB) {
	db.MustExec(schema)
}

const schema = `
	CREATE SCHEMA IF NOT EXISTS storage;

	CREATE TABLE storage.chunks (
		hash_id BYTEA NOT NULL UNIQUE,
		size INT8 NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);	
`
