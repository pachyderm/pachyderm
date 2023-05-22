package server

import (
	"context"
	"path"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	authserver "github.com/pachyderm/pachyderm/v2/src/server/auth"
	ppsserver "github.com/pachyderm/pachyderm/v2/src/server/pps"
	etcd "go.etcd.io/etcd/client/v3"
)

// Env is the dependencies needed to run the PFS API server
type Env struct {
	ObjectClient obj.Client
	DB           *pachsql.DB
	EtcdPrefix   string
	EtcdClient   *etcd.Client
	TaskService  task.Service
	TxnEnv       *txnenv.TransactionEnv
	Listener     col.PostgresListener

	AuthServer authserver.APIServer
	// TODO: a reasonable repo metadata solution would let us get rid of this circular dependency
	// permissions might also work.
	GetPPSServer func() ppsserver.APIServer
	// TODO: remove this, the load tests need a pachClient
	GetPachClient func(ctx context.Context) *client.APIClient

	BackgroundContext context.Context
	StorageConfig     serviceenv.StorageConfiguration
	PachwInSidecar    bool
}

func EnvFromServiceEnv(env serviceenv.ServiceEnv, txnEnv *txnenv.TransactionEnv) (*Env, error) {
	// Setup etcd, object storage, and database clients.
	objClient, err := obj.NewClient(env.Context(), env.Config().StorageBackend, env.Config().StorageRoot)
	if err != nil {
		return nil, err
	}
	etcdPrefix := path.Join(env.Config().EtcdPrefix, env.Config().PFSEtcdPrefix)
	if env.AuthServer() == nil {
		panic("auth server cannot be nil")
	}
	return &Env{
		ObjectClient: objClient,
		DB:           env.GetDBClient(),
		TxnEnv:       txnEnv,
		Listener:     env.GetPostgresListener(),
		EtcdPrefix:   etcdPrefix,
		EtcdClient:   env.GetEtcdClient(),
		TaskService:  env.GetTaskService(etcdPrefix),

		AuthServer:    env.AuthServer(),
		GetPPSServer:  env.PpsServer,
		GetPachClient: env.GetPachClient,

		BackgroundContext: pctx.Child(env.Context(), "PFS"),
		StorageConfig:     env.Config().StorageConfiguration,
		PachwInSidecar:    env.Config().PachwInSidecars,
	}, nil
}
