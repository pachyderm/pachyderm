package clusterstate

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
)

var state_2_1_0 migrations.State = state_2_0_0.
	Apply("create pfs cache v1", func(ctx context.Context, env migrations.Env) error {
		return fileset.CreatePostgresCacheV1(ctx, env.Tx)
	}, migrations.Squash)

	// DO NOT MODIFY THIS STATE
	// IT HAS ALREADY SHIPPED IN A RELEASE
