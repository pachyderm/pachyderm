package v2_12_0

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
)

func Migrate(state migrations.State) migrations.State {
	return state.
		Apply("Create PJS Schema", createPJSSchema).
		Apply("Alter Branch Provenance Table", alterBranchProvenanceTable).
		Apply("Create Chunkset Schema", createChunksetSchema).
		Apply("Create snapshot schema", createSnapshotSchema).
		Apply("Create storage.fileset_pins Table", func(ctx context.Context, env migrations.Env) error {
			return fileset.CreatePinsTable(ctx, env.Tx)
		}).
		Apply("Create admin schema + restarts table", createPachydermRestartSchema)
}
