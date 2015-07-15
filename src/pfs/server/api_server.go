package server

import (
	"bytes"
	"io"

	"golang.org/x/net/context"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/drive"
	"github.com/pachyderm/pachyderm/src/pfs/route"
	"github.com/pachyderm/pachyderm/src/pfs/shard"
)

type apiServer struct {
	sharder shard.Sharder
	router  route.Router
	driver  drive.Driver
}

func newAPIServer(
	sharder shard.Sharder,
	router route.Router,
	driver drive.Driver,
) *apiServer {
	return &apiServer{
		sharder,
		router,
		driver,
	}
}

func (a *apiServer) InitRepository(ctx context.Context, initRepositoryRequest *pfs.InitRepositoryRequest) (*pfs.InitRepositoryResponse, error) {
	shards, err := a.getMasterShards()
	if err != nil {
		return nil, err
	}
	for shard := range shards {
		if err := a.driver.InitRepository(initRepositoryRequest.Repository, shard); err != nil {
			return nil, err
		}
	}
	return &pfs.InitRepositoryResponse{}, nil
}

func (a *apiServer) GetFile(getFileRequest *pfs.GetFileRequest, apiGetFileServer pfs.Api_GetFileServer) (retErr error) {
	shard, err := a.sharder.GetShard(getFileRequest.Path)
	if err != nil {
		return err
	}
	ok, err := a.router.IsLocalMasterShard(shard)
	if err != nil {
		return err
	}
	if !ok {
		apiClient, err := a.router.GetAPIClient(shard)
		if err != nil {
			return err
		}
		apiGetFileClient, err := apiClient.GetFile(context.Background(), getFileRequest)
		if err != nil {
			return err
		}
		for bytesValue, err := apiGetFileClient.Recv(); err != io.EOF; bytesValue, err = apiGetFileClient.Recv() {
			if err != nil {
				return err
			}
			if sendErr := apiGetFileServer.Send(bytesValue); sendErr != nil {
				return sendErr
			}
		}
		return nil
	}
	readCloser, err := a.driver.GetFile(getFileRequest.Path, shard)
	if err != nil {
		return err
	}
	defer func() {
		if err := readCloser.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	return writeToStreamingBytesServer(readCloser, apiGetFileServer)
}

func (a *apiServer) PutFile(ctx context.Context, putFileRequest *pfs.PutFileRequest) (*pfs.PutFileResponse, error) {
	shard, err := a.sharder.GetShard(putFileRequest.Path)
	if err != nil {
		return nil, err
	}
	ok, err := a.router.IsLocalMasterShard(shard)
	if err != nil {
		return nil, err
	}
	if !ok {
		apiClient, err := a.router.GetAPIClient(shard)
		if err != nil {
			return nil, err
		}
		return apiClient.PutFile(ctx, putFileRequest)
	}
	if err := a.driver.PutFile(putFileRequest.Path, shard, bytes.NewReader(putFileRequest.Value)); err != nil {
		return nil, err
	}
	return &pfs.PutFileResponse{}, nil
}

func (a *apiServer) ListFiles(ctx context.Context, listFilesRequest *pfs.ListFilesRequest) (*pfs.ListFilesResponse, error) {
	return &pfs.ListFilesResponse{}, nil
}

func (a *apiServer) GetParent(ctx context.Context, getParentRequest *pfs.GetParentRequest) (*pfs.GetParentResponse, error) {
	return &pfs.GetParentResponse{}, nil
}

func (a *apiServer) GetChildren(ctx context.Context, getChildrenRequest *pfs.GetChildrenRequest) (*pfs.GetChildrenResponse, error) {
	return &pfs.GetChildrenResponse{}, nil
}

func (a *apiServer) Branch(ctx context.Context, branchRequest *pfs.BranchRequest) (*pfs.BranchResponse, error) {
	return &pfs.BranchResponse{}, nil
}

func (a *apiServer) Commit(ctx context.Context, commitRequest *pfs.CommitRequest) (*pfs.CommitResponse, error) {
	return &pfs.CommitResponse{}, nil
}

func (a *apiServer) PullDiff(pullDiffRequest *pfs.PullDiffRequest, apiPullDiffServer pfs.Api_PullDiffServer) error {
	return nil
}

func (a *apiServer) PushDiff(ctx context.Context, pushDiffRequest *pfs.PushDiffRequest) (*pfs.PushDiffResponse, error) {
	return &pfs.PushDiffResponse{}, nil
}

func (a *apiServer) GetRepositoryInfo(ctx context.Context, getRepositoryInfoRequest *pfs.GetRepositoryInfoRequest) (*pfs.GetRepositoryInfoResponse, error) {
	return &pfs.GetRepositoryInfoResponse{}, nil
}

func (a *apiServer) GetCommitInfo(ctx context.Context, getCommitInfoRequest *pfs.GetCommitInfoRequest) (*pfs.GetCommitInfoResponse, error) {
	return &pfs.GetCommitInfoResponse{}, nil
}

// TODO(pedge)
func (a *apiServer) getMasterShards() (map[int]bool, error) {
	m := make(map[int]bool)
	numShards := a.sharder.NumShards()
	for i := 0; i < numShards; i++ {
		ok, err := a.router.IsLocalMasterShard(i)
		if err != nil {
			return nil, err
		}
		if ok {
			m[i] = true
		}
	}
	return m, nil
}
