package collection_test

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"

	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

func TestPostgresCollections(suite *testing.T) {
	suite.Parallel()
	postgres := testutil.NewPostgresDeployment(suite)

	newCollection := func(ctx context.Context, t *testing.T) (ReadCallback, WriteCallback) {
		db, listener := postgres.NewDatabase(t)
		testCol, err := col.NewPostgresCollection(ctx, db, listener, &col.TestItem{}, []*col.Index{TestSecondaryIndex})
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

	// TODO: postgres-specific collection tests:
	suite.Run("With", func(subsuite *testing.T) {
		subsuite.Run("ReadOnly", func(subsuite *testing.T) {
			subsuite.Run("Get", func(t *testing.T) {
			})
			subsuite.Run("List", func(t *testing.T) {
			})
		})
		subsuite.Run("ReadWrite", func(subsuite *testing.T) {
			subsuite.Run("Get", func(t *testing.T) {
			})
		})
	})
}
