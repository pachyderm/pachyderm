package server

import (
	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	authserver "github.com/pachyderm/pachyderm/v2/src/server/auth/server"
	"golang.org/x/net/context"
)

// TODO: Block tmp repo writes.

var _ APIServer = &validatedAPIServer{}

type validatedAPIServer struct {
	APIServer
	env serviceenv.ServiceEnv
}

func newValidatedAPIServer(embeddedServer APIServer, env serviceenv.ServiceEnv) *validatedAPIServer {
	return &validatedAPIServer{
		APIServer: embeddedServer,
		env:       env,
	}
}

// DeleteRepoInTransaction is identical to DeleteRepo except that it can run
// inside an existing etcd STM transaction.  This is not an RPC.
func (a *validatedAPIServer) DeleteRepoInTransaction(txnCtx *txnenv.TransactionContext, request *pfs.DeleteRepoRequest) error {
	// TODO(2.0 required): How should auth be applied when using all?
	if !request.All {
		repo := request.Repo
		// Check if the caller is authorized to delete this repo
		if err := authserver.CheckRepoIsAuthorizedInTransaction(txnCtx, repo.Name, auth.Permission_REPO_DELETE); err != nil {
			return err
		}
	}
	return a.APIServer.DeleteRepoInTransaction(txnCtx, request)
}

// FinishCommitInTransaction is identical to FinishCommit except that it can run
// inside an existing etcd STM transaction.  This is not an RPC.
func (a *validatedAPIServer) FinishCommitInTransaction(txnCtx *txnenv.TransactionContext, request *pfs.FinishCommitRequest) error {
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
	if err := authserver.CheckRepoIsAuthorizedInTransaction(txnCtx, userCommit.Branch.Repo.Name, auth.Permission_REPO_WRITE); err != nil {
		return err
	}
	return a.APIServer.FinishCommitInTransaction(txnCtx, request)
}

// SquashCommitInTransaction is identical to SquashCommit except that it can run
// inside an existing etcd STM transaction.  This is not an RPC.
func (a *validatedAPIServer) SquashCommitInTransaction(txnCtx *txnenv.TransactionContext, request *pfs.SquashCommitRequest) error {
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
	if err := authserver.CheckRepoIsAuthorizedInTransaction(txnCtx, userCommit.Branch.Repo.Name, auth.Permission_REPO_DELETE_COMMIT); err != nil {
		return err
	}
	return a.APIServer.SquashCommitInTransaction(txnCtx, request)
}

// InspectFile implements the protobuf pfs.InspectFile RPC
func (a *validatedAPIServer) InspectFile(ctx context.Context, request *pfs.InspectFileRequest) (response *pfs.FileInfo, retErr error) {
	if err := validateFile(request.File); err != nil {
		return nil, err
	}
	if err := authserver.CheckRepoIsAuthorized(a.env.GetPachClient(ctx), request.File.Commit.Branch.Repo.Name, auth.Permission_REPO_INSPECT_FILE); err != nil {
		return nil, err
	}
	return a.APIServer.InspectFile(ctx, request)
}

// ListFile implements the protobuf pfs.ListFile RPC
func (a *validatedAPIServer) ListFile(request *pfs.ListFileRequest, server pfs.API_ListFileServer) (retErr error) {
	if err := validateFile(request.File); err != nil {
		return err
	}
	if err := authserver.CheckRepoIsAuthorized(a.env.GetPachClient(server.Context()), request.File.Commit.Branch.Repo.Name, auth.Permission_REPO_LIST_FILE); err != nil {
		return err
	}
	return a.APIServer.ListFile(request, server)
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
	if err := authserver.CheckRepoIsAuthorized(a.env.GetPachClient(server.Context()), file.Commit.Branch.Repo.Name, auth.Permission_REPO_READ, auth.Permission_REPO_LIST_FILE); err != nil {
		return err
	}
	return a.APIServer.WalkFile(request, server)
}

// FlushJob implements the protobuf pfs.FlushJob RPC
func (a *validatedpiServer) FlushJob(request *pfs.FlushJobRequest, server pfs.API_FlushJobServer) error {
	if request.Job == nil {
		return errors.New("job cannot be nil")
	}
	return a.APIServer.FlushJob(request, server)
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
	if err := authserver.CheckRepoIsAuthorized(a.env.GetPachClient(server.Context()), commit.Branch.Repo.Name, auth.Permission_REPO_READ, auth.Permission_REPO_LIST_FILE); err != nil {
		return err
	}
	return a.APIServer.GlobFile(request, server)
}

func (a *validatedAPIServer) ClearCommit(ctx context.Context, req *pfs.ClearCommitRequest) (*types.Empty, error) {
	if req.Commit == nil {
		return nil, errors.Errorf("commit cannot be nil")
	}
	if err := authserver.CheckRepoIsAuthorized(a.env.GetPachClient(ctx), req.Commit.Branch.Repo.Name, auth.Permission_REPO_WRITE); err != nil {
		return nil, err
	}
	return a.APIServer.ClearCommit(ctx, req)
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
