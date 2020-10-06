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
	ErrNoTTLSet      = errors.Errorf("no ttl set on path")
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

var _ Store = &pgStore{}

type pgStore struct {
	db *sqlx.DB
}

func NewPGStore(db *sqlx.DB) Store {
	return &pgStore{db: db}
}

func (s *pgStore) PutIndex(ctx context.Context, p string, idx *Index, ttl time.Duration) error {
	if idx == nil {
		idx = &Index{}
	}
	data, err := proto.Marshal(idx)
	if err != nil {
		return err
	}
	if ttl <= 0 {
		_, err := s.db.ExecContext(ctx, `INSERT INTO storage.paths (path, index_pb) VALUES ($1, $2)`, p, data)
		return err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO storage.paths (path, index_pb, expires_at) VALUES ($1, $2, CURRENT_TIMESTAMP + interval '1 microsecond' * $3)`, p, data, ttl.Microseconds())
	return err
}

func (s *pgStore) GetIndex(ctx context.Context, p string) (*Index, error) {
	var indexData []byte
	if err := s.db.GetContext(ctx, &indexData, `SELECT index_pb FROM storage.paths WHERE path = $1`, p); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPathNotExists
		}
		return nil, err
	}
	idx := &Index{}
	if err := proto.Unmarshal(indexData, idx); err != nil {
		return nil, err
	}
	return idx, nil
}

func (s *pgStore) Walk(ctx context.Context, prefix string, cb func(string) error) error {
	rows, err := s.db.QueryContext(ctx, `SELECT path from storage.paths WHERE path LIKE $1 || '%'`, prefix)
	if err != nil {
		return err
	}
	defer rows.Close()
	var p string
	for rows.Next() {
		if err := rows.Scan(&p); err != nil {
			return err
		}
		if err := cb(p); err != nil {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	return nil
}

func (s *pgStore) Delete(ctx context.Context, p string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM storage.paths WHERE path = $1`, p)
	return err
}

func (s *pgStore) SetTTL(ctx context.Context, p string, ttl time.Duration) (time.Time, error) {
	var expiresAt time.Time
	err := s.db.GetContext(ctx, &expiresAt, `
	UPDATE storage.paths
	SET expires_at = (CURRENT_TIMESTAMP + interval '1 microsecond' * $1)
	WHERE path = $2
	RETURNING expires_at`, ttl.Microseconds(), p)
	if err != nil {
		return time.Time{}, err
	}
	return expiresAt, nil
}

func (s *pgStore) GetExpiresAt(ctx context.Context, p string) (time.Time, error) {
	var expiresAt sql.NullTime
	if err := s.db.GetContext(ctx, &expiresAt, `SELECT expires_at FROM storage.paths WHERE path = $1`, p); err != nil {
		return time.Time{}, err
	}
	if !expiresAt.Valid {
		return time.Time{}, ErrNoTTLSet
	}
	return expiresAt.Time, nil
}

const schema = `
	CREATE SCHEMA IF NOT EXISTS storage;

	CREATE TABLE IF NOT EXISTS storage.paths (
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
