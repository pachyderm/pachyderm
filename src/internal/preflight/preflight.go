// Package preflight offers checks that can be run by pachd in preflight mode.
package preflight

import (
	"context"
	"database/sql"
	"net/url"
	"os"
	"path/filepath"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/clusterstate"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/server/v3/embed"
	"go.uber.org/zap"
)

func TestMigrations(ctx context.Context, db *pachsql.DB) (retErr error) {
	// Create test dirs for etcd data
	dir, err := os.MkdirTemp("", "test-migrations")
	if err != nil {
		return errors.Wrap(err, "create etcd server tmpdir")
	}
	defer os.RemoveAll(dir)

	etcdConfig := embed.NewConfig()
	etcdConfig.MaxTxnOps = 10000
	etcdConfig.Dir = filepath.Join(dir, "dir")
	etcdConfig.WalDir = filepath.Join(dir, "wal")
	etcdConfig.InitialElectionTickAdvance = false
	etcdConfig.TickMs = 10
	etcdConfig.ElectionMs = 50
	etcdConfig.ListenPeerUrls = []url.URL{}
	etcdConfig.ListenClientUrls = []url.URL{{
		Scheme: "http",
		Host:   "localhost:7777",
	}}
	log.AddLoggerToEtcdServer(ctx, etcdConfig)
	etcd, err := embed.StartEtcd(etcdConfig)
	if err != nil {
		return errors.Wrap(err, "start etcd")
	}
	defer etcd.Close()

	etcdCfg := log.GetEtcdClientConfig(ctx)
	etcdCfg.Endpoints = []string{"http://localhost:7777"}
	etcdCfg.DialOptions = client.DefaultDialOptions()
	etcdClient, err := clientv3.New(etcdCfg)
	if err != nil {
		return errors.Wrap(err, "connect to etcd")
	}
	defer etcdClient.Close()

	txx, err := db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
	})
	if err != nil {
		return errors.Wrap(err, "start tx")
	}
	defer func() {
		if err := txx.Rollback(); err != nil {
			errors.JoinInto(&retErr, errors.Wrap(err, "rollback"))
			return
		}
		log.Info(ctx, "txn rolled back ok")
	}()
	states := migrations.CollectStates(nil, clusterstate.DesiredClusterState)
	env := migrations.MakeEnv(nil, etcdClient)
	env.Tx = txx
	env.WithTableLocks = false
	for _, s := range states {
		if err := migrations.ApplyMigrationTx(ctx, env, s); err != nil {
			log.Error(ctx, "migration did not apply; continuing", zap.Error(err))
			errors.JoinInto(&retErr, errors.Wrapf(err, "migration %v", s.Number()))
		}
	}
	log.Info(ctx, "done applying migrations")
	return
}
