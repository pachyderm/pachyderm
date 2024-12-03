package clusterstate

import (
	"context"
	ec "github.com/pachyderm/pachyderm/v2/src/enterprise"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
)

const configKey = "config"

func EnterpriseConfigCollection(db *pachsql.DB, listener col.PostgresListener) col.PostgresCollection {
	return col.NewPostgresCollection(
		"enterpriseConfig",
		db,
		listener,
		&ec.EnterpriseConfig{},
		nil,
	)
}

// EnterpriseConfigPostgresMigration is not properly documented.
//
// The enterpriseConfig collection stores the information necessary for the enterprise-service to
// heartbeat to the license service for ongoing license validity checks. For clusters with enterprise,
// if this information were lost, the cluster would eventually become locked out. We migrate
// This data is migrated to postgres so that the data stored in etcd can truly be considered ephemeral.
//
// TODO: document.
func EnterpriseConfigPostgresMigration(ctx context.Context, tx *pachsql.Tx, etcd *clientv3.Client) error {
	if err := col.SetupPostgresCollections(ctx, tx, EnterpriseConfigCollection(nil, nil)); err != nil {
		return err
	}
	config, err := checkForEtcdRecord(ctx, etcd)
	if err != nil {
		return err
	}
	if config != nil {
		return errors.EnsureStack(EnterpriseConfigCollection(nil, nil).ReadWrite(tx).Put(ctx, configKey, config))
	}
	return nil
}

func checkForEtcdRecord(ctx context.Context, etcd *clientv3.Client) (*ec.EnterpriseConfig, error) {
	etcdConfigCol := col.NewEtcdCollection(etcd, "", nil, &ec.EnterpriseConfig{}, nil, nil)
	var config ec.EnterpriseConfig
	if err := etcdConfigCol.ReadOnly().Get(ctx, configKey, &config); err != nil {
		if col.IsErrNotFound(err) {
			return nil, nil
		}
		return nil, errors.EnsureStack(err)
	}
	return &config, nil
}

func DeleteEnterpriseConfigFromEtcd(ctx context.Context, etcd *clientv3.Client) error {
	if _, err := col.NewSTM(ctx, etcd, func(stm col.STM) error {
		etcdConfigCol := col.NewEtcdCollection(etcd, "", nil, &ec.EnterpriseConfig{}, nil, nil)
		return errors.EnsureStack(etcdConfigCol.ReadWrite(stm).Delete(ctx, configKey))
	}); err != nil {
		if !col.IsErrNotFound(err) {
			return err
		}
	}
	return nil
}

var state_2_1_0 migrations.State = state_2_0_0.
	Apply("Move EnterpriseConfig from etcd -> postgres", func(ctx context.Context, env migrations.Env) error {
		return EnterpriseConfigPostgresMigration(ctx, env.Tx, env.EtcdClient)
	}).
	Apply("Remove old EnterpriseConfig record from etcd", func(ctx context.Context, env migrations.Env) error {
		return DeleteEnterpriseConfigFromEtcd(ctx, env.EtcdClient)
	}).
	Apply("create pfs cache v1", func(ctx context.Context, env migrations.Env) error {
		return fileset.CreatePostgresCacheV1(ctx, env.Tx)
	}, migrations.Squash)

	// DO NOT MODIFY THIS STATE
	// IT HAS ALREADY SHIPPED IN A RELEASE
