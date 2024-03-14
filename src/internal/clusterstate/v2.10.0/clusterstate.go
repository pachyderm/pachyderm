package v2_10_0

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
)

func Migrate(state migrations.State) migrations.State {
	return state.
		Apply("Add metadata to projects", addMetadataToProjects, migrations.Squash)
}
