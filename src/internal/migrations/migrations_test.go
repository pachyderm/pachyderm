package migrations

import (
	"context"
	"testing"
	"time"

	ec "github.com/pachyderm/pachyderm/v2/src/enterprise"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/dockertestenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/testetcd"
	enterpriseserver "github.com/pachyderm/pachyderm/v2/src/server/enterprise/server"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
)

func TestMigration(t *testing.T) {
	db := dockertestenv.NewTestDB(t)
	state := InitialState().
		Apply("test 1", func(ctx context.Context, env Env) error {
			// NoOp
			return nil
		}).
		Apply("test 2", func(ctx context.Context, env Env) error {
			_, err := env.Tx.ExecContext(ctx, `CREATE TABLE test_table1 (id BIGSERIAL PRIMARY KEY, field1 TEXT, field2 TEXT);`)
			return err
		}).
		Apply("test 3", func(ctx context.Context, env Env) error {
			_, err := env.Tx.ExecContext(ctx, `CREATE TABLE test_table2 (id BIGSERIAL PRIMARY KEY, field1 TEXT, field2 TEXT);`)
			return err
		})
	ctx := context.Background()
	func() {
		eg, ctx := errgroup.WithContext(ctx)
		const numWaiters = 10
		for i := 0; i < numWaiters; i++ {
			eg.Go(func() error {
				return BlockUntil(ctx, db, state)
			})
		}
		eg.Go(func() error {
			time.Sleep(time.Second)
			return ApplyMigrations(ctx, db, Env{}, state)
		})
		require.NoError(t, eg.Wait())
	}()
	var count int
	require.NoError(t, db.GetContext(ctx, &count, `SELECT count(*) FROM migrations`))
	assert.Equal(t, state.Number()+1, count)
	var max int
	require.NoError(t, db.GetContext(ctx, &max, `SELECT max(id) FROM migrations`))
	assert.Equal(t, state.Number(), max)
}

func TestEnterpriseConfigMigration(t *testing.T) {
	db := dockertestenv.NewTestDB(t)
	etcd := testetcd.NewEnv(t).EtcdClient

	config := &ec.EnterpriseConfig{
		Id:            "id",
		LicenseServer: "server",
		Secret:        "secret",
	}

	etcdConfigCol := col.NewEtcdCollection(etcd, "", nil, &ec.EnterpriseConfig{}, nil, nil)
	_, err := col.NewSTM(context.Background(), etcd, func(stm col.STM) error {
		return etcdConfigCol.ReadWrite(stm).Put("config", config)
	})
	require.NoError(t, err)

	env := Env{EtcdClient: etcd}
	// create enterprise config record in etcd
	state := InitialState().
		// the following two state changes were shipped in v2.0.0
		Apply("create collections schema", func(ctx context.Context, env Env) error {
			return col.CreatePostgresSchema(ctx, env.Tx)
		}).
		Apply("create collections trigger functions", func(ctx context.Context, env Env) error {
			return col.SetupPostgresV0(ctx, env.Tx)
		}).
		// the following two state changes are shipped in v2.1.0 to migrate EnterpriseConfig from etcd -> postgres
		Apply("Move EnterpriseConfig from etcd -> postgres", func(ctx context.Context, env Env) error {
			return enterpriseserver.EnterpriseConfigPostgresMigration(ctx, env.Tx, env.EtcdClient)
		}).
		Apply("Remove old EnterpriseConfig record from etcd", func(ctx context.Context, env Env) error {
			return enterpriseserver.DeleteEnterpriseConfigFromEtcd(ctx, env.EtcdClient)
		})
	// run the migration
	err = ApplyMigrations(context.Background(), db, env, state)
	require.NoError(t, err)
	err = BlockUntil(context.Background(), db, state)
	require.NoError(t, err)

	pgCol := enterpriseserver.EnterpriseConfigCollection(db, nil)
	result := &ec.EnterpriseConfig{}
	require.NoError(t, pgCol.ReadOnly(context.Background()).Get("config", result))
	require.Equal(t, config.Id, result.Id)
	require.Equal(t, config.LicenseServer, result.LicenseServer)
	require.Equal(t, config.Secret, result.Secret)

	err = etcdConfigCol.ReadOnly(context.Background()).Get("config", &ec.EnterpriseConfig{})
	require.YesError(t, err)
	require.True(t, col.IsErrNotFound(err))
}
