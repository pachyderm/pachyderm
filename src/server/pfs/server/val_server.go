package server

import (
	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/src/auth"
	"github.com/pachyderm/pachyderm/src/internal/errors"
	"github.com/pachyderm/pachyderm/src/internal/serviceenv"
	txnenv "github.com/pachyderm/pachyderm/src/internal/transactionenv"
	"github.com/pachyderm/pachyderm/src/pfs"
	authserver "github.com/pachyderm/pachyderm/src/server/auth/server"
	"golang.org/x/net/context"
)

// TODO: Block tmp repo writes.

var _ APIServer = &validatedAPIServer{}

type validatedAPIServer struct {
	APIServer
	env *serviceenv.ServiceEnv
}

func newValidatedAPIServer(embeddedServer APIServer, env *serviceenv.ServiceEnv) *validatedAPIServer {
	return &validatedAPIServer{
		APIServer: embeddedServer,
		env:       env,
	}
}

// DeleteRepoInTransaction is identical to DeleteRepo except that it can run
// inside an existing etcd STM transaction.  This is not an RPC.
func (a *validatedAPIServer) DeleteRepoInTransaction(txnCtx *txnenv.TransactionContext, request *pfs.DeleteRepoRequest) error {
	repo := request.Repo
	// Check if the caller is authorized to delete this repo
	if err := authserver.CheckIsAuthorizedInTransaction(txnCtx, repo, auth.Scope_OWNER); err != nil {
		return err
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
	if userCommit.Repo == nil {
		return errors.New("commit repo cannot be nil")
	}
	if err := authserver.CheckIsAuthorizedInTransaction(txnCtx, userCommit.Repo, auth.Scope_WRITER); err != nil {
		return err
	}
	return a.APIServer.FinishCommitInTransaction(txnCtx, request)
}

// DeleteCommitInTransaction is identical to DeleteCommit except that it can run
// inside an existing etcd STM transaction.  This is not an RPC.
func (a *validatedAPIServer) DeleteCommitInTransaction(txnCtx *txnenv.TransactionContext, request *pfs.DeleteCommitRequest) error {
	userCommit := request.Commit
	// Validate arguments
	if userCommit == nil {
		return errors.New("commit cannot be nil")
	}
	if userCommit.Repo == nil {
		return errors.New("commit repo cannot be nil")
	}
	if err := authserver.CheckIsAuthorizedInTransaction(txnCtx, userCommit.Repo, auth.Scope_WRITER); err != nil {
		return err
	}
	return a.APIServer.DeleteCommitInTransaction(txnCtx, request)
}

// CopyFile implements the protobuf pfs.CopyFile RPC
func (a *validatedAPIServer) CopyFile(ctx context.Context, request *pfs.CopyFileRequest) (response *types.Empty, retErr error) {
	src, dst := request.Src, request.Dst
	// Validate arguments
	if src == nil {
		return nil, errors.New("src cannot be nil")
	}
	if src.Commit == nil {
		return nil, errors.New("src commit cannot be nil")
	}
	if src.Commit.Repo == nil {
		return nil, errors.New("src commit repo cannot be nil")
	}
	if dst == nil {
		return nil, errors.New("dst cannot be nil")
	}
	if dst.Commit == nil {
		return nil, errors.New("dst commit cannot be nil")
	}
	if dst.Commit.Repo == nil {
		return nil, errors.New("dst commit repo cannot be nil")
	}
	// authorization
	if err := authserver.CheckIsAuthorized(a.env.GetPachClient(ctx), src.Commit.Repo, auth.Scope_READER); err != nil {
		return nil, err
	}
	if err := authserver.CheckIsAuthorized(a.env.GetPachClient(ctx), dst.Commit.Repo, auth.Scope_WRITER); err != nil {
		return nil, err
	}
	if err := checkFilePath(dst.Path); err != nil {
		return nil, err
	}
	return a.APIServer.CopyFile(ctx, request)
}

// InspectFile implements the protobuf pfs.InspectFile RPC
func (a *validatedAPIServer) InspectFile(ctx context.Context, request *pfs.InspectFileRequest) (response *pfs.FileInfo, retErr error) {
	if err := validateFile(request.File); err != nil {
		return nil, err
	}
	if err := authserver.CheckIsAuthorized(a.env.GetPachClient(ctx), request.File.Commit.Repo, auth.Scope_READER); err != nil {
		return nil, err
	}
	return a.APIServer.InspectFile(ctx, request)
}

// ListFile implements the protobuf pfs.ListFile RPC
func (a *validatedAPIServer) ListFile(request *pfs.ListFileRequest, server pfs.API_ListFileServer) (retErr error) {
	if err := validateFile(request.File); err != nil {
		return err
	}
	if err := authserver.CheckIsAuthorized(a.env.GetPachClient(server.Context()), request.File.Commit.Repo, auth.Scope_READER); err != nil {
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
	if file.Commit.Repo == nil {
		return errors.New("file commit repo cannot be nil")
	}
	if err := authserver.CheckIsAuthorized(a.env.GetPachClient(server.Context()), file.Commit.Repo, auth.Scope_READER); err != nil {
		return err
	}
	return a.APIServer.WalkFile(request, server)
}

// GlobFile implements the protobuf pfs.GlobFile RPC
func (a *validatedAPIServer) GlobFile(request *pfs.GlobFileRequest, server pfs.API_GlobFileServer) (retErr error) {
	commit := request.Commit
	// Validate arguments
	if commit == nil {
		return errors.New("commit cannot be nil")
	}
	if commit.Repo == nil {
		return errors.New("commit repo cannot be nil")
	}
	if err := authserver.CheckIsAuthorized(a.env.GetPachClient(server.Context()), commit.Repo, auth.Scope_READER); err != nil {
		return err
	}
	return a.APIServer.GlobFile(request, server)
}

func (a *validatedAPIServer) ClearCommit(ctx context.Context, req *pfs.ClearCommitRequest) (*types.Empty, error) {
	if req.Commit == nil {
		return nil, errors.Errorf("commit cannot be nil")
	}
	if err := authserver.CheckIsAuthorized(a.env.GetPachClient(ctx), req.Commit.Repo, auth.Scope_WRITER); err != nil {
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
	if file.Commit.Repo == nil {
		return errors.New("file commit repo cannot be nil")
	}
	return nil
}
