package server

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/client"
	ec "github.com/pachyderm/pachyderm/v2/src/enterprise"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	"github.com/pachyderm/pachyderm/v2/src/server/auth"
	logrus "github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type Env struct {
	DB       *pachsql.DB
	Listener col.PostgresListener
	TxnEnv   *txnenv.TransactionEnv

	EtcdClient *clientv3.Client
	EtcdPrefix string

	AuthServer    auth.APIServer
	GetPachClient func(context.Context) *client.APIClient

	Logger            *logrus.Logger
	BackgroundContext context.Context
}

func EnvFromServiceEnv(senv serviceenv.ServiceEnv, etcdPrefix string) Env {
	return Env{
		DB:       senv.GetDBClient(),
		Listener: senv.GetPostgresListener(),

		EtcdClient: senv.GetEtcdClient(),
		EtcdPrefix: etcdPrefix,

		AuthServer:    senv.AuthServer(),
		GetPachClient: senv.GetPachClient,

		Logger:            senv.Logger(),
		BackgroundContext: senv.Context(),
	}
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

func TryEnterpriseConfigPostgresMigration(ctx context.Context, migrationEnv migrations.Env) error {
	config, err := checkForEtcdRecord(ctx, migrationEnv)
	if err != nil {
		return err
	}
	if config != nil {
		return EnterpriseConfigCollection(nil, nil).ReadWrite(migrationEnv.Tx).Put(configKey, config)
	}
	return nil
}

func checkForEtcdRecord(ctx context.Context, migrationEnv migrations.Env) (*ec.EnterpriseConfig, error) {
	etcdConfigCol := col.NewEtcdCollection(migrationEnv.EtcdClient, migrationEnv.EtcdPrefix, nil, &ec.EnterpriseConfig{}, nil, nil)
	var config ec.EnterpriseConfig
	if err := etcdConfigCol.ReadOnly(ctx).Get(configKey, &config); err != nil {
		if col.IsErrNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return &config, nil
}
