package v2_10_0

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"strings"
)

func normalizeCommitTotals(ctx context.Context, env migrations.Env) error {
	ctx = pctx.Child(ctx, "normalizeCommitTotals")
	tx := env.Tx
	if _, err := tx.ExecContext(ctx, `ALTER TABLE pfs.commits ADD COLUMN total_fileset_id UUID REFERENCES storage.filesets(id)`); err != nil {
		return errors.Wrap(err, "add total_fileset column to pfs.commits")
	}
	log.Info(ctx, "normalizing pfs.commit_totals")
	// If there are commits in the pfs.commit_totals table that do not exist in the pfs.commits table, bail on the migration.
	danglingCommits := []string{}
	if err := sqlx.SelectContext(ctx, tx, &danglingCommits,
		`SELECT t.commit_id FROM pfs.commit_totals t EXCEPT (SELECT commit_id FROM pfs.commits);`); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return errors.Wrap(err, "getting commits exclusively in pfs.commit_totals")
	}
	if len(danglingCommits) > 0 {
		log.Info(ctx, fmt.Sprintf("the following %d commits do not exist in the pfs.commits table and will not be migrated: (%s)", len(danglingCommits), strings.Join(danglingCommits, ",")))
	}
	_, err := tx.ExecContext(ctx, "UPDATE pfs.commits c SET total_fileset_id = fileset_id FROM pfs.commit_totals t WHERE c.commit_id = t.commit_id")
	if err != nil {
		return errors.Wrap(err, "migrate totals")
	}
	if _, err := tx.ExecContext(ctx, "DROP TABLE pfs.commit_totals;"); err != nil {
		return errors.Wrap(err, "drop the pfs.commit_totals")
	}
	return nil
}
