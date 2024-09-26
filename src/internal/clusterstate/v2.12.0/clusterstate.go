package v2_12_0

import (
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
)

func Migrate(state migrations.State) migrations.State {
	return state.
		Apply("Create PJS Schema", createPJSSchema).
		Apply("Alter Branch Provenance Table", alterBranchProvenanceTable).
		Apply("Create Chunkset Schema", createChunksetSchema)
}
