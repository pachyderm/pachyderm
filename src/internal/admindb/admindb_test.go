package admindb

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/clusterstate"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
)

type restart struct {
	RestartAt time.Time `db:"restart_at"`
	Reason    string    `db:"reason"`
	CreatedBy string    `db:"created_by"`
}

func TestRestart(t *testing.T) {
	ctx := pctx.TestContext(t)
	db := dockertestenv.NewMigratedTestDB(t, clusterstate.DesiredClusterState)
	now := time.Now().Truncate(time.Microsecond)

	testData := []struct {
		name               string
		startup            time.Time
		restarts           []restart
		wantMessage        string
		wantRestart        bool
		wantRestartContent []restart
	}{
		{
			name:    "startup in the past",
			startup: now.Add(-time.Hour),
			restarts: []restart{
				{now, "because it's fun", "me@1.2.3.4"},
			},
			wantMessage: "because it's fun by me@1.2.3.4",
			wantRestart: true,
			wantRestartContent: []restart{
				{now, "because it's fun", "me@1.2.3.4"},
			},
		},
		{
			name:    "already restarted",
			startup: now,
			restarts: []restart{
				{now.Add(-30 * time.Second), "", ""},
			},
			wantRestartContent: []restart{
				{now.Add(-30 * time.Second), "", ""},
			},
		},
		{
			name:    "many pending restarts",
			startup: now.Add(-2*time.Hour - 30*time.Minute),
			restarts: []restart{
				{now.Add(-3 * time.Hour), "first", ""},
				{now.Add(-2 * time.Hour), "", "root@localhost"},
				{now.Add(-1 * time.Hour), "third", ""},
			},
			wantMessage: "third; restart requested by root@localhost",
			wantRestart: true,
			wantRestartContent: []restart{
				{now.Add(-1 * time.Hour), "third", ""},
				{now.Add(-2 * time.Hour), "", "root@localhost"},
				{now.Add(-3 * time.Hour), "first", ""},
			},
		},
		{
			name:    "pending restart plus cleanup",
			startup: now.Add(-2*time.Hour - 30*time.Minute),
			restarts: []restart{
				{now.Add(-100 * 24 * time.Hour), "stale", ""},
				{now.Add(-1 * time.Hour), "", ""},
			},
			wantMessage: "restart requested",
			wantRestart: true,
			wantRestartContent: []restart{
				{now.Add(-1 * time.Hour), "", ""},
			},
		},
		{
			name:    "no pending restart plus cleanup",
			startup: now,
			restarts: []restart{
				{now.Add(-100 * 24 * time.Hour), "stale", ""},
				{now.Add(-1 * time.Hour), "", ""},
			},
			wantRestartContent: []restart{
				{now.Add(-1 * time.Hour), "", ""},
			},
		},
	}

	for _, test := range testData {
		t.Run(test.name, func(t *testing.T) {
			var gotRestart bool
			var gotMessage string
			if err := dbutil.WithTx(ctx, db, func(ctx context.Context, tx *pachsql.Tx) error {
				if _, err := tx.ExecContext(ctx, `truncate table admin.restarts`); err != nil {
					return errors.Wrap(err, "truncate table")
				}
				for _, r := range test.restarts {
					if err := ScheduleRestart(ctx, tx, r.RestartAt, r.Reason, r.CreatedBy); err != nil {
						return errors.Wrap(err, "ScheduleRestart")
					}
				}
				var err error
				gotRestart, gotMessage, err = ShouldRestart(ctx, tx, test.startup, now)
				if err != nil {
					return errors.Wrap(err, "ShouldRestart")
				}
				return nil
			}); err != nil {
				t.Fatalf("WithTx(schedule/check): %v", err)
			}
			if got, want := gotRestart, test.wantRestart; got != want {
				t.Errorf("restart?\n  got: %v\n want: %v", got, want)
			}
			if got, want := gotMessage, test.wantMessage; got != want {
				t.Errorf("reason:\n  got: %v\n want: %v", got, want)
			}

			var gotRestarts []restart
			if err := dbutil.WithTx(ctx, db, func(ctx context.Context, tx *pachsql.Tx) error {
				if err := sqlx.SelectContext(ctx, tx, &gotRestarts, `select restart_at,reason,created_by from admin.restarts order by restart_at desc`); err != nil {
					return errors.Wrap(err, "select restarts")
				}
				return nil
			}, dbutil.WithReadOnly()); err != nil {
				t.Fatalf("WithTx(check): %v", err)
			}
			if diff := cmp.Diff(test.wantRestartContent, gotRestarts); diff != "" {
				t.Errorf("restarts (-want +got):\n%v", diff)
			}
		})
	}
}
