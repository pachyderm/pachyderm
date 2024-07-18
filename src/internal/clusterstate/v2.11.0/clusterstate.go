package v2_11_0

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
)

func Migrate(state migrations.State) migrations.State {
	return state.
		Apply("Normalize commit totals", normalizeCommitTotals).
		Apply("Normalize commit diffs", normalizeCommitDiffs, migrations.Squash).
		Apply("Add auth principals table", addAuthPrincipals, migrations.Squash).
		Apply("Add project created_by", addProjectCreatedBy, migrations.Squash).
		Apply("Create PJS schema", createPJSSchema, migrations.Squash).
		Apply("Add commit created_by", addCommitCreatedBy, migrations.Squash).
		Apply("Add branch created_by", addBranchCreatedBy, migrations.Squash)
}
