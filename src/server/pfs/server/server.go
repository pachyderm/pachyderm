package server

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachconfig"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"
	etcd "go.etcd.io/etcd/client/v3"
)

type PipelineInspector interface {
	InspectPipelineInTransaction(context.Context, *txncontext.TransactionContext, *pps.Pipeline) (*pps.PipelineInfo, error)
}

// PFSAuth contains the auth methods called by PFS.
// It is a subset of what the Auth Service provides.
type PFSAuth interface {
	CheckRepoIsAuthorized(ctx context.Context, repo *pfs.Repo, p ...auth.Permission) error
	WhoAmI(ctx context.Context, req *auth.WhoAmIRequest) (*auth.WhoAmIResponse, error)
	GetPermissions(ctx context.Context, req *auth.GetPermissionsRequest) (*auth.GetPermissionsResponse, error)

	CheckProjectIsAuthorizedInTransaction(txnCtx *txncontext.TransactionContext, project *pfs.Project, p ...auth.Permission) error
	CheckRepoIsAuthorizedInTransaction(txnCtx *txncontext.TransactionContext, repo *pfs.Repo, p ...auth.Permission) error
	CreateRoleBindingInTransaction(txnCtx *txncontext.TransactionContext, principal string, roleSlice []string, resource *auth.Resource) error
	DeleteRoleBindingInTransaction(transactionContext *txncontext.TransactionContext, resource *auth.Resource) error
	GetPermissionsInTransaction(txnCtx *txncontext.TransactionContext, req *auth.GetPermissionsRequest) (*auth.GetPermissionsResponse, error)
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

	Auth                 PFSAuth
	GetPipelineInspector func() PipelineInspector

	BackgroundContext context.Context
	StorageConfig     pachconfig.StorageConfiguration
}

// NewAPIServer creates an APIServer.
func NewAPIServer(env Env) (pfsserver.APIServer, error) {
	a, err := newAPIServer(env)
	if err != nil {
		return nil, err
	}
	return newValidatedAPIServer(a, env.Auth), nil
}
