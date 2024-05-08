package v2_10_0

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
)

func Migrate(state migrations.State) migrations.State {
	return Migrate_v2_10_BeforeDuplicates(state).
		Apply("Fix duplicate pipeline versions", DeduplicatePipelineVersions, migrations.Squash)
}

func Migrate_v2_10_BeforeDuplicates(state migrations.State) migrations.State {
	return state.
		Apply("Add metadata to projects", addMetadataToProjects, migrations.Squash).
		Apply("Add metadata to commits", addMetadataToCommits, migrations.Squash).
		Apply("Add metadata to branches", addMetadataToBranches, migrations.Squash).
		Apply("Add metadata to repos", addMetadataToRepos, migrations.Squash).
		Apply("Add cluster metadata table", addClusterMetadata, migrations.Squash)
}
