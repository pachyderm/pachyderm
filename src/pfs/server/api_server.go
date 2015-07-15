package server

import (
	"golang.org/x/net/context"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/route"
	"github.com/pachyderm/pachyderm/src/pfs/shard"
)

type apiServer struct {
	sharder shard.Sharder
	router  route.Router
}

func newAPIServer(
	sharder shard.Sharder,
	router route.Router,
) *apiServer {
	return &apiServer{
		sharder,
		router,
	}
}

func (a *apiServer) InitRepository(ctx context.Context, initRepositoryRequest *pfs.InitRepositoryRequest) (*pfs.InitRepositoryResponse, error) {
	return &pfs.InitRepositoryResponse{}, nil
}

func (a *apiServer) GetFile(getFileRequest *pfs.GetFileRequest, apiGetFileServer pfs.Api_GetFileServer) error {
	return nil
}

func (a *apiServer) PutFile(ctx context.Context, putFileRequest *pfs.PutFileRequest) (*pfs.PutFileResponse, error) {
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
