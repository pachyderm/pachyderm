package tracker

import (
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/pachyderm/pachyderm/src/server/pkg/dbutil"
)

func TestPGTracker(t *testing.T) {
	t.Parallel()
	TestTracker(t, func(cb func(tracker Tracker)) {
		dbutil.WithTestDB(t, func(db *sqlx.DB) {
			db.MustExec("CREATE SCHEMA storage")
			db.MustExec(schema)
			tr := NewPGTracker(db)
			cb(tr)
		})
	})
}
