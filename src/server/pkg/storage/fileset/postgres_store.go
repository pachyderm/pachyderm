package fileset

import (
	"context"
	"database/sql"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/jmoiron/sqlx"
)

var _ Store = &postgresStore{}

type postgresStore struct {
	db *sqlx.DB
}

// NewPostgresStore returns a Store backed by db
func NewPostgresStore(db *sqlx.DB) Store {
	return &postgresStore{db: db}
}

func (s *postgresStore) Set(ctx context.Context, p string, md *Metadata) error {
	if md == nil {
		md = &Metadata{}
	}
	data, err := proto.Marshal(md)
	if err != nil {
		return err
	}
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO storage.filesets (path, metadata_pb)
		VALUES ($1, $2)
		ON CONFLICT (path) DO NOTHING
		`, p, data)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrPathExists
	}
	return nil
}

func (s *postgresStore) Get(ctx context.Context, p string) (*Metadata, error) {
	var mdData []byte
	if err := s.db.GetContext(ctx, &mdData, `SELECT metadata_pb FROM storage.filesets WHERE path = $1`, p); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrPathNotExists
		}
		return nil, err
	}
	md := &Metadata{}
	if err := proto.Unmarshal(mdData, md); err != nil {
		return nil, err
	}
	return md, nil
}

func (s *postgresStore) Walk(ctx context.Context, prefix string, cb func(string) error) (retErr error) {
	rows, err := s.db.QueryContext(ctx, `SELECT path from storage.filesets WHERE path LIKE $1 || '%'`, prefix)
	if err != nil {
		return err
	}
	defer func() {
		if err := rows.Close(); retErr == nil {
			retErr = err
		}
	}()
	var p string
	for rows.Next() {
		if err := rows.Scan(&p); err != nil {
			return err
		}
		if err := cb(p); err != nil {
			return err
		}
	}
	return rows.Err()
}

func (s *postgresStore) Delete(ctx context.Context, p string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM storage.filesets WHERE path = $1`, p)
	return err
}

const schema = `
	CREATE SCHEMA IF NOT EXISTS storage;

	CREATE TABLE IF NOT EXISTS storage.filesets (
		path VARCHAR(250) PRIMARY KEY,
		metadata_pb BYTEA NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
`

// SetupPostgresStore sets up the tables for a Store
func SetupPostgresStore(db *sqlx.DB) {
	db.MustExec(schema)
}

// NewTestStore returns a Store scoped to the lifetime of the test.
func NewTestStore(t testing.TB, db *sqlx.DB) Store {
	SetupPostgresStore(db)
	return NewPostgresStore(db)
}
