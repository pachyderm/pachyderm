package v2_10_0

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
)

func Migrate(state migrations.State) migrations.State {
	return MigratePartial(state, 0)
}

type migrationStep struct {
	desc string
	f    func(ctx context.Context, env migrations.Env) error
}

var migrationSteps = []migrationStep{
	{
		desc: "Add metadata to projects",
		f:    addMetadataToProjects,
	},
	{
		desc: "Add metadata to commits",
		f:    addMetadataToCommits,
	},
	{
		desc: "Add metadata to branches",
		f:    addMetadataToBranches,
	},
	{
		desc: "Add metadata to repos",
		f:    addMetadataToRepos,
	},
	{
		desc: "Add cluster metadata table",
		f:    addClusterMetadata,
	},
	{
		desc: "Fix duplicate pipeline versions",
		f:    DeduplicatePipelineVersions,
	},
}

func MigratePartial(state migrations.State, trimRight int) migrations.State {
	until := len(migrationSteps) - trimRight
	finalSteps := migrationSteps[:until]
	lastState := state
	for _, s := range finalSteps {
		lastState = lastState.Apply(s.desc, s.f, migrations.Squash)
	}
	return lastState
}
