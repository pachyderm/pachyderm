package pfsdb

import (
	"context"

	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
)

// returns collections released in v2.4.0 - specifically the projects collection
// DO NOT MODIFY THIS FUNCTION
// IT HAS BEEN USED IN A RELEASED MIGRATION
func CollectionsV2_4_0() []col.PostgresCollection {
	return []col.PostgresCollection{
		col.NewPostgresCollection(projectsCollectionName, nil, nil, nil, nil),
	}
}

// MigrateV2_4_0 migrates PFS to be fully project-aware with a default project.
// It uses some internal knowledge about how cols.PostgresCollection works to do
// so.  It does not migrate the stored protos themselves; those must be updated
// at read time.
func MigrateV2_4_0(ctx context.Context, tx *pachsql.Tx) error {
	if _, err := tx.ExecContext(ctx, `UPDATE collections.repos SET key = 'default/' || key, idx_name = 'default/' || idx_name`); err != nil {
		return errors.Wrap(err, "could not update repos")
	}
	if _, err := tx.ExecContext(ctx, `UPDATE collections.branches SET key = 'default/' || key, idx_repo = 'default/' || idx_repo`); err != nil {
		return errors.Wrap(err, "could not update branches")
	}
	if _, err := tx.ExecContext(ctx, `UPDATE collections.commit SET key = 'default/' || key, idx_repo = 'default/' || idx_repo, idx_branchless = 'default/' || idx_branchless`); err != nil {
		return errors.Wrap(err, "could not update commits")
	}
	if _, err := tx.ExecContext(ctx, `UPDATE collections.projects SET key = 'default' where key == ''`); err != nil {
		return errors.Wrap(err, "could not update projects")
	}
	return nil
}
