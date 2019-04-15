package server

import (
	"fmt"
	"sync"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	"github.com/pachyderm/pachyderm/src/server/pkg/hashtree"
	"github.com/pachyderm/pachyderm/src/server/pkg/log"
	"github.com/pachyderm/pachyderm/src/server/pkg/serviceenv"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	grpcErrorf = grpc.Errorf // needed to get passed govet
)

// APIServer implements the public interface of the Pachyderm File System,
// including all RPCs defined in the protobuf spec.  Implementation details
// occur in the 'driver' code, and this layer serves to translate the protobuf
// request structures into normal function calls.
type APIServer struct {
	log.Logger
	driver *driver

	// env generates clients for pachyderm's downstream services
	env *serviceenv.ServiceEnv
	// pachClientOnce ensures that _pachClient is only initialized once
	pachClientOnce sync.Once
	// pachClient is a cached Pachd client that connects to Pachyderm's object
	// store API and auth API. Instead of accessing it directly, functions should
	// call a.env.GetPachClient()
	_pachClient *client.APIClient
}

func newAPIServer(env *serviceenv.ServiceEnv, etcdPrefix string, treeCache *hashtree.Cache, storageRoot string, memoryRequest int64) (*APIServer, error) {
	d, err := newDriver(env, etcdPrefix, treeCache, storageRoot, memoryRequest)
	if err != nil {
		return nil, err
	}
	s := &APIServer{
		Logger: log.NewLogger("pfs.API"),
		driver: d,
		env:    env,
	}
	go func() { s.env.GetPachClient(context.Background()) }() // Begin dialing connection on startup
	return s, nil
}

// CreateRepoInTransaction is identical to CreateRepo except that it can run
// inside an existing etcd STM transaction.  This is not an RPC.
func (a *APIServer) CreateRepoInTransaction(
	pachClient *client.APIClient,
	stm col.STM,
	request *pfs.CreateRepoRequest,
) error {
	return a.driver.createRepo(pachClient, stm, request.Repo, request.Description, request.Update)
}

// CreateRepo implements the protobuf pfs.CreateRepo RPC
func (a *APIServer) CreateRepo(ctx context.Context, request *pfs.CreateRepoRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	_, err := col.NewSTM(ctx, a.driver.etcdClient, func(stm col.STM) error {
		return a.CreateRepoInTransaction(a.env.GetPachClient(ctx), stm, request)
	})
	if err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

// InspectRepo implements the protobuf pfs.InspectRepo RPC
func (a *APIServer) InspectRepo(ctx context.Context, request *pfs.InspectRepoRequest) (response *pfs.RepoInfo, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	return a.driver.inspectRepo(a.env.GetPachClient(ctx), request.Repo, true)
}

// ListRepo implements the protobuf pfs.ListRepo RPC
func (a *APIServer) ListRepo(ctx context.Context, request *pfs.ListRepoRequest) (response *pfs.ListRepoResponse, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	repoInfos, err := a.driver.listRepo(a.env.GetPachClient(ctx), true)
	return repoInfos, err
}

// DeleteRepoInTransaction is identical to DeleteRepo except that it can run
// inside an existing etcd STM transaction.  This is not an RPC.
func (a *APIServer) DeleteRepoInTransaction(
	pachClient *client.APIClient,
	stm col.STM,
	request *pfs.DeleteRepoRequest,
) error {
	if request.All {
		if err := a.driver.deleteAll(pachClient, stm); err != nil {
			return err
		}
	} else {
		if err := a.driver.deleteRepo(pachClient, stm, request.Repo, request.Force); err != nil {
			return err
		}
	}

	return nil
}

// DeleteRepo implements the protobuf pfs.DeleteRepo RPC
func (a *APIServer) DeleteRepo(ctx context.Context, request *pfs.DeleteRepoRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	_, err := col.NewSTM(ctx, a.driver.etcdClient, func(stm col.STM) error {
		return a.DeleteRepoInTransaction(a.env.GetPachClient(ctx), stm, request)
	})
	if err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

// StartCommitInTransaction is identical to StartCommit except that it can run
// inside an existing etcd STM transaction.  This is not an RPC.  The target
// commit can be specified but is optional.  This is so that the transaction can
// report the commit ID back to the client before the transaction has finished
// and it can be used in future commands inside the same transaction.
func (a *APIServer) StartCommitInTransaction(
	pachClient *client.APIClient,
	stm col.STM,
	request *pfs.StartCommitRequest,
	commit *pfs.Commit,
) (*pfs.Commit, error) {
	id := ""
	if commit != nil {
		id = commit.ID
	}
	return a.driver.startCommit(pachClient, stm, id, request.Parent, request.Branch, request.Provenance, request.Description)
}

// StartCommit implements the protobuf pfs.StartCommit RPC
func (a *APIServer) StartCommit(ctx context.Context, request *pfs.StartCommitRequest) (response *pfs.Commit, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	var err error
	commit := &pfs.Commit{}
	_, err = col.NewSTM(ctx, a.driver.etcdClient, func(stm col.STM) error {
		commit, err = a.StartCommitInTransaction(a.env.GetPachClient(ctx), stm, request, nil)
		return err
	})
	if err != nil {
		return nil, err
	}
	return commit, nil
}

// BuildCommit implements the protobuf pfs.BuildCommit RPC
func (a *APIServer) BuildCommit(ctx context.Context, request *pfs.BuildCommitRequest) (response *pfs.Commit, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	commit, err := a.driver.buildCommit(a.env.GetPachClient(ctx), request.ID, request.Parent, request.Branch, request.Provenance, request.Tree)
	if err != nil {
		return nil, err
	}
	return commit, nil
}

// FinishCommitInTransaction is identical to FinishCommit except that it can run
// inside an existing etcd STM transaction.  This is not an RPC.
func (a *APIServer) FinishCommitInTransaction(
	pachClient *client.APIClient,
	stm col.STM,
	request *pfs.FinishCommitRequest,
) error {
	if request.Trees != nil {
		return a.driver.finishOutputCommit(pachClient, stm, request.Commit, request.Trees, request.Datums, request.SizeBytes)
	}
	return a.driver.finishCommit(pachClient, stm, request.Commit, request.Tree, request.Empty, request.Description)
}

// FinishCommit implements the protobuf pfs.FinishCommit RPC
func (a *APIServer) FinishCommit(ctx context.Context, request *pfs.FinishCommitRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	_, err := col.NewSTM(ctx, a.driver.etcdClient, func(stm col.STM) error {
		return a.FinishCommitInTransaction(a.env.GetPachClient(ctx), stm, request)
	})
	if err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

// InspectCommit implements the protobuf pfs.InspectCommit RPC
func (a *APIServer) InspectCommit(ctx context.Context, request *pfs.InspectCommitRequest) (response *pfs.CommitInfo, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	return a.driver.inspectCommit(a.env.GetPachClient(ctx), request.Commit, request.BlockState)
}

// ListCommit implements the protobuf pfs.ListCommit RPC
func (a *APIServer) ListCommit(ctx context.Context, request *pfs.ListCommitRequest) (response *pfs.CommitInfos, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	commitInfos, err := a.driver.listCommit(a.env.GetPachClient(ctx), request.Repo, request.To, request.From, request.Number)
	if err != nil {
		return nil, err
	}
	return &pfs.CommitInfos{
		CommitInfo: commitInfos,
	}, nil
}

// ListCommitStream implements the protobuf pfs.ListCommitStream RPC
func (a *APIServer) ListCommitStream(req *pfs.ListCommitRequest, respServer pfs.API_ListCommitStreamServer) (retErr error) {
	func() { a.Log(req, nil, nil, 0) }()
	sent := 0
	defer func(start time.Time) {
		a.Log(req, fmt.Sprintf("stream containing %d commits", sent), retErr, time.Since(start))
	}(time.Now())
	return a.driver.listCommitF(a.env.GetPachClient(respServer.Context()), req.Repo, req.To, req.From, req.Number, func(ci *pfs.CommitInfo) error {
		sent++
		return respServer.Send(ci)
	})
}

// CreateBranchInTransaction is identical to CreateBranch except that it can run
// inside an existing etcd STM transaction.  This is not an RPC.
func (a *APIServer) CreateBranchInTransaction(
	pachClient *client.APIClient,
	stm col.STM,
	request *pfs.CreateBranchRequest,
) error {
	return a.driver.createBranch(pachClient, stm, request.Branch, request.Head, request.Provenance)
}

// CreateBranch implements the protobuf pfs.CreateBranch RPC
func (a *APIServer) CreateBranch(ctx context.Context, request *pfs.CreateBranchRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	_, err := col.NewSTM(ctx, a.driver.etcdClient, func(stm col.STM) error {
		return a.CreateBranchInTransaction(a.env.GetPachClient(ctx), stm, request)
	})
	if err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

// InspectBranch implements the protobuf pfs.InspectBranch RPC
func (a *APIServer) InspectBranch(ctx context.Context, request *pfs.InspectBranchRequest) (response *pfs.BranchInfo, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	return a.driver.inspectBranch(a.env.GetPachClient(ctx), request.Branch)
}

// ListBranch implements the protobuf pfs.ListBranch RPC
func (a *APIServer) ListBranch(ctx context.Context, request *pfs.ListBranchRequest) (response *pfs.BranchInfos, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	branches, err := a.driver.listBranch(a.env.GetPachClient(ctx), request.Repo)
	if err != nil {
		return nil, err
	}
	return &pfs.BranchInfos{BranchInfo: branches}, nil
}

// DeleteBranchInTransaction is identical to DeleteBranch except that it can run
// inside an existing etcd STM transaction.  This is not an RPC.
func (a *APIServer) DeleteBranchInTransaction(
	pachClient *client.APIClient,
	stm col.STM,
	request *pfs.DeleteBranchRequest,
) error {
	return a.driver.deleteBranch(pachClient, stm, request.Branch, request.Force)
}

// DeleteBranch implements the protobuf pfs.DeleteBranch RPC
func (a *APIServer) DeleteBranch(ctx context.Context, request *pfs.DeleteBranchRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	_, err := col.NewSTM(ctx, a.driver.etcdClient, func(stm col.STM) error {
		return a.DeleteBranchInTransaction(a.env.GetPachClient(ctx), stm, request)
	})
	if err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

// DeleteCommitInTransaction is identical to DeleteCommit except that it can run
// inside an existing etcd STM transaction.  This is not an RPC.
func (a *APIServer) DeleteCommitInTransaction(
	pachClient *client.APIClient,
	stm col.STM,
	request *pfs.DeleteCommitRequest,
) error {
	return a.driver.deleteCommit(pachClient, stm, request.Commit)
}

// DeleteCommit implements the protobuf pfs.DeleteCommit RPC
func (a *APIServer) DeleteCommit(ctx context.Context, request *pfs.DeleteCommitRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	_, err := col.NewSTM(ctx, a.driver.etcdClient, func(stm col.STM) error {
		return a.DeleteCommitInTransaction(a.env.GetPachClient(ctx), stm, request)
	})
	if err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

// FlushCommit implements the protobuf pfs.FlushCommit RPC
func (a *APIServer) FlushCommit(request *pfs.FlushCommitRequest, stream pfs.API_FlushCommitServer) (retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, nil, retErr, time.Since(start)) }(time.Now())

	return a.driver.flushCommit(a.env.GetPachClient(stream.Context()), request.Commits, request.ToRepos, stream.Send)
}

// SubscribeCommit implements the protobuf pfs.SubscribeCommit RPC
func (a *APIServer) SubscribeCommit(request *pfs.SubscribeCommitRequest, stream pfs.API_SubscribeCommitServer) (retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, nil, retErr, time.Since(start)) }(time.Now())

	return a.driver.subscribeCommit(a.env.GetPachClient(stream.Context()), request.Repo, request.Branch, request.From, request.State, stream.Send)
}

// PutFile implements the protobuf pfs.PutFile RPC
func (a *APIServer) PutFile(putFileServer pfs.API_PutFileServer) (retErr error) {
	s := newPutFileServer(putFileServer)
	r, err := s.Peek()
	if err != nil {
		return err
	}
	request := *r
	request.Value = nil
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	defer drainFileServer(putFileServer)
	defer func() {
		if err := putFileServer.SendAndClose(&types.Empty{}); err != nil && retErr == nil {
			retErr = err
		}
	}()
	pachClient := a.env.GetPachClient(s.Context())
	return a.driver.putFiles(pachClient, s)
}

// CopyFileInTransaction is identical to CopyFile except that it can run inside
// an existing etcd STM transaction.  This is not an RPC.
func (a *APIServer) CopyFileInTransaction(
	pachClient *client.APIClient,
	stm col.STM,
	request *pfs.CopyFileRequest,
) error {
	return a.driver.copyFile(pachClient, stm, request.Src, request.Dst, request.Overwrite)
}

// CopyFile implements the protobuf pfs.CopyFile RPC
func (a *APIServer) CopyFile(ctx context.Context, request *pfs.CopyFileRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	_, err := col.NewSTM(ctx, a.driver.etcdClient, func(stm col.STM) error {
		return a.CopyFileInTransaction(a.env.GetPachClient(ctx), stm, request)
	})
	if err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

// GetFile implements the protobuf pfs.GetFile RPC
func (a *APIServer) GetFile(request *pfs.GetFileRequest, apiGetFileServer pfs.API_GetFileServer) (retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, nil, retErr, time.Since(start)) }(time.Now())

	file, err := a.driver.getFile(a.env.GetPachClient(apiGetFileServer.Context()), request.File, request.OffsetBytes, request.SizeBytes)
	if err != nil {
		return err
	}
	return grpcutil.WriteToStreamingBytesServer(file, apiGetFileServer)
}

// InspectFile implements the protobuf pfs.InspectFile RPC
func (a *APIServer) InspectFile(ctx context.Context, request *pfs.InspectFileRequest) (response *pfs.FileInfo, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	return a.driver.inspectFile(a.env.GetPachClient(ctx), request.File)
}

// ListFile implements the protobuf pfs.ListFile RPC
func (a *APIServer) ListFile(ctx context.Context, request *pfs.ListFileRequest) (response *pfs.FileInfos, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) {
		if response != nil && len(response.FileInfo) > client.MaxListItemsLog {
			logrus.Infof("Response contains %d objects; logging the first %d", len(response.FileInfo), client.MaxListItemsLog)
			a.Log(request, &pfs.FileInfos{FileInfo: response.FileInfo[:client.MaxListItemsLog]}, retErr, time.Since(start))
		} else {
			a.Log(request, response, retErr, time.Since(start))
		}
	}(time.Now())

	var fileInfos []*pfs.FileInfo
	if err := a.driver.listFile(a.env.GetPachClient(ctx), request.File, request.Full, request.History, func(fi *pfs.FileInfo) error {
		fileInfos = append(fileInfos, fi)
		return nil
	}); err != nil {
		return nil, err
	}
	return &pfs.FileInfos{
		FileInfo: fileInfos,
	}, nil
}

// ListFileStream implements the protobuf pfs.ListFileStream RPC
func (a *APIServer) ListFileStream(request *pfs.ListFileRequest, respServer pfs.API_ListFileStreamServer) (retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	var sent int
	defer func(start time.Time) {
		a.Log(request, fmt.Sprintf("response stream with %d objects", sent), retErr, time.Since(start))
	}(time.Now())
	return a.driver.listFile(a.env.GetPachClient(respServer.Context()), request.File, request.Full, request.History, func(fi *pfs.FileInfo) error {
		sent++
		return respServer.Send(fi)
	})
}

// WalkFile implements the protobuf pfs.WalkFile RPC
func (a *APIServer) WalkFile(request *pfs.WalkFileRequest, server pfs.API_WalkFileServer) (retErr error) {
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
func (a *APIServer) GlobFile(ctx context.Context, request *pfs.GlobFileRequest) (response *pfs.FileInfos, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) {
		if response != nil && len(response.FileInfo) > client.MaxListItemsLog {
			logrus.Infof("Response contains %d objects; logging the first %d", len(response.FileInfo), client.MaxListItemsLog)
			a.Log(request, &pfs.FileInfos{FileInfo: response.FileInfo[:client.MaxListItemsLog]}, retErr, time.Since(start))
		} else {
			a.Log(request, response, retErr, time.Since(start))
		}
	}(time.Now())

	var fileInfos []*pfs.FileInfo
	if err := a.driver.globFile(a.env.GetPachClient(ctx), request.Commit, request.Pattern, func(fi *pfs.FileInfo) error {
		fileInfos = append(fileInfos, fi)
		return nil
	}); err != nil {
		return nil, err
	}
	return &pfs.FileInfos{
		FileInfo: fileInfos,
	}, nil
}

// GlobFileStream implements the protobuf pfs.GlobFileStream RPC
func (a *APIServer) GlobFileStream(request *pfs.GlobFileRequest, respServer pfs.API_GlobFileStreamServer) (retErr error) {
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
func (a *APIServer) DiffFile(ctx context.Context, request *pfs.DiffFileRequest) (response *pfs.DiffFileResponse, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) {
		if response != nil && (len(response.NewFiles) > client.MaxListItemsLog || len(response.OldFiles) > client.MaxListItemsLog) {
			logrus.Infof("Response contains too many objects; truncating.")
			a.Log(request, &pfs.DiffFileResponse{
				NewFiles: truncateFiles(response.NewFiles),
				OldFiles: truncateFiles(response.OldFiles),
			}, retErr, time.Since(start))
		} else {
			a.Log(request, response, retErr, time.Since(start))
		}
	}(time.Now())
	newFileInfos, oldFileInfos, err := a.driver.diffFile(a.env.GetPachClient(ctx), request.NewFile, request.OldFile, request.Shallow)
	if err != nil {
		return nil, err
	}
	return &pfs.DiffFileResponse{
		NewFiles: newFileInfos,
		OldFiles: oldFileInfos,
	}, nil
}

// DeleteFileInTransaction is identical to DeleteFile except that it can run
// inside an existing etcd STM transaction.  This is not an RPC.
func (a *APIServer) DeleteFileInTransaction(
	pachClient *client.APIClient,
	stm col.STM,
	request *pfs.DeleteFileRequest,
) error {
	return a.driver.deleteFile(pachClient, stm, request.File)
}

// DeleteFile implements the protobuf pfs.DeleteFile RPC
func (a *APIServer) DeleteFile(ctx context.Context, request *pfs.DeleteFileRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	_, err := col.NewSTM(ctx, a.driver.etcdClient, func(stm col.STM) error {
		return a.DeleteFileInTransaction(a.env.GetPachClient(ctx), stm, request)
	})
	if err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

// DeleteAllInTransaction is identical to DeleteAll except that it can run
// inside an existing etcd STM transaction.  This is not an RPC.
func (a *APIServer) DeleteAllInTransaction(
	pachClient *client.APIClient,
	stm col.STM,
	request *pfs.DeleteAllRequest,
) error {
	return a.driver.deleteAll(pachClient, stm)
}

// DeleteAll implements the protobuf pfs.DeleteAll RPC
func (a *APIServer) DeleteAll(ctx context.Context, request *pfs.DeleteAllRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	_, err := col.NewSTM(ctx, a.driver.etcdClient, func(stm col.STM) error {
		return a.DeleteAllInTransaction(a.env.GetPachClient(ctx), stm, request)
	})
	if err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

func drainFileServer(putFileServer interface {
	Recv() (*pfs.PutFileRequest, error)
}) {
	for {
		if _, err := putFileServer.Recv(); err != nil {
			break
		}
	}
}

func truncateFiles(fileInfos []*pfs.FileInfo) []*pfs.FileInfo {
	if len(fileInfos) > client.MaxListItemsLog {
		return fileInfos[:client.MaxListItemsLog]
	}
	return fileInfos
}
