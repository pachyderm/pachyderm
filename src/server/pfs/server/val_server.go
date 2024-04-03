package server

import (
	"context"
	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"

	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"
)

// TODO: Block tmp repo writes.

type validatedAPIServer struct {
	*apiServer
	auth PFSAuth
}

func newValidatedAPIServer(embeddedServer *apiServer, auth PFSAuth) *validatedAPIServer {
	return &validatedAPIServer{
		apiServer: embeddedServer,
		auth:      auth,
	}
}

// DeleteRepoInTransaction is identical to DeleteRepo except that it can run
// inside an existing etcd STM transaction.  This is not an RPC.
func (a *validatedAPIServer) DeleteRepoInTransaction(ctx context.Context, txnCtx *txncontext.TransactionContext, request *pfs.DeleteRepoRequest) (bool, error) {
	if request.Repo == nil {
		return false, errors.New("must specify repo")
	}
	return a.apiServer.DeleteRepoInTransaction(ctx, txnCtx, request)
}

// FinishCommitInTransaction is identical to FinishCommit except that it can run
// inside an existing postgres transaction.  This is not an RPC.
func (a *validatedAPIServer) FinishCommitInTransaction(ctx context.Context, txnCtx *txncontext.TransactionContext, request *pfs.FinishCommitRequest) error {
	userCommit := request.Commit
	// Validate arguments
	if err := checkCommit(userCommit); err != nil {
		return errors.Wrap(err, "check new file commit")
	}
	if err := a.auth.CheckRepoIsAuthorizedInTransaction(txnCtx, userCommit.Repo, auth.Permission_REPO_WRITE); err != nil {
		return errors.EnsureStack(err)
	}
	return a.apiServer.FinishCommitInTransaction(ctx, txnCtx, request)
}

// InspectFile implements the protobuf pfs.InspectFile RPC
func (a *validatedAPIServer) InspectFile(ctx context.Context, request *pfs.InspectFileRequest) (response *pfs.FileInfo, retErr error) {
	if err := validateFile(request.File); err != nil {
		return nil, err
	}
	if err := a.auth.CheckRepoIsAuthorized(ctx, request.File.Commit.Repo, auth.Permission_REPO_INSPECT_FILE); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return a.apiServer.InspectFile(ctx, request)
}

// ListFile implements the protobuf pfs.ListFile RPC
func (a *validatedAPIServer) ListFile(request *pfs.ListFileRequest, server pfs.API_ListFileServer) (retErr error) {
	if err := validateFile(request.File); err != nil {
		return err
	}
	if err := a.auth.CheckRepoIsAuthorized(server.Context(), request.File.Commit.Repo, auth.Permission_REPO_LIST_FILE); err != nil {
		return errors.EnsureStack(err)
	}
	return a.apiServer.ListFile(request, server)
}

// GetFileSet implements the protobuf pfs.GetFileSet RPC
func (a *validatedAPIServer) GetFileSet(ctx context.Context, req *pfs.GetFileSetRequest) (resp *pfs.CreateFileSetResponse, retErr error) {
	if err := checkCommit(req.Commit); err != nil {
		return nil, err
	}
	return a.apiServer.GetFileSet(ctx, req)
}

// WalkFile implements the protobuf pfs.WalkFile RPC
func (a *validatedAPIServer) WalkFile(request *pfs.WalkFileRequest, server pfs.API_WalkFileServer) (retErr error) {
	file := request.File
	// Validate arguments
	if err := validateFile(file); err != nil {
		return err
	}
	if err := a.auth.CheckRepoIsAuthorized(server.Context(), file.Commit.Repo, auth.Permission_REPO_READ, auth.Permission_REPO_LIST_FILE); err != nil {
		return errors.EnsureStack(err)
	}
	return a.apiServer.WalkFile(request, server)
}

// GlobFile implements the protobuf pfs.GlobFile RPC
func (a *validatedAPIServer) GlobFile(request *pfs.GlobFileRequest, server pfs.API_GlobFileServer) (retErr error) {
	commit := request.Commit
	// Validate arguments
	if err := checkCommit(commit); err != nil {
		return err
	}
	if err := a.auth.CheckRepoIsAuthorized(server.Context(), commit.Repo, auth.Permission_REPO_READ, auth.Permission_REPO_LIST_FILE); err != nil {
		return errors.EnsureStack(err)
	}
	return a.apiServer.GlobFile(request, server)
}

func (a *validatedAPIServer) ClearCommit(ctx context.Context, req *pfs.ClearCommitRequest) (*emptypb.Empty, error) {
	if err := checkCommit(req.Commit); err != nil {
		return nil, err
	}
	if err := a.auth.CheckRepoIsAuthorized(ctx, req.Commit.Repo, auth.Permission_REPO_WRITE); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return a.apiServer.ClearCommit(ctx, req)
}

func (a *validatedAPIServer) InspectCommit(ctx context.Context, req *pfs.InspectCommitRequest) (response *pfs.CommitInfo, retErr error) {
	if err := checkCommit(req.Commit); err != nil {
		return nil, err
	}
	return a.apiServer.InspectCommit(ctx, req)
}

func (a *validatedAPIServer) InspectCommitSet(request *pfs.InspectCommitSetRequest, server pfs.API_InspectCommitSetServer) error {
	if request.CommitSet == nil {
		return errors.New("commitset cannot be nil")
	}
	return a.apiServer.InspectCommitSet(request, server)
}

// ListCommit implements the protobuf pfs.ListCommit RPC
func (a *validatedAPIServer) ListCommit(req *pfs.ListCommitRequest, respServer pfs.API_ListCommitServer) (retErr error) {
	if req.To != nil {
		if err := checkCommit(req.To); err != nil {
			return err
		}
	}
	if req.From != nil {
		if err := checkCommit(req.From); err != nil {
			return err
		}
	}
	return a.apiServer.ListCommit(req, respServer)
}

func (a *validatedAPIServer) SquashCommitSet(ctx context.Context, request *pfs.SquashCommitSetRequest) (*emptypb.Empty, error) {
	if request.CommitSet == nil {
		return nil, errors.New("commitset cannot be nil")
	}
	return a.apiServer.SquashCommitSet(ctx, request)
}

func (a *validatedAPIServer) GetFile(request *pfs.GetFileRequest, server pfs.API_GetFileServer) error {
	if err := validateFile(request.File); err != nil {
		return err
	}
	return a.apiServer.GetFile(request, server)
}

func (a *validatedAPIServer) GetFileTAR(request *pfs.GetFileRequest, server pfs.API_GetFileTARServer) error {
	if err := validateFile(request.File); err != nil {
		return err
	}
	return a.apiServer.GetFileTAR(request, server)
}

func (a *validatedAPIServer) CreateBranchInTransaction(ctx context.Context, txnCtx *txncontext.TransactionContext, request *pfs.CreateBranchRequest) error {
	if request.Head != nil {
		if err := checkCommit(request.Head); err != nil {
			return err
		}
		if request.Branch.Repo.Name != request.Head.Repo.Name {
			return errors.New("branch and head commit must belong to the same repo")
		}
	}
	return a.apiServer.CreateBranchInTransaction(ctx, txnCtx, request)
}

func (a *validatedAPIServer) Egress(ctx context.Context, request *pfs.EgressRequest) (*pfs.EgressResponse, error) {
	if err := pfsserver.ValidateSQLDatabaseEgress(request.GetSqlDatabase()); err != nil {
		return nil, err
	}
	if err := checkCommit(request.Commit); err != nil {
		return nil, err
	}
	return a.apiServer.Egress(ctx, request)
}

func (a *validatedAPIServer) DiffFile(request *pfs.DiffFileRequest, server pfs.API_DiffFileServer) error {
	if request.NewFile == nil {
		return errors.New("file cannot be nil")
	}
	if err := checkCommit(request.NewFile.Commit); err != nil {
		return errors.Wrap(err, "check new file commit")
	}
	if request.OldFile != nil {
		if err := checkCommit(request.OldFile.Commit); err != nil {
			return errors.Wrap(err, "check old file commit")
		}
	}
	return a.apiServer.DiffFile(request, server)
}

func (a *validatedAPIServer) AddFileSet(ctx context.Context, req *pfs.AddFileSetRequest) (_ *emptypb.Empty, retErr error) {
	if err := checkCommit(req.Commit); err != nil {
		return nil, err
	}
	return a.apiServer.AddFileSet(ctx, req)
}

func (a *validatedAPIServer) SubscribeCommit(request *pfs.SubscribeCommitRequest, stream pfs.API_SubscribeCommitServer) (retErr error) {
	if request.From != nil {
		if err := checkCommit(request.From); err != nil {
			return err
		}
	}
	return a.apiServer.SubscribeCommit(request, stream)
}

func (a *validatedAPIServer) StartCommit(ctx context.Context, request *pfs.StartCommitRequest) (response *pfs.Commit, retErr error) {
	if request.Parent != nil {
		if err := checkCommit(request.Parent); err != nil {
			return nil, err
		}
	}
	return a.apiServer.StartCommit(ctx, request)
}

func (a *validatedAPIServer) FindCommits(request *pfs.FindCommitsRequest, srv pfs.API_FindCommitsServer) error {
	if request.Start != nil {
		if err := checkCommit(request.Start); err != nil {
			return err
		}
	}
	return a.apiServer.FindCommits(request, srv)
}

// WalkCommitProvenance implements the protobuf pfs.WalkCommitProvenance RPC
func (a *validatedAPIServer) WalkCommitProvenance(request *pfs.WalkCommitProvenanceRequest, srv pfs.API_WalkCommitProvenanceServer) error {
	return errors.Wrap(dbutil.WithTx(srv.Context(), a.driver.env.DB, func(ctx context.Context, tx *pachsql.Tx) error {
		commits, err := a.pickCommits(ctx, tx, request.Start)
		if err != nil {
			return errors.Wrap(err, "walk commit provenance")
		}
		return a.apiServer.WalkCommitProvenanceTx(ctx, tx,
			&WalkCommitProvenanceRequest{
				StartWithID:                 commits,
				WalkCommitProvenanceRequest: request,
			}, srv)
	}), "walk commit provenance")
}

// WalkCommitSubvenance implements the protobuf pfs.WalkCommitSubvenance RPC
func (a *validatedAPIServer) WalkCommitSubvenance(request *pfs.WalkCommitSubvenanceRequest, srv pfs.API_WalkCommitSubvenanceServer) error {
	return errors.Wrap(dbutil.WithTx(srv.Context(), a.driver.env.DB, func(ctx context.Context, tx *pachsql.Tx) error {
		commits, err := a.pickCommits(ctx, tx, request.Start)
		if err != nil {
			return errors.Wrap(err, "walk commit subvenance")
		}
		return a.apiServer.WalkCommitSubvenanceTx(ctx, tx,
			&WalkCommitSubvenanceRequest{
				StartWithID:                 commits,
				WalkCommitSubvenanceRequest: request,
			}, srv)
	}), "walk commit subvenance")
}

// WalkBranchProvenance implements the protobuf pfs.WalkBranchProvenance RPC
func (a *validatedAPIServer) WalkBranchProvenance(request *pfs.WalkBranchProvenanceRequest, srv pfs.API_WalkBranchProvenanceServer) error {
	return errors.Wrap(dbutil.WithTx(srv.Context(), a.driver.env.DB, func(ctx context.Context, tx *pachsql.Tx) error {
		branches, err := a.pickBranches(ctx, tx, request.Start)
		if err != nil {
			return err
		}
		return a.apiServer.WalkBranchProvenanceTx(ctx, tx,
			&WalkBranchProvenanceRequest{
				StartWithID:                 branches,
				WalkBranchProvenanceRequest: request,
			}, srv)
	}), "walk branch provenance")
}

// WalkBranchSubvenance implements the protobuf pfs.WalkBranchSubvenance RPC
func (a *validatedAPIServer) WalkBranchSubvenance(request *pfs.WalkBranchSubvenanceRequest, srv pfs.API_WalkBranchSubvenanceServer) error {
	return errors.Wrap(dbutil.WithTx(srv.Context(), a.driver.env.DB, func(ctx context.Context, tx *pachsql.Tx) error {
		branches, err := a.pickBranches(ctx, tx, request.Start)
		if err != nil {
			return err
		}
		return a.apiServer.WalkBranchSubvenanceTx(ctx, tx,
			&WalkBranchSubvenanceRequest{
				StartWithID:                 branches,
				WalkBranchSubvenanceRequest: request,
			}, srv)
	}), "walk branch subvenance")
}

func (a *validatedAPIServer) pickCommits(ctx context.Context, tx *pachsql.Tx, pickers []*pfs.CommitPicker) ([]*pfsdb.CommitWithID, error) {
	commits := make([]*pfsdb.CommitWithID, 0)
	for _, picker := range pickers {
		commit, err := pfsdb.PickCommit(ctx, picker, tx)
		if err != nil {
			return nil, errors.Wrap(err, "pick commits")
		}
		commits = append(commits, commit)
	}
	return commits, nil
}

func (a *validatedAPIServer) pickBranches(ctx context.Context, tx *pachsql.Tx, pickers []*pfs.BranchPicker) ([]*pfsdb.BranchInfoWithID, error) {
	branches := make([]*pfsdb.BranchInfoWithID, 0)
	for _, picker := range pickers {
		branch, err := pfsdb.PickBranch(ctx, picker, tx)
		if err != nil {
			return nil, errors.Wrap(err, "pick branches")
		}
		branches = append(branches, branch)
	}
	return branches, nil
}

func validateFile(file *pfs.File) error {
	if file == nil {
		return errors.New("file cannot be nil")
	}
	return checkCommit(file.Commit)
}

// popualtes c.Repo using c.Branch.Repo if necessary
func checkCommit(c *pfs.Commit) error {
	if c == nil {
		return errors.New("commit cannot be nil")
	}
	c.Repo = c.AccessRepo()
	if c.Repo == nil {
		return errors.Errorf("commit must have a repo")
	}
	c.GetBranch().GetRepo().EnsureProject()
	c.GetRepo().EnsureProject()
	return nil
}
