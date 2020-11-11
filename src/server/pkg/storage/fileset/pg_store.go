package fileset

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/src/server/pkg/dbutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/errutil"
)

var _ Store = &postgresStore{}

type postgresStore struct {
	db *sqlx.DB
}

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
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO storage.filesets (path, metadata_pb)
		VALUES ($1, $2)
		ON CONFLICT (path) DO UPDATE SET metadata_pb = $2
		`, p, data)
	return err
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
	defer func() { errutil.SetRetErrIfNil(&retErr, rows.Close()) }()
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

type pathRow struct {
	Path       string    `db:"path"`
	MetadataPB []byte    `db:"metadata_pb"`
	CreatedAt  time.Time `db:"created_at"`
}

func SetupPostgresStore(db *sqlx.DB) {
	db.MustExec(schema)
}

func WithTestStore(t testing.TB, cb func(Store)) {
	dbutil.WithTestDB(t, func(db *sqlx.DB) {
		SetupPostgresStore(db)
		s := NewPostgresStore(db)
		cb(s)
	})
}
