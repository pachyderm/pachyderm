package server

import (
	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/serviceenv"
	txnenv "github.com/pachyderm/pachyderm/src/server/pkg/transactionenv"
	"golang.org/x/net/context"
)

var _ APIServer = &validatedAPIServer{}

type validatedAPIServer struct {
	APIServer
	env *serviceenv.ServiceEnv
}

func newValidatedAPIServer(inner APIServer, env *serviceenv.ServiceEnv) *validatedAPIServer {
	return &validatedAPIServer{
		APIServer: inner,
		env:       env,
	}
}

// CopyFile implements the protobuf pfs.CopyFile RPC
func (a *validatedAPIServer) CopyFile(ctx context.Context, req *pfs.CopyFileRequest) (response *types.Empty, retErr error) {
	src, dst := req.Src, req.Dst
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
	if err := a.checkIsAuthorized(ctx, src.Commit.Repo, auth.Scope_READER); err != nil {
		return nil, err
	}
	if err := a.checkIsAuthorized(ctx, dst.Commit.Repo, auth.Scope_WRITER); err != nil {
		return nil, err
	}
	if err := checkFilePath(dst.Path); err != nil {
		return nil, err
	}
	return a.APIServer.CopyFile(ctx, req)
}

// InspectFileV2 returns info about a file.
func (a *validatedAPIServer) InspectFileV2(ctx context.Context, req *pfs.InspectFileRequest) (*pfs.FileInfo, error) {
	if err := validateFile(req.File); err != nil {
		return nil, err
	}
	if err := a.checkIsAuthorized(ctx, req.File.Commit.Repo, auth.Scope_READER); err != nil {
		return nil, err
	}
	return a.APIServer.InspectFileV2(ctx, req)
}

// WalkFileV2 walks over all the files under a directory, including children of children.
func (a *validatedAPIServer) WalkFileV2(req *pfs.WalkFileRequest, server pfs.API_WalkFileV2Server) error {
	file := req.File
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
	if err := a.checkIsAuthorized(server.Context(), file.Commit.Repo, auth.Scope_READER); err != nil {
		return err
	}
	return a.APIServer.WalkFileV2(req, server)
}

func (a *validatedAPIServer) GlobFileV2(request *pfs.GlobFileRequest, server pfs.API_GlobFileV2Server) (retErr error) {
	commit := request.Commit
	// Validate arguments
	if commit == nil {
		return errors.New("commit cannot be nil")
	}
	if commit.Repo == nil {
		return errors.New("commit repo cannot be nil")
	}
	if err := a.checkIsAuthorized(server.Context(), commit.Repo, auth.Scope_READER); err != nil {
		return err
	}
	return a.APIServer.GlobFileV2(request, server)
}

func (a *validatedAPIServer) ListFileV2(req *pfs.ListFileRequest, server pfs.API_ListFileV2Server) error {
	if err := validateFile(req.File); err != nil {
		return err
	}
	if err := a.checkIsAuthorized(server.Context(), req.File.Commit.Repo, auth.Scope_READER); err != nil {
		return err
	}
	return a.APIServer.ListFileV2(req, server)
}

func (a *validatedAPIServer) ClearCommit(ctx context.Context, req *pfs.ClearCommitRequest) (*types.Empty, error) {
	if req.Commit == nil {
		return nil, errors.Errorf("commit cannot be nil")
	}
	if err := a.checkIsAuthorized(ctx, req.Commit.Repo, auth.Scope_WRITER); err != nil {
		return nil, err
	}
	return a.APIServer.ClearCommit(ctx, req)
}

// DeleteRepoInTransaction is identical to DeleteRepo except that it can run
// inside an existing etcd STM transaction.  This is not an RPC.
func (a *validatedAPIServer) DeleteRepoInTransaction(
	txnCtx *txnenv.TransactionContext,
	request *pfs.DeleteRepoRequest,
) error {
	repo := request.Repo
	// Check if the caller is authorized to delete this repo
	if err := a.checkIsAuthorizedInTransaction(txnCtx, repo, auth.Scope_OWNER); err != nil {
		return err
	}
	return a.APIServer.DeleteRepoInTransaction(txnCtx, request)
}

// DeleteCommitInTransaction is identical to DeleteCommit except that it can run
// inside an existing etcd STM transaction.  This is not an RPC.
func (a *validatedAPIServer) DeleteCommitInTransaction(
	txnCtx *txnenv.TransactionContext,
	request *pfs.DeleteCommitRequest,
) error {
	userCommit := request.Commit
	// Validate arguments
	if userCommit == nil {
		return errors.New("commit cannot be nil")
	}
	if userCommit.Repo == nil {
		return errors.New("commit repo cannot be nil")
	}
	if err := a.checkIsAuthorizedInTransaction(txnCtx, userCommit.Repo, auth.Scope_WRITER); err != nil {
		return err
	}
	return a.APIServer.DeleteCommitInTransaction(txnCtx, request)
}

func (a *validatedAPIServer) getAuth(ctx context.Context) client.AuthAPIClient {
	return a.env.GetPachClient(ctx)
}

func (a *validatedAPIServer) checkIsAuthorized(ctx context.Context, r *pfs.Repo, s auth.Scope) error {
	client := a.getAuth(ctx)
	me, err := client.WhoAmI(ctx, &auth.WhoAmIRequest{})
	if auth.IsErrNotActivated(err) {
		return nil
	}
	req := &auth.AuthorizeRequest{Repo: r.Name, Scope: s}
	resp, err := client.Authorize(ctx, req)
	if err != nil {
		return errors.Wrapf(grpcutil.ScrubGRPC(err), "error during authorization check for operation on \"%s\"", r.Name)
	}
	if !resp.Authorized {
		return &auth.ErrNotAuthorized{Subject: me.Username, Repo: r.Name, Required: s}
	}
	return nil
}

func (a *validatedAPIServer) checkIsAuthorizedInTransaction(txnCtx *txnenv.TransactionContext, r *pfs.Repo, s auth.Scope) error {
	me, err := txnCtx.Client.WhoAmI(txnCtx.ClientContext, &auth.WhoAmIRequest{})
	if auth.IsErrNotActivated(err) {
		return nil
	}

	req := &auth.AuthorizeRequest{Repo: r.Name, Scope: s}
	resp, err := txnCtx.Auth().AuthorizeInTransaction(txnCtx, req)
	if err != nil {
		return errors.Wrapf(grpcutil.ScrubGRPC(err), "error during authorization check for operation on \"%s\"", r.Name)
	}
	if !resp.Authorized {
		return &auth.ErrNotAuthorized{Subject: me.Username, Repo: r.Name, Required: s}
	}
	return nil
}
