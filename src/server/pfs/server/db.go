package server

import (
	"github.com/pachyderm/pachyderm/src/server/pkg/migrations"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/track"
	"golang.org/x/net/context"
)

var state migrations.State

func init() {
	state = migrations.InitialState().
		Apply(func(ctx context.Context, env migrations.Env) error {
			return track.SetupPostgresTrackerV0(ctx, env.Tx)
		}).
		Apply(func(ctx context.Context, env migrations.Env) error {
			return chunk.SetupPostgresStoreV0(ctx, env.Tx)
		})
}
