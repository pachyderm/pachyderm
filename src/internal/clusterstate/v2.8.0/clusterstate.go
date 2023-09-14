package v2_8_0

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
)

func Migrate(state migrations.State) migrations.State {
	return state.
		Apply("Create pfs schema", createPFSSchema).
		Apply("Migrate collections.repos to pfs.repos", migrateRepos, migrations.Squash).
		Apply("Migrate collections.branches to pfs.branches", migrateBranches, migrations.Squash).
		Apply("Migrate collections.commits to pfs.commits", migrateCommits).
		Apply("Synthesize cluster defaults from environment variables", synthesizeClusterDefaults).
		Apply("Synthesize user and effective specs from their pipeline details", synthesizeSpecs, migrations.Squash)
}

func PostMigrate(state migrations.State) migrations.State {
	return state.Apply("alter pfs.commits schema post data migration", alterCommitsTablePostDataMigration)
}
