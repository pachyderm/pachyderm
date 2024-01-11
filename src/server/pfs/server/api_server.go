package server

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/dbutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/pachsql"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsdb"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/metrics"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"
	taskapi "github.com/pachyderm/pachyderm/v2/src/task"
)

// apiServer implements the public interface of the Pachyderm File System,
// including all RPCs defined in the protobuf spec.  Implementation details
// occur in the 'driver' code, and this layer serves to translate the protobuf
// request structures into normal function calls.
type apiServer struct {
	pfs.UnimplementedAPIServer
	env    Env
	driver *driver
}

func newAPIServer(env Env) (*apiServer, error) {
	d, err := newDriver(env)
	if err != nil {
		return nil, err
	}
	s := &apiServer{
		env:    env,
		driver: d,
	}
	return s, nil
}

// ActivateAuth implements the protobuf pfs.ActivateAuth RPC
func (a *apiServer) ActivateAuth(ctx context.Context, request *pfs.ActivateAuthRequest) (response *pfs.ActivateAuthResponse, retErr error) {
	var resp *pfs.ActivateAuthResponse
	if err := a.env.TxnEnv.WithWriteContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		var err error
		resp, err = a.ActivateAuthInTransaction(ctx, txnCtx, request)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return resp, nil
}

func (a *apiServer) ActivateAuthInTransaction(ctx context.Context, txnCtx *txncontext.TransactionContext, request *pfs.ActivateAuthRequest) (response *pfs.ActivateAuthResponse, retErr error) {
	// Create role bindings for projects created before auth activation
	if err := pfsdb.ForEachProject(ctx, txnCtx.SqlTx, func(proj pfsdb.ProjectWithID) error {
		var principal string
		var roleSlice []string
		if proj.ProjectInfo.Project.Name == pfs.DefaultProjectName {
			// Grant all users ProjectWriter role for default project.
			principal = auth.AllClusterUsersSubject
			roleSlice = []string{auth.ProjectWriterRole}
		}
		err := a.env.Auth.CreateRoleBindingInTransaction(txnCtx, principal, roleSlice,
			&auth.Resource{Type: auth.ResourceType_PROJECT, Name: proj.ProjectInfo.Project.Name})
		if err != nil && !col.IsErrExists(err) {
			return errors.Wrap(err, "activate auth in transaction")
		}
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "activate auth in transaction")
	}
	// Create role bindings for repos created before auth activation
	if err := pfsdb.ForEachRepo(ctx, txnCtx.SqlTx, nil, func(repoInfoWithID pfsdb.RepoInfoWithID) error {
		err := a.env.Auth.CreateRoleBindingInTransaction(txnCtx, "", nil, repoInfoWithID.RepoInfo.Repo.AuthResource())
		if err != nil && !col.IsErrExists(err) {
			return errors.EnsureStack(err)
		}
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "activate auth in transaction")
	}
	return &pfs.ActivateAuthResponse{}, nil
}

// CreateRepoInTransaction is identical to CreateRepo except that it can run
// inside an existing postgres transaction.  This is not an RPC.
func (a *apiServer) CreateRepoInTransaction(ctx context.Context, txnCtx *txncontext.TransactionContext, request *pfs.CreateRepoRequest) error {
	if repo := request.GetRepo(); repo != nil && repo.Name == fileSetsRepo {
		return errors.Errorf("%s is a reserved name", fileSetsRepo)
	}
	return a.driver.createRepo(ctx, txnCtx, request.Repo, request.Description, request.Update)
}

// CreateRepo implements the protobuf pfs.CreateRepo RPC
func (a *apiServer) CreateRepo(ctx context.Context, request *pfs.CreateRepoRequest) (response *emptypb.Empty, retErr error) {
	request.Repo.EnsureProject()
	if err := a.env.TxnEnv.WithTransaction(ctx, func(txn txnenv.Transaction) error {
		return errors.EnsureStack(txn.CreateRepo(request))
	}); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// InspectRepoInTransaction is identical to InspectRepo except that it can run
// inside an existing postgres transaction.  This is not an RPC.
func (a *apiServer) InspectRepoInTransaction(ctx context.Context, txnCtx *txncontext.TransactionContext, originalRequest *pfs.InspectRepoRequest) (*pfs.RepoInfo, error) {
	request := proto.Clone(originalRequest).(*pfs.InspectRepoRequest)
	return a.driver.inspectRepo(ctx, txnCtx, request.Repo, true)
}

// InspectRepo implements the protobuf pfs.InspectRepo RPC
func (a *apiServer) InspectRepo(ctx context.Context, request *pfs.InspectRepoRequest) (response *pfs.RepoInfo, retErr error) {
	request.Repo.EnsureProject()
	var repoInfo *pfs.RepoInfo
	var size int64
	if err := a.env.TxnEnv.WithReadContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		var err error
		repoInfo, err = a.InspectRepoInTransaction(ctx, txnCtx, request)
		if err != nil {
			return err
		}
		size, err = a.driver.repoSize(ctx, txnCtx, repoInfo)
		return err
	}); err != nil {
		return nil, err
	}
	if repoInfo.Details == nil {
		repoInfo.Details = &pfs.RepoInfo_Details{}
	}
	repoInfo.Details.SizeBytes = size
	return repoInfo, nil
}

// ListRepo implements the protobuf pfs.ListRepo RPC
func (a *apiServer) ListRepo(request *pfs.ListRepoRequest, srv pfs.API_ListRepoServer) (retErr error) {
	var repos []*pfs.RepoInfo
	var err error
	if err := errors.Wrap(dbutil.WithTx(srv.Context(), a.env.DB, func(ctx context.Context, tx *pachsql.Tx) error {
		return a.driver.txnEnv.WithReadContext(ctx, func(txnCxt *txncontext.TransactionContext) error {
			repos, err = a.driver.listRepoInTransaction(srv.Context(), txnCxt, true, request.Type, request.Projects)
			if err != nil {
				return err
			}
			return nil
		})
	}, dbutil.WithReadOnly()), "list repo"); err != nil {
		return err
	}
	for _, repo := range repos {
		if err := errors.Wrap(srv.Send(repo), "sending repo"); err != nil {
			return err
		}
	}
	return nil
}

// DeleteRepoInTransaction is identical to DeleteRepo except that it can run
// inside an existing postgres transaction.  This is not an RPC.
func (a *apiServer) DeleteRepoInTransaction(ctx context.Context, txnCtx *txncontext.TransactionContext, request *pfs.DeleteRepoRequest) (bool, error) {
	return a.driver.deleteRepo(ctx, txnCtx, request.Repo, request.Force)
}

// DeleteRepoInTransaction is identical to DeleteRepo except that it can run
// inside an existing postgres transaction.  This is not an RPC.
func (a *apiServer) DeleteReposInTransaction(ctx context.Context, txnCtx *txncontext.TransactionContext, repos []*pfs.Repo, force bool) error {
	var ris []*pfs.RepoInfo
	for _, r := range repos {
		ri, err := a.driver.inspectRepo(ctx, txnCtx, r, false) // evaluate auth in d.deleteReposHelper()
		if err != nil {
			return errors.Wrap(err, "list repos for delete")
		}
		ris = append(ris, ri)
	}
	if _, err := a.driver.deleteReposHelper(ctx, txnCtx, ris, force); err != nil {
		return err
	}
	return nil
}

// DeleteRepo implements the protobuf pfs.DeleteRepo RPC
func (a *apiServer) DeleteRepo(ctx context.Context, request *pfs.DeleteRepoRequest) (response *pfs.DeleteRepoResponse, retErr error) {
	request.GetRepo().EnsureProject()
	if request.GetRepo() == nil {
		return nil, status.Error(codes.InvalidArgument, "no repo specified")
	}
	result := &pfs.DeleteRepoResponse{}
	if err := a.env.TxnEnv.WithTransaction(ctx, func(txn txnenv.Transaction) error {
		repoDeleted, err := txn.DeleteRepo(request)
		if err != nil {
			return errors.Wrap(err, "delete repo")
		}
		result.Deleted = repoDeleted
		return nil
	}); err != nil {
		return nil, err
	}
	return result, nil
}

// DeleteRepos implements the pfs.DeleteRepo RPC.  It deletes more than one repo at once.
func (a *apiServer) DeleteRepos(ctx context.Context, request *pfs.DeleteReposRequest) (resp *pfs.DeleteReposResponse, err error) {
	var repos []*pfs.Repo
	switch {
	case request.All:
		repos, err = a.driver.deleteRepos(ctx, nil, request.Force)
	case len(request.Projects) > 0:
		repos, err = a.driver.deleteRepos(ctx, request.Projects, request.Force)
	}
	if err != nil {
		return nil, err
	}
	return &pfs.DeleteReposResponse{
		Repos: repos,
	}, nil
}

// StartCommitInTransaction is identical to StartCommit except that it can run
// inside an existing postgres transaction.  This is not an RPC.
func (a *apiServer) StartCommitInTransaction(ctx context.Context, txnCtx *txncontext.TransactionContext, request *pfs.StartCommitRequest) (*pfs.Commit, error) {
	return a.driver.startCommit(ctx, txnCtx, request.Parent, request.Branch, request.Description)
}

// StartCommit implements the protobuf pfs.StartCommit RPC
func (a *apiServer) StartCommit(ctx context.Context, request *pfs.StartCommitRequest) (response *pfs.Commit, retErr error) {
	var err error
	request.GetBranch().GetRepo().EnsureProject()
	commit := &pfs.Commit{}
	if err = a.env.TxnEnv.WithTransaction(ctx, func(txn txnenv.Transaction) error {
		commit, err = txn.StartCommit(request)
		return errors.EnsureStack(err)
	}); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return commit, nil
}

// FinishCommitInTransaction is identical to FinishCommit except that it can run
// inside an existing postgres transaction.  This is not an RPC.
func (a *apiServer) FinishCommitInTransaction(ctx context.Context, txnCtx *txncontext.TransactionContext, request *pfs.FinishCommitRequest) error {
	return metrics.ReportRequest(func() error {
		return a.driver.finishCommit(ctx, txnCtx, request.Commit, request.Description, request.Error, request.Force)
	})
}

// FinishCommit implements the protobuf pfs.FinishCommit RPC
func (a *apiServer) FinishCommit(ctx context.Context, request *pfs.FinishCommitRequest) (response *emptypb.Empty, retErr error) {
	if err := a.env.TxnEnv.WithTransaction(ctx, func(txn txnenv.Transaction) error {
		return errors.EnsureStack(txn.FinishCommit(request))
	}); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// InspectCommitInTransaction is identical to InspectCommit (some features
// excluded) except that it can run inside an existing postgres transaction.
// This is not an RPC.
func (a *apiServer) InspectCommitInTransaction(ctx context.Context, txnCtx *txncontext.TransactionContext, request *pfs.InspectCommitRequest) (*pfs.CommitInfo, error) {
	return a.driver.resolveCommit(ctx, txnCtx.SqlTx, request.Commit)
}

// InspectCommit implements the protobuf pfs.InspectCommit RPC
func (a *apiServer) InspectCommit(ctx context.Context, request *pfs.InspectCommitRequest) (response *pfs.CommitInfo, retErr error) {
	return a.driver.inspectCommit(ctx, request.Commit, request.Wait)
}

// ListCommit implements the protobuf pfs.ListCommit RPC
func (a *apiServer) ListCommit(request *pfs.ListCommitRequest, respServer pfs.API_ListCommitServer) (retErr error) {
	request.GetRepo().EnsureProject()
	return a.driver.listCommit(respServer.Context(), request.Repo, request.To, request.From, request.StartedTime, request.Number, request.Reverse, request.All, request.OriginKind, func(ci *pfs.CommitInfo) error {
		return errors.EnsureStack(respServer.Send(ci))
	})
}

func (a *apiServer) SquashCommit(ctx context.Context, request *pfs.SquashCommitRequest) (*pfs.SquashCommitResponse, error) {
	if err := a.env.TxnEnv.WithWriteContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		return a.driver.squashCommit(ctx, txnCtx, request.Commit, request.Recursive)
	}); err != nil {
		return nil, err
	}
	return &pfs.SquashCommitResponse{}, nil
}

func (a *apiServer) DropCommit(ctx context.Context, request *pfs.DropCommitRequest) (*pfs.DropCommitResponse, error) {
	if err := a.env.TxnEnv.WithWriteContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		return a.driver.dropCommit(ctx, txnCtx, request.Commit, request.Recursive)
	}); err != nil {
		return nil, err
	}
	return &pfs.DropCommitResponse{}, nil
}

// InspectCommitSetInTransaction performs the same job as InspectCommitSet
// without the option of blocking for commits to finish so that it can run
// inside an existing postgres transaction.  This is not an RPC.
func (a *apiServer) InspectCommitSetInTransaction(ctx context.Context, txnCtx *txncontext.TransactionContext, commitset *pfs.CommitSet, includeAliases bool) ([]*pfs.CommitInfo, error) {
	return a.driver.inspectCommitSetImmediateTx(ctx, txnCtx, commitset, includeAliases)
}

// InspectCommitSet implements the protobuf pfs.InspectCommitSet RPC
func (a *apiServer) InspectCommitSet(request *pfs.InspectCommitSetRequest, server pfs.API_InspectCommitSetServer) (retErr error) {
	var count int
	if err := a.driver.inspectCommitSet(server.Context(), request.CommitSet, request.Wait, func(ci *pfs.CommitInfo) error {
		count++
		return server.Send(ci)
	}); err != nil {
		return err
	}
	if count == 0 {
		return pfsserver.ErrCommitSetNotFound{CommitSet: request.CommitSet}
	}
	return nil
}

// ListCommitSet implements the protobuf pfs.ListCommitSet RPC
func (a *apiServer) ListCommitSet(request *pfs.ListCommitSetRequest, serv pfs.API_ListCommitSetServer) (retErr error) {
	return a.driver.listCommitSet(serv.Context(), request.Project, func(commitSetInfo *pfs.CommitSetInfo) error {
		return errors.EnsureStack(serv.Send(commitSetInfo))
	})
}

// SquashCommitSetInTransaction is identical to SquashCommitSet except that it can run
// inside an existing postgres transaction.  This is not an RPC.
func (a *apiServer) SquashCommitSetInTransaction(ctx context.Context, txnCtx *txncontext.TransactionContext, request *pfs.SquashCommitSetRequest) error {
	return a.driver.squashCommitSet(ctx, txnCtx, request.CommitSet)
}

// SquashCommitSet implements the protobuf pfs.SquashCommitSet RPC
func (a *apiServer) SquashCommitSet(ctx context.Context, request *pfs.SquashCommitSetRequest) (response *emptypb.Empty, retErr error) {
	if err := a.env.TxnEnv.WithTransaction(ctx, func(txn txnenv.Transaction) error {
		return errors.EnsureStack(txn.SquashCommitSet(request))
	}); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// DropCommitSet implements the protobuf pfs.DropCommitSet RPC
func (a *apiServer) DropCommitSet(ctx context.Context, request *pfs.DropCommitSetRequest) (response *emptypb.Empty, retErr error) {
	if err := a.env.TxnEnv.WithWriteContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		return a.driver.dropCommitSet(ctx, txnCtx, request.CommitSet)
	}); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// SubscribeCommit implements the protobuf pfs.SubscribeCommit RPC
func (a *apiServer) SubscribeCommit(request *pfs.SubscribeCommitRequest, stream pfs.API_SubscribeCommitServer) (retErr error) {
	request.GetRepo().EnsureProject()
	return a.driver.subscribeCommit(stream.Context(), request.Repo, request.Branch, request.From, request.State, request.All, request.OriginKind, stream.Send)
}

// ClearCommit deletes all data in the commit.
func (a *apiServer) ClearCommit(ctx context.Context, request *pfs.ClearCommitRequest) (_ *emptypb.Empty, retErr error) {
	return &emptypb.Empty{}, a.driver.clearCommit(ctx, request.Commit)
}

// FindCommits searches for commits that reference a supplied file being modified in a branch.
func (a *apiServer) FindCommits(request *pfs.FindCommitsRequest, srv pfs.API_FindCommitsServer) error {
	var cancel context.CancelFunc
	ctx := srv.Context()
	deadline, ok := srv.Context().Deadline()
	if ok {
		// new context's deadline is shorter to give time to return last searched commit.
		ctx, cancel = context.WithTimeout(ctx, time.Duration(float64(time.Until(deadline))*0.85))
		defer cancel()
	}
	return a.driver.findCommits(ctx, request, srv.Send)
}

// CreateBranchInTransaction is identical to CreateBranch except that it can run
// inside an existing postgres transaction.  This is not an RPC.
func (a *apiServer) CreateBranchInTransaction(ctx context.Context, txnCtx *txncontext.TransactionContext, request *pfs.CreateBranchRequest) error {
	return a.driver.createBranch(ctx, txnCtx, request.Branch, request.Head, request.Provenance, request.Trigger)
}

// CreateBranch implements the protobuf pfs.CreateBranch RPC
func (a *apiServer) CreateBranch(ctx context.Context, request *pfs.CreateBranchRequest) (response *emptypb.Empty, retErr error) {
	request.GetBranch().GetRepo().EnsureProject()
	for _, b := range request.Provenance {
		b.GetRepo().EnsureProject()
	}
	if err := a.env.TxnEnv.WithTransaction(ctx, func(txn txnenv.Transaction) error {
		return errors.EnsureStack(txn.CreateBranch(request))
	}); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// InspectBranch implements the protobuf pfs.InspectBranch RPC
func (a *apiServer) InspectBranch(ctx context.Context, request *pfs.InspectBranchRequest) (response *pfs.BranchInfo, retErr error) {
	request.GetBranch().GetRepo().EnsureProject()
	return a.driver.inspectBranch(ctx, request.Branch)
}

func (a *apiServer) InspectBranchInTransaction(ctx context.Context, txnCtx *txncontext.TransactionContext, request *pfs.InspectBranchRequest) (*pfs.BranchInfo, error) {
	request.GetBranch().GetRepo().EnsureProject()
	return a.driver.inspectBranchInTransaction(ctx, txnCtx, request.Branch)
}

// ListBranch implements the protobuf pfs.ListBranch RPC
func (a *apiServer) ListBranch(request *pfs.ListBranchRequest, srv pfs.API_ListBranchServer) (retErr error) {
	request.GetRepo().EnsureProject()
	if request.Repo == nil {
		return a.driver.listBranch(srv.Context(), request.Reverse, srv.Send)
	}
	return a.env.TxnEnv.WithReadContext(srv.Context(), func(txnCtx *txncontext.TransactionContext) error {
		return a.driver.listBranchInTransaction(srv.Context(), txnCtx, request.Repo, request.Reverse, srv.Send)
	})
}

// DeleteBranchInTransaction is identical to DeleteBranch except that it can run
// inside an existing postgres transaction.  This is not an RPC.
func (a *apiServer) DeleteBranchInTransaction(ctx context.Context, txnCtx *txncontext.TransactionContext, request *pfs.DeleteBranchRequest) error {
	return a.driver.deleteBranch(ctx, txnCtx, request.Branch, request.Force)
}

// DeleteBranch implements the protobuf pfs.DeleteBranch RPC
func (a *apiServer) DeleteBranch(ctx context.Context, request *pfs.DeleteBranchRequest) (response *emptypb.Empty, retErr error) {
	request.GetBranch().GetRepo().EnsureProject()
	if err := a.env.TxnEnv.WithTransaction(ctx, func(txn txnenv.Transaction) error {
		return errors.EnsureStack(txn.DeleteBranch(request))
	}); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// CreateProject implements the protobuf pfs.CreateProject RPC
func (a *apiServer) CreateProject(ctx context.Context, request *pfs.CreateProjectRequest) (*emptypb.Empty, error) {
	if err := a.driver.createProject(ctx, request); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// InspectProject implements the protobuf pfs.InspectProject RPC
func (a *apiServer) InspectProject(ctx context.Context, request *pfs.InspectProjectRequest) (*pfs.ProjectInfo, error) {
	return a.driver.inspectProject(ctx, request.Project)
}

// InspectProjectV2 implements the protobuf pfs.InspectProjectV2 RPC
func (a *apiServer) InspectProjectV2(ctx context.Context, request *pfs.InspectProjectV2Request) (*pfs.InspectProjectV2Response, error) {
	info, err := a.driver.inspectProject(ctx, request.Project)
	if err != nil {
		return nil, errors.Wrapf(err, "could not inspect project %q", request.Project.String())
	}
	ppsServer := a.driver.env.GetPPSServer()
	if ppsServer == nil {
		return nil, errors.New("no PPS client set")
	}
	resp, err := ppsServer.GetProjectDefaults(ctx, &pps.GetProjectDefaultsRequest{Project: request.GetProject()})
	if err != nil {
		return nil, errors.Wrapf(err, "could not get project defaults for %q", request.Project.String())
	}
	return &pfs.InspectProjectV2Response{
		Info:         info,
		DefaultsJson: resp.GetProjectDefaultsJson(),
	}, nil
}

// ListProject implements the protobuf pfs.ListProject RPC
func (a *apiServer) ListProject(request *pfs.ListProjectRequest, srv pfs.API_ListProjectServer) error {
	return a.driver.listProject(srv.Context(), srv.Send)
}

// DeleteProject implements the protobuf pfs.DeleteProject RPC
func (a *apiServer) DeleteProject(ctx context.Context, request *pfs.DeleteProjectRequest) (*emptypb.Empty, error) {
	if err := a.env.TxnEnv.WithWriteContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		return a.driver.deleteProject(ctx, txnCtx, request.Project, request.Force)
	}); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (a *apiServer) ModifyFile(server pfs.API_ModifyFileServer) (retErr error) {
	commit, err := readCommit(server)
	if err != nil {
		return err
	}
	return metrics.ReportRequestWithThroughput(func() (int64, error) {
		var bytesRead int64
		if err := a.driver.modifyFile(server.Context(), commit, func(uw *fileset.UnorderedWriter) error {
			n, err := a.modifyFile(server.Context(), uw, server)
			if err != nil {
				return err
			}
			bytesRead += n
			return nil
		}); err != nil {
			return bytesRead, err
		}
		return bytesRead, errors.EnsureStack(server.SendAndClose(&emptypb.Empty{}))
	})
}

type modifyFileSource interface {
	Recv() (*pfs.ModifyFileRequest, error)
}

// modifyFile reads from a modifyFileSource until io.EOF and writes changes to an UnorderedWriter.
// SetCommit messages will result in an error.
func (a *apiServer) modifyFile(ctx context.Context, uw *fileset.UnorderedWriter, server modifyFileSource) (int64, error) {
	var bytesRead int64
	for {
		msg, err := server.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			return bytesRead, errors.EnsureStack(err)
		}
		switch mod := msg.Body.(type) {
		case *pfs.ModifyFileRequest_AddFile:
			var err error
			var n int64
			p := mod.AddFile.Path
			t := mod.AddFile.Datum
			switch src := mod.AddFile.Source.(type) {
			case *pfs.AddFile_Raw:
				n, err = putFileRaw(ctx, uw, p, t, src.Raw)
			case *pfs.AddFile_Url:
				n, err = putFileURL(ctx, a.env.TaskService, uw, p, t, src.Url)
			default:
				// need to write empty data to path
				n, err = putFileRaw(ctx, uw, p, t, &wrapperspb.BytesValue{})
			}
			if err != nil {
				return bytesRead, err
			}
			bytesRead += n
		case *pfs.ModifyFileRequest_DeleteFile:
			if err := deleteFile(ctx, uw, mod.DeleteFile); err != nil {
				return bytesRead, err
			}
		case *pfs.ModifyFileRequest_CopyFile:
			cf := mod.CopyFile
			if err := a.driver.copyFile(ctx, uw, cf.Dst, cf.Src, cf.Append, cf.Datum); err != nil {
				return bytesRead, err
			}
		case *pfs.ModifyFileRequest_SetCommit:
			return bytesRead, errors.Errorf("cannot set commit")
		default:
			return bytesRead, errors.Errorf("unrecognized message type")
		}
	}
	return bytesRead, nil
}

func putFileRaw(ctx context.Context, uw *fileset.UnorderedWriter, path, tag string, src *wrapperspb.BytesValue) (int64, error) {
	if err := uw.Put(ctx, path, tag, true, bytes.NewReader(src.Value)); err != nil {
		return 0, err
	}
	return int64(len(src.Value)), nil
}

func deleteFile(ctx context.Context, uw *fileset.UnorderedWriter, request *pfs.DeleteFile) error {
	return uw.Delete(ctx, request.Path, request.Datum)
}

// GetFileTAR implements the protobuf pfs.GetFileTAR RPC
func (a *apiServer) GetFileTAR(request *pfs.GetFileRequest, server pfs.API_GetFileTARServer) (retErr error) {
	return metrics.ReportRequestWithThroughput(func() (int64, error) {
		ctx := server.Context()
		if request.URL != "" {
			return a.driver.getFileURL(ctx, a.env.TaskService, request.URL, request.File, request.PathRange)
		}
		src, err := a.driver.getFile(ctx, request.File, request.PathRange)
		if err != nil {
			return 0, err
		}
		var bytesWritten int64
		err = grpcutil.WithStreamingBytesWriter(server, func(w io.Writer) error {
			var err error
			bytesWritten, err = withGetFileWriter(w, func(w io.Writer) error {
				return getFileTar(ctx, w, src)
			})
			return err
		})
		return bytesWritten, err
	})
}

// GetFile implements the protobuf pfs.GetFile RPC
func (a *apiServer) GetFile(request *pfs.GetFileRequest, server pfs.API_GetFileServer) (retErr error) {
	return metrics.ReportRequestWithThroughput(func() (int64, error) {
		ctx := server.Context()
		if request.URL != "" {
			return a.driver.getFileURL(ctx, a.env.TaskService, request.URL, request.File, request.PathRange)
		}
		src, err := a.driver.getFile(ctx, request.File, request.PathRange)
		if err != nil {
			return 0, err
		}
		if err := checkSingleFile(ctx, src); err != nil {
			return 0, err
		}
		var n int64
		if err := src.Iterate(ctx, func(fi *pfs.FileInfo, file fileset.File) error {
			n = fileset.SizeFromIndex(file.Index())
			return grpcutil.WithStreamingBytesWriter(server, func(w io.Writer) error {
				return errors.EnsureStack(file.Content(ctx, w, chunk.WithOffsetBytes(request.Offset)))
			})
		}); err != nil {
			return 0, errors.EnsureStack(err)
		}
		return n, nil
	})
}

func withGetFileWriter(w io.Writer, cb func(io.Writer) error) (int64, error) {
	gfw := &getFileWriter{w: w}
	err := cb(gfw)
	return gfw.bytesWritten, err
}

type getFileWriter struct {
	w            io.Writer
	bytesWritten int64
}

func (gfw *getFileWriter) Write(data []byte) (int, error) {
	n, err := gfw.w.Write(data)
	gfw.bytesWritten += int64(n)
	return n, errors.EnsureStack(err)
}

func getFileTar(ctx context.Context, w io.Writer, src Source) error {
	// TODO: remove absolute paths on the way out?
	// nonAbsolute := &fileset.HeaderMapper{
	// 	R: filter,
	// 	F: func(th *tar.Header) *tar.Header {
	// 		th.Name = "." + th.Name
	// 		return th
	// 	},
	// }
	if err := src.Iterate(ctx, func(fi *pfs.FileInfo, file fileset.File) error {
		return fileset.WriteTarEntry(ctx, w, file)
	}); err != nil {
		return errors.EnsureStack(err)
	}
	return errors.EnsureStack(tar.NewWriter(w).Close())
}

// InspectFile implements the protobuf pfs.InspectFile RPC
func (a *apiServer) InspectFile(ctx context.Context, request *pfs.InspectFileRequest) (response *pfs.FileInfo, retErr error) {
	request.GetFile().GetCommit().GetBranch().GetRepo().EnsureProject()
	request.GetFile().GetCommit().GetRepo().EnsureProject()
	return a.driver.inspectFile(ctx, request.File)
}

// ListFile implements the protobuf pfs.ListFile RPC
func (a *apiServer) ListFile(request *pfs.ListFileRequest, server pfs.API_ListFileServer) (retErr error) {
	return a.driver.listFile(server.Context(), request.File, request.PaginationMarker, request.Number, request.Reverse, func(fi *pfs.FileInfo) error {
		return errors.EnsureStack(server.Send(fi))
	})
}

// WalkFile implements the protobuf pfs.WalkFile RPC
func (a *apiServer) WalkFile(request *pfs.WalkFileRequest, server pfs.API_WalkFileServer) (retErr error) {
	return a.driver.walkFile(server.Context(), request.File, request.PaginationMarker, request.Number, request.Reverse, func(fi *pfs.FileInfo) error {
		return errors.EnsureStack(server.Send(fi))
	})
}

// GlobFile implements the protobuf pfs.GlobFile RPC
func (a *apiServer) GlobFile(request *pfs.GlobFileRequest, respServer pfs.API_GlobFileServer) (retErr error) {
	return a.driver.globFile(respServer.Context(), request.Commit, request.Pattern, request.PathRange, func(fi *pfs.FileInfo) error {
		return errors.EnsureStack(respServer.Send(fi))
	})
}

// DiffFile implements the protobuf pfs.DiffFile RPC
func (a *apiServer) DiffFile(request *pfs.DiffFileRequest, server pfs.API_DiffFileServer) (retErr error) {
	return a.driver.diffFile(server.Context(), request.OldFile, request.NewFile, func(oldFi, newFi *pfs.FileInfo) error {
		return errors.EnsureStack(server.Send(&pfs.DiffFileResponse{
			OldFile: oldFi,
			NewFile: newFi,
		}))
	})
}

// DeleteAll implements the protobuf pfs.DeleteAll RPC
func (a *apiServer) DeleteAll(ctx context.Context, request *emptypb.Empty) (response *emptypb.Empty, retErr error) {
	if err := a.driver.deleteAll(ctx); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// Fsck implements the protobuf pfs.Fsck RPC
func (a *apiServer) Fsck(request *pfs.FsckRequest, fsckServer pfs.API_FsckServer) (retErr error) {
	ctx := fsckServer.Context()
	if err := a.driver.fsck(ctx, request.Fix, func(resp *pfs.FsckResponse) error {
		return errors.EnsureStack(fsckServer.Send(resp))
	}); err != nil {
		return err
	}
	if target := request.GetZombieTarget(); target != nil {
		target.GetBranch().GetRepo().EnsureProject()
		return a.driver.detectZombie(ctx, target, fsckServer.Send)
	}
	var repos []*pfs.RepoInfo
	var err error
	if request.GetZombieAll() {
		if err := dbutil.WithTx(fsckServer.Context(), a.env.DB, func(ctx context.Context, tx *pachsql.Tx) error {
			return a.driver.txnEnv.WithWriteContext(ctx, func(txnCxt *txncontext.TransactionContext) error {
				// list meta repos as a proxy for finding pipelines
				repos, err = a.driver.listRepoInTransaction(ctx, txnCxt, false, pfs.MetaRepoType, nil)
				return errors.Wrap(err, "list repos in tx by meta repo type")
			})
		}, dbutil.WithReadOnly()); err != nil {
			return err
		}
		for _, info := range repos {
			// TODO: actually derive output branch from job/pipeline, currently that coupling causes issues
			output := client.NewCommit(info.Repo.Project.GetName(), info.Repo.Name, "master", "")
			for output != nil {
				info, err := a.driver.inspectCommit(ctx, output, pfs.CommitState_STARTED)
				if err != nil {
					return err
				}
				// we will be reading the whole file system, so unfinished commits would be very slow
				if info.Error == "" && info.Finished != nil {
					break
				}
				output = info.ParentCommit
			}
			if output == nil {
				return nil
			}
			if err := a.driver.detectZombie(ctx, output, fsckServer.Send); err != nil {
				return errors.Wrap(err, "fsck")
			}
		}
	}
	return nil
}

// CreateFileSet implements the pfs.CreateFileset RPC
func (a *apiServer) CreateFileSet(server pfs.API_CreateFileSetServer) (retErr error) {
	fsID, err := a.driver.createFileSet(server.Context(), func(uw *fileset.UnorderedWriter) error {
		_, err := a.modifyFile(server.Context(), uw, server)
		return err
	})
	if err != nil {
		return err
	}
	return errors.EnsureStack(server.SendAndClose(&pfs.CreateFileSetResponse{
		FileSetId: fsID.HexString(),
	}))
}

func (a *apiServer) GetFileSet(ctx context.Context, req *pfs.GetFileSetRequest) (resp *pfs.CreateFileSetResponse, retErr error) {
	if req.Type == pfs.GetFileSetRequest_DIFF {
		diff, err := a.driver.commitStore.GetDiffFileSet(ctx, req.Commit)
		if err != nil {
			return nil, err
		}
		return &pfs.CreateFileSetResponse{
			FileSetId: diff.HexString(),
		}, nil
	}
	filesetID, err := a.driver.getFileSet(ctx, req.Commit)
	if err != nil {
		return nil, err
	}
	return &pfs.CreateFileSetResponse{
		FileSetId: filesetID.HexString(),
	}, nil
}

func (a *apiServer) ShardFileSet(ctx context.Context, req *pfs.ShardFileSetRequest) (*pfs.ShardFileSetResponse, error) {
	fsid, err := fileset.ParseID(req.FileSetId)
	if err != nil {
		return nil, err
	}
	shards, err := a.driver.shardFileSet(ctx, *fsid)
	if err != nil {
		return nil, err
	}
	return &pfs.ShardFileSetResponse{
		Shards: shards,
	}, nil
}

func (a *apiServer) AddFileSet(ctx context.Context, req *pfs.AddFileSetRequest) (_ *emptypb.Empty, retErr error) {
	if err := a.env.TxnEnv.WithWriteContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		return a.AddFileSetInTransaction(ctx, txnCtx, req)
	}); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (a *apiServer) AddFileSetInTransaction(ctx context.Context, txnCtx *txncontext.TransactionContext, request *pfs.AddFileSetRequest) error {
	fsid, err := fileset.ParseID(request.FileSetId)
	if err != nil {
		return err
	}
	if err := a.driver.addFileSet(ctx, txnCtx, request.Commit, *fsid); err != nil {
		return err
	}
	return nil
}

// RenewFileSet implements the pfs.RenewFileSet RPC
func (a *apiServer) RenewFileSet(ctx context.Context, req *pfs.RenewFileSetRequest) (_ *emptypb.Empty, retErr error) {
	fsid, err := fileset.ParseID(req.FileSetId)
	if err != nil {
		return nil, err
	}
	if err := a.driver.renewFileSet(ctx, *fsid, time.Duration(req.TtlSeconds)*time.Second); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// ComposeFileSet implements the pfs.ComposeFileSet RPC
func (a *apiServer) ComposeFileSet(ctx context.Context, req *pfs.ComposeFileSetRequest) (resp *pfs.CreateFileSetResponse, retErr error) {
	var fsids []fileset.ID
	for _, id := range req.FileSetIds {
		fsid, err := fileset.ParseID(id)
		if err != nil {
			return nil, err
		}
		fsids = append(fsids, *fsid)
	}
	filesetID, err := a.driver.composeFileSet(ctx, fsids, time.Duration(req.TtlSeconds)*time.Second, req.Compact)
	if err != nil {
		return nil, err
	}
	return &pfs.CreateFileSetResponse{
		FileSetId: filesetID.HexString(),
	}, nil
}

func (a *apiServer) CheckStorage(ctx context.Context, req *pfs.CheckStorageRequest) (*pfs.CheckStorageResponse, error) {
	chunks := a.driver.storage.Chunks
	count, err := chunks.Check(ctx, req.ChunkBegin, req.ChunkEnd, req.ReadChunkData)
	if err != nil {
		return nil, err
	}
	return &pfs.CheckStorageResponse{
		ChunkObjectCount: int64(count),
	}, nil
}

func (a *apiServer) PutCache(ctx context.Context, req *pfs.PutCacheRequest) (resp *emptypb.Empty, retErr error) {
	var fsids []fileset.ID
	for _, id := range req.FileSetIds {
		fsid, err := fileset.ParseID(id)
		if err != nil {
			return nil, err
		}
		fsids = append(fsids, *fsid)
	}
	if err := a.driver.putCache(ctx, req.Key, req.Value, fsids, req.Tag); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (a *apiServer) GetCache(ctx context.Context, req *pfs.GetCacheRequest) (resp *pfs.GetCacheResponse, retErr error) {
	value, err := a.driver.getCache(ctx, req.Key)
	if err != nil {
		return nil, err
	}
	return &pfs.GetCacheResponse{Value: value}, nil
}

func (a *apiServer) ClearCache(ctx context.Context, req *pfs.ClearCacheRequest) (resp *emptypb.Empty, retErr error) {
	if err := a.driver.clearCache(ctx, req.TagPrefix); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (a *apiServer) ListTask(req *taskapi.ListTaskRequest, server pfs.API_ListTaskServer) error {
	return task.List(server.Context(), a.env.TaskService, req, server.Send)
}

func readCommit(srv pfs.API_ModifyFileServer) (*pfs.Commit, error) {
	msg, err := srv.Recv()
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	switch x := msg.Body.(type) {
	case *pfs.ModifyFileRequest_SetCommit:
		if err := checkCommit(x.SetCommit); err != nil {
			return nil, err
		}
		return x.SetCommit, nil
	default:
		return nil, errors.Errorf("first message must be a commit")
	}
}

func (a *apiServer) Egress(ctx context.Context, req *pfs.EgressRequest) (*pfs.EgressResponse, error) {
	file := req.Commit.NewFile("/")
	switch target := req.Target.(type) {
	case *pfs.EgressRequest_ObjectStorage:
		result, err := a.driver.copyToObjectStorage(ctx, a.env.TaskService, file, target.ObjectStorage.Url)
		if err != nil {
			return nil, errors.EnsureStack(err)
		}
		return &pfs.EgressResponse{Result: &pfs.EgressResponse_ObjectStorage{ObjectStorage: result}}, nil

	case *pfs.EgressRequest_SqlDatabase:
		src, err := a.driver.getFile(ctx, file, nil)
		if err != nil {
			return nil, err
		}
		result, err := copyToSQLDB(ctx, src, target.SqlDatabase.Url, target.SqlDatabase.FileFormat)
		if err != nil {
			return nil, errors.EnsureStack(err)
		}
		return &pfs.EgressResponse{Result: &pfs.EgressResponse_SqlDatabase{SqlDatabase: result}}, nil
	}
	return nil, errors.Errorf("egress failed")
}
