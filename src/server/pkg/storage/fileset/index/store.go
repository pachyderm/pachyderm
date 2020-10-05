package index

import (
	"context"
	"database/sql"
	"testing"
	"time"

	proto "github.com/gogo/protobuf/proto"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/server/pkg/dbutil"
)

var (
	ErrPathNotExists = errors.Errorf("path does not exist")
)

type Store interface {
	PutIndex(ctx context.Context, p string, idx *Index, ttl time.Duration) error
	GetIndex(ctx context.Context, p string) (*Index, error)
	Delete(ctx context.Context, p string) error
	Walk(ctx context.Context, prefix string, cb func(string) error) error

	SetTTL(ctx context.Context, p string, ttl time.Duration) (time.Time, error)
	GetExpiresAt(ctx context.Context, p string) (time.Time, error)
}

func Copy(ctx context.Context, src, dst Store, srcPath, dstPath string) error {
	idx, err := src.GetIndex(ctx, srcPath)
	if err != nil {
		return err
	}
	return dst.PutIndex(ctx, dstPath, idx, 0)
}

func WithTestStore(t *testing.T, cb func(Store)) {
	dbutil.WithTestDB(t, func(db *sqlx.DB) {
		PGStoreApplySchema(db)
		s := NewPGStore(db)
		cb(s)
	})
}

var _ Store = &PGStore{}

// PGStore is an index store backed by postgres
type PGStore struct {
	db *sqlx.DB
}

func NewPGStore(db *sqlx.DB) *PGStore {
	return &PGStore{db: db}
}

func (s *PGStore) PutIndex(ctx context.Context, p string, idx *Index, ttl time.Duration) error {
	data, err := proto.Marshal(idx)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO storage.paths (path, index_pb) VALUES ($1, $2)`, p, data)
	return err
}

func (s *PGStore) GetIndex(ctx context.Context, p string) (*Index, error) {
	var indexData []byte
	if err := s.db.GetContext(ctx, &indexData, `SELECT index_pb FROM storage.paths WHERE path = $1`, p); err != nil {
		return nil, err
	}
	idx := &Index{}
	if err := proto.Unmarshal(indexData, idx); err != nil {
		return nil, err
	}
	return idx, nil
}

func (s *PGStore) Walk(ctx context.Context, prefix string, cb func(string) error) error {
	panic("not implemented")
}

func (s *PGStore) Delete(ctx context.Context, p string) error {
	panic("not implemented")
}

func (s *PGStore) SetTTL(ctx context.Context, prefix string, ttl time.Duration) (time.Time, error) {
	panic("not implemented")
}

func (s *PGStore) GetExpiresAt(ctx context.Context, p string) (time.Time, error) {
	panic("not implemented")
}

const schema = `
	CREATE SCHEMA IF NOT EXISTS storage;

	CREATE TABLE storage.paths (
		path VARCHAR(250) PRIMARY KEY,
		index_pb BYTEA NOT NULL,
		root_chunk INT8,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
		expires_at TIMESTAMP
	);
`

type pathRow struct {
	Path      string       `db:"path"`
	IndexPB   []byte       `db:"index_pb"`
	CreatedAt time.Time    `db:"created_at"`
	ExpiresAt sql.NullTime `db:"expires_at"`
}

func PGStoreApplySchema(db *sqlx.DB) {
	db.MustExec(schema)
}
