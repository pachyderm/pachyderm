package fileset

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/src/server/pkg/dbutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/index"
)

var _ PathStore = &pgPathStore{}

type pgPathStore struct {
	db *sqlx.DB
}

func NewPGPathStore(db *sqlx.DB) PathStore {
	return &pgPathStore{db: db}
}

func (s *pgPathStore) PutIndex(ctx context.Context, p string, idx *index.Index, ttl time.Duration) error {
	if idx == nil {
		idx = &index.Index{}
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

func (s *pgPathStore) GetIndex(ctx context.Context, p string) (*index.Index, error) {
	var indexData []byte
	if err := s.db.GetContext(ctx, &indexData, `SELECT index_pb FROM storage.paths WHERE path = $1`, p); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPathNotExists
		}
		return nil, err
	}
	idx := &index.Index{}
	if err := proto.Unmarshal(indexData, idx); err != nil {
		return nil, err
	}
	return idx, nil
}

func (s *pgPathStore) Walk(ctx context.Context, prefix string, cb func(string) error) error {
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

func (s *pgPathStore) Delete(ctx context.Context, p string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM storage.paths WHERE path = $1`, p)
	return err
}

func (s *pgPathStore) SetTTL(ctx context.Context, p string, ttl time.Duration) (time.Time, error) {
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

func (s *pgPathStore) GetExpiresAt(ctx context.Context, p string) (time.Time, error) {
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

func PGPathStoreApplySchema(db *sqlx.DB) {
	db.MustExec(schema)
}

func WithTestPathStore(t *testing.T, cb func(PathStore)) {
	dbutil.WithTestDB(t, func(db *sqlx.DB) {
		PGPathStoreApplySchema(db)
		s := NewPGPathStore(db)
		cb(s)
	})
}
