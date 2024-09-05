package v2_12_0

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
)

// TODO: Inverted index?
func createBranchPropagationSpecsTable(ctx context.Context, env migrations.Env) error {
	ctx = pctx.Child(ctx, "createBranchPropagationSpecsTable")
	_, err := env.Tx.ExecContext(ctx, `
		CREATE TABLE pfs.branch_propagation_specs (
			from_id BIGINT NOT NULL,
			to_id BIGINT NOT NULL,
			FOREIGN KEY (from_id, to_id) REFERENCES pfs.branch_provenance (from_id, to_id) ON DELETE CASCADE,
			PRIMARY KEY (from_id, to_id),
			never BOOLEAN
		);
	`)
	if err != nil {
		return errors.Wrap(err, "create branch propagation specs table")
	}
	return nil
}
