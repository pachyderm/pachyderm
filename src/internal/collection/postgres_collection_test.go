package collection_test

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"

	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

func TestPostgresCollections(suite *testing.T) {
	suite.Parallel()

	newCollection := func(ctx context.Context, t *testing.T) (ReadCallback, WriteCallback) {
		config := serviceenv.ConfigFromOptions(testutil.NewTestDBConfig(t))
		options := []dbutil.Option{
			dbutil.WithHostPort(config.PostgresServiceHost, config.PostgresServicePort),
			dbutil.WithDBName(config.PostgresDBName),
		}

		db, err := dbutil.NewDB(options...)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, db.Close())
		})

		dsn := dbutil.GetDSN(options...)
		listener := col.NewPostgresListener(dsn)
		t.Cleanup(func() {
			require.NoError(t, listener.Close())
		})

		testCol, err := col.NewPostgresCollection(ctx, db, listener, &col.TestItem{}, []*col.Index{TestSecondaryIndex}, nil)
		require.NoError(t, err)

		readCallback := func(ctx context.Context) col.ReadOnlyCollection {
			return testCol.ReadOnly(ctx)
		}

		writeCallback := func(ctx context.Context, f func(col.ReadWriteCollection) error) error {
			return col.NewSQLTx(ctx, db, func(tx *sqlx.Tx) error {
				return f(testCol.ReadWrite(tx))
			})
		}

		return readCallback, writeCallback
	}

	collectionTests(suite, newCollection)
	watchTests(suite, newCollection)

	// TODO: postgres-specific collection tests:
	// GetRevByIndex(index *Index, indexVal string, val proto.Message, opts *Options, f func(int64) error) error
	// DeleteByIndex(index *Index, indexVal string) error

	// TODO: test keycheck function
}
