package fileset

import (
	"context"
	"database/sql"
	"testing"
	"time"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/backoff"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/miscutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pacherr"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"google.golang.org/protobuf/proto"
)

var _ MetadataStore = &postgresStore{}

type postgresStore struct {
	db      *pachsql.DB
	cache   *lru.Cache[ID, *Metadata]
	deduper *miscutil.WorkDeduper[ID]
}

// NewPostgresStore returns a Store backed by db
// TODO: Expose configuration for cache size?
func NewPostgresStore(db *pachsql.DB) MetadataStore {
	mc, err := lru.New[ID, *Metadata](100)
	if err != nil {
		panic(err)
	}
	return &postgresStore{
		db:      db,
		cache:   mc,
		deduper: &miscutil.WorkDeduper[ID]{},
	}
}

func (s *postgresStore) DB() *pachsql.DB {
	return s.db
}

func (s *postgresStore) SetTx(tx *pachsql.Tx, id ID, md *Metadata) error {
	if md == nil {
		md = &Metadata{}
	}
	data, err := proto.Marshal(md)
	if err != nil {
		return errors.EnsureStack(err)
	}
	res, err := tx.Exec(
		`INSERT INTO storage.filesets (id, metadata_pb)
		VALUES ($1, $2)
		ON CONFLICT (id) DO NOTHING
		`, id, data)
	if err != nil {
		return errors.EnsureStack(err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return errors.EnsureStack(err)
	}
	if n == 0 {
		return errors.WithStack(ErrFileSetExists)
	}
	return nil
}

func (s *postgresStore) get(ctx context.Context, q sqlx.QueryerContext, id ID) (*Metadata, error) {
	var mdData []byte
	if err := sqlx.GetContext(ctx, q, &mdData, `SELECT metadata_pb FROM storage.filesets WHERE id = $1`, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.WithStack(ErrFileSetNotExists)
		}
		return nil, errors.EnsureStack(err)
	}
	md := &Metadata{}
	if err := proto.Unmarshal(mdData, md); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return md, nil
}

func (s *postgresStore) Get(ctx context.Context, id ID) (*Metadata, error) {
	md, err := s.getFromCache(id)
	if err == nil {
		return md, err
	}
	b := backoff.NewExponentialBackOff()
	b.InitialInterval = 1 * time.Millisecond
	if err := backoff.RetryUntilCancel(ctx, func() error {
		var err error
		md, err = s.getFromCache(id)
		return err
	}, b, func(err error, _ time.Duration) error {
		if !pacherr.IsNotExist(err) {
			return err
		}
		return s.deduper.Do(ctx, id, func() error {
			md, err := s.get(ctx, s.db, id)
			if err != nil {
				return err
			}
			s.putInCache(id, md)
			return nil
		})
	}); err != nil {
		return nil, err
	}
	return md, nil
}

func (s *postgresStore) getFromCache(id ID) (*Metadata, error) {
	md, ok := s.cache.Get(id)
	if !ok {
		return nil, pacherr.NewNotExist("memory-filesets", id.HexString())
	}
	return proto.Clone(md).(*Metadata), nil
}

func (s *postgresStore) putInCache(id ID, md *Metadata) {
	md = proto.Clone(md).(*Metadata)
	s.cache.Add(id, md)
}

func (s *postgresStore) GetTx(tx *pachsql.Tx, id ID) (*Metadata, error) {
	return s.get(context.Background(), tx, id)
}

func (s *postgresStore) DeleteTx(tx *pachsql.Tx, id ID) error {
	_, err := tx.Exec(`DELETE FROM storage.filesets WHERE id = $1`, id)
	return errors.EnsureStack(err)
}

func (s *postgresStore) Exists(ctx context.Context, id ID) (bool, error) {
	_, err := s.get(ctx, s.db, id)
	if err != nil {
		if errors.Is(err, ErrFileSetNotExists) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// SetupPostgresStoreV0 sets up the tables for a Store
// DO NOT MODIFY THIS FUNCTION
// IT HAS BEEN USED IN A RELEASED MIGRATION
func SetupPostgresStoreV0(ctx context.Context, tx *pachsql.Tx) error {
	const schema = `
	CREATE TABLE storage.filesets (
		id UUID NOT NULL PRIMARY KEY,
		metadata_pb BYTEA NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
	);
`
	_, err := tx.ExecContext(ctx, schema)
	return errors.EnsureStack(err)
}

// NewTestStore returns a Store scoped to the lifetime of the test.
func NewTestStore(ctx context.Context, t testing.TB, db *pachsql.DB) MetadataStore {
	tx := db.MustBegin()
	tx.MustExec(`CREATE SCHEMA IF NOT EXISTS storage`)
	require.NoError(t, SetupPostgresStoreV0(ctx, tx))
	require.NoError(t, tx.Commit())
	return NewPostgresStore(db)
}
