package server

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/ghodss/yaml"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"

	"github.com/pachyderm/pachyderm/v2/src/client"
	col "github.com/pachyderm/pachyderm/v2/src/internal/collection"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/miscutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsload"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/metrics"
	"github.com/pachyderm/pachyderm/v2/src/internal/task"
	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"
	taskapi "github.com/pachyderm/pachyderm/v2/src/task"
)

// apiServer implements the public interface of the Pachyderm File System,
// including all RPCs defined in the protobuf spec.  Implementation details
// occur in the 'driver' code, and this layer serves to translate the protobuf
// request structures into normal function calls.
type apiServer struct {
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
		resp, err = a.ActivateAuthInTransaction(txnCtx, request)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return resp, nil
}

func (a *apiServer) ActivateAuthInTransaction(txnCtx *txncontext.TransactionContext, request *pfs.ActivateAuthRequest) (response *pfs.ActivateAuthResponse, retErr error) {
	var repoInfo pfs.RepoInfo
	if err := a.driver.repos.ReadWrite(txnCtx.SqlTx).List(&repoInfo, col.DefaultOptions(), func(string) error {
		err := a.env.AuthServer.CreateRoleBindingInTransaction(txnCtx, "", nil, repoInfo.Repo.AuthResource())
		if err != nil && !col.IsErrExists(err) {
			return errors.EnsureStack(err)
		}
		return nil
	}); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return &pfs.ActivateAuthResponse{}, nil
}

// CreateRepoInTransaction is identical to CreateRepo except that it can run
// inside an existing postgres transaction.  This is not an RPC.
func (a *apiServer) CreateRepoInTransaction(txnCtx *txncontext.TransactionContext, request *pfs.CreateRepoRequest) error {
	if repo := request.GetRepo(); repo != nil && repo.Name == fileSetsRepo {
		return errors.Errorf("%s is a reserved name", fileSetsRepo)
	}
	return a.driver.createRepo(txnCtx, request.Repo, request.Description, request.Update)
}

// CreateRepo implements the protobuf pfs.CreateRepo RPC
func (a *apiServer) CreateRepo(ctx context.Context, request *pfs.CreateRepoRequest) (response *types.Empty, retErr error) {
	request.Repo.EnsureProject()
	if err := a.env.TxnEnv.WithTransaction(ctx, func(txn txnenv.Transaction) error {
		return errors.EnsureStack(txn.CreateRepo(request))
	}, nil); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

// InspectRepoInTransaction is identical to InspectRepo except that it can run
// inside an existing postgres transaction.  This is not an RPC.
func (a *apiServer) InspectRepoInTransaction(txnCtx *txncontext.TransactionContext, originalRequest *pfs.InspectRepoRequest) (*pfs.RepoInfo, error) {
	request := proto.Clone(originalRequest).(*pfs.InspectRepoRequest)
	return a.driver.inspectRepo(txnCtx, request.Repo, true)
}

// InspectRepo implements the protobuf pfs.InspectRepo RPC
func (a *apiServer) InspectRepo(ctx context.Context, request *pfs.InspectRepoRequest) (response *pfs.RepoInfo, retErr error) {
	request.Repo.EnsureProject()
	var repoInfo *pfs.RepoInfo
	err := a.env.TxnEnv.WithReadContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		var err error
		repoInfo, err = a.InspectRepoInTransaction(txnCtx, request)
		return err
	})
	if err != nil {
		return nil, err
	}
	size, err := a.driver.repoSize(ctx, repoInfo.Repo)
	if err != nil {
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
	return a.driver.listRepo(srv.Context(), true, request.Type, srv.Send)
}

// DeleteRepoInTransaction is identical to DeleteRepo except that it can run
// inside an existing postgres transaction.  This is not an RPC.
func (a *apiServer) DeleteRepoInTransaction(txnCtx *txncontext.TransactionContext, request *pfs.DeleteRepoRequest) error {
	return a.driver.deleteRepo(txnCtx, request.Repo, request.Force)
}

// DeleteRepo implements the protobuf pfs.DeleteRepo RPC
func (a *apiServer) DeleteRepo(ctx context.Context, request *pfs.DeleteRepoRequest) (response *types.Empty, retErr error) {
	request.GetRepo().EnsureProject()
	if err := a.env.TxnEnv.WithTransaction(ctx, func(txn txnenv.Transaction) error {
		return errors.EnsureStack(txn.DeleteRepo(request))
	}, nil); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

// StartCommitInTransaction is identical to StartCommit except that it can run
// inside an existing postgres transaction.  This is not an RPC.
func (a *apiServer) StartCommitInTransaction(txnCtx *txncontext.TransactionContext, request *pfs.StartCommitRequest) (*pfs.Commit, error) {
	return a.driver.startCommit(txnCtx, request.Parent, request.Branch, request.Description)
}

// StartCommit implements the protobuf pfs.StartCommit RPC
func (a *apiServer) StartCommit(ctx context.Context, request *pfs.StartCommitRequest) (response *pfs.Commit, retErr error) {
	var err error
	request.GetParent().GetBranch().GetRepo().EnsureProject()
	request.GetBranch().GetRepo().EnsureProject()
	commit := &pfs.Commit{}
	if err = a.env.TxnEnv.WithTransaction(ctx, func(txn txnenv.Transaction) error {
		commit, err = txn.StartCommit(request)
		return errors.EnsureStack(err)
	}, nil); err != nil {
		return nil, errors.EnsureStack(err)
	}
	return commit, nil
}

// FinishCommitInTransaction is identical to FinishCommit except that it can run
// inside an existing postgres transaction.  This is not an RPC.
func (a *apiServer) FinishCommitInTransaction(txnCtx *txncontext.TransactionContext, request *pfs.FinishCommitRequest) error {
	return metrics.ReportRequest(func() error {
		return a.driver.finishCommit(txnCtx, request.Commit, request.Description, request.Error, request.Force)
	})
}

// FinishCommit implements the protobuf pfs.FinishCommit RPC
func (a *apiServer) FinishCommit(ctx context.Context, request *pfs.FinishCommitRequest) (response *types.Empty, retErr error) {
	request.GetCommit().GetBranch().GetRepo().EnsureProject()
	if err := a.env.TxnEnv.WithTransaction(ctx, func(txn txnenv.Transaction) error {
		return errors.EnsureStack(txn.FinishCommit(request))
	}, nil); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

// InspectCommitInTransaction is identical to InspectCommit (some features
// excluded) except that it can run inside an existing postgres transaction.
// This is not an RPC.
func (a *apiServer) InspectCommitInTransaction(txnCtx *txncontext.TransactionContext, request *pfs.InspectCommitRequest) (*pfs.CommitInfo, error) {
	return a.driver.resolveCommit(txnCtx.SqlTx, request.Commit)
}

// InspectCommit implements the protobuf pfs.InspectCommit RPC
func (a *apiServer) InspectCommit(ctx context.Context, request *pfs.InspectCommitRequest) (response *pfs.CommitInfo, retErr error) {
	request.GetCommit().GetBranch().GetRepo().EnsureProject()
	return a.driver.inspectCommit(ctx, request.Commit, request.Wait)
}

// ListCommit implements the protobuf pfs.ListCommit RPC
func (a *apiServer) ListCommit(request *pfs.ListCommitRequest, respServer pfs.API_ListCommitServer) (retErr error) {
	request.GetRepo().EnsureProject()
	request.GetFrom().GetBranch().GetRepo().EnsureProject()
	request.GetTo().GetBranch().GetRepo().EnsureProject()
	return a.driver.listCommit(respServer.Context(), request.Repo, request.To, request.From, request.StartedTime, request.Number, request.Reverse, request.All, request.OriginKind, func(ci *pfs.CommitInfo) error {
		return errors.EnsureStack(respServer.Send(ci))
	})
}

// InspectCommitSetInTransaction performs the same job as InspectCommitSet
// without the option of blocking for commits to finish so that it can run
// inside an existing postgres transaction.  This is not an RPC.
func (a *apiServer) InspectCommitSetInTransaction(txnCtx *txncontext.TransactionContext, commitset *pfs.CommitSet) ([]*pfs.CommitInfo, error) {
	return a.driver.inspectCommitSetImmediate(txnCtx, commitset)
}

// InspectCommitSet implements the protobuf pfs.InspectCommitSet RPC
func (a *apiServer) InspectCommitSet(request *pfs.InspectCommitSetRequest, server pfs.API_InspectCommitSetServer) (retErr error) {
	return a.driver.inspectCommitSet(server.Context(), request.CommitSet, request.Wait, server.Send)
}

// ListCommitSet implements the protobuf pfs.ListCommitSet RPC
func (a *apiServer) ListCommitSet(request *pfs.ListCommitSetRequest, serv pfs.API_ListCommitSetServer) (retErr error) {
	return a.driver.listCommitSet(serv.Context(), func(commitSetInfo *pfs.CommitSetInfo) error {
		return errors.EnsureStack(serv.Send(commitSetInfo))
	})
}

// SquashCommitSetInTransaction is identical to SquashCommitSet except that it can run
// inside an existing postgres transaction.  This is not an RPC.
func (a *apiServer) SquashCommitSetInTransaction(txnCtx *txncontext.TransactionContext, request *pfs.SquashCommitSetRequest) error {
	return a.driver.squashCommitSet(txnCtx, request.CommitSet)
}

// SquashCommitSet implements the protobuf pfs.SquashCommitSet RPC
func (a *apiServer) SquashCommitSet(ctx context.Context, request *pfs.SquashCommitSetRequest) (response *types.Empty, retErr error) {
	if err := a.env.TxnEnv.WithTransaction(ctx, func(txn txnenv.Transaction) error {
		return errors.EnsureStack(txn.SquashCommitSet(request))
	}, nil); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

// DropCommitSet implements the protobuf pfs.DropCommitSet RPC
func (a *apiServer) DropCommitSet(ctx context.Context, request *pfs.DropCommitSetRequest) (response *types.Empty, retErr error) {
	if err := a.env.TxnEnv.WithWriteContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		return a.driver.dropCommitSet(txnCtx, request.CommitSet)
	}); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

// SubscribeCommit implements the protobuf pfs.SubscribeCommit RPC
func (a *apiServer) SubscribeCommit(request *pfs.SubscribeCommitRequest, stream pfs.API_SubscribeCommitServer) (retErr error) {
	request.GetRepo().EnsureProject()
	request.GetFrom().GetBranch().GetRepo().EnsureProject()
	return a.driver.subscribeCommit(stream.Context(), request.Repo, request.Branch, request.From, request.State, request.All, request.OriginKind, stream.Send)
}

// ClearCommit deletes all data in the commit.
func (a *apiServer) ClearCommit(ctx context.Context, request *pfs.ClearCommitRequest) (_ *types.Empty, retErr error) {
	request.GetCommit().GetBranch().GetRepo().EnsureProject()
	return &types.Empty{}, a.driver.clearCommit(ctx, request.Commit)
}

// CreateBranchInTransaction is identical to CreateBranch except that it can run
// inside an existing postgres transaction.  This is not an RPC.
func (a *apiServer) CreateBranchInTransaction(txnCtx *txncontext.TransactionContext, request *pfs.CreateBranchRequest) error {
	return a.driver.createBranch(txnCtx, request.Branch, request.Head, request.Provenance, request.Trigger)
}

// CreateBranch implements the protobuf pfs.CreateBranch RPC
func (a *apiServer) CreateBranch(ctx context.Context, request *pfs.CreateBranchRequest) (response *types.Empty, retErr error) {
	request.GetHead().GetBranch().GetRepo().EnsureProject()
	request.GetBranch().GetRepo().EnsureProject()
	for _, b := range request.Provenance {
		b.GetRepo().EnsureProject()
	}
	if err := a.env.TxnEnv.WithTransaction(ctx, func(txn txnenv.Transaction) error {
		return errors.EnsureStack(txn.CreateBranch(request))
	}, func(txnCtx *txncontext.TransactionContext) (string, error) {
		if request.Head == nil || request.NewCommitSet {
			return "", nil
		}
		// CreateBranch in a transaction by itself has special handling with regards
		// to its CommitSet ID.  In order to better support a 'deferred processing'
		// workflow with global IDs, it is useful for moving a branch head to be
		// done in the same CommitSet as the parent commit of the new branch head -
		// this is similar to how we handle triggers when finishing a commit.
		// Therefore we override the CommitSet ID being used by this operation, and
		// propagateBranches will update the existing CommitSet structure.  As an
		// escape hatch in case of an unexpected workload, this behavior can be
		// overridden by setting NewCommitSet=true in the request.
		// if request.Head != nil && !request.NewCommitSet {
		commitInfo, err := a.driver.resolveCommit(txnCtx.SqlTx, request.Head)
		if err != nil {
			return "", err
		}
		return commitInfo.Commit.ID, nil
	}); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

// InspectBranch implements the protobuf pfs.InspectBranch RPC
func (a *apiServer) InspectBranch(ctx context.Context, request *pfs.InspectBranchRequest) (response *pfs.BranchInfo, retErr error) {
	request.GetBranch().GetRepo().EnsureProject()
	branchInfo := &pfs.BranchInfo{}
	if err := a.env.TxnEnv.WithReadContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		var err error
		branchInfo, err = a.driver.inspectBranch(txnCtx, request.Branch)
		return err
	}); err != nil {
		return nil, err
	}
	return branchInfo, nil
}

func (a *apiServer) InspectBranchInTransaction(txnCtx *txncontext.TransactionContext, request *pfs.InspectBranchRequest) (*pfs.BranchInfo, error) {
	request.GetBranch().GetRepo().EnsureProject()
	return a.driver.inspectBranch(txnCtx, request.Branch)
}

// ListBranch implements the protobuf pfs.ListBranch RPC
func (a *apiServer) ListBranch(request *pfs.ListBranchRequest, srv pfs.API_ListBranchServer) (retErr error) {
	request.GetRepo().EnsureProject()
	if request.Repo == nil {
		return a.driver.listBranch(srv.Context(), request.Reverse, srv.Send)
	}
	return a.env.TxnEnv.WithReadContext(srv.Context(), func(txnCtx *txncontext.TransactionContext) error {
		return a.driver.listBranchInTransaction(txnCtx, request.Repo, request.Reverse, srv.Send)
	})
}

// DeleteBranchInTransaction is identical to DeleteBranch except that it can run
// inside an existing postgres transaction.  This is not an RPC.
func (a *apiServer) DeleteBranchInTransaction(txnCtx *txncontext.TransactionContext, request *pfs.DeleteBranchRequest) error {
	return a.driver.deleteBranch(txnCtx, request.Branch, request.Force)
}

// DeleteBranch implements the protobuf pfs.DeleteBranch RPC
func (a *apiServer) DeleteBranch(ctx context.Context, request *pfs.DeleteBranchRequest) (response *types.Empty, retErr error) {
	request.GetBranch().GetRepo().EnsureProject()
	if err := a.env.TxnEnv.WithTransaction(ctx, func(txn txnenv.Transaction) error {
		return errors.EnsureStack(txn.DeleteBranch(request))
	}, nil); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

// CreateProject implements the protobuf pfs.CreateProject RPC
func (a *apiServer) CreateProject(ctx context.Context, request *pfs.CreateProjectRequest) (*types.Empty, error) {
	if err := a.driver.createProject(ctx, request); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

// InspectProject implements the protobuf pfs.InspectProject RPC
func (a *apiServer) InspectProject(ctx context.Context, request *pfs.InspectProjectRequest) (*pfs.ProjectInfo, error) {
	return a.driver.inspectProject(ctx, request.Project)
}

// ListProject implements the protobuf pfs.ListProject RPC
func (a *apiServer) ListProject(request *pfs.ListProjectRequest, srv pfs.API_ListProjectServer) error {
	return a.driver.listProject(srv.Context(), srv.Send)
}

// DeleteProject implements the protobuf pfs.DeleteProject RPC
func (a *apiServer) DeleteProject(ctx context.Context, request *pfs.DeleteProjectRequest) (*types.Empty, error) {
	if err := a.env.TxnEnv.WithWriteContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		return a.driver.deleteProject(txnCtx, request.Project, request.Force)
	}); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
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
		return bytesRead, errors.EnsureStack(server.SendAndClose(&types.Empty{}))
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
				n, err = putFileURL(ctx, uw, p, t, src.Url)
			default:
				// need to write empty data to path
				n, err = putFileRaw(ctx, uw, p, t, &types.BytesValue{})
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

func putFileRaw(ctx context.Context, uw *fileset.UnorderedWriter, path, tag string, src *types.BytesValue) (int64, error) {
	if err := uw.Put(ctx, path, tag, true, bytes.NewReader(src.Value)); err != nil {
		return 0, err
	}
	return int64(len(src.Value)), nil
}

func putFileURL(ctx context.Context, uw *fileset.UnorderedWriter, dstPath, tag string, src *pfs.AddFile_URLSource) (n int64, retErr error) {
	url, err := url.Parse(src.URL)
	if err != nil {
		return 0, errors.EnsureStack(err)
	}
	switch url.Scheme {
	case "http":
		fallthrough
	case "https":
		resp, err := http.Get(src.URL)
		if err != nil {
			return 0, errors.EnsureStack(err)
		} else if resp.StatusCode >= 400 {
			return 0, errors.Errorf("error retrieving content from %q: %s", src.URL, resp.Status)
		}
		defer func() {
			if err := resp.Body.Close(); retErr == nil {
				retErr = err
			}
		}()
		return 0, uw.Put(ctx, dstPath, tag, true, resp.Body)
	default:
		url, err := obj.ParseURL(src.URL)
		if err != nil {
			return 0, errors.Wrapf(err, "error parsing url %v", src)
		}
		objClient, err := obj.NewClientFromURLAndSecret(url, false)
		if err != nil {
			return 0, err
		}
		if src.Recursive {
			path := strings.TrimPrefix(url.Object, "/")
			err := objClient.Walk(ctx, path, func(name string) error {
				return miscutil.WithPipe(func(w io.Writer) error {
					return errors.EnsureStack(objClient.Get(ctx, name, w))
				}, func(r io.Reader) error {
					return uw.Put(ctx, filepath.Join(dstPath, strings.TrimPrefix(name, path)), tag, true, r)
				})
			})
			return 0, errors.EnsureStack(err)
		}
		return 0, miscutil.WithPipe(func(w io.Writer) error {
			return errors.EnsureStack(objClient.Get(ctx, url.Object, w))
		}, func(r io.Reader) error {
			return uw.Put(ctx, dstPath, tag, true, r)
		})
	}
}

func deleteFile(ctx context.Context, uw *fileset.UnorderedWriter, request *pfs.DeleteFile) error {
	return uw.Delete(ctx, request.Path, request.Datum)
}

// GetFileTAR implements the protobuf pfs.GetFileTAR RPC
func (a *apiServer) GetFileTAR(request *pfs.GetFileRequest, server pfs.API_GetFileTARServer) (retErr error) {
	request.GetFile().GetCommit().GetBranch().Repo.EnsureProject()
	return metrics.ReportRequestWithThroughput(func() (int64, error) {
		ctx := server.Context()
		src, err := a.driver.getFile(ctx, request.File, request.PathRange)
		if err != nil {
			return 0, err
		}
		if request.URL != "" {
			return getFileURL(ctx, request.URL, src)
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
	request.GetFile().GetCommit().GetBranch().GetRepo().EnsureProject()
	return metrics.ReportRequestWithThroughput(func() (int64, error) {
		ctx := server.Context()
		src, err := a.driver.getFile(ctx, request.File, request.PathRange)
		if err != nil {
			return 0, err
		}
		if request.URL != "" {
			return getFileURL(ctx, request.URL, src)
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

// TODO: Parallelize and decide on appropriate config.
func getFileURL(ctx context.Context, URL string, src Source) (int64, error) {
	parsedURL, err := obj.ParseURL(URL)
	if err != nil {
		return 0, err
	}
	objClient, err := obj.NewClientFromURLAndSecret(parsedURL, false)
	if err != nil {
		return 0, err
	}
	var bytesWritten int64
	err = src.Iterate(ctx, func(fi *pfs.FileInfo, file fileset.File) (retErr error) {
		if fi.FileType != pfs.FileType_FILE {
			return nil
		}
		if err := miscutil.WithPipe(func(w io.Writer) error {
			return errors.EnsureStack(file.Content(ctx, w))
		}, func(r io.Reader) error {
			return errors.EnsureStack(objClient.Put(ctx, filepath.Join(parsedURL.Object, fi.File.Path), r))
		}); err != nil {
			return err
		}
		bytesWritten += int64(fi.SizeBytes)
		return nil
	})
	return bytesWritten, errors.EnsureStack(err)
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
	return a.driver.inspectFile(ctx, request.File)
}

// ListFile implements the protobuf pfs.ListFile RPC
func (a *apiServer) ListFile(request *pfs.ListFileRequest, server pfs.API_ListFileServer) (retErr error) {
	request.GetFile().GetCommit().GetBranch().GetRepo().EnsureProject()
	return a.driver.listFile(server.Context(), request.File, request.PaginationMarker, request.Number, request.Reverse, func(fi *pfs.FileInfo) error {
		return errors.EnsureStack(server.Send(fi))
	})
}

// WalkFile implements the protobuf pfs.WalkFile RPC
func (a *apiServer) WalkFile(request *pfs.WalkFileRequest, server pfs.API_WalkFileServer) (retErr error) {
	request.GetFile().GetCommit().GetBranch().GetRepo().EnsureProject()
	return a.driver.walkFile(server.Context(), request.File, func(fi *pfs.FileInfo) error {
		return errors.EnsureStack(server.Send(fi))
	})
}

// GlobFile implements the protobuf pfs.GlobFile RPC
func (a *apiServer) GlobFile(request *pfs.GlobFileRequest, respServer pfs.API_GlobFileServer) (retErr error) {
	request.GetCommit().GetBranch().GetRepo().EnsureProject()
	return a.driver.globFile(respServer.Context(), request.Commit, request.Pattern, request.PathRange, func(fi *pfs.FileInfo) error {
		return errors.EnsureStack(respServer.Send(fi))
	})
}

// DiffFile implements the protobuf pfs.DiffFile RPC
func (a *apiServer) DiffFile(request *pfs.DiffFileRequest, server pfs.API_DiffFileServer) (retErr error) {
	request.GetNewFile().GetCommit().GetBranch().GetRepo().EnsureProject()
	request.GetOldFile().GetCommit().GetBranch().GetRepo().EnsureProject()
	return a.driver.diffFile(server.Context(), request.OldFile, request.NewFile, func(oldFi, newFi *pfs.FileInfo) error {
		return errors.EnsureStack(server.Send(&pfs.DiffFileResponse{
			OldFile: oldFi,
			NewFile: newFi,
		}))
	})
}

// DeleteAll implements the protobuf pfs.DeleteAll RPC
func (a *apiServer) DeleteAll(ctx context.Context, request *types.Empty) (response *types.Empty, retErr error) {
	if err := a.driver.deleteAll(ctx); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

// Fsckimplements the protobuf pfs.Fsck RPC
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
	if request.GetZombieAll() {
		// list meta repos as a proxy for finding pipelines
		return a.driver.listRepo(ctx, false, pfs.MetaRepoType, func(info *pfs.RepoInfo) error {
			// TODO: actually derive output branch from job/pipeline, currently that coupling causes issues
			output := client.NewProjectCommit(info.Repo.Project.GetName(), info.Repo.Name, "master", "")
			for output != nil {
				info, err := a.driver.inspectCommit(ctx, output, pfs.CommitState_STARTED)
				if err != nil {
					return err
				}
				// we will be reading the whole file system, so unfinished commits would be very slow
				if info.Error == "" && info.Finished != nil && info.Origin.Kind != pfs.OriginKind_ALIAS {
					break
				}
				output = info.ParentCommit
			}
			if output == nil {
				return nil
			}
			return a.driver.detectZombie(ctx, output, fsckServer.Send)
		})
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
	req.GetCommit().GetBranch().GetRepo().EnsureProject()
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

func (a *apiServer) AddFileSet(ctx context.Context, req *pfs.AddFileSetRequest) (_ *types.Empty, retErr error) {
	req.GetCommit().GetBranch().GetRepo().EnsureProject()
	if err := a.env.TxnEnv.WithWriteContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		return a.AddFileSetInTransaction(txnCtx, req)
	}); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

func (a *apiServer) AddFileSetInTransaction(txnCtx *txncontext.TransactionContext, request *pfs.AddFileSetRequest) error {
	fsid, err := fileset.ParseID(request.FileSetId)
	if err != nil {
		return err
	}
	if err := a.driver.addFileSet(txnCtx, request.Commit, *fsid); err != nil {
		return err
	}
	return nil
}

// RenewFileSet implements the pfs.RenewFileSet RPC
func (a *apiServer) RenewFileSet(ctx context.Context, req *pfs.RenewFileSetRequest) (_ *types.Empty, retErr error) {
	fsid, err := fileset.ParseID(req.FileSetId)
	if err != nil {
		return nil, err
	}
	if err := a.driver.renewFileSet(ctx, *fsid, time.Duration(req.TtlSeconds)*time.Second); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
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
	chunks := a.driver.storage.ChunkStorage()
	count, err := chunks.Check(ctx, req.ChunkBegin, req.ChunkEnd, req.ReadChunkData)
	if err != nil {
		return nil, err
	}
	return &pfs.CheckStorageResponse{
		ChunkObjectCount: int64(count),
	}, nil
}

func (a *apiServer) PutCache(ctx context.Context, req *pfs.PutCacheRequest) (resp *types.Empty, retErr error) {
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
	return &types.Empty{}, nil
}

func (a *apiServer) GetCache(ctx context.Context, req *pfs.GetCacheRequest) (resp *pfs.GetCacheResponse, retErr error) {
	value, err := a.driver.getCache(ctx, req.Key)
	if err != nil {
		return nil, err
	}
	return &pfs.GetCacheResponse{Value: value}, nil
}

func (a *apiServer) ClearCache(ctx context.Context, req *pfs.ClearCacheRequest) (resp *types.Empty, retErr error) {
	if err := a.driver.clearCache(ctx, req.TagPrefix); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

// RunLoadTest implements the pfs.RunLoadTest RPC
func (a *apiServer) RunLoadTest(ctx context.Context, req *pfs.RunLoadTestRequest) (_ *pfs.RunLoadTestResponse, retErr error) {
	pachClient := a.env.GetPachClient(ctx)
	taskService := a.env.TaskService
	var project string
	repo := "load_test"
	if req.Branch != nil {
		project = req.Branch.Repo.Project.GetName()
		repo = req.Branch.Repo.Name
	}
	if err := pachClient.CreateProjectRepo(project, repo); err != nil && !pfsserver.IsRepoExistsErr(err) {
		return nil, err
	}
	branch := uuid.New()
	if req.Branch != nil {
		branch = req.Branch.Name
	}
	if err := pachClient.CreateProjectBranch(project, repo, branch, "", "", nil); err != nil {
		return nil, err
	}
	seed := time.Now().UTC().UnixNano()
	if req.Seed > 0 {
		seed = req.Seed
	}
	resp := &pfs.RunLoadTestResponse{
		Spec:   req.Spec,
		Branch: client.NewProjectBranch(req.Branch.GetRepo().GetProject().GetName(), repo, branch),
		Seed:   seed,
	}
	start := time.Now()
	var err error
	resp.StateId, err = a.runLoadTest(pachClient, taskService, resp.Branch, req.Spec, seed, req.StateId)
	if err != nil {
		resp.Error = err.Error()
	}
	resp.Duration = types.DurationProto(time.Since(start))
	return resp, nil
}

func (a *apiServer) runLoadTest(pachClient *client.APIClient, taskService task.Service, branch *pfs.Branch, specStr string, seed int64, stateID string) (string, error) {
	jsonBytes, err := yaml.YAMLToJSON([]byte(specStr))
	if err != nil {
		return "", errors.EnsureStack(err)
	}
	spec := &pfsload.CommitSpec{}
	if err := jsonpb.Unmarshal(bytes.NewReader(jsonBytes), spec); err != nil {
		return "", errors.EnsureStack(err)
	}
	return pfsload.Commit(pachClient, taskService, branch, spec, seed, stateID)
}

func (a *apiServer) RunLoadTestDefault(ctx context.Context, _ *types.Empty) (resp *pfs.RunLoadTestResponse, retErr error) {
	for _, spec := range defaultLoadSpecs {
		var err error
		resp, err = a.RunLoadTest(ctx, &pfs.RunLoadTestRequest{
			Spec: spec,
		})
		if err != nil {
			return nil, err
		}
		if resp.Error != "" {
			return resp, nil
		}
	}
	return resp, nil
}

var defaultLoadSpecs = []string{`
count: 3 
modifications:
  - count: 5
    putFile:
      count: 5
      source: "random"
validator: {}
fileSources:
  - name: "random"
    random:
      directory:
        depth: 
          min: 0
          max: 3
        run: 3
      sizes:
        - min: 1000
          max: 10000
          prob: 30 
        - min: 10000
          max: 100000
          prob: 30 
        - min: 1000000
          max: 10000000
          prob: 30 
        - min: 10000000
          max: 100000000
          prob: 10 
`, `
count: 3 
modifications:
  - count: 5
    putFile:
      count: 10000
      source: "random"
validator: {}
fileSources:
  - name: "random"
    random:
      sizes:
        - min: 100
          max: 1000
          prob: 100
`, `
count: 3 
modifications:
  - count: 5
    putFile:
      count: 1
      source: "random"
validator: {}
fileSources:
  - name: "random"
    random:
      sizes:
        - min: 10000000
          max: 100000000
          prob: 100 
`}

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
		if projectName := x.SetCommit.Branch.Repo.Project.GetName(); projectName == "" {
			x.SetCommit.Branch.Repo.Project = &pfs.Project{Name: pfs.DefaultProjectName}
		}
		return x.SetCommit, nil
	default:
		return nil, errors.Errorf("first message must be a commit")
	}
}

func (a *apiServer) Egress(ctx context.Context, req *pfs.EgressRequest) (*pfs.EgressResponse, error) {
	src, err := a.driver.getFile(ctx, req.Commit.NewFile("/"), nil)
	if err != nil {
		return nil, err
	}
	switch target := req.Target.(type) {
	case *pfs.EgressRequest_ObjectStorage:
		result, err := copyToObjectStorage(ctx, src, target.ObjectStorage.Url)
		if err != nil {
			return nil, errors.EnsureStack(err)
		}
		return &pfs.EgressResponse{Result: &pfs.EgressResponse_ObjectStorage{ObjectStorage: result}}, nil

	case *pfs.EgressRequest_SqlDatabase:
		result, err := copyToSQLDB(ctx, src, target.SqlDatabase.Url, target.SqlDatabase.FileFormat)
		if err != nil {
			return nil, errors.EnsureStack(err)
		}
		return &pfs.EgressResponse{Result: &pfs.EgressResponse_SqlDatabase{SqlDatabase: result}}, nil
	}
	return nil, errors.Errorf("egress failed")
}
