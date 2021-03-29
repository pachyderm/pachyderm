package testing

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"

	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	dbtesting "github.com/pachyderm/pachyderm/v2/src/internal/dbutil/testing"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
)

func TestPostgresCollections(suite *testing.T) {
	suite.Parallel()
	postgres := dbtesting.NewPostgresDeployment(suite)

	newCollection := func(t *testing.T) (col.ReadOnlyCollection, WriteCallback) {
		db, listener := postgres.NewDatabase(t)
		testCol, err := col.NewPostgresCollection(db, listener, &TestItem{}, []*col.Index{TestSecondaryIndex})
		require.NoError(t, err)

		writeCallback := func(f func(col.ReadWriteCollection) error) error {
			return col.NewSQLTx(context.Background(), db, func(tx *sqlx.Tx) error {
				return f(testCol.ReadWrite(tx))
			})
		}

		return testCol.ReadOnly(context.Background()), writeCallback
	}

	collectionTests(suite, newCollection)

	// TODO: postgres-specific collection tests - With()
}
