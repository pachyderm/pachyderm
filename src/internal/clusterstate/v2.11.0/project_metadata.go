package v2_11_0

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
)

func addProjectMetadata(ctx context.Context, env migrations.Env) error {
	ctx = pctx.Child(ctx, "addProjectMetadata")
	tx := env.Tx
	if _, err := tx.ExecContext(ctx, `ALTER TABLE core.projects ADD COLUMN created_by TEXT NOT NULL DEFAULT ''`); err != nil {
		return errors.Wrap(err, "add created_by column to core.projects")
	}
	return nil
}

func addCommitCreatedBy(ctx context.Context, env migrations.Env) error {
	ctx = pctx.Child(ctx, "addCommitCreationInfo")
	if _, err := env.Tx.ExecContext(ctx, `ALTER TABLE pfs.commits ADD COLUMN created_by TEXT NOT NULL DEFAULT ''`); err != nil {
		return errors.Wrap(err, "add created_by and created_at to pfs.commits")
	}
	return nil
}
