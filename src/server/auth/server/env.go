package server

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/identity"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	"github.com/pachyderm/pachyderm/v2/src/server/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/server/pfs"
	"github.com/pachyderm/pachyderm/v2/src/server/pps"
	logrus "github.com/sirupsen/logrus"
	etcd "go.etcd.io/etcd/client/v3"
)

// Env is the environment required for an apiServer
type Env struct {
	DB         *pachsql.DB
	EtcdClient *etcd.Client
	Listener   col.PostgresListener
	TxnEnv     *txnenv.TransactionEnv

	// circular dependency
	GetEnterpriseServer func() enterprise.APIServer
	GetIdentityServer   func() identity.APIServer
	GetPfsServer        func() pfs.APIServer
	GetPpsServer        func() pps.APIServer

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
		GetIdentityServer:   senv.IdentityServer,
		GetPfsServer:        senv.PfsServer,
		GetPpsServer:        senv.PpsServer,

		BackgroundContext: senv.Context(),
		Logger:            senv.Logger(),
		Config:            *senv.Config(),
	}
}
