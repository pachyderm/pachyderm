package collection_test

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"

	"github.com/pachyderm/pachyderm/v2/src/client"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/testpachd"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
	"github.com/pachyderm/pachyderm/v2/src/proxy"
)

func TestPostgresCollections(suite *testing.T) {
	suite.Parallel()

	PostgresCollectionTests(suite, func(ctx context.Context, t *testing.T) (*sqlx.DB, col.PostgresListener) {
		config := serviceenv.ConfigFromOptions(testutil.NewTestDBConfig(t))
		options := []dbutil.Option{
			dbutil.WithHostPort(config.PostgresHost, config.PostgresPort),
			dbutil.WithDBName(config.PostgresDBName),
			dbutil.WithMaxOpenConns(1), // All tests should be able to run on a single connection
		}

		db, err := dbutil.NewDB(options...)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, db.Close())
		})

		require.NoError(t, dbutil.WithTx(ctx, db, func(sqlTx *sqlx.Tx) error {
			if err := col.CreatePostgresSchema(ctx, sqlTx); err != nil {
				return err
			}
			return col.SetupPostgresV0(ctx, sqlTx)
		}))

		dsn := dbutil.GetDSN(options...)
		listener := col.NewPostgresListener(dsn)
		t.Cleanup(func() {
			require.NoError(t, listener.Close())
		})
		return db, listener
	})
}

// TODO: Add test for filling up watcher buffer.
func TestPostgresCollectionsProxy(suite *testing.T) {
	suite.Parallel()

	PostgresCollectionTests(suite, func(_ context.Context, t *testing.T) (*sqlx.DB, col.PostgresListener) {
		dbConfig := testutil.NewTestDBConfig(t)
		config := serviceenv.ConfigFromOptions(dbConfig)
		options := []dbutil.Option{
			dbutil.WithHostPort(config.PostgresHost, config.PostgresPort),
			dbutil.WithDBName(config.PostgresDBName),
			dbutil.WithMaxOpenConns(1), // All tests should be able to run on a single connection
		}

		db, err := dbutil.NewDB(options...)
		require.NoError(t, err)
		t.Cleanup(func() {
			require.NoError(t, db.Close())
		})

		env := testpachd.NewRealEnv(t, dbConfig)
		listener := client.NewProxyPostgresListener(func() (proxy.APIClient, error) { return env.PachClient.ProxyClient, nil })
		t.Cleanup(func() {
			require.NoError(t, listener.Close())
		})
		return db, listener
	})
}

func PostgresCollectionTests(suite *testing.T, setup func(context.Context, *testing.T) (*sqlx.DB, col.PostgresListener)) {
	newCollection := func(ctx context.Context, t *testing.T) (ReadCallback, WriteCallback) {
		db, listener := setup(ctx, t)

		testCol := col.NewPostgresCollection("test_items", db, listener, &col.TestItem{}, []*col.Index{TestSecondaryIndex}, nil)
		require.NoError(t, dbutil.WithTx(ctx, db, func(sqlTx *sqlx.Tx) error {
			return col.SetupPostgresCollections(ctx, sqlTx, testCol)
		}))

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

	// Postgres collections support getting multiple rows by a secondary index,
	// although it requires loading the entire result set into memory to prevent
	// multiple queries from using the transaction at once.  Consequently, it
	// should only be used for small result sets.
	suite.Run("ReadWriteGetByIndex", func(subsuite *testing.T) {
		subsuite.Parallel()
		emptyRead, emptyWriter := newCollection(context.Background(), subsuite)
		defaultRead, defaultWriter := initCollection(subsuite, newCollection)

		subsuite.Run("Empty", func(t *testing.T) {
			t.Parallel()
			err := emptyWriter(context.Background(), func(rw col.ReadWriteCollection) error {
				pgrw := rw.(col.PostgresReadWriteCollection)
				return pgrw.GetByIndex(TestSecondaryIndex, "foo", &col.TestItem{}, col.DefaultOptions(), func(string) error {
					return errors.New("GetByIndex callback should not have been called for an empty collection")
				})
			})
			require.NoError(t, err)
			count, err := emptyRead(context.Background()).Count()
			require.NoError(t, err)
			require.Equal(t, int64(0), count)
		})

		subsuite.Run("Success", func(t *testing.T) {
			t.Parallel()
			keys := []string{}
			err := defaultWriter(context.Background(), func(rw col.ReadWriteCollection) error {
				testProto := &col.TestItem{}
				pgrw := rw.(col.PostgresReadWriteCollection)
				return pgrw.GetByIndex(TestSecondaryIndex, originalValue, testProto, col.DefaultOptions(), func(key string) error {
					require.Equal(t, testProto.ID, key)
					require.Equal(t, testProto.Value, originalValue)
					keys = append(keys, key)
					// Clear testProto.ID and testProto.Value just to make sure they get overwritten each time
					testProto.ID = ""
					testProto.Value = ""
					return nil
				})
			})
			require.NoError(t, err)
			require.ElementsEqual(t, keys, idRange(0, defaultCollectionSize))
			checkDefaultCollection(t, defaultRead, RowDiff{})
		})

		subsuite.Run("NoResults", func(t *testing.T) {
			t.Parallel()
			err := defaultWriter(context.Background(), func(rw col.ReadWriteCollection) error {
				testProto := &col.TestItem{}
				pgrw := rw.(col.PostgresReadWriteCollection)
				return pgrw.GetByIndex(TestSecondaryIndex, changedValue, testProto, col.DefaultOptions(), func(string) error {
					return errors.New("GetByIndex callback should not have been called for an index value with no rows")
				})
			})
			require.NoError(t, err)
			checkDefaultCollection(t, defaultRead, RowDiff{})
		})

		subsuite.Run("InvalidIndex", func(t *testing.T) {
			t.Parallel()
			err := defaultWriter(context.Background(), func(rw col.ReadWriteCollection) error {
				pgrw := rw.(col.PostgresReadWriteCollection)
				return pgrw.GetByIndex(&col.Index{}, "", &col.TestItem{}, col.DefaultOptions(), func(key string) error {
					return errors.New("GetByIndex callback should not have been called when using an invalid index")
				})
			})
			require.YesError(t, err)
			require.Matches(t, "Unknown collection index", err.Error())
			checkDefaultCollection(t, defaultRead, RowDiff{})
		})

		subsuite.Run("Nested", func(t *testing.T) {
			t.Parallel()
			outerKeys := []string{}
			innerKeys := []string{}
			innerID := makeID(3)
			err := defaultWriter(context.Background(), func(rw col.ReadWriteCollection) error {
				testProto := &col.TestItem{}
				pgrw := rw.(col.PostgresReadWriteCollection)
				return pgrw.GetByIndex(TestSecondaryIndex, originalValue, testProto, col.DefaultOptions(), func(key string) error {
					outerKeys = append(outerKeys, testProto.ID)
					if err := pgrw.Get(innerID, testProto); err != nil {
						return err
					}
					innerKeys = append(innerKeys, testProto.ID)
					// Clear testProto.ID and testProto.Value just to make sure they get overwritten each time
					testProto.ID = ""
					testProto.Value = ""
					return nil
				})
			})
			require.NoError(t, err)
			require.ElementsEqual(t, idRange(0, defaultCollectionSize), outerKeys)

			expectedInnerKeys := []string{}
			for i := 0; i < defaultCollectionSize; i++ {
				expectedInnerKeys = append(expectedInnerKeys, innerID)
			}

			require.ElementsEqual(t, expectedInnerKeys, innerKeys)
			checkDefaultCollection(t, defaultRead, RowDiff{})
		})
	})

	// TODO: postgres-specific collection tests:
	// GetRevByIndex(index *Index, indexVal string, val proto.Message, opts *Options, f func(int64) error) error
	// DeleteByIndex(index *Index, indexVal string) error

	// TODO: test keycheck function
}
