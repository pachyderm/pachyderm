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
	"go.uber.org/zap"
	"strings"
)

func normalizeCommitTotals(ctx context.Context, env migrations.Env) error {
	ctx = pctx.Child(ctx, "normalizeCommitTotals")
	tx := env.Tx
	if _, err := tx.ExecContext(ctx, `ALTER TABLE pfs.commits ADD COLUMN total_fileset_id UUID REFERENCES storage.filesets(id)`); err != nil {
		return errors.Wrap(err, "add total_fileset column to pfs.commits")
	}
	count := int64(0)
	if err := tx.GetContext(ctx, &count, `SELECT count(fileset_id) FROM pfs.commit_totals;`); err != nil {
		return errors.Wrap(err, "counting rows in pfs.commit_totals")
	}
	if count == 0 {
		return nil
	}
	log.Info(ctx, "normalizing pfs.commit_totals", zap.Int64("total", count))
	// If there are commits in the pfs.commit_totals table that do not exist in the pfs.commits table, bail on the migration.
	missingCommits := []string{}
	if err := sqlx.SelectContext(ctx, tx, &missingCommits,
		`SELECT t.commit_id FROM pfs.commit_totals t EXCEPT (SELECT commit_id FROM pfs.commits);`); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return errors.Wrap(err, "getting commits exclusively in pfs.commit_totals")
	}
	if len(missingCommits) > 0 {
		return errors.New(fmt.Sprintf("following %d commits do not exist in the pfs.commits table: (%s)", len(missingCommits), strings.Join(missingCommits, ",")))
	}
	result, err := tx.ExecContext(ctx, "UPDATE pfs.commits c SET total_fileset_id = fileset_id FROM pfs.commit_totals t WHERE c.commit_id = t.commit_id")
	if err != nil {
		return errors.Wrap(err, "migrate totals")
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "migrate totals")
	}
	if count != rowsAffected {
		return errors.New(fmt.Sprintf("not all commits migrated (expected: %d, got: %d)", count, rowsAffected))
	}
	if _, err := tx.ExecContext(ctx, "DROP TABLE pfs.commit_totals;"); err != nil {
		return errors.Wrap(err, "drop the pfs.commit_totals")
	}
	return nil
}
