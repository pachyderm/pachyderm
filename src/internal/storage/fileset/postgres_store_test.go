package fileset

import (
	"testing"

	dbtesting "github.com/pachyderm/pachyderm/v2/src/internal/dbutil/testing"
)

func TestPostgresStore(t *testing.T) {
	StoreTestSuite(t, func(t testing.TB) Store {
		db := dbtesting.NewTestDB(t)
		return NewTestStore(t, db)
	})
}
