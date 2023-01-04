package fileset

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
)

func TestPostgresStore(t *testing.T) {
	StoreTestSuite(t, func(t testing.TB) MetadataStore {
		ctx := pctx.TestContext(t)
		db := dockertestenv.NewTestDB(t)
		return NewTestStore(ctx, t, db)
	})
}
