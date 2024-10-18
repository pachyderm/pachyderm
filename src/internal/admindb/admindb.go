// Package admindb provides functions for interacting with the admin.* database schema.
package admindb

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
)

// ScheduleRestart schedules every Pachyderm pod to restart at `when`.
func ScheduleRestart(ctx context.Context, tx *pachsql.Tx, when time.Time, reason, who string) error {
	if _, err := tx.ExecContext(ctx, `insert into admin.restarts (restart_at, reason, created_by) values ($1, $2, $3)`, when, reason, who); err != nil {
		return errors.Wrap(err, "insert")
	}
	return nil
}

type restartMessage struct {
	Reason    string `db:"reason"`
	CreatedBy string `db:"created_by"`
}

func (m restartMessage) String() string {
	var msg strings.Builder
	if m.Reason != "" {
		msg.WriteString(m.Reason)
	} else {
		msg.WriteString("restart requested")
	}
	if m.CreatedBy != "" {
		fmt.Fprintf(&msg, " by %v", m.CreatedBy)
	}
	return msg.String()
}

func message(messages ...restartMessage) string {
	parts := make([]string, len(messages))
	for i, m := range messages {
		parts[i] = m.String()
	}
	return strings.Join(parts, "; ")
}

// ShouldRestart returns true (and the reason) if a pod that started at `startup` should restart.
// All times are compared from the perspective of the database server.  It also cleans up outdated
// restart requests (90 days since `now`).
func ShouldRestart(ctx context.Context, tx *pachsql.Tx, startup, now time.Time) (restart bool, msg string, _ error) {
	var result []restartMessage
	if err := sqlx.SelectContext(ctx, tx, &result,
		`select reason, created_by from admin.restarts where restart_at >= $1 order by restart_at desc`, startup); err != nil {
		return false, "", errors.Wrap(err, "query pending restarts")
	}
	if _, err := tx.ExecContext(ctx, `delete from admin.restarts where restart_at < $1::timestamptz - interval '90 days'`, now); err != nil {
		return false, "", errors.Wrap(err, "cleanup very old restarts")
	}
	if len(result) > 0 {
		restart = true
	}
	msg = message(result...)
	return restart, msg, nil
}
