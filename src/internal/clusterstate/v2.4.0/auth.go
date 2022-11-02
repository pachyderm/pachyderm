package v2_4_0

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
)

// migrateAuth migrates auth to be fully project-aware with a default project.
// It uses some internal knowledge about how cols.PostgresCollection works to do
// so.
func migrateAuth(ctx context.Context, tx *pachsql.Tx) error {
	if _, err := tx.ExecContext(ctx, `UPDATE collections.role_bindings SET key = regexp_replace(key, '^REPO:([-a-zA-Z0-9_]+)$', 'REPO:default/\1') where key ~ '^REPO:([-a-zA-Z0-9_]+)'`); err != nil {
		return errors.Wrap(err, "could not update role bindings")
	}
	if _, err := tx.ExecContext(ctx, `UPDATE collections.role_bindings SET key = regexp_replace(key, '^SPEC_REPO:([-a-zA-Z0-9_]+)$', 'SPEC_REPO:default/\1') where key ~ '^SPEC_REPO:([-a-zA-Z0-9_]+)'`); err != nil {
		return errors.Wrap(err, "could not update role bindings")
	}
	return nil
}
