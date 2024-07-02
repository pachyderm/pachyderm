package v2_11_0

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
)

func normalizeCommitDiffs(ctx context.Context, env migrations.Env) error {
	ctx = pctx.Child(ctx, "normalizeCommitDiffs")
	tx := env.Tx
	_, err := tx.ExecContext(ctx,
		`ALTER TABLE pfs.commit_diffs ADD COLUMN commit_int_id BIGINT REFERENCES pfs.commits(int_id) ON DELETE CASCADE`)
	if err != nil {
		return errors.Wrap(err, "add commit_int_id column to pfs.commit_diffs")
	}
	log.Info(ctx, "normalizing pfs.commit_diffs")
	_, err = tx.ExecContext(ctx, "UPDATE pfs.commit_diffs d SET commit_int_id = int_id FROM pfs.commits c WHERE d.commit_id = c.commit_id")
	if err != nil {
		return errors.Wrap(err, "migrate diffs")
	}
	_, err = tx.ExecContext(ctx, `
		ALTER TABLE pfs.commit_diffs DROP CONSTRAINT commit_diffs_pkey;
		ALTER TABLE pfs.commit_diffs ADD PRIMARY KEY (commit_int_id, num);
		ALTER TABLE pfs.commit_diffs DROP COLUMN commit_id; 
		`)
	if err != nil {
		return errors.Wrap(err, "change the primary key column of pfs.commit_diffs")
	}
	return nil
}
