package server

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"strings"

	"google.golang.org/grpc"

	"golang.org/x/net/context"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/drive"
	"github.com/pachyderm/pachyderm/src/pfs/route"
	"go.pedge.io/google-protobuf"
	"go.pedge.io/proto/stream"
)

var (
	emptyInstance = &google_protobuf.Empty{}
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

func (a *combinedAPIServer) RepoCreate(ctx context.Context, request *pfs.RepoCreateRequest) (*google_protobuf.Empty, error) {
	masterShards, err := a.router.GetMasterShards()
	if err != nil {
		return nil, err
	}
	if err := a.driver.RepoCreate(request.Repo, masterShards); err != nil {
		return nil, err
	}
	replicaShards, err := a.router.GetReplicaShards()
	if err != nil {
		return nil, err
	}
	if err := a.driver.RepoCreate(request.Repo, replicaShards); err != nil {
		return nil, err
	}
	if !request.Redirect {
		clientConns, err := a.router.GetAllClientConns()
		if err != nil {
			return nil, err
		}
		for _, clientConn := range clientConns {
			if _, err := pfs.NewApiClient(clientConn).RepoCreate(
				ctx,
				&pfs.RepoCreateRequest{
					Repo:     request.Repo,
					Redirect: true,
				},
			); err != nil {
				return nil, err
			}
		}
		// Create the initial commit
		if _, err = a.CommitStart(ctx, &pfs.CommitStartRequest{
			Parent: nil,
			Commit: &pfs.Commit{
				Repo: request.Repo,
				Id:   InitialCommitID,
			},
			Redirect: false,
		}); err != nil {
			return nil, err
		}
		if _, err = a.CommitFinish(ctx, &pfs.CommitFinishRequest{
			Commit: &pfs.Commit{
				Repo: request.Repo,
				Id:   InitialCommitID,
			},
			Redirect: false,
		}); err != nil {
			return nil, err
		}
	}
	return emptyInstance, nil
}

func (a *combinedAPIServer) RepoInspect(ctx context.Context, request *pfs.RepoInspectRequest) (*pfs.RepoInspectResponse, error) {
	masterShards, err := a.router.GetMasterShards()
	if err != nil {
		return nil, err
	}
	for shard := range masterShards {
		repoInfo, ok, err := a.driver.RepoInspect(request.Repo, shard)
		if err != nil {
			return nil, err
		}
		if !ok {
			return &pfs.RepoInspectResponse{}, nil
		}
		return &pfs.RepoInspectResponse{RepoInfo: repoInfo}, nil
	}
	if !request.Redirect {
		clientConns, err := a.router.GetAllClientConns()
		if err != nil {
			return nil, err
		}
		for _, clientConn := range clientConns {
			return pfs.NewApiClient(clientConn).RepoInspect(
				ctx,
				&pfs.RepoInspectRequest{
					Repo:     request.Repo,
					Redirect: true,
				},
			)
		}
	}
	return nil, fmt.Errorf("pachyderm: no available masters")
}

func (a *combinedAPIServer) RepoList(ctx context.Context, request *pfs.RepoListRequest) (*pfs.RepoListResponse, error) {
	masterShards, err := a.router.GetMasterShards()
	if err != nil {
		return nil, err
	}
	for shard := range masterShards {
		repos, err := a.driver.RepoList(shard)
		if err != nil {
			return nil, err
		}
		return &pfs.RepoListResponse{RepoInfo: repos}, nil
	}
	if !request.Redirect {
		clientConns, err := a.router.GetAllClientConns()
		if err != nil {
			return nil, err
		}
		for _, clientConn := range clientConns {
			return pfs.NewApiClient(clientConn).RepoList(ctx, &pfs.RepoListRequest{Redirect: true})
		}
	}
	return nil, fmt.Errorf("pachyderm: no available masters")
}

func (a *combinedAPIServer) RepoDelete(ctx context.Context, request *pfs.RepoDeleteRequest) (*google_protobuf.Empty, error) {
	masterShards, err := a.router.GetMasterShards()
	if err != nil {
		return nil, err
	}
	if err := a.driver.RepoDelete(request.Repo, masterShards); err != nil {
		return nil, err
	}
	replicaShards, err := a.router.GetReplicaShards()
	if err != nil {
		return nil, err
	}
	if err := a.driver.RepoDelete(request.Repo, replicaShards); err != nil {
		return nil, err
	}
	if !request.Redirect {
		clientConns, err := a.router.GetAllClientConns()
		if err != nil {
			return nil, err
		}
		for _, clientConn := range clientConns {
			if _, err := pfs.NewApiClient(clientConn).RepoCreate(
				ctx,
				&pfs.RepoCreateRequest{
					Repo:     request.Repo,
					Redirect: true,
				},
			); err != nil {
				return nil, err
			}
		}
	}
	return emptyInstance, nil

}

func (a *combinedAPIServer) CommitStart(ctx context.Context, request *pfs.CommitStartRequest) (*pfs.CommitStartResponse, error) {
	if request.Redirect && request.Commit == nil {
		return nil, fmt.Errorf("must set a commit for redirect %+v", request)
	}
	shards, err := a.getAllShards(false)
	if err != nil {
		return nil, err
	}
	commit, err := a.driver.CommitStart(request.Parent, request.Commit, shards)
	if err != nil {
		return nil, err
	}
	if !request.Redirect {
		clientConns, err := a.router.GetAllClientConns()
		if err != nil {
			return nil, err
		}
		for _, clientConn := range clientConns {
			if _, err := pfs.NewApiClient(clientConn).CommitStart(
				ctx,
				&pfs.CommitStartRequest{
					Parent:   request.Parent,
					Commit:   commit,
					Redirect: true,
				},
			); err != nil {
				return nil, err
			}
		}
	}
	return &pfs.CommitStartResponse{
		Commit: commit,
	}, nil
}

func (a *combinedAPIServer) CommitFinish(ctx context.Context, request *pfs.CommitFinishRequest) (*google_protobuf.Empty, error) {
	shards, err := a.router.GetMasterShards()
	if err != nil {
		return nil, err
	}
	if err := a.driver.CommitFinish(request.Commit, shards); err != nil {
		return nil, err
	}
	if err := a.commitToReplicas(ctx, request.Commit); err != nil {
		return nil, err
	}
	if !request.Redirect {
		clientConns, err := a.router.GetAllClientConns()
		if err != nil {
			return nil, err
		}
		for _, clientConn := range clientConns {
			if _, err := pfs.NewApiClient(clientConn).CommitFinish(
				ctx,
				&pfs.CommitFinishRequest{
					Commit:   request.Commit,
					Redirect: true,
				},
			); err != nil {
				return nil, err
			}
		}
	}
	return emptyInstance, nil
}

// TODO(pedge): race on Branch
func (a *combinedAPIServer) CommitInspect(ctx context.Context, request *pfs.CommitInspectRequest) (*pfs.CommitInspectResponse, error) {
	shard, clientConn, err := a.getMasterShardOrMasterClientConnIfNecessary()
	if err != nil {
		return nil, err
	}
	if clientConn != nil {
		return pfs.NewApiClient(clientConn).CommitInspect(ctx, request)
	}
	commitInfo, ok, err := a.driver.CommitInspect(request.Commit, shard)
	if err != nil {
		return nil, err
	}
	if !ok {
		return &pfs.CommitInspectResponse{}, nil
	}
	return &pfs.CommitInspectResponse{
		CommitInfo: commitInfo,
	}, nil
}

func (a *combinedAPIServer) CommitList(ctx context.Context, request *pfs.CommitListRequest) (*pfs.CommitListResponse, error) {
	shard, clientConn, err := a.getMasterShardOrMasterClientConnIfNecessary()
	if err != nil {
		return nil, err
	}
	if clientConn != nil {
		return pfs.NewApiClient(clientConn).CommitList(ctx, request)
	}
	commitInfos, err := a.driver.CommitList(request.Repo, shard)
	if err != nil {
		return nil, err
	}
	return &pfs.CommitListResponse{
		CommitInfo: commitInfos,
	}, nil
}

func (a *combinedAPIServer) CommitDelete(ctx context.Context, request *pfs.CommitDeleteRequest) (*google_protobuf.Empty, error) {
	shards, err := a.router.GetMasterShards()
	if err != nil {
		return nil, err
	}
	if err := a.driver.CommitDelete(request.Commit, shards); err != nil {
		return nil, err
	}
	// TODO push delete to replicas
	if !request.Redirect {
		clientConns, err := a.router.GetAllClientConns()
		if err != nil {
			return nil, err
		}
		for _, clientConn := range clientConns {
			if _, err := pfs.NewApiClient(clientConn).CommitDelete(
				ctx,
				&pfs.CommitDeleteRequest{
					Commit:   request.Commit,
					Redirect: true,
				},
			); err != nil {
				return nil, err
			}
		}
	}
	return emptyInstance, nil
}

func (a *combinedAPIServer) FilePut(ctx context.Context, request *pfs.FilePutRequest) (*google_protobuf.Empty, error) {
	if request.FileType == pfs.FileType_FILE_TYPE_DIR {
		if len(request.Value) > 0 {
			return emptyInstance, fmt.Errorf("FilePutRequest shouldn't have type dir and a value")
		}
		shards, err := a.getAllShards(false)
		if err != nil {
			return nil, err
		}
		if err := a.driver.MakeDirectory(request.File, shards); err != nil {
			return nil, err
		}
		if !request.Redirect {
			clientConns, err := a.router.GetAllClientConns()
			if err != nil {
				return nil, err
			}
			for _, clientConn := range clientConns {
				if _, err := pfs.NewApiClient(clientConn).FilePut(
					ctx,
					&pfs.FilePutRequest{
						File:     request.File,
						FileType: pfs.FileType_FILE_TYPE_DIR,
						Redirect: true,
					},
				); err != nil {
					return nil, err
				}
			}
		}
		return emptyInstance, nil
	}
	if strings.HasPrefix(request.File.Path, "/") {
		// This is a subtle error case, the paths foo and /foo will hash to
		// different shards but will produce the same change once they get to
		// those shards due to how path.Join. This can go wrong in a number of
		// ways so we forbid leading slashes.
		return nil, fmt.Errorf("pachyderm: leading slash in path: %s", request.File.Path)
	}
	shard, clientConn, err := a.getShardAndClientConnIfNecessary(request.File, false)
	if err != nil {
		return nil, err
	}
	if clientConn != nil {
		return pfs.NewApiClient(clientConn).FilePut(ctx, request)
	}
	if err := a.driver.FilePut(request.File, shard, request.OffsetBytes, bytes.NewReader(request.Value)); err != nil {
		return nil, err
	}
	return emptyInstance, nil
}

func (a *combinedAPIServer) FileGet(request *pfs.FileGetRequest, apiFileGetServer pfs.Api_FileGetServer) (retErr error) {
	shard, clientConn, err := a.getShardAndClientConnIfNecessary(request.File, false)
	if err != nil {
		return err
	}
	if clientConn != nil {
		apiFileGetClient, err := pfs.NewApiClient(clientConn).FileGet(context.Background(), request)
		if err != nil {
			return err
		}
		return protostream.RelayFromStreamingBytesClient(apiFileGetClient, apiFileGetServer)
	}
	file, err := a.driver.FileGet(request.File, shard)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	return protostream.WriteToStreamingBytesServer(
		io.NewSectionReader(file, request.OffsetBytes, request.SizeBytes),
		apiFileGetServer,
	)
}

func (a *combinedAPIServer) FileInspect(ctx context.Context, request *pfs.FileInspectRequest) (*pfs.FileInspectResponse, error) {
	shard, clientConn, err := a.getShardAndClientConnIfNecessary(request.File, false)
	if err != nil {
		return nil, err
	}
	if clientConn != nil {
		return pfs.NewApiClient(clientConn).FileInspect(context.Background(), request)
	}
	fileInfo, ok, err := a.driver.FileInspect(request.File, shard)
	if err != nil {
		return nil, err
	}
	if !ok {
		return &pfs.FileInspectResponse{}, nil
	}
	return &pfs.FileInspectResponse{
		FileInfo: fileInfo,
	}, nil
}

func (a *combinedAPIServer) FileList(ctx context.Context, request *pfs.FileListRequest) (*pfs.FileListResponse, error) {
	shards, err := a.getAllShards(false)
	if err != nil {
		return nil, err
	}
	dynamicShard := request.Shard
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
		subFileInfos, err := a.driver.FileList(request.File, shard)
		if err != nil {
			return nil, err
		}
		for _, fileInfo := range subFileInfos {
			if fileInfo.FileType == pfs.FileType_FILE_TYPE_DIR {
				if seenDirectories[fileInfo.File.Path] {
					continue
				}
				seenDirectories[fileInfo.File.Path] = true
			}
			fileInfos = append(fileInfos, fileInfo)
		}
	}
	if !request.Redirect {
		clientConns, err := a.router.GetAllClientConns()
		if err != nil {
			return nil, err
		}
		for _, clientConn := range clientConns {
			listFilesResponse, err := pfs.NewApiClient(clientConn).FileList(
				ctx,
				&pfs.FileListRequest{
					File:     request.File,
					Shard:    request.Shard,
					Redirect: true,
				},
			)
			if err != nil {
				return nil, err
			}
			for _, fileInfo := range listFilesResponse.FileInfo {
				if fileInfo.FileType == pfs.FileType_FILE_TYPE_DIR {
					if seenDirectories[fileInfo.File.Path] {
						continue
					}
					seenDirectories[fileInfo.File.Path] = true
				}
				fileInfos = append(fileInfos, fileInfo)
			}
		}
	}
	return &pfs.FileListResponse{
		FileInfo: fileInfos,
	}, nil
}

func (a *combinedAPIServer) FileDelete(ctx context.Context, request *pfs.FileDeleteRequest) (*google_protobuf.Empty, error) {
	if strings.HasPrefix(request.File.Path, "/") {
		// This is a subtle error case, the paths foo and /foo will hash to
		// different shards but will produce the same change once they get to
		// those shards due to how path.Join. This can go wrong in a number of
		// ways so we forbid leading slashes.
		return nil, fmt.Errorf("pachyderm: leading slash in path: %s", request.File.Path)
	}
	shard, clientConn, err := a.getShardAndClientConnIfNecessary(request.File, false)
	if err != nil {
		return nil, err
	}
	if clientConn != nil {
		return pfs.NewApiClient(clientConn).FileDelete(ctx, request)
	}
	if err := a.driver.FileDelete(request.File, shard); err != nil {
		return nil, err
	}
	return emptyInstance, nil
}

func (a *combinedAPIServer) PullDiff(request *pfs.PullDiffRequest, apiPullDiffServer pfs.InternalApi_PullDiffServer) error {
	clientConn, err := a.getClientConnIfNecessary(int(request.Shard), false)
	if err != nil {
		return err
	}
	if clientConn != nil {
		apiPullDiffClient, err := pfs.NewInternalApiClient(clientConn).PullDiff(context.Background(), request)
		if err != nil {
			return err
		}
		return protostream.RelayFromStreamingBytesClient(apiPullDiffClient, apiPullDiffServer)
	}
	var buffer bytes.Buffer
	a.driver.DiffPull(request.Commit, int(request.Shard), &buffer)
	return protostream.WriteToStreamingBytesServer(
		&buffer,
		apiPullDiffServer,
	)
}

func (a *combinedAPIServer) PushDiff(ctx context.Context, pushDiffRequest *pfs.PushDiffRequest) (*google_protobuf.Empty, error) {
	ok, err := a.isLocalReplicaShard(int(pushDiffRequest.Shard))
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("pachyderm: illegal PushDiffRequest for unknown shard %d", pushDiffRequest.Shard)
	}
	return emptyInstance, a.driver.DiffPush(pushDiffRequest.Commit, bytes.NewReader(pushDiffRequest.Value))
}

func (a *combinedAPIServer) Master(shard int) error {
	clientConns, err := a.router.GetReplicaClientConns(shard)
	if err != nil {
		return err
	}
	for _, clientConn := range clientConns {
		apiClient := pfs.NewApiClient(clientConn)
		response, err := apiClient.RepoList(context.Background(), &pfs.RepoListRequest{})
		if err != nil {
			return err
		}
		for _, repoInfo := range response.RepoInfo {
			if err := a.driver.RepoCreate(repoInfo.Repo, map[int]bool{shard: true}); err != nil {
				return err
			}
			response, err := apiClient.CommitList(context.Background(), &pfs.CommitListRequest{Repo: repoInfo.Repo})
			if err != nil {
				return err
			}
			localCommitInfo, err := a.driver.CommitList(repoInfo.Repo, shard)
			if err != nil {
				return err
			}
			for i, commitInfo := range response.CommitInfo {
				if i < len(localCommitInfo) {
					if *commitInfo != *localCommitInfo[i] {
						return fmt.Errorf("divergent data")
					}
					continue
				}
				pullDiffClient, err := pfs.NewInternalApiClient(clientConn).PullDiff(
					context.Background(),
					&pfs.PullDiffRequest{
						Commit: commitInfo.Commit,
						Shard:  uint64(shard),
					},
				)
				if err != nil {
					return err
				}
				if err := a.driver.DiffPush(commitInfo.Commit, protostream.NewStreamingBytesReader(pullDiffClient)); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (a *combinedAPIServer) Replica(shard int) error {
	return nil
}

func (a *combinedAPIServer) Clear(shard int) error {
	return nil
}

func (a *combinedAPIServer) getShardAndClientConnIfNecessary(file *pfs.File, replicaOk bool) (int, *grpc.ClientConn, error) {
	shard, err := a.sharder.GetShard(file)
	if err != nil {
		return shard, nil, err
	}
	clientConn, err := a.getClientConnIfNecessary(shard, replicaOk)
	return shard, clientConn, err
}

func (a *combinedAPIServer) getClientConnIfNecessary(shard int, replicaOk bool) (*grpc.ClientConn, error) {
	ok, err := a.isLocalMasterShard(shard)
	if err != nil {
		return nil, err
	}
	if !ok {
		if !replicaOk {
			clientConn, err := a.router.GetMasterClientConn(shard)
			return clientConn, err
		}
		ok, err = a.isLocalReplicaShard(shard)
		if err != nil {
			return nil, err
		}
		if !ok {
			clientConn, err := a.router.GetMasterOrReplicaClientConn(shard)
			return clientConn, err
		}
	}
	return nil, nil
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

func (a *combinedAPIServer) getAllShards(replicaToo bool) (map[int]bool, error) {
	shards, err := a.router.GetMasterShards()
	if err != nil {
		return nil, err
	}
	if replicaToo {
		replicaShards, err := a.router.GetReplicaShards()
		if err != nil {
			return nil, err
		}
		for replicaShard := range replicaShards {
			shards[replicaShard] = true
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

func (a *combinedAPIServer) isLocalReplicaShard(shard int) (bool, error) {
	shards, err := a.router.GetReplicaShards()
	if err != nil {
		return false, err
	}
	_, ok := shards[shard]
	return ok, nil
}

func (a *combinedAPIServer) commitToReplicas(ctx context.Context, commit *pfs.Commit) error {
	shards, err := a.router.GetMasterShards()
	if err != nil {
		return err
	}
	for shard := range shards {
		clientConns, err := a.router.GetReplicaClientConns(shard)
		if err != nil {
			return err
		}
		var diff bytes.Buffer
		if err = a.driver.DiffPull(commit, shard, &diff); err != nil {
			return err
		}
		for _, clientConn := range clientConns {
			if _, err = pfs.NewInternalApiClient(clientConn).PushDiff(
				ctx,
				&pfs.PushDiffRequest{
					Commit: commit,
					Shard:  uint64(shard),
					Value:  diff.Bytes(),
				},
			); err != nil {
				return err
			}
		}
	}
	return nil
}
