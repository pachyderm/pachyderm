package fileset

import (
	"context"
	"database/sql"
	"testing"

	"github.com/gogo/protobuf/proto"
	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

var _ MetadataStore = &postgresStore{}

type postgresStore struct {
	db *sqlx.DB
}

// NewPostgresStore returns a Store backed by db
func NewPostgresStore(db *sqlx.DB) MetadataStore {
	return &postgresStore{db: db}
}

func (s *postgresStore) DB() *sqlx.DB {
	return s.db
}

func (s *postgresStore) SetTx(tx *sqlx.Tx, id ID, md *Metadata) error {
	if md == nil {
		md = &Metadata{}
	}
	data, err := proto.Marshal(md)
	if err != nil {
		return err
	}
	res, err := tx.Exec(
		`INSERT INTO storage.filesets (path, metadata_pb)
		VALUES ($1, $2)
		ON CONFLICT (path) DO NOTHING
		`, id.HexString(), data)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrFileSetExists
	}
	return nil
}

func (s *postgresStore) Get(ctx context.Context, id ID) (*Metadata, error) {
	var mdData []byte
	if err := s.db.GetContext(ctx, &mdData, `SELECT metadata_pb FROM storage.filesets WHERE path = $1`, id.HexString()); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrFileSetNotExists
		}
		return nil, err
	}
	md := &Metadata{}
	if err := proto.Unmarshal(mdData, md); err != nil {
		return nil, err
	}
	return md, nil
}

func (s *postgresStore) DeleteTx(tx *sqlx.Tx, id ID) error {
	_, err := tx.Exec(`DELETE FROM storage.filesets WHERE path = $1`, id.HexString())
	return err
}

const schema = `
	CREATE TABLE IF NOT EXISTS storage.filesets (
		path VARCHAR(250) PRIMARY KEY,
		metadata_pb BYTEA NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
`

// SetupPostgresStoreV0 sets up the tables for a Store
func SetupPostgresStoreV0(ctx context.Context, tx *sqlx.Tx) error {
	_, err := tx.ExecContext(ctx, schema)
	return err
}

// NewTestStore returns a Store scoped to the lifetime of the test.
func NewTestStore(t testing.TB, db *sqlx.DB) MetadataStore {
	ctx := context.Background()
	tx := db.MustBegin()
	tx.MustExec(`CREATE SCHEMA IF NOT EXISTS storage`)
	require.NoError(t, SetupPostgresStoreV0(ctx, tx))
	require.NoError(t, tx.Commit())
	return NewPostgresStore(db)
}
