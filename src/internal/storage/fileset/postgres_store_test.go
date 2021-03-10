package fileset

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
)

func TestPostgresStore(t *testing.T) {
	StoreTestSuite(t, func(t testing.TB) MetadataStore {
		db := dbutil.NewTestDB(t)
		return NewTestStore(t, db)
	})
}
