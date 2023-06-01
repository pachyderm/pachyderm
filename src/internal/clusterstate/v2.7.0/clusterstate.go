package v2_7_0

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
)

func Migrate(state migrations.State) migrations.State {
	return state.Apply("Setup core.* tables", func(ctx context.Context, env migrations.Env) error {
		if err := SetupCore(ctx, env.Tx); err != nil {
			return err
		}
		return nil
	})
}
