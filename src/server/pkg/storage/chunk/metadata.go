package chunk

import (
	"context"
	"crypto/sha512"
	"encoding/hex"

	"github.com/jmoiron/sqlx"
)

type ChunkID []byte

func Hash(data []byte) ChunkID {
	h := sha512.New()
	h.Write(data)
	return h.Sum(nil)[:32]
}

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

type MetadataStore interface {
	// SetChunkInfo adds chunk metadata to the tracker
	SetChunkMetadata(ctx context.Context, chunkID ChunkID, md ChunkMetadata) error
	// GetChunkInfo returns info about the chunk if it exists
	GetChunkMetadata(ctx context.Context, chunkID ChunkID) (*ChunkMetadata, error)
	// DeleteChunkInfo removes chunk metadata from the tracker
	DeleteChunkMetadata(ctx context.Context, chunkID ChunkID) error
}

var _ MetadataStore = &PGStore{}

type PGStore struct {
	db *sqlx.DB
}

func NewPGStore(db *sqlx.DB) *PGStore {
	return &PGStore{db: db}
}

func (s *PGStore) SetChunkMetadata(ctx context.Context, chunkID ChunkID, md ChunkMetadata) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO storage.chunks (hash_id, size) VALUES ($1, $2)
		ON CONFLICT DO UPDATE SET size = $2 WHERE id = $1
		`, chunkID, md.Size)
	return err
}

func (s *PGStore) GetChunkMetadata(ctx context.Context, chunkID ChunkID) (*ChunkMetadata, error) {
	type chunkRow struct {
		size int `db:"size"`
	}
	var x chunkRow
	if err := s.db.GetContext(ctx, &x, `SELECT size FROM `); err != nil {
		return nil, err
	}
	return &ChunkMetadata{
		Size: x.size,
	}, nil
}

func (s *PGStore) DeleteChunkMetadata(ctx context.Context, chunkID ChunkID) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM storage.chunks WHERE hash_id = $1`, chunkID)
	return err
}

const schema = `
	CREATE SCHEMA IF NOT EXISTS storage;

	CREATE TABLE storage.chunks (
		id BYTEA(32) NOT NULL UNIQUE,
		size INT8 NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);	
`
