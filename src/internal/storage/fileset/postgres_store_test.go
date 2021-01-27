package fileset

import (
	"testing"

	"github.com/pachyderm/pachyderm/src/internal/dbutil"
)

func TestPostgresStore(t *testing.T) {
	StoreTestSuite(t, func(t testing.TB) Store {
		db := dbutil.NewTestDB(t)
		return NewTestStore(t, db)
	})
}
