package clusterstate

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	enterpriseserver "github.com/pachyderm/pachyderm/v2/src/server/enterprise/server"
)

var state_2_1_0 migrations.State = state_2_0_0.
	Apply("Move EnterpriseConfig from etcd -> postgres", func(ctx context.Context, env migrations.Env) error {
		return enterpriseserver.EnterpriseConfigPostgresMigration(ctx, env.Tx, env.EtcdClient)
	}).
	Apply("Remove old EnterpriseConfig record from etcd", func(ctx context.Context, env migrations.Env) error {
		return enterpriseserver.DeleteEnterpriseConfigFromEtcd(ctx, env.EtcdClient)
	})
