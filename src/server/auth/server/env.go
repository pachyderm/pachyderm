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
	env    serviceenv.ServiceEnv
	txnEnv *txnenv.TransactionEnv
}

func EnvFromServiceEnv(senv serviceenv.ServiceEnv, txnEnv *txnenv.TransactionEnv) Env {
	return Env{
		env:    senv,
		txnEnv: txnEnv,
	}
}

// Delegations to the service environment.  These are explicit in order to make
// it clear which parts of the service environment are relied upon.
func (e Env) DB() *pachsql.DB                           { return e.env.GetDBClient() }
func (e Env) EtcdClient() *etcd.Client                  { return e.env.GetEtcdClient() }
func (e Env) Listener() col.PostgresListener            { return e.env.GetPostgresListener() }
func (e Env) GetEnterpriseServer() enterprise.APIServer { return e.env.EnterpriseServer() }
func (e Env) GetIdentityServer() identity.APIServer     { return e.env.IdentityServer() }
func (e Env) GetPfsServer() pfs.APIServer               { return e.env.PfsServer() }
func (e Env) GetPpsServer() pps.APIServer               { return e.env.PpsServer() }
func (e Env) BackgroundContext() context.Context        { return e.env.Context() }
func (e Env) Logger() *logrus.Logger                    { return e.env.Logger() }
func (e Env) Config() serviceenv.Configuration          { return *e.env.Config() }
