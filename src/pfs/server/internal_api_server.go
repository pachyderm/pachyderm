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

func (a *internalAPIServer) RepoCreate(ctx context.Context, request *pfs.RepoCreateRequest) (*google_protobuf.Empty, error) {
	if err := a.driver.RepoCreate(request.Repo); err != nil {
		return nil, err
	}
	return emptyInstance, nil
}

func (a *internalAPIServer) RepoInspect(ctx context.Context, request *pfs.RepoInspectRequest) (*pfs.RepoInfo, error) {
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shards, err := a.router.GetAllShards(version)
	if err != nil {
		return nil, err
	}
	for shard := range shards {
		return a.driver.RepoInspect(request.Repo, shard)
	}
	return nil, fmt.Errorf("pachyderm: RepoInspect on server with no shards")
}

func (a *internalAPIServer) RepoList(ctx context.Context, request *pfs.RepoListRequest) (*pfs.RepoInfos, error) {
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shards, err := a.router.GetAllShards(version)
	if err != nil {
		return nil, err
	}
	for shard := range shards {
		repoInfos, err := a.driver.RepoList(shard)
		return &pfs.RepoInfos{RepoInfo: repoInfos}, err
	}
	return nil, fmt.Errorf("pachyderm: RepoList on server with no shards")
}

func (a *internalAPIServer) RepoDelete(ctx context.Context, request *pfs.RepoDeleteRequest) (*google_protobuf.Empty, error) {
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shards, err := a.router.GetAllShards(version)
	if err != nil {
		return nil, err
	}
	if err := a.driver.RepoDelete(request.Repo, shards); err != nil {
		return nil, err
	}
	return emptyInstance, nil

}

func (a *internalAPIServer) CommitStart(ctx context.Context, request *pfs.CommitStartRequest) (*pfs.Commit, error) {
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shards, err := a.router.GetMasterShards(version)
	if err != nil {
		return nil, err
	}
	return a.driver.CommitStart(request.Parent, request.Commit, shards)
}

func (a *internalAPIServer) CommitFinish(ctx context.Context, request *pfs.CommitFinishRequest) (*google_protobuf.Empty, error) {
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shards, err := a.router.GetMasterShards(version)
	if err != nil {
		return nil, err
	}
	if err := a.driver.CommitFinish(request.Commit, shards); err != nil {
		return nil, err
	}
	if err := a.commitToReplicas(ctx, request.Commit); err != nil {
		return nil, err
	}
	return emptyInstance, nil
}

// TODO(pedge): race on Branch
func (a *internalAPIServer) CommitInspect(ctx context.Context, request *pfs.CommitInspectRequest) (*pfs.CommitInfo, error) {
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shards, err := a.router.GetAllShards(version)
	if err != nil {
		return nil, err
	}
	for shard := range shards {
		return a.driver.CommitInspect(request.Commit, shard)
	}
	return nil, fmt.Errorf("pachyderm: CommitInspect on server with no shards")
}

func (a *internalAPIServer) CommitList(ctx context.Context, request *pfs.CommitListRequest) (*pfs.CommitInfos, error) {
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shards, err := a.router.GetAllShards(version)
	if err != nil {
		return nil, err
	}
	for shard := range shards {
		commitInfos, err := a.driver.CommitList(request.Repo, shard)
		if err != nil {
			return nil, err
		}
		return &pfs.CommitInfos{
			CommitInfo: commitInfos,
		}, nil
	}
	return nil, fmt.Errorf("pachyderm: CommitList on server with no shards")
}

func (a *internalAPIServer) CommitDelete(ctx context.Context, request *pfs.CommitDeleteRequest) (*google_protobuf.Empty, error) {
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shards, err := a.router.GetAllShards(version)
	if err != nil {
		return nil, err
	}
	if err := a.driver.CommitDelete(request.Commit, shards); err != nil {
		return nil, err
	}
	// TODO push delete to replicas
	return emptyInstance, nil
}

func (a *internalAPIServer) FilePut(ctx context.Context, request *pfs.FilePutRequest) (*google_protobuf.Empty, error) {
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
			return emptyInstance, fmt.Errorf("FilePutRequest shouldn't have type dir and a value")
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
	if err := a.driver.FilePut(request.File, shard, request.OffsetBytes, bytes.NewReader(request.Value)); err != nil {
		return nil, err
	}
	return emptyInstance, nil
}

func (a *internalAPIServer) FileGet(request *pfs.FileGetRequest, apiFileGetServer pfs.InternalApi_FileGetServer) (retErr error) {
	version, err := a.getVersion(apiFileGetServer.Context())
	if err != nil {
		return err
	}
	shard, err := a.getShardForFile(request.File, version)
	if err != nil {
		return err
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

func (a *internalAPIServer) FileInspect(ctx context.Context, request *pfs.FileInspectRequest) (*pfs.FileInfo, error) {
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shard, err := a.getShardForFile(request.File, version)
	if err != nil {
		return nil, err
	}
	return a.driver.FileInspect(request.File, shard)
}

func (a *internalAPIServer) FileList(ctx context.Context, request *pfs.FileListRequest) (*pfs.FileInfos, error) {
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
	return &pfs.FileInfos{
		FileInfo: fileInfos,
	}, nil
}

func (a *internalAPIServer) FileDelete(ctx context.Context, request *pfs.FileDeleteRequest) (*google_protobuf.Empty, error) {
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
	if err := a.driver.FileDelete(request.File, shard); err != nil {
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
	a.driver.DiffPull(request.Commit, request.Shard, &buffer)
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
	return emptyInstance, a.driver.DiffPush(request.Commit, bytes.NewReader(request.Value))
}

func (a *internalAPIServer) AddShard(shard uint64) error {
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
	shards, err := a.router.GetAllShards(version)
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
