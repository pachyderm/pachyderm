package v2_10_0

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
)

func addMetadataToProjects(ctx context.Context, env migrations.Env) error {
	tx := env.Tx
	if _, err := tx.ExecContext(ctx, `ALTER TABLE core.projects ADD COLUMN metadata JSONB NOT NULL DEFAULT '{}' CHECK (jsonb_typeof(metadata) = 'object')`); err != nil {
		return errors.Wrap(err, "add metadata column to core.projects")
	}
	return nil
}
