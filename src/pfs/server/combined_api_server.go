package server

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"strings"

	"google.golang.org/grpc"

	"golang.org/x/net/context"

	"github.com/pachyderm/pachyderm"
	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/drive"
	"github.com/pachyderm/pachyderm/src/pfs/route"
	"github.com/pachyderm/pachyderm/src/pkg/protoutil"
	"github.com/peter-edge/go-google-protobuf"
)

var (
	emptyInstance              = &google_protobuf.Empty{}
	getVersionResponseInstance = &pfs.GetVersionResponse{
		Version: &pfs.Version{
			Major:      pachyderm.MajorVersion,
			Minor:      pachyderm.MinorVersion,
			Micro:      pachyderm.MicroVersion,
			Additional: pachyderm.AdditionalVersion,
		},
	}
)

type combinedAPIServer struct {
	sharder route.Sharder
	router  route.Router
	driver  drive.Driver
}

func newCombinedAPIServer(
	sharder route.Sharder,
	router route.Router,
	driver drive.Driver,
) *combinedAPIServer {
	return &combinedAPIServer{
		sharder,
		router,
		driver,
	}
}

func (a *combinedAPIServer) GetVersion(ctx context.Context, empty *google_protobuf.Empty) (*pfs.GetVersionResponse, error) {
	return getVersionResponseInstance, nil
}

func (a *combinedAPIServer) InitRepository(ctx context.Context, initRepositoryRequest *pfs.InitRepositoryRequest) (*google_protobuf.Empty, error) {
	shards, err := a.getAllShards(true)
	if err != nil {
		return nil, err
	}
	if err := a.driver.InitRepository(initRepositoryRequest.Repository, shards); err != nil {
		return nil, err
	}
	if !initRepositoryRequest.Redirect {
		clientConns, err := a.router.GetAllClientConns()
		if err != nil {
			return nil, err
		}
		for _, clientConn := range clientConns {
			if _, err := pfs.NewApiClient(clientConn).InitRepository(
				ctx,
				&pfs.InitRepositoryRequest{
					Repository: initRepositoryRequest.Repository,
					Redirect:   true,
				},
			); err != nil {
				return nil, err
			}
		}
	}
	return emptyInstance, nil
}

func (a *combinedAPIServer) GetFile(getFileRequest *pfs.GetFileRequest, apiGetFileServer pfs.Api_GetFileServer) (retErr error) {
	shard, clientConn, err := a.getShardAndClientConnIfNecessary(getFileRequest.Path, true)
	if err != nil {
		return err
	}
	if clientConn != nil {
		apiGetFileClient, err := pfs.NewApiClient(clientConn).GetFile(context.Background(), getFileRequest)
		if err != nil {
			return err
		}
		return protoutil.RelayFromStreamingBytesClient(apiGetFileClient, apiGetFileServer)
	}
	file, err := a.driver.GetFile(getFileRequest.Path, shard)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	return protoutil.WriteToStreamingBytesServer(
		io.NewSectionReader(file, getFileRequest.OffsetBytes, getFileRequest.SizeBytes),
		apiGetFileServer,
	)
}

func (a *combinedAPIServer) GetFileInfo(ctx context.Context, getFileInfoRequest *pfs.GetFileInfoRequest) (*pfs.GetFileInfoResponse, error) {
	shard, clientConn, err := a.getShardAndClientConnIfNecessary(getFileInfoRequest.Path, true)
	if err != nil {
		return nil, err
	}
	if clientConn != nil {
		return pfs.NewApiClient(clientConn).GetFileInfo(context.Background(), getFileInfoRequest)
	}
	fileInfo, ok, err := a.driver.GetFileInfo(getFileInfoRequest.Path, shard)
	if err != nil {
		return nil, err
	}
	if !ok {
		return &pfs.GetFileInfoResponse{}, nil
	}
	return &pfs.GetFileInfoResponse{
		FileInfo: fileInfo,
	}, nil
}

func (a *combinedAPIServer) MakeDirectory(ctx context.Context, makeDirectoryRequest *pfs.MakeDirectoryRequest) (*google_protobuf.Empty, error) {
	shards, err := a.getAllShards(true)
	if err != nil {
		return nil, err
	}
	if err := a.driver.MakeDirectory(makeDirectoryRequest.Path, shards); err != nil {
		return nil, err
	}
	if !makeDirectoryRequest.Redirect {
		clientConns, err := a.router.GetAllClientConns()
		if err != nil {
			return nil, err
		}
		for _, clientConn := range clientConns {
			if _, err := pfs.NewApiClient(clientConn).MakeDirectory(
				ctx,
				&pfs.MakeDirectoryRequest{
					Path:     makeDirectoryRequest.Path,
					Redirect: true,
				},
			); err != nil {
				return nil, err
			}
		}
	}
	return emptyInstance, nil
}

func (a *combinedAPIServer) PutFile(ctx context.Context, putFileRequest *pfs.PutFileRequest) (*google_protobuf.Empty, error) {
	if strings.HasPrefix(putFileRequest.Path.Path, "/") {
		// This is a subtle error case, the paths foo and /foo will hash to
		// different shards but will produce the same change once they get to
		// those shards due to how path.Join. This can go wrong in a number of
		// ways so we forbid leading slashes.
		return nil, fmt.Errorf("pachyderm: leading slash in path: %s", putFileRequest.Path.Path)
	}
	shard, clientConn, err := a.getShardAndClientConnIfNecessary(putFileRequest.Path, false)
	if err != nil {
		return nil, err
	}
	if clientConn != nil {
		return pfs.NewApiClient(clientConn).PutFile(ctx, putFileRequest)
	}
	if err := a.driver.PutFile(putFileRequest.Path, shard, putFileRequest.OffsetBytes, bytes.NewReader(putFileRequest.Value)); err != nil {
		return nil, err
	}
	return emptyInstance, nil
}

func (a *combinedAPIServer) ListFiles(ctx context.Context, listFilesRequest *pfs.ListFilesRequest) (*pfs.ListFilesResponse, error) {
	shards, err := a.getAllShards(false)
	if err != nil {
		return nil, err
	}
	dynamicShard := listFilesRequest.Shard
	if dynamicShard == nil {
		dynamicShard = &pfs.Shard{Number: 0, Modulo: 1}
	}
	filteredShards := make(map[int]bool)
	for shard := range shards {
		if uint64(shard)%dynamicShard.Modulo == dynamicShard.Number {
			filteredShards[shard] = true
		}
	}
	var fileInfos []*pfs.FileInfo
	seenDirectories := make(map[string]bool)
	for shard := range filteredShards {
		subFileInfos, err := a.driver.ListFiles(listFilesRequest.Path, shard)
		if err != nil {
			return nil, err
		}
		for _, fileInfo := range subFileInfos {
			if fileInfo.FileType == pfs.FileType_FILE_TYPE_DIR {
				if seenDirectories[fileInfo.Path.Path] {
					continue
				}
				seenDirectories[fileInfo.Path.Path] = true
			}
			fileInfos = append(fileInfos, fileInfo)
		}
	}
	if !listFilesRequest.Redirect {
		clientConns, err := a.router.GetAllClientConns()
		if err != nil {
			return nil, err
		}
		for _, clientConn := range clientConns {
			listFilesResponse, err := pfs.NewApiClient(clientConn).ListFiles(
				ctx,
				&pfs.ListFilesRequest{
					Path:     listFilesRequest.Path,
					Shard:    listFilesRequest.Shard,
					Redirect: true,
				},
			)
			if err != nil {
				return nil, err
			}
			for _, fileInfo := range listFilesResponse.FileInfo {
				if fileInfo.FileType == pfs.FileType_FILE_TYPE_DIR {
					if seenDirectories[fileInfo.Path.Path] {
						continue
					}
					seenDirectories[fileInfo.Path.Path] = true
				}
				fileInfos = append(fileInfos, fileInfo)
			}
		}
	}
	return &pfs.ListFilesResponse{
		FileInfo: fileInfos,
	}, nil
}

func (a *combinedAPIServer) Branch(ctx context.Context, branchRequest *pfs.BranchRequest) (*pfs.BranchResponse, error) {
	if branchRequest.Redirect {
		if branchRequest.NewCommit == nil {
			return nil, fmt.Errorf("must set a new commit for redirect %+v", branchRequest)
		}
	} else {
		if branchRequest.NewCommit != nil {
			return nil, fmt.Errorf("cannot set a new commit for non-redirect %+v", branchRequest)
		}
	}
	shards, err := a.getAllShards(false)
	if err != nil {
		return nil, err
	}
	newCommit, err := a.driver.Branch(branchRequest.Commit, branchRequest.NewCommit, shards)
	if err != nil {
		return nil, err
	}
	if !branchRequest.Redirect {
		clientConns, err := a.router.GetAllClientConns()
		if err != nil {
			return nil, err
		}
		for _, clientConn := range clientConns {
			if _, err := pfs.NewApiClient(clientConn).Branch(
				ctx,
				&pfs.BranchRequest{
					Commit:    branchRequest.Commit,
					Redirect:  true,
					NewCommit: newCommit,
				},
			); err != nil {
				return nil, err
			}
		}
	}
	return &pfs.BranchResponse{
		Commit: newCommit,
	}, nil
}

func (a *combinedAPIServer) Commit(ctx context.Context, commitRequest *pfs.CommitRequest) (*google_protobuf.Empty, error) {
	shards, err := a.getAllShards(false)
	if err != nil {
		return nil, err
	}
	if err := a.driver.Commit(commitRequest.Commit, shards); err != nil {
		return nil, err
	}
	if !commitRequest.Redirect {
		clientConns, err := a.router.GetAllClientConns()
		if err != nil {
			return nil, err
		}
		for _, clientConn := range clientConns {
			if _, err := pfs.NewApiClient(clientConn).Commit(
				ctx,
				&pfs.CommitRequest{
					Commit:   commitRequest.Commit,
					Redirect: true,
				},
			); err != nil {
				return nil, err
			}
		}
	}
	return emptyInstance, nil
}

func (a *combinedAPIServer) PullDiff(pullDiffRequest *pfs.PullDiffRequest, apiPullDiffServer pfs.InternalApi_PullDiffServer) error {
	return nil
}

func (a *combinedAPIServer) PushDiff(ctx context.Context, pushDiffRequest *pfs.PushDiffRequest) (*google_protobuf.Empty, error) {
	return emptyInstance, nil
}

// TODO(pedge): race on Branch
func (a *combinedAPIServer) GetCommitInfo(ctx context.Context, getCommitInfoRequest *pfs.GetCommitInfoRequest) (*pfs.GetCommitInfoResponse, error) {
	shard, clientConn, err := a.getMasterShardOrMasterClientConnIfNecessary()
	if err != nil {
		return nil, err
	}
	if clientConn != nil {
		return pfs.NewApiClient(clientConn).GetCommitInfo(ctx, getCommitInfoRequest)
	}
	commitInfo, ok, err := a.driver.GetCommitInfo(getCommitInfoRequest.Commit, shard)
	if err != nil {
		return nil, err
	}
	if !ok {
		return &pfs.GetCommitInfoResponse{}, nil
	}
	return &pfs.GetCommitInfoResponse{
		CommitInfo: commitInfo,
	}, nil
}

func (a *combinedAPIServer) ListCommits(ctx context.Context, listCommitsRequest *pfs.ListCommitsRequest) (*pfs.ListCommitsResponse, error) {
	shard, clientConn, err := a.getMasterShardOrMasterClientConnIfNecessary()
	if err != nil {
		return nil, err
	}
	if clientConn != nil {
		return pfs.NewApiClient(clientConn).ListCommits(ctx, listCommitsRequest)
	}
	commitInfos, err := a.driver.ListCommits(listCommitsRequest.Repository, shard)
	if err != nil {
		return nil, err
	}
	return &pfs.ListCommitsResponse{
		CommitInfo: commitInfos,
	}, nil
}

func (a *combinedAPIServer) getShardAndClientConnIfNecessary(path *pfs.Path, slaveOk bool) (int, *grpc.ClientConn, error) {
	shard, err := a.sharder.GetShard(path)
	if err != nil {
		return shard, nil, err
	}
	ok, err := a.isLocalMasterShard(shard)
	if err != nil {
		return shard, nil, err
	}
	if !ok {
		if !slaveOk {
			clientConn, err := a.router.GetMasterClientConn(shard)
			return shard, clientConn, err
		}
		ok, err = a.isLocalSlaveShard(shard)
		if err != nil {
			return shard, nil, err
		}
		if !ok {
			clientConn, err := a.router.GetMasterOrSlaveClientConn(shard)
			return shard, clientConn, err
		}
	}
	return shard, nil, nil
}

func (a *combinedAPIServer) getMasterShardOrMasterClientConnIfNecessary() (int, *grpc.ClientConn, error) {
	shards, err := a.router.GetMasterShards()
	if err != nil {
		return -1, nil, err
	}
	if len(shards) > 0 {
		for shard := range shards {
			return shard, nil, nil
		}
	}
	clientConn, err := a.router.GetMasterClientConn(int(rand.Uint32()) % a.sharder.NumShards())
	return -1, clientConn, err
}

func (a *combinedAPIServer) getAllShards(slaveToo bool) (map[int]bool, error) {
	shards, err := a.router.GetMasterShards()
	if err != nil {
		return nil, err
	}
	if slaveToo {
		slaveShards, err := a.router.GetSlaveShards()
		if err != nil {
			return nil, err
		}
		for slaveShard := range slaveShards {
			shards[slaveShard] = true
		}
	}
	return shards, nil
}

func (a *combinedAPIServer) isLocalMasterShard(shard int) (bool, error) {
	shards, err := a.router.GetMasterShards()
	if err != nil {
		return false, err
	}
	_, ok := shards[shard]
	return ok, nil
}

func (a *combinedAPIServer) isLocalSlaveShard(shard int) (bool, error) {
	shards, err := a.router.GetSlaveShards()
	if err != nil {
		return false, err
	}
	_, ok := shards[shard]
	return ok, nil
}
