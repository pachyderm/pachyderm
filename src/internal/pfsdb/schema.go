package pfsdb

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
)

// SchemaCommitStoreV0 runs SQL to setup the commit store.
// DO NOT MODIFY THIS FUNCTION
// IT HAS BEEN USED IN A RELEASED MIGRATION
func SchemaCommitStoreV0(ctx context.Context, tx *pachsql.Tx) error {
	_, err := tx.ExecContext(ctx, `
		CREATE TABLE pfs.commit_diffs (
			commit_id TEXT NOT NULL,
			num BIGSERIAL NOT NULL,
			fileset_id UUID NOT NULL,
			PRIMARY KEY(commit_id, num)
		);

		CREATE TABLE pfs.commit_totals (
			commit_id TEXT NOT NULL,
			fileset_id UUID NOT NULL,
			PRIMARY KEY(commit_id)
		);
	`)
	return errors.EnsureStack(err)
}
