package clusterstate

import (
	"context"

	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	enterpriseserver "github.com/pachyderm/pachyderm/v2/src/server/enterprise/server"
)

var state_2_1_0 migrations.State = state_2_0_0.
	Apply("Move EnterpriseConfig from etcd -> postgres", func(ctx context.Context, env migrations.Env) error {
		if err := col.SetupPostgresCollections(ctx, env.Tx, enterpriseserver.EnterpriseConfigCollection(nil, nil)); err != nil {
			return err
		}
		return enterpriseserver.TryEnterpriseConfigPostgresMigration(ctx, env)
	})
