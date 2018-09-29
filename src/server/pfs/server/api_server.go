package server

import (
	"fmt"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/hashtree"
	"github.com/pachyderm/pachyderm/src/server/pkg/log"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	grpcErrorf = grpc.Errorf // needed to get passed govet
)

type apiServer struct {
	log.Logger
	driver *driver
}

func newAPIServer(address string, etcdAddresses []string, etcdPrefix string, treeCache *hashtree.Cache, storageRoot string, memoryRequest int64) (*apiServer, error) {
	d, err := newDriver(address, etcdAddresses, etcdPrefix, treeCache, storageRoot, memoryRequest)
	if err != nil {
		return nil, err
	}
	return &apiServer{
		Logger: log.NewLogger("pfs.API"),
		driver: d,
	}, nil
}

func (a *apiServer) CreateRepo(ctx context.Context, request *pfs.CreateRepoRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	if err := a.driver.createRepo(ctx, request.Repo, request.Description, request.Update); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

func (a *apiServer) InspectRepo(ctx context.Context, request *pfs.InspectRepoRequest) (response *pfs.RepoInfo, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	return a.driver.inspectRepo(ctx, request.Repo, true)
}

func (a *apiServer) ListRepo(ctx context.Context, request *pfs.ListRepoRequest) (response *pfs.ListRepoResponse, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	repoInfos, err := a.driver.listRepo(ctx, true)
	return repoInfos, err
}

func (a *apiServer) DeleteRepo(ctx context.Context, request *pfs.DeleteRepoRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	if request.All {
		if err := a.driver.deleteAll(ctx); err != nil {
			return nil, err
		}
	} else {
		if err := a.driver.deleteRepo(ctx, request.Repo, request.Force); err != nil {
			return nil, err
		}
	}

	return &types.Empty{}, nil
}

func (a *apiServer) StartCommit(ctx context.Context, request *pfs.StartCommitRequest) (response *pfs.Commit, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	commit, err := a.driver.startCommit(ctx, request.Parent, request.Branch, request.Provenance, request.Description)
	if err != nil {
		return nil, err
	}
	return commit, nil
}

func (a *apiServer) BuildCommit(ctx context.Context, request *pfs.BuildCommitRequest) (response *pfs.Commit, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	commit, err := a.driver.buildCommit(ctx, request.ID, request.Parent, request.Branch, request.Provenance, request.Tree)
	if err != nil {
		return nil, err
	}
	return commit, nil
}

func (a *apiServer) FinishCommit(ctx context.Context, request *pfs.FinishCommitRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	if request.Trees != nil {
		if err := a.driver.finishOutputCommit(ctx, request.Commit, request.Trees, request.SizeBytes); err != nil {
			return nil, err
		}
	} else if err := a.driver.finishCommit(ctx, request.Commit, request.Tree, request.Empty, request.Description); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

func (a *apiServer) InspectCommit(ctx context.Context, request *pfs.InspectCommitRequest) (response *pfs.CommitInfo, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	return a.driver.inspectCommit(ctx, request.Commit, request.BlockState)
}

func (a *apiServer) ListCommit(ctx context.Context, request *pfs.ListCommitRequest) (response *pfs.CommitInfos, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	commitInfos, err := a.driver.listCommit(ctx, request.Repo, request.To, request.From, request.Number)
	if err != nil {
		return nil, err
	}
	return &pfs.CommitInfos{
		CommitInfo: commitInfos,
	}, nil
}

func (a *apiServer) ListCommitStream(req *pfs.ListCommitRequest, respServer pfs.API_ListCommitStreamServer) (retErr error) {
	func() { a.Log(req, nil, nil, 0) }()
	sent := 0
	defer func(start time.Time) {
		a.Log(req, fmt.Sprintf("stream containing %d commits", sent), retErr, time.Since(start))
	}(time.Now())
	return a.driver.listCommitF(respServer.Context(), req.Repo, req.To, req.From, req.Number, func(ci *pfs.CommitInfo) error {
		sent++
		return respServer.Send(ci)
	})
}

func (a *apiServer) CreateBranch(ctx context.Context, request *pfs.CreateBranchRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	if err := a.driver.createBranch(ctx, request.Branch, request.Head, request.Provenance); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

func (a *apiServer) InspectBranch(ctx context.Context, request *pfs.InspectBranchRequest) (response *pfs.BranchInfo, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	return a.driver.inspectBranch(ctx, request.Branch)
}

func (a *apiServer) ListBranch(ctx context.Context, request *pfs.ListBranchRequest) (response *pfs.BranchInfos, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	branches, err := a.driver.listBranch(ctx, request.Repo)
	if err != nil {
		return nil, err
	}
	return &pfs.BranchInfos{BranchInfo: branches}, nil
}

func (a *apiServer) DeleteBranch(ctx context.Context, request *pfs.DeleteBranchRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	if err := a.driver.deleteBranch(ctx, request.Branch, request.Force); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

func (a *apiServer) DeleteCommit(ctx context.Context, request *pfs.DeleteCommitRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	if err := a.driver.deleteCommit(ctx, request.Commit); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

func (a *apiServer) FlushCommit(request *pfs.FlushCommitRequest, stream pfs.API_FlushCommitServer) (retErr error) {
	ctx := stream.Context()
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, nil, retErr, time.Since(start)) }(time.Now())

	return a.driver.flushCommit(ctx, request.Commits, request.ToRepos, stream.Send)
}

func (a *apiServer) SubscribeCommit(request *pfs.SubscribeCommitRequest, stream pfs.API_SubscribeCommitServer) (retErr error) {
	ctx := stream.Context()
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, nil, retErr, time.Since(start)) }(time.Now())

	return a.driver.subscribeCommit(ctx, request.Repo, request.Branch, request.From, request.State, stream.Send)
}

func (a *apiServer) PutFile(putFileServer pfs.API_PutFileServer) (retErr error) {
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
	return a.driver.putFiles(s)
}

func (a *apiServer) CopyFile(ctx context.Context, request *pfs.CopyFileRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	if err := a.driver.copyFile(ctx, request.Src, request.Dst, request.Overwrite); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

func (a *apiServer) GetFile(request *pfs.GetFileRequest, apiGetFileServer pfs.API_GetFileServer) (retErr error) {
	ctx := apiGetFileServer.Context()
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, nil, retErr, time.Since(start)) }(time.Now())

	file, err := a.driver.getFile(ctx, request.File, request.OffsetBytes, request.SizeBytes)
	if err != nil {
		return err
	}
	return grpcutil.WriteToStreamingBytesServer(file, apiGetFileServer)
}

func (a *apiServer) InspectFile(ctx context.Context, request *pfs.InspectFileRequest) (response *pfs.FileInfo, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	return a.driver.inspectFile(ctx, request.File)
}

func (a *apiServer) ListFile(ctx context.Context, request *pfs.ListFileRequest) (response *pfs.FileInfos, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) {
		if response != nil && len(response.FileInfo) > client.MaxListItemsLog {
			logrus.Infof("Response contains %d objects; logging the first %d", len(response.FileInfo), client.MaxListItemsLog)
			a.Log(request, &pfs.FileInfos{response.FileInfo[:client.MaxListItemsLog]}, retErr, time.Since(start))
		} else {
			a.Log(request, response, retErr, time.Since(start))
		}
	}(time.Now())

	var fileInfos []*pfs.FileInfo
	if err := a.driver.listFile(ctx, request.File, request.Full, func(fi *pfs.FileInfo) error {
		fileInfos = append(fileInfos, fi)
		return nil
	}); err != nil {
		return nil, err
	}
	return &pfs.FileInfos{
		FileInfo: fileInfos,
	}, nil
}

func (a *apiServer) ListFileStream(request *pfs.ListFileRequest, respServer pfs.API_ListFileStreamServer) (retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	var sent int
	defer func(start time.Time) {
		a.Log(request, fmt.Sprintf("response stream with %d objects", sent), retErr, time.Since(start))
	}(time.Now())
	return a.driver.listFile(respServer.Context(), request.File, request.Full, func(fi *pfs.FileInfo) error {
		sent++
		return respServer.Send(fi)
	})
}

func (a *apiServer) WalkFile(request *pfs.WalkFileRequest, server pfs.API_WalkFileServer) (retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	var sent int
	defer func(start time.Time) {
		a.Log(request, fmt.Sprintf("response stream with %d objects", sent), retErr, time.Since(start))
	}(time.Now())
	return a.driver.walkFile(server.Context(), request.File, func(fi *pfs.FileInfo) error {
		sent++
		return server.Send(fi)
	})
}

func (a *apiServer) GlobFile(ctx context.Context, request *pfs.GlobFileRequest) (response *pfs.FileInfos, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) {
		if response != nil && len(response.FileInfo) > client.MaxListItemsLog {
			logrus.Infof("Response contains %d objects; logging the first %d", len(response.FileInfo), client.MaxListItemsLog)
			a.Log(request, &pfs.FileInfos{response.FileInfo[:client.MaxListItemsLog]}, retErr, time.Since(start))
		} else {
			a.Log(request, response, retErr, time.Since(start))
		}
	}(time.Now())

	var fileInfos []*pfs.FileInfo
	if err := a.driver.globFile(ctx, request.Commit, request.Pattern, func(fi *pfs.FileInfo) error {
		fileInfos = append(fileInfos, fi)
		return nil
	}); err != nil {
		return nil, err
	}
	return &pfs.FileInfos{
		FileInfo: fileInfos,
	}, nil
}

func (a *apiServer) GlobFileStream(request *pfs.GlobFileRequest, respServer pfs.API_GlobFileStreamServer) (retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	var sent int
	defer func(start time.Time) {
		a.Log(request, fmt.Sprintf("response stream with %d objects", sent), retErr, time.Since(start))
	}(time.Now())
	return a.driver.globFile(respServer.Context(), request.Commit, request.Pattern, func(fi *pfs.FileInfo) error {
		sent++
		return respServer.Send(fi)
	})
}

func (a *apiServer) DiffFile(ctx context.Context, request *pfs.DiffFileRequest) (response *pfs.DiffFileResponse, retErr error) {
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
	newFileInfos, oldFileInfos, err := a.driver.diffFile(ctx, request.NewFile, request.OldFile, request.Shallow)
	if err != nil {
		return nil, err
	}
	return &pfs.DiffFileResponse{
		NewFiles: newFileInfos,
		OldFiles: oldFileInfos,
	}, nil
}

func (a *apiServer) DeleteFile(ctx context.Context, request *pfs.DeleteFileRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	err := a.driver.deleteFile(ctx, request.File)
	if err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

func (a *apiServer) DeleteAll(ctx context.Context, request *types.Empty) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	if err := a.driver.deleteAll(ctx); err != nil {
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
