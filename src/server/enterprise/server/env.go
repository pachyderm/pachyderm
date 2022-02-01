package server

import (
	"context"

	clientv3 "go.etcd.io/etcd/client/v3"
	kube "k8s.io/client-go/kubernetes"

	"github.com/pachyderm/pachyderm/v2/src/client"
	ec "github.com/pachyderm/pachyderm/v2/src/enterprise"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	"github.com/pachyderm/pachyderm/v2/src/server/auth"
)

type Env struct {
	DB       *pachsql.DB
	Listener col.PostgresListener
	TxnEnv   *txnenv.TransactionEnv

	EtcdClient *clientv3.Client
	EtcdPrefix string

	AuthServer    auth.APIServer
	GetPachClient func(context.Context) *client.APIClient
	GetKubeClient func() *kube.Clientset

	BackgroundContext context.Context
	Namespace         string
}

func EnvFromServiceEnv(senv serviceenv.ServiceEnv, etcdPrefix string, txEnv *txnenv.TransactionEnv) Env {
	e := Env{
		DB:       senv.GetDBClient(),
		Listener: senv.GetPostgresListener(),
		TxnEnv:   txEnv,

		EtcdClient: senv.GetEtcdClient(),
		EtcdPrefix: etcdPrefix,

		AuthServer:    senv.AuthServer(),
		GetPachClient: senv.GetPachClient,
		GetKubeClient: senv.GetKubeClient,

		BackgroundContext: senv.Context(),
		Namespace:         senv.Config().Namespace,
	}
	if e.Namespace == "" {
		e.Namespace = "default"
	}
	return e
}

func EnterpriseConfigCollection(db *pachsql.DB, listener col.PostgresListener) col.PostgresCollection {
	return col.NewPostgresCollection(
		"enterpriseConfig",
		db,
		listener,
		&ec.EnterpriseConfig{},
		nil,
	)
}

// The enterpriseConfig collection stores the information necessary for the enterprise-service to
// heartbeat to the license service for ongoing license validity checks. For clusters with enterprise,
// if this information were lost, the cluster would eventually become locked out. We migrate
// This data is migrated to postgres so that the data stored in etcd can truly be considered ephemeral.
func EnterpriseConfigPostgresMigration(ctx context.Context, tx *pachsql.Tx, etcd *clientv3.Client) error {
	if err := col.SetupPostgresCollections(ctx, tx, EnterpriseConfigCollection(nil, nil)); err != nil {
		return err
	}
	config, err := checkForEtcdRecord(ctx, etcd)
	if err != nil {
		return err
	}
	if config != nil {
		return errors.EnsureStack(EnterpriseConfigCollection(nil, nil).ReadWrite(tx).Put(configKey, config))
	}
	return nil
}

func checkForEtcdRecord(ctx context.Context, etcd *clientv3.Client) (*ec.EnterpriseConfig, error) {
	etcdConfigCol := col.NewEtcdCollection(etcd, "", nil, &ec.EnterpriseConfig{}, nil, nil)
	var config ec.EnterpriseConfig
	if err := etcdConfigCol.ReadOnly(ctx).Get(configKey, &config); err != nil {
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
		return errors.EnsureStack(etcdConfigCol.ReadWrite(stm).Delete(configKey))
	}); err != nil {
		if !col.IsErrNotFound(err) {
			return err
		}
	}
	return nil
}

// IsPaused returns true if the enterprise configuration indicates that this
// cluster should be paused.
func (env Env) IsPaused(ctx context.Context) (bool, error) {
	var (
		config    ec.EnterpriseConfig
		configCol = EnterpriseConfigCollection(env.DB, env.Listener)
	)
	if err := configCol.ReadOnly(ctx).Get(configKey, &config); err != nil {
		if col.IsErrNotFound(err) {
			return false, nil
		}
		return false, errors.EnsureStack(err)
	}
	return config.Paused, nil
}

// StopWorkers stops all workers
func (env Env) StopWorkers(ctx context.Context) error {
	return scaleDownWorkers(ctx, env.GetKubeClient(), env.Namespace)
}
