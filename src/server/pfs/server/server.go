package server

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsload"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	authserver "github.com/pachyderm/pachyderm/v2/src/server/auth"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"
	etcd "go.etcd.io/etcd/client/v3"
)

type PipelineInspector interface {
	InspectPipelineInTransaction(context.Context, *txncontext.TransactionContext, *pps.Pipeline) (*pps.PipelineInfo, error)
}

// Env is the dependencies needed to run the PFS API server
type Env struct {
	ObjectClient obj.Client
	DB           *pachsql.DB
	EtcdPrefix   string
	EtcdClient   *etcd.Client
	TaskService  task.Service
	TxnEnv       *txnenv.TransactionEnv
	Listener     col.PostgresListener

	AuthServer           authserver.APIServer
	GetPipelineInspector func() PipelineInspector
	// TODO: remove this, the load tests need a pachClient
	GetPachClient func(ctx context.Context) *client.APIClient

	BackgroundContext context.Context
	StorageConfig     pachconfig.StorageConfiguration
	PachwInSidecar    bool
}

// NewAPIServer creates an APIServer.
func NewAPIServer(env Env) (pfsserver.APIServer, error) {
	a, err := newAPIServer(env)
	if err != nil {
		return nil, err
	}
	go func() {
		pfsload.Worker(env.GetPachClient(pctx.Child(env.BackgroundContext, "pfsload")), env.TaskService) //nolint:errcheck
	}()
	return newValidatedAPIServer(a, env.AuthServer), nil
}

// TODO: remove this after load tests have been moved to Debug Server
func NewSidecarAPIServer(env Env) (pfsserver.APIServer, error) {
	a, err := newAPIServer(env)
	if err != nil {
		return nil, err
	}
	return newValidatedAPIServer(a, env.AuthServer), nil
}

// NewPachwAPIServer is used when running pachd in Pachw Mode.
// In Pachw Mode, a pachd instance processes storage and URl related tasks via the task service.
// TODO: remove this after load tests have been moved to Debug server.
func NewPachwAPIServer(env Env) (pfsserver.APIServer, error) {
	a, err := newAPIServer(env)
	if err != nil {
		return nil, err
	}
	go func() { pfsload.Worker(env.GetPachClient(env.BackgroundContext), env.TaskService) }() //nolint:errcheck
	return newValidatedAPIServer(a, env.AuthServer), nil
}
