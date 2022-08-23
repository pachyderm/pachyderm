package server

import (
	"context"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	authserver "github.com/pachyderm/pachyderm/v2/src/server/auth"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"
)

// TODO: Block tmp repo writes.

type validatedAPIServer struct {
	*apiServer
	auth authserver.APIServer
}

func newValidatedAPIServer(embeddedServer *apiServer, auth authserver.APIServer) *validatedAPIServer {
	return &validatedAPIServer{
		apiServer: embeddedServer,
		auth:      auth,
	}
}

// DeleteRepoInTransaction is identical to DeleteRepo except that it can run
// inside an existing etcd STM transaction.  This is not an RPC.
func (a *validatedAPIServer) DeleteRepoInTransaction(txnCtx *txncontext.TransactionContext, request *pfs.DeleteRepoRequest) error {
	if request.Repo == nil {
		return errors.New("must specify repo")
	}
	return a.apiServer.DeleteRepoInTransaction(txnCtx, request)
}

// FinishCommitInTransaction is identical to FinishCommit except that it can run
// inside an existing etcd STM transaction.  This is not an RPC.
func (a *validatedAPIServer) FinishCommitInTransaction(txnCtx *txncontext.TransactionContext, request *pfs.FinishCommitRequest) error {
	userCommit := request.Commit
	// Validate arguments
	if userCommit == nil {
		return errors.New("commit cannot be nil")
	}
	if userCommit.Branch == nil {
		return errors.New("commit branch cannot be nil")
	}
	if userCommit.Branch.Repo == nil {
		return errors.New("commit repo cannot be nil")
	}
	if err := a.auth.CheckRepoIsAuthorizedInTransaction(txnCtx, userCommit.Branch.Repo, auth.Permission_REPO_WRITE); err != nil {
		return errors.EnsureStack(err)
	}
	return a.apiServer.FinishCommitInTransaction(txnCtx, request)
}

// InspectFile implements the protobuf pfs.InspectFile RPC
func (a *validatedAPIServer) InspectFile(ctx context.Context, request *pfs.InspectFileRequest) (response *pfs.FileInfo, retErr error) {
	if err := validateFile(request.File); err != nil {
		return nil, err
	}
	if err := a.auth.CheckRepoIsAuthorized(ctx, request.File.Commit.Branch.Repo, auth.Permission_REPO_INSPECT_FILE); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return a.apiServer.InspectFile(ctx, request)
}

// ListFile implements the protobuf pfs.ListFile RPC
func (a *validatedAPIServer) ListFile(request *pfs.ListFileRequest, server pfs.API_ListFileServer) (retErr error) {
	if err := validateFile(request.File); err != nil {
		return err
	}
	if err := a.auth.CheckRepoIsAuthorized(server.Context(), request.File.Commit.Branch.Repo, auth.Permission_REPO_LIST_FILE); err != nil {
		return errors.EnsureStack(err)
	}
	return a.apiServer.ListFile(request, server)
}

// WalkFile implements the protobuf pfs.WalkFile RPC
func (a *validatedAPIServer) WalkFile(request *pfs.WalkFileRequest, server pfs.API_WalkFileServer) (retErr error) {
	file := request.File
	// Validate arguments
	if file == nil {
		return errors.New("file cannot be nil")
	}
	if file.Commit == nil {
		return errors.New("file commit cannot be nil")
	}
	if file.Commit.Branch == nil {
		return errors.New("file branch cannot be nil")
	}
	if file.Commit.Branch.Repo == nil {
		return errors.New("file commit repo cannot be nil")
	}
	if err := a.auth.CheckRepoIsAuthorized(server.Context(), file.Commit.Branch.Repo, auth.Permission_REPO_READ, auth.Permission_REPO_LIST_FILE); err != nil {
		return errors.EnsureStack(err)
	}
	return a.apiServer.WalkFile(request, server)
}

// GlobFile implements the protobuf pfs.GlobFile RPC
func (a *validatedAPIServer) GlobFile(request *pfs.GlobFileRequest, server pfs.API_GlobFileServer) (retErr error) {
	commit := request.Commit
	// Validate arguments
	if commit == nil {
		return errors.New("commit cannot be nil")
	}
	if commit.Branch == nil {
		return errors.New("commit branch cannot be nil")
	}
	if commit.Branch.Repo == nil {
		return errors.New("commit repo cannot be nil")
	}
	if err := a.auth.CheckRepoIsAuthorized(server.Context(), commit.Branch.Repo, auth.Permission_REPO_READ, auth.Permission_REPO_LIST_FILE); err != nil {
		return errors.EnsureStack(err)
	}
	return a.apiServer.GlobFile(request, server)
}

func (a *validatedAPIServer) ClearCommit(ctx context.Context, req *pfs.ClearCommitRequest) (*types.Empty, error) {
	if req.Commit == nil {
		return nil, errors.Errorf("commit cannot be nil")
	}
	if err := a.auth.CheckRepoIsAuthorized(ctx, req.Commit.Branch.Repo, auth.Permission_REPO_WRITE); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return a.apiServer.ClearCommit(ctx, req)
}

func (a *validatedAPIServer) InspectCommit(ctx context.Context, req *pfs.InspectCommitRequest) (response *pfs.CommitInfo, retErr error) {
	if req.Commit == nil {
		return nil, errors.New("commit cannot be nil")
	}
	return a.apiServer.InspectCommit(ctx, req)
}

func (a *validatedAPIServer) InspectCommitSet(request *pfs.InspectCommitSetRequest, server pfs.API_InspectCommitSetServer) error {
	if request.CommitSet == nil {
		return errors.New("commitset cannot be nil")
	}
	return a.apiServer.InspectCommitSet(request, server)
}

func (a *validatedAPIServer) SquashCommitSet(ctx context.Context, request *pfs.SquashCommitSetRequest) (*types.Empty, error) {
	if request.CommitSet == nil {
		return nil, errors.New("commitset cannot be nil")
	}
	return a.apiServer.SquashCommitSet(ctx, request)
}

func (a *validatedAPIServer) GetFile(request *pfs.GetFileRequest, server pfs.API_GetFileServer) error {
	if request.File == nil {
		return errors.New("file cannot be nil")
	}
	if request.File.Commit == nil {
		return errors.New("commit cannot be nil")
	}
	if request.File.Commit.Branch == nil {
		return errors.New("branch cannot be nil")
	}
	if request.File.Commit.Branch.Repo == nil {
		return errors.New("repo cannot be nil")
	}
	return a.apiServer.GetFile(request, server)
}

func (a *validatedAPIServer) GetFileTAR(request *pfs.GetFileRequest, server pfs.API_GetFileTARServer) error {
	if request.File == nil {
		return errors.New("file cannot be nil")
	}
	if request.File.Commit == nil {
		return errors.New("commit cannot be nil")
	}
	if request.File.Commit.Branch == nil {
		return errors.New("branch cannot be nil")
	}
	if request.File.Commit.Branch.Repo == nil {
		return errors.New("repo cannot be nil")
	}
	return a.apiServer.GetFileTAR(request, server)
}

func (a *validatedAPIServer) CreateBranchInTransaction(txnCtx *txncontext.TransactionContext, request *pfs.CreateBranchRequest) error {
	if request.Head != nil && request.Branch.Repo.Name != request.Head.Branch.Repo.Name {
		return errors.New("branch and head commit must belong to the same repo")
	}
	return a.apiServer.CreateBranchInTransaction(txnCtx, request)
}

func (a *validatedAPIServer) Egress(ctx context.Context, request *pfs.EgressRequest) (*pfs.EgressResponse, error) {
	err := pfsserver.ValidateSQLDatabaseEgress(request.GetSqlDatabase())
	if err != nil {
		return nil, err
	}
	return a.apiServer.Egress(ctx, request)
}

func validateFile(file *pfs.File) error {
	if file == nil {
		return errors.New("file cannot be nil")
	}
	if file.Commit == nil {
		return errors.New("file commit cannot be nil")
	}
	if file.Commit.Branch == nil {
		return errors.New("file branch cannot be nil")
	}
	if file.Commit.Branch.Repo == nil {
		return errors.New("file commit repo cannot be nil")
	}
	return nil
}
