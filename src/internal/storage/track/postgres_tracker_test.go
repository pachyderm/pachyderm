package track

import (
	"testing"

	_ "github.com/lib/pq"
	"github.com/pachyderm/pachyderm/v2/src/internal/testutil"
)

func TestPostgresTracker(t *testing.T) {
	t.Parallel()
	TestTracker(t, func(testing.TB) Tracker {
		db := testutil.NewTestDB(t)
		db.MustExec("CREATE SCHEMA storage")
		db.MustExec(schema)
		return NewPostgresTracker(db)
	})
}
