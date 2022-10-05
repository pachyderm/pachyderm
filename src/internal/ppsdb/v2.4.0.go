package ppsdb

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
)

func MigrateV2_4_0(ctx context.Context, tx *pachsql.Tx) error {
	if _, err := tx.ExecContext(ctx, `UPDATE collections.pipelines SET key = 'default/' || key, idx_name = 'default/' || idx_name, idx_version = 'default/' || idx_version`); err != nil {
		return errors.Wrap(err, "could not update pipelines")
	}
	if _, err := tx.ExecContext(ctx, `UPDATE collections.jobs SET key = 'default/' || key, idx_pipeline = 'default/' || idx_pipeline, idx_job_state = 'default/' || idx_job_state, idx_jobset = 'default/' || idx_jobset`); err != nil {
		return errors.Wrap(err, "could not update jobs")
	}
	return nil
}
