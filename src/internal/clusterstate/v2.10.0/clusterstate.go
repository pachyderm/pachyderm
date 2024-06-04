package v2_10_0

import (
	"context"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"

	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
)

func Migrate(state migrations.State) migrations.State {
	return Migrate_v2_10_BeforeDuplicates(state).
		Apply("Fix duplicate pipeline versions", DeduplicatePipelineVersions, migrations.Squash).
		Apply("Create compaction steps table", func(ctx context.Context, env migrations.Env) error {
			_, err := env.Tx.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS storage.compaction_steps (
    			id UUID NOT NULL PRIMARY KEY,
    			compacted_id UUID NOT NULL,
    			step BIGSERIAL NOT NULL
			);`)
			if err != nil {
				return errors.Wrap(err, "adding compaction steps migration")
			}
			return nil
		})
}

func Migrate_v2_10_BeforeDuplicates(state migrations.State) migrations.State {
	return state.
		Apply("Add metadata to projects", addMetadataToProjects, migrations.Squash).
		Apply("Add metadata to commits", addMetadataToCommits, migrations.Squash).
		Apply("Add metadata to branches", addMetadataToBranches, migrations.Squash).
		Apply("Add metadata to repos", addMetadataToRepos, migrations.Squash).
		Apply("Add cluster metadata table", addClusterMetadata, migrations.Squash)
}
