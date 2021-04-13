package fileset

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

func TestPostgresStore(t *testing.T) {
	StoreTestSuite(t, func(t testing.TB) MetadataStore {
		db := testutil.NewTestDB(t)
		return NewTestStore(t, db)
	})
}
