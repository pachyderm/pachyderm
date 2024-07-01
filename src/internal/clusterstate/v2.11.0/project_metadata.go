package v2_11_0

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
)

func addProjectCreatedBy(ctx context.Context, env migrations.Env) error {
	ctx = pctx.Child(ctx, "addProjectMetadata")
	tx := env.Tx
	if _, err := tx.ExecContext(ctx, `ALTER TABLE core.projects ADD COLUMN created_by TEXT REFERENCES auth.principals(subject)`); err != nil {
		return errors.Wrap(err, "add created_by column to core.projects")
	}
	return nil
}
