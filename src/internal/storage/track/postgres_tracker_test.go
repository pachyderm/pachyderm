package track

import (
	"testing"

	_ "github.com/lib/pq"
	dbtesting "github.com/pachyderm/pachyderm/v2/src/internal/dbutil/testing"
)

func TestPostgresTracker(t *testing.T) {
	t.Parallel()
	TestTracker(t, func(testing.TB) Tracker {
		db := dbtesting.NewTestDB(t)
		db.MustExec("CREATE SCHEMA storage")
		db.MustExec(schema)
		return NewPostgresTracker(db)
	})
}
