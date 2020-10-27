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

var _ Store = &pgStore{}

type pgStore struct {
	db *sqlx.DB
}

func NewPGStore(db *sqlx.DB) Store {
	return &pgStore{db: db}
}

func (s *pgStore) PutIndex(ctx context.Context, p string, idx *index.Index) error {
	if idx == nil {
		idx = &index.Index{}
	}
	data, err := proto.Marshal(idx)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO storage.paths (path, index_pb)
		VALUES ($1, $2)
		ON CONFLICT (path) DO UPDATE SET index_pb = $2
		`, p, data)
	return err
}

func (s *pgStore) GetIndex(ctx context.Context, p string) (*index.Index, error) {
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

const schema = `
	CREATE SCHEMA IF NOT EXISTS storage;

	CREATE TABLE IF NOT EXISTS storage.paths (
		path VARCHAR(250) PRIMARY KEY,
		index_pb BYTEA NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
`

type pathRow struct {
	Path      string    `db:"path"`
	IndexPB   []byte    `db:"index_pb"`
	CreatedAt time.Time `db:"created_at"`
}

func PGStoreApplySchema(db *sqlx.DB) {
	db.MustExec(schema)
}

func WithTestStore(t *testing.T, cb func(Store)) {
	dbutil.WithTestDB(t, func(db *sqlx.DB) {
		PGStoreApplySchema(db)
		s := NewPGStore(db)
		cb(s)
	})
}
