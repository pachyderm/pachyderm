package fileset

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
)

func TestPostgresStore(t *testing.T) {
	StoreTestSuite(t, func(t testing.TB) MetadataStore {
		db := dockertestenv.NewTestDB(t)
		return NewTestStore(t, db)
	})
}
