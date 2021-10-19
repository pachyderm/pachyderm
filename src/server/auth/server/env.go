package server

import (
	etcd "github.com/coreos/etcd/clientv3"
	"github.com/jmoiron/sqlx"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	"github.com/pachyderm/pachyderm/v2/src/server/enterprise"
	logrus "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

// Env is the environment required for an apiServer
type Env struct {
	DB         *sqlx.DB
	EtcdClient *etcd.Client
	Listener   col.PostgresListener
	TxnEnv     *txnenv.TransactionEnv

	// circular dependency
	GetEnterpriseServer func() enterprise.APIServer

	BackgroundContext context.Context
	Logger            *logrus.Logger
	Config            serviceenv.Configuration
}

func EnvFromServiceEnv(senv serviceenv.ServiceEnv, txnEnv *txnenv.TransactionEnv) Env {
	return Env{
		DB:         senv.GetDBClient(),
		EtcdClient: senv.GetEtcdClient(),
		Listener:   senv.GetPostgresListener(),
		TxnEnv:     txnEnv,

		GetEnterpriseServer: senv.EnterpriseServer,

		BackgroundContext: senv.Context(),
		Logger:            senv.Logger(),
		Config:            *senv.Config(),
	}
}
