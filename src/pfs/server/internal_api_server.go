package server

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"

	"google.golang.org/grpc/metadata"

	"golang.org/x/net/context"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/drive"
	"github.com/pachyderm/pachyderm/src/pfs/route"
	"go.pedge.io/google-protobuf"
	"go.pedge.io/proto/stream"
)

type internalAPIServer struct {
	sharder route.Sharder
	router  route.Router
	driver  drive.Driver
}

func newInternalAPIServer(
	sharder route.Sharder,
	router route.Router,
	driver drive.Driver,
) *internalAPIServer {
	return &internalAPIServer{
		sharder,
		router,
		driver,
	}
}

func (a *internalAPIServer) CreateRepo(ctx context.Context, request *pfs.CreateRepoRequest) (*google_protobuf.Empty, error) {
	if err := a.driver.CreateRepo(request.Repo); err != nil {
		return nil, err
	}
	return emptyInstance, nil
}

func (a *internalAPIServer) InspectRepo(ctx context.Context, request *pfs.InspectRepoRequest) (*pfs.RepoInfo, error) {
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shards, err := a.router.GetAllShards(version)
	if err != nil {
		return nil, err
	}
	for shard := range shards {
		return a.driver.InspectRepo(request.Repo, shard)
	}
	return nil, fmt.Errorf("pachyderm: InspectRepo on server with no shards")
}

func (a *internalAPIServer) ListRepo(ctx context.Context, request *pfs.ListRepoRequest) (*pfs.RepoInfos, error) {
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shards, err := a.router.GetAllShards(version)
	if err != nil {
		return nil, err
	}
	for shard := range shards {
		repoInfos, err := a.driver.ListRepo(shard)
		return &pfs.RepoInfos{RepoInfo: repoInfos}, err
	}
	return nil, fmt.Errorf("pachyderm: ListRepo on server with no shards")
}

func (a *internalAPIServer) DeleteRepo(ctx context.Context, request *pfs.DeleteRepoRequest) (*google_protobuf.Empty, error) {
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shards, err := a.router.GetAllShards(version)
	if err != nil {
		return nil, err
	}
	if err := a.driver.DeleteRepo(request.Repo, shards); err != nil {
		return nil, err
	}
	return emptyInstance, nil

}

func (a *internalAPIServer) StartCommit(ctx context.Context, request *pfs.StartCommitRequest) (*pfs.Commit, error) {
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shards, err := a.router.GetMasterShards(version)
	if err != nil {
		return nil, err
	}
	return a.driver.StartCommit(request.Parent, request.Commit, shards)
}

func (a *internalAPIServer) FinishCommit(ctx context.Context, request *pfs.FinishCommitRequest) (*google_protobuf.Empty, error) {
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shards, err := a.router.GetMasterShards(version)
	if err != nil {
		return nil, err
	}
	if err := a.driver.FinishCommit(request.Commit, shards); err != nil {
		return nil, err
	}
	if err := a.commitToReplicas(ctx, request.Commit); err != nil {
		return nil, err
	}
	return emptyInstance, nil
}

// TODO(pedge): race on Branch
func (a *internalAPIServer) InspectCommit(ctx context.Context, request *pfs.InspectCommitRequest) (*pfs.CommitInfo, error) {
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shards, err := a.router.GetMasterShards(version)
	if err != nil {
		return nil, err
	}
	for shard := range shards {
		return a.driver.InspectCommit(request.Commit, shard)
	}
	return nil, fmt.Errorf("pachyderm: InspectCommit on server with no shards")
}

func (a *internalAPIServer) ListCommit(ctx context.Context, request *pfs.ListCommitRequest) (*pfs.CommitInfos, error) {
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shards, err := a.router.GetMasterShards(version)
	if err != nil {
		return nil, err
	}
	for shard := range shards {
		commitInfos, err := a.driver.ListCommit(request.Repo, request.From, shard)
		if err != nil {
			return nil, err
		}
		return &pfs.CommitInfos{
			CommitInfo: commitInfos,
		}, nil
	}
	return nil, fmt.Errorf("pachyderm: ListCommit on server with no shards")
}

func (a *internalAPIServer) DeleteCommit(ctx context.Context, request *pfs.DeleteCommitRequest) (*google_protobuf.Empty, error) {
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shards, err := a.router.GetMasterShards(version)
	if err != nil {
		return nil, err
	}
	if err := a.driver.DeleteCommit(request.Commit, shards); err != nil {
		return nil, err
	}
	// TODO push delete to replicas
	return emptyInstance, nil
}

func (a *internalAPIServer) PutFile(ctx context.Context, request *pfs.PutFileRequest) (*google_protobuf.Empty, error) {
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(request.File.Path, "/") {
		// This is a subtle error case, the paths foo and /foo will hash to
		// different shards but will produce the same change once they get to
		// those shards due to how path.Join. This can go wrong in a number of
		// ways so we forbid leading slashes.
		return nil, fmt.Errorf("pachyderm: leading slash in path: %s", request.File.Path)
	}
	if request.FileType == pfs.FileType_FILE_TYPE_DIR {
		if len(request.Value) > 0 {
			return emptyInstance, fmt.Errorf("PutFileRequest shouldn't have type dir and a value")
		}
		shards, err := a.router.GetMasterShards(version)
		if err != nil {
			return nil, err
		}
		if err := a.driver.MakeDirectory(request.File, shards); err != nil {
			return nil, err
		}
		return emptyInstance, nil
	}
	shard, err := a.getMasterShardForFile(request.File, version)
	if err != nil {
		return nil, err
	}
	if err := a.driver.PutFile(request.File, shard, request.OffsetBytes, bytes.NewReader(request.Value)); err != nil {
		return nil, err
	}
	return emptyInstance, nil
}

func (a *internalAPIServer) GetFile(request *pfs.GetFileRequest, apiGetFileServer pfs.InternalApi_GetFileServer) (retErr error) {
	version, err := a.getVersion(apiGetFileServer.Context())
	if err != nil {
		return err
	}
	shard, err := a.getShardForFile(request.File, version)
	if err != nil {
		return err
	}
	file, err := a.driver.GetFile(request.File, shard)
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
		apiGetFileServer,
	)
}

func (a *internalAPIServer) InspectFile(ctx context.Context, request *pfs.InspectFileRequest) (*pfs.FileInfo, error) {
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shard, err := a.getShardForFile(request.File, version)
	if err != nil {
		return nil, err
	}
	return a.driver.InspectFile(request.File, shard)
}

func (a *internalAPIServer) ListFile(ctx context.Context, request *pfs.ListFileRequest) (*pfs.FileInfos, error) {
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shards, err := a.router.GetMasterShards(version)
	if err != nil {
		return nil, err
	}
	dynamicShard := request.Shard
	if dynamicShard == nil {
		dynamicShard = &pfs.Shard{Number: 0, Modulo: 1}
	}
	filteredShards := make(map[uint64]bool)
	for shard := range shards {
		if uint64(shard)%dynamicShard.Modulo == dynamicShard.Number {
			filteredShards[shard] = true
		}
	}
	var fileInfos []*pfs.FileInfo
	seenDirectories := make(map[string]bool)
	for shard := range filteredShards {
		subFileInfos, err := a.driver.ListFile(request.File, shard)
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
	return &pfs.FileInfos{
		FileInfo: fileInfos,
	}, nil
}

func (a *internalAPIServer) DeleteFile(ctx context.Context, request *pfs.DeleteFileRequest) (*google_protobuf.Empty, error) {
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(request.File.Path, "/") {
		// This is a subtle error case, the paths foo and /foo will hash to
		// different shards but will produce the same change once they get to
		// those shards due to how path.Join. This can go wrong in a number of
		// ways so we forbid leading slashes.
		return nil, fmt.Errorf("pachyderm: leading slash in path: %s", request.File.Path)
	}
	shard, err := a.getMasterShardForFile(request.File, version)
	if err != nil {
		return nil, err
	}
	if err := a.driver.DeleteFile(request.File, shard); err != nil {
		return nil, err
	}
	return emptyInstance, nil
}

func (a *internalAPIServer) PullDiff(request *pfs.PullDiffRequest, apiPullDiffServer pfs.InternalApi_PullDiffServer) error {
	version, err := a.getVersion(apiPullDiffServer.Context())
	if err != nil {
		return err
	}
	ok, err := a.isLocalShard(request.Shard, version)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("pachyderm: illegal PullDiffRequest for unknown shard %d", request.Shard)
	}
	var buffer bytes.Buffer
	a.driver.PullDiff(request.Commit, request.Shard, &buffer)
	return protostream.WriteToStreamingBytesServer(
		&buffer,
		apiPullDiffServer,
	)
}

func (a *internalAPIServer) PushDiff(ctx context.Context, request *pfs.PushDiffRequest) (*google_protobuf.Empty, error) {
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	ok, err := a.isLocalReplicaShard(request.Shard, version)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("pachyderm: illegal PushDiffRequest for unknown shard %d", request.Shard)
	}
	return emptyInstance, a.driver.PushDiff(request.Commit, bytes.NewReader(request.Value))
}

func (a *internalAPIServer) AddShard(shard uint64) error {
	version, ctx, err := a.versionAndCtx(context.Background())
	if err != nil {
		return err
	}
	if version == route.InvalidVersion {
		return nil
	}
	clientConn, err := a.router.GetMasterOrReplicaClientConn(shard, version)
	if err != nil {
		return err
	}
	repoInfos, err := pfs.NewInternalApiClient(clientConn).ListRepo(ctx, &pfs.ListRepoRequest{})
	if err != nil {
		return err
	}
	for _, repoInfo := range repoInfos.RepoInfo {
		if err := a.driver.CreateRepo(repoInfo.Repo); err != nil {
			return err
		}
		commitInfos, err := pfs.NewInternalApiClient(clientConn).ListCommit(ctx, &pfs.ListCommitRequest{Repo: repoInfo.Repo})
		if err != nil {
			return err
		}
		for i := range commitInfos.CommitInfo {
			commit := commitInfos.CommitInfo[len(commitInfos.CommitInfo)-(i+1)].Commit
			commitInfo, err := a.driver.InspectCommit(commit, shard)
			if err != nil {
				return err
			}
			if commitInfo != nil {
				// we already have the commit so nothing to do
				continue
			}
			pullDiffRequest := &pfs.PullDiffRequest{
				Commit: commit,
				Shard:  shard,
			}
			pullDiffClient, err := pfs.NewInternalApiClient(clientConn).PullDiff(ctx, pullDiffRequest)
			if err != nil {
				return err
			}
			diffReader := protostream.NewStreamingBytesReader(pullDiffClient)
			a.driver.PushDiff(commit, diffReader)
		}
	}
	return nil
}

func (a *internalAPIServer) RemoveShard(shard uint64) error {
	return nil
}

func (a *internalAPIServer) LocalShards() (map[uint64]bool, error) {
	return nil, nil
}

func (a *internalAPIServer) getMasterShardForFile(file *pfs.File, version int64) (uint64, error) {
	shard, err := a.sharder.GetShard(file)
	if err != nil {
		return 0, err
	}
	shards, err := a.router.GetMasterShards(version)
	if err != nil {
		return 0, err
	}
	_, ok := shards[shard]
	if !ok {
		return 0, fmt.Errorf("pachyderm: shard %d not found locally", shard)
	}
	return shard, nil
}

func (a *internalAPIServer) getShardForFile(file *pfs.File, version int64) (uint64, error) {
	shard, err := a.sharder.GetShard(file)
	if err != nil {
		return 0, err
	}
	shards, err := a.router.GetMasterShards(version)
	if err != nil {
		return 0, err
	}
	_, ok := shards[shard]
	if !ok {
		return 0, fmt.Errorf("pachyderm: shard %d not found locally", shard)
	}
	return shard, nil
}

func (a *internalAPIServer) isLocalMasterShard(shard uint64, version int64) (bool, error) {
	shards, err := a.router.GetMasterShards(version)
	if err != nil {
		return false, err
	}
	_, ok := shards[shard]
	return ok, nil
}

func (a *internalAPIServer) isLocalReplicaShard(shard uint64, version int64) (bool, error) {
	shards, err := a.router.GetReplicaShards(version)
	if err != nil {
		return false, err
	}
	_, ok := shards[shard]
	return ok, nil
}

func (a *internalAPIServer) isLocalShard(shard uint64, version int64) (bool, error) {
	shards, err := a.router.GetAllShards(version)
	if err != nil {
		return false, err
	}
	_, ok := shards[shard]
	return ok, nil
}

func (a *internalAPIServer) commitToReplicas(ctx context.Context, commit *pfs.Commit) error {
	version, err := a.getVersion(ctx)
	if err != nil {
		return err
	}
	shards, err := a.router.GetMasterShards(version)
	if err != nil {
		return err
	}
	for shard := range shards {
		clientConns, err := a.router.GetReplicaClientConns(shard, version)
		if err != nil {
			return err
		}
		var diff bytes.Buffer
		if err = a.driver.PullDiff(commit, shard, &diff); err != nil {
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

func (a *internalAPIServer) getVersion(ctx context.Context) (int64, error) {
	md, ok := metadata.FromContext(ctx)
	if !ok {
		return 0, fmt.Errorf("version not found in context")
	}
	encodedVersion, ok := md["version"]
	if !ok {
		return 0, fmt.Errorf("version not found in context")
	}
	if len(encodedVersion) != 1 {
		return 0, fmt.Errorf("version not found in context")
	}
	return strconv.ParseInt(encodedVersion[0], 10, 64)
}

func (a *internalAPIServer) versionAndCtx(ctx context.Context) (int64, context.Context, error) {
	version, err := a.router.Version()
	if err != nil {
		return 0, nil, err
	}
	newCtx := metadata.NewContext(
		ctx,
		metadata.Pairs("version", fmt.Sprint(version)),
	)
	return version, newCtx, nil
}
