package server

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/log"
	"github.com/pachyderm/pachyderm/src/server/pkg/serviceenv"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/metrics"
	txnenv "github.com/pachyderm/pachyderm/src/server/pkg/transactionenv"

	"golang.org/x/net/context"
)

var _ APIServer = &apiServer{}

// apiServer implements the public interface of the Pachyderm File System,
// including all RPCs defined in the protobuf spec.  Implementation details
// occur in the 'driver' code, and this layer serves to translate the protobuf
// request structures into normal function calls.
type apiServer struct {
	log.Logger
	driver *driver
	txnEnv *txnenv.TransactionEnv

	// env generates clients for pachyderm's downstream services
	env *serviceenv.ServiceEnv
}

func newAPIServer(env *serviceenv.ServiceEnv, txnEnv *txnenv.TransactionEnv, etcdPrefix string, db *sqlx.DB) (*apiServer, error) {
	d, err := newDriver(env, txnEnv, etcdPrefix, db)
	if err != nil {
		return nil, err
	}
	s := &apiServer{
		Logger: log.NewLogger("pfs.API"),
		driver: d,
		env:    env,
		txnEnv: txnEnv,
	}
	go func() { s.env.GetPachClient(context.Background()) }() // Begin dialing connection on startup
	return s, nil
}

// CreateRepoInTransaction is identical to CreateRepo except that it can run
// inside an existing etcd STM transaction.  This is not an RPC.
func (a *apiServer) CreateRepoInTransaction(txnCtx *txnenv.TransactionContext, request *pfs.CreateRepoRequest) error {
	if repo := request.GetRepo(); repo != nil && repo.Name == tmpRepo {
		return errors.Errorf("%s is a reserved name", tmpRepo)
	}
	return a.driver.createRepo(txnCtx, request.Repo, request.Description, request.Update)
}

// CreateRepo implements the protobuf pfs.CreateRepo RPC
func (a *apiServer) CreateRepo(ctx context.Context, request *pfs.CreateRepoRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	if err := a.txnEnv.WithTransaction(ctx, func(txn txnenv.Transaction) error {
		return txn.CreateRepo(request)
	}); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

// InspectRepoInTransaction is identical to InspectRepo except that it can run
// inside an existing etcd STM transaction.  This is not an RPC.
func (a *apiServer) InspectRepoInTransaction(txnCtx *txnenv.TransactionContext, originalRequest *pfs.InspectRepoRequest) (*pfs.RepoInfo, error) {
	request := proto.Clone(originalRequest).(*pfs.InspectRepoRequest)
	return a.driver.inspectRepo(txnCtx, request.Repo, true)
}

// InspectRepo implements the protobuf pfs.InspectRepo RPC
func (a *apiServer) InspectRepo(ctx context.Context, request *pfs.InspectRepoRequest) (response *pfs.RepoInfo, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	var info *pfs.RepoInfo
	err := a.txnEnv.WithReadContext(ctx, func(txnCtx *txnenv.TransactionContext) error {
		var err error
		info, err = a.InspectRepoInTransaction(txnCtx, request)
		return err
	})
	if err != nil {
		return nil, err
	}
	return info, nil
}

// ListRepo implements the protobuf pfs.ListRepo RPC
func (a *apiServer) ListRepo(ctx context.Context, request *pfs.ListRepoRequest) (response *pfs.ListRepoResponse, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	repoInfos, err := a.driver.listRepo(a.env.GetPachClient(ctx), true)
	return repoInfos, err
}

// DeleteRepoInTransaction is identical to DeleteRepo except that it can run
// inside an existing etcd STM transaction.  This is not an RPC.
func (a *apiServer) DeleteRepoInTransaction(txnCtx *txnenv.TransactionContext, request *pfs.DeleteRepoRequest) error {
	if request.All {
		return a.driver.deleteAll(txnCtx)
	}
	return a.driver.deleteRepo(txnCtx, request.Repo, request.Force)
}

// DeleteRepo implements the protobuf pfs.DeleteRepo RPC
func (a *apiServer) DeleteRepo(ctx context.Context, request *pfs.DeleteRepoRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	if request.SplitTransaction {
		if err := func() error {
			if request.All {
				return a.driver.deleteAllSplitTransaction(a.env.GetPachClient(ctx))
			}
			return a.driver.deleteRepoSplitTransaction(ctx, request.Repo, request.Force)
		}(); err != nil {
			return nil, err
		}
		return &types.Empty{}, nil
	}
	if err := a.txnEnv.WithTransaction(ctx, func(txn txnenv.Transaction) error {
		return txn.DeleteRepo(request)
	}); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

// StartCommitInTransaction is identical to StartCommit except that it can run
// inside an existing etcd STM transaction.  This is not an RPC.  The target
// commit can be specified but is optional.  This is so that the transaction can
// report the commit ID back to the client before the transaction has finished
// and it can be used in future commands inside the same transaction.
func (a *apiServer) StartCommitInTransaction(txnCtx *txnenv.TransactionContext, request *pfs.StartCommitRequest, commit *pfs.Commit) (*pfs.Commit, error) {
	id := ""
	if commit != nil {
		id = commit.ID
	}
	return a.driver.startCommit(txnCtx, id, request.Parent, request.Branch, request.Provenance, request.Description)
}

// StartCommit implements the protobuf pfs.StartCommit RPC
func (a *apiServer) StartCommit(ctx context.Context, request *pfs.StartCommitRequest) (response *pfs.Commit, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	var err error
	commit := &pfs.Commit{}
	if err = a.txnEnv.WithTransaction(ctx, func(txn txnenv.Transaction) error {
		commit, err = txn.StartCommit(request, nil)
		return err
	}); err != nil {
		return nil, err
	}
	return commit, nil
}

// FinishCommitInTransaction is identical to FinishCommit except that it can run
// inside an existing etcd STM transaction.  This is not an RPC.
func (a *apiServer) FinishCommitInTransaction(txnCtx *txnenv.TransactionContext, request *pfs.FinishCommitRequest) error {
	return metrics.ReportRequest(func() error {
		if request.Empty {
			request.Description += pfs.EmptyStr
		}
		return a.driver.finishCommit(txnCtx, request.Commit, request.Description)
	})
}

// FinishCommit implements the protobuf pfs.FinishCommit RPC
func (a *apiServer) FinishCommit(ctx context.Context, request *pfs.FinishCommitRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	if err := a.txnEnv.WithTransaction(ctx, func(txn txnenv.Transaction) error {
		return txn.FinishCommit(request)
	}); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

// InspectCommit implements the protobuf pfs.InspectCommit RPC
func (a *apiServer) InspectCommit(ctx context.Context, request *pfs.InspectCommitRequest) (response *pfs.CommitInfo, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	return a.driver.inspectCommit(a.env.GetPachClient(ctx), request.Commit, request.BlockState)
}

// ListCommit implements the protobuf pfs.ListCommit RPC
func (a *apiServer) ListCommit(request *pfs.ListCommitRequest, respServer pfs.API_ListCommitServer) (retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	sent := 0
	defer func(start time.Time) {
		a.Log(request, fmt.Sprintf("stream containing %d commits", sent), retErr, time.Since(start))
	}(time.Now())
	return a.driver.listCommit(a.env.GetPachClient(respServer.Context()), request.Repo, request.To, request.From, request.Number, request.Reverse, func(ci *pfs.CommitInfo) error {
		sent++
		return respServer.Send(ci)
	})
}

// DeleteCommitInTransaction is identical to DeleteCommit except that it can run
// inside an existing etcd STM transaction.  This is not an RPC.
func (a *apiServer) DeleteCommitInTransaction(txnCtx *txnenv.TransactionContext, request *pfs.DeleteCommitRequest) error {
	return a.driver.deleteCommit(txnCtx, request.Commit)
}

// DeleteCommit implements the protobuf pfs.DeleteCommit RPC
func (a *apiServer) DeleteCommit(ctx context.Context, request *pfs.DeleteCommitRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	if err := a.txnEnv.WithTransaction(ctx, func(txn txnenv.Transaction) error {
		return txn.DeleteCommit(request)
	}); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

// FlushCommit implements the protobuf pfs.FlushCommit RPC
func (a *apiServer) FlushCommit(request *pfs.FlushCommitRequest, stream pfs.API_FlushCommitServer) (retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	return a.driver.flushCommit(a.env.GetPachClient(stream.Context()), request.Commits, request.ToRepos, stream.Send)
}

// SubscribeCommit implements the protobuf pfs.SubscribeCommit RPC
func (a *apiServer) SubscribeCommit(request *pfs.SubscribeCommitRequest, stream pfs.API_SubscribeCommitServer) (retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	return a.driver.subscribeCommit(a.env.GetPachClient(stream.Context()), request.Repo, request.Branch, request.Prov, request.From, request.State, stream.Send)
}

// ClearCommit deletes all data in the commit.
func (a *apiServer) ClearCommit(ctx context.Context, request *pfs.ClearCommitRequest) (_ *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	return nil, a.driver.clearCommit(a.env.GetPachClient(ctx), request.Commit)
}

// CreateBranchInTransaction is identical to CreateBranch except that it can run
// inside an existing etcd STM transaction.  This is not an RPC.
func (a *apiServer) CreateBranchInTransaction(txnCtx *txnenv.TransactionContext, request *pfs.CreateBranchRequest) error {
	return a.driver.createBranch(txnCtx, request.Branch, request.Head, request.Provenance, request.Trigger)
}

// CreateBranch implements the protobuf pfs.CreateBranch RPC
func (a *apiServer) CreateBranch(ctx context.Context, request *pfs.CreateBranchRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	if err := a.txnEnv.WithTransaction(ctx, func(txn txnenv.Transaction) error {
		return txn.CreateBranch(request)
	}); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

// InspectBranch implements the protobuf pfs.InspectBranch RPC
func (a *apiServer) InspectBranch(ctx context.Context, request *pfs.InspectBranchRequest) (response *pfs.BranchInfo, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	branchInfo := &pfs.BranchInfo{}
	if err := a.txnEnv.WithReadContext(ctx, func(txnCtx *txnenv.TransactionContext) error {
		var err error
		branchInfo, err = a.driver.inspectBranch(txnCtx, request.Branch)
		return err
	}); err != nil {
		return nil, err
	}
	return branchInfo, nil
}

// ListBranch implements the protobuf pfs.ListBranch RPC
func (a *apiServer) ListBranch(ctx context.Context, request *pfs.ListBranchRequest) (response *pfs.BranchInfos, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	branches, err := a.driver.listBranch(a.env.GetPachClient(ctx), request.Repo, request.Reverse)
	if err != nil {
		return nil, err
	}
	return &pfs.BranchInfos{BranchInfo: branches}, nil
}

// DeleteBranchInTransaction is identical to DeleteBranch except that it can run
// inside an existing etcd STM transaction.  This is not an RPC.
func (a *apiServer) DeleteBranchInTransaction(txnCtx *txnenv.TransactionContext, request *pfs.DeleteBranchRequest) error {
	return a.driver.deleteBranch(txnCtx, request.Branch, request.Force)
}

// DeleteBranch implements the protobuf pfs.DeleteBranch RPC
func (a *apiServer) DeleteBranch(ctx context.Context, request *pfs.DeleteBranchRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	if err := a.txnEnv.WithTransaction(ctx, func(txn txnenv.Transaction) error {
		return txn.DeleteBranch(request)
	}); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

func (a *apiServer) FileOperation(server pfs.API_FileOperationServer) (retErr error) {
	request, err := server.Recv()
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	return metrics.ReportRequestWithThroughput(func() (int64, error) {
		if err != nil {
			return 0, err
		}
		var bytesRead int64
		if err := a.driver.fileOperation(a.env.GetPachClient(server.Context()), request.Commit, func(uw *fileset.UnorderedWriter) error {
			for {
				request, err := server.Recv()
				if err != nil {
					if errors.Is(err, io.EOF) {
						return nil
					}
					return err
				}
				// TODO Validation.
				switch op := request.Operation.(type) {
				case *pfs.FileOperationRequest_AppendFile:
					n, err := appendFile(uw, server, op.AppendFile)
					bytesRead += n
					if err != nil {
						return err
					}
				case *pfs.FileOperationRequest_DeleteFile:
					if err := deleteFile(uw, op.DeleteFile); err != nil {
						return err
					}
				}
			}
		}); err != nil {
			return bytesRead, err
		}
		return bytesRead, server.SendAndClose(&types.Empty{})
	})
}

type fileOpSource interface {
	Recv() (*pfs.FileOperationRequest, error)
}

func appendFile(uw *fileset.UnorderedWriter, server fileOpSource, req *pfs.AppendFileRequest) (int64, error) {
	afr := &appendFileReader{
		server: server,
		r:      bytes.NewReader(req.Data),
	}
	err := uw.Append(afr, req.Overwrite, req.Tag)
	return afr.bytesRead, err
}

type appendFileReader struct {
	server    fileOpSource
	r         *bytes.Reader
	bytesRead int64
}

func (afr *appendFileReader) Read(data []byte) (int, error) {
	for afr.r.Len() == 0 {
		request, err := afr.server.Recv()
		if err != nil {
			return 0, err
		}
		op := request.Operation.(*pfs.FileOperationRequest_AppendFile)
		appendFileReq := op.AppendFile
		afr.r = bytes.NewReader(appendFileReq.Data)
	}
	n, err := afr.r.Read(data)
	afr.bytesRead += int64(n)
	return n, err
}

func deleteFile(uw *fileset.UnorderedWriter, request *pfs.DeleteFileRequest) error {
	for _, file := range request.Files {
		uw.Delete(file, request.Tag)
	}
	return nil
}

// CopyFile implements the protobuf pfs.CopyFile RPC
func (a *apiServer) CopyFile(ctx context.Context, request *pfs.CopyFileRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	if err := a.driver.copyFile(a.env.GetPachClient(ctx), request.Src, request.Dst, request.Overwrite); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

// GetFile implements the protobuf pfs.GetFile RPC
func (a *apiServer) GetFile(request *pfs.GetFileRequest, server pfs.API_GetFileServer) (retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	return metrics.ReportRequestWithThroughput(func() (int64, error) {
		commit := request.File.Commit
		glob := request.File.Path
		gfw := newGetFileWriter(grpcutil.NewStreamingBytesWriter(server))
		err := a.driver.getFile(a.env.GetPachClient(server.Context()), commit, glob, gfw)
		return gfw.bytesWritten, err
	})
}

type getFileWriter struct {
	w            io.Writer
	bytesWritten int64
}

func newGetFileWriter(w io.Writer) *getFileWriter {
	return &getFileWriter{w: w}
}

func (gfw *getFileWriter) Write(data []byte) (int, error) {
	n, err := gfw.w.Write(data)
	gfw.bytesWritten += int64(n)
	return n, err
}

// InspectFile implements the protobuf pfs.InspectFile RPC
func (a *apiServer) InspectFile(ctx context.Context, request *pfs.InspectFileRequest) (response *pfs.FileInfo, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	return a.driver.inspectFile(a.env.GetPachClient(ctx), request.File)
}

// ListFile implements the protobuf pfs.ListFile RPC
func (a *apiServer) ListFile(request *pfs.ListFileRequest, server pfs.API_ListFileServer) (retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	var sent int
	defer func(start time.Time) {
		a.Log(request, fmt.Sprintf("response stream with %d objects", sent), retErr, time.Since(start))
	}(time.Now())
	return a.driver.listFile(a.env.GetPachClient(server.Context()), request.File, request.Full, func(fi *pfs.FileInfo) error {
		sent++
		return server.Send(fi)
	})
}

// WalkFile implements the protobuf pfs.WalkFile RPC
func (a *apiServer) WalkFile(request *pfs.WalkFileRequest, server pfs.API_WalkFileServer) (retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	var sent int
	defer func(start time.Time) {
		a.Log(request, fmt.Sprintf("response stream with %d objects", sent), retErr, time.Since(start))
	}(time.Now())
	return a.driver.walkFile(a.env.GetPachClient(server.Context()), request.File, func(fi *pfs.FileInfo) error {
		sent++
		return server.Send(fi)
	})
}

// GlobFile implements the protobuf pfs.GlobFile RPC
func (a *apiServer) GlobFile(request *pfs.GlobFileRequest, respServer pfs.API_GlobFileServer) (retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	var sent int
	defer func(start time.Time) {
		a.Log(request, fmt.Sprintf("response stream with %d objects", sent), retErr, time.Since(start))
	}(time.Now())
	return a.driver.globFile(a.env.GetPachClient(respServer.Context()), request.Commit, request.Pattern, func(fi *pfs.FileInfo) error {
		sent++
		return respServer.Send(fi)
	})
}

// DiffFile implements the protobuf pfs.DiffFile RPC
func (a *apiServer) DiffFile(request *pfs.DiffFileRequest, server pfs.API_DiffFileServer) (retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	var sent int
	defer func(start time.Time) {
		a.Log(request, fmt.Sprintf("response stream with %d objects", sent), retErr, time.Since(start))
	}(time.Now())
	return a.driver.diffFile(a.env.GetPachClient(server.Context()), request.OldFile, request.NewFile, func(oldFi, newFi *pfs.FileInfo) error {
		sent++
		return server.Send(&pfs.DiffFileResponse{
			OldFile: oldFi,
			NewFile: newFi,
		})
	})
}

// DeleteAll implements the protobuf pfs.DeleteAll RPC
func (a *apiServer) DeleteAll(ctx context.Context, request *types.Empty) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	err := a.txnEnv.WithWriteContext(ctx, func(txnCtx *txnenv.TransactionContext) error {
		return a.driver.deleteAll(txnCtx)
	})
	if err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

// Fsckimplements the protobuf pfs.Fsck RPC
func (a *apiServer) Fsck(request *pfs.FsckRequest, fsckServer pfs.API_FsckServer) (retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	sent := 0
	defer func(start time.Time) {
		a.Log(request, fmt.Sprintf("stream containing %d messages", sent), retErr, time.Since(start))
	}(time.Now())
	if err := a.driver.fsck(a.env.GetPachClient(fsckServer.Context()), request.Fix, func(resp *pfs.FsckResponse) error {
		sent++
		return fsckServer.Send(resp)
	}); err != nil {
		return err
	}
	return nil
}

// CreateFileset implements the pfs.CreateFileset RPC
func (a *apiServer) CreateFileset(server pfs.API_CreateFilesetServer) error {
	fsID, err := a.driver.createFileset(server)
	if err != nil {
		return err
	}
	return server.SendAndClose(&pfs.CreateFilesetResponse{
		FilesetId: fsID,
	})
}

// RenewFileset implements the pfs.RenewFileset RPC
func (a *apiServer) RenewFileset(ctx context.Context, req *pfs.RenewFilesetRequest) (*types.Empty, error) {
	if err := a.driver.renewFileset(ctx, req.FilesetId, time.Duration(req.TtlSeconds)*time.Second); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}
