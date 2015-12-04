package server

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.pedge.io/google-protobuf"
	"go.pedge.io/proto/rpclog"
	"go.pedge.io/proto/stream"
	"go.pedge.io/protolog"
	"golang.org/x/net/context"
	"google.golang.org/grpc/metadata"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/drive"
	"github.com/pachyderm/pachyderm/src/pfs/route"
	"github.com/pachyderm/pachyderm/src/pkg/shard"
)

type internalAPIServer struct {
	protorpclog.Logger
	sharder           route.Sharder
	router            route.Router
	driver            drive.Driver
	commitWaiters     map[commitWait][]chan *pfs.CommitInfo
	commitWaitersLock sync.Mutex
}

func newInternalAPIServer(
	sharder route.Sharder,
	router route.Router,
	driver drive.Driver,
) *internalAPIServer {
	return &internalAPIServer{
		protorpclog.NewLogger("pachyderm.pfs.InternalAPI"),
		sharder,
		router,
		driver,
		make(map[commitWait][]chan *pfs.CommitInfo),
		sync.Mutex{},
	}
}

func (a *internalAPIServer) CreateRepo(ctx context.Context, request *pfs.CreateRepoRequest) (response *google_protobuf.Empty, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	if err := a.driver.CreateRepo(request.Repo); err != nil {
		return nil, err
	}
	return google_protobuf.EmptyInstance, nil
}

func (a *internalAPIServer) InspectRepo(ctx context.Context, request *pfs.InspectRepoRequest) (response *pfs.RepoInfo, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
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

func (a *internalAPIServer) ListRepo(ctx context.Context, request *pfs.ListRepoRequest) (response *pfs.RepoInfos, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
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

func (a *internalAPIServer) DeleteRepo(ctx context.Context, request *pfs.DeleteRepoRequest) (response *google_protobuf.Empty, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
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
	return google_protobuf.EmptyInstance, nil

}

func (a *internalAPIServer) StartCommit(ctx context.Context, request *pfs.StartCommitRequest) (response *google_protobuf.Empty, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shards, err := a.router.GetMasterShards(version)
	if err != nil {
		return nil, err
	}
	if err := a.driver.StartCommit(request.Parent, request.Commit, shards); err != nil {
		return nil, err
	}
	if err := a.pulseCommitWaiters(request.Commit, pfs.CommitType_COMMIT_TYPE_WRITE, shards); err != nil {
		return nil, err
	}
	return google_protobuf.EmptyInstance, nil
}

func (a *internalAPIServer) FinishCommit(ctx context.Context, request *pfs.FinishCommitRequest) (response *google_protobuf.Empty, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
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
	if err := a.pulseCommitWaiters(request.Commit, pfs.CommitType_COMMIT_TYPE_READ, shards); err != nil {
		return nil, err
	}
	if err := a.commitToReplicas(ctx, request.Commit); err != nil {
		return nil, err
	}
	return google_protobuf.EmptyInstance, nil
}

func (a *internalAPIServer) InspectCommit(ctx context.Context, request *pfs.InspectCommitRequest) (response *pfs.CommitInfo, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shards, err := a.router.GetMasterShards(version)
	if err != nil {
		return nil, err
	}
	return a.driver.InspectCommit(request.Commit, shards)
}

func (a *internalAPIServer) ListCommit(ctx context.Context, request *pfs.ListCommitRequest) (response *pfs.CommitInfos, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shards, err := a.router.GetMasterShards(version)
	if err != nil {
		return nil, err
	}
	commitInfos, err := a.filteredListCommits(request.Repo, request.From, request.CommitType, shards)
	if err != nil {
		return nil, err
	}
	if len(commitInfos) == 0 && request.Block {
		commitChan := make(chan *pfs.CommitInfo)
		if err := a.registerCommitWaiter(request, shards, commitChan); err != nil {
			return nil, err
		}
		for commitInfo := range commitChan {
			commitInfos = append(commitInfos, commitInfo)
		}
	}
	return &pfs.CommitInfos{
		CommitInfo: commitInfos,
	}, nil
}

func (a *internalAPIServer) DeleteCommit(ctx context.Context, request *pfs.DeleteCommitRequest) (response *google_protobuf.Empty, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
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
	return google_protobuf.EmptyInstance, nil
}

func (a *internalAPIServer) PutBlock(ctx context.Context, request *pfs.PutBlockRequest) (response *google_protobuf.Empty, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	block := a.sharder.GetBlock(request.Value)
	shard, err := a.getMasterShardForBlock(block, version)
	if err != nil {
		return nil, err
	}
	return google_protobuf.EmptyInstance, a.driver.PutBlock(request.File, block, shard, bytes.NewReader(request.Value))
}

func (a *internalAPIServer) GetBlock(request *pfs.GetBlockRequest, apiGetBlockServer pfs.InternalAPI_GetBlockServer) (retErr error) {
	defer func(start time.Time) { a.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	version, err := a.getVersion(apiGetBlockServer.Context())
	if err != nil {
		return err
	}
	shard, err := a.getShardForBlock(request.Block, version)
	if err != nil {
		return err
	}
	block, err := a.driver.GetBlock(request.Block, shard)
	if err != nil {
		return err
	}
	defer func() {
		if err := block.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	return protostream.WriteToStreamingBytesServer(block, apiGetBlockServer)
}

func (a *internalAPIServer) InspectBlock(ctx context.Context, request *pfs.InspectBlockRequest) (response *pfs.BlockInfo, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shard, err := a.getShardForBlock(request.Block, version)
	if err != nil {
		return nil, err
	}
	return a.driver.InspectBlock(request.Block, shard)
}

func (a *internalAPIServer) ListBlock(ctx context.Context, request *pfs.ListBlockRequest) (response *pfs.BlockInfos, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shards, err := a.router.GetMasterShards(version)
	if err != nil {
		return nil, err
	}
	if request.Shard == nil {
		request.Shard = &pfs.Shard{Number: 0, Modulus: 1}
	}
	sharder := route.NewSharder(request.Shard.Modulus, 0)
	var wg sync.WaitGroup
	var lock sync.Mutex
	var blockInfos []*pfs.BlockInfo
	var loopErr error
	for shard := range shards {
		wg.Add(1)
		go func(shard uint64) {
			defer wg.Done()
			subBlockInfos, err := a.driver.ListBlock(shard)
			lock.Lock()
			defer lock.Unlock()
			if err != nil {
				if loopErr == nil {
					loopErr = err
				}
				return
			}
			for _, blockInfo := range subBlockInfos {
				if request.Shard == nil || sharder.GetBlockShard(blockInfo.Block) == request.Shard.Number {
					blockInfos = append(blockInfos, blockInfo)
				}
			}
		}(shard)
	}
	wg.Wait()
	if loopErr != nil {
		return nil, loopErr
	}
	return &pfs.BlockInfos{
		BlockInfo: blockInfos,
	}, nil
}

func (a *internalAPIServer) PutFile(putFileServer pfs.InternalAPI_PutFileServer) (retErr error) {
	var request *pfs.PutFileRequest
	defer func(start time.Time) { a.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	version, err := a.getVersion(putFileServer.Context())
	if err != nil {
		return err
	}
	defer func() {
		if err := putFileServer.SendAndClose(google_protobuf.EmptyInstance); err != nil && retErr == nil {
			retErr = err
		}
	}()
	request, err = putFileServer.Recv()
	if err != nil {
		return err
	}
	if strings.HasPrefix(request.File.Path, "/") {
		// This is a subtle error case, the paths foo and /foo will hash to
		// different shards but will produce the same change once they get to
		// those shards due to how path.Join. This can go wrong in a number of
		// ways so we forbid leading slashes.
		return fmt.Errorf("pachyderm: leading slash in path: %s", request.File.Path)
	}
	if request.FileType == pfs.FileType_FILE_TYPE_DIR {
		if len(request.Value) > 0 {
			return fmt.Errorf("PutFileRequest shouldn't have type dir and a value")
		}
		shards, err := a.router.GetMasterShards(version)
		if err != nil {
			return err
		}
		if err := a.driver.MakeDirectory(request.File, shards); err != nil {
			return err
		}
		return nil
	}
	shard, err := a.getMasterShardForFile(request.File, version)
	if err != nil {
		return err
	}
	reader := putFileReader{
		server: putFileServer,
	}
	_, err = reader.buffer.Write(request.Value)
	if err != nil {
		return err
	}
	if err := a.driver.PutFile(request.File, shard, request.OffsetBytes, &reader); err != nil {
		return err
	}
	return nil
}

func (a *internalAPIServer) GetFile(request *pfs.GetFileRequest, apiGetFileServer pfs.InternalAPI_GetFileServer) (retErr error) {
	defer func(start time.Time) { a.Log(request, nil, retErr, time.Since(start)) }(time.Now())
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

func (a *internalAPIServer) InspectFile(ctx context.Context, request *pfs.InspectFileRequest) (response *pfs.FileInfo, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
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

func (a *internalAPIServer) ListFile(ctx context.Context, request *pfs.ListFileRequest) (response *pfs.FileInfos, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shards, err := a.router.GetMasterShards(version)
	if err != nil {
		return nil, err
	}
	if request.Shard == nil {
		request.Shard = &pfs.Shard{Number: 0, Modulus: 1}
	}
	sharder := route.NewSharder(request.Shard.Modulus, 0)
	var wg sync.WaitGroup
	var lock sync.Mutex
	var fileInfos []*pfs.FileInfo
	seenDirectories := make(map[string]bool)
	var loopErr error
	for shard := range shards {
		wg.Add(1)
		go func(shard uint64) {
			defer wg.Done()
			subFileInfos, err := a.driver.ListFile(request.File, shard)
			lock.Lock()
			defer lock.Unlock()
			if err != nil {
				if loopErr == nil {
					loopErr = err
				}
				return
			}
			for _, fileInfo := range subFileInfos {
				if fileInfo.FileType == pfs.FileType_FILE_TYPE_DIR {
					if seenDirectories[fileInfo.File.Path] {
						continue
					}
					seenDirectories[fileInfo.File.Path] = true
				}
				if sharder.GetShard(fileInfo.File) == request.Shard.Number {
					fileInfos = append(fileInfos, fileInfo)
				}
			}
		}(shard)
	}
	wg.Wait()
	if loopErr != nil {
		return nil, loopErr
	}
	return &pfs.FileInfos{
		FileInfo: fileInfos,
	}, nil
}

func (a *internalAPIServer) ListChange(ctx context.Context, request *pfs.ListChangeRequest) (response *pfs.Changes, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shards, err := a.router.GetMasterShards(version)
	if err != nil {
		return nil, err
	}
	if request.Shard == nil {
		request.Shard = &pfs.Shard{Number: 0, Modulus: 1}
	}
	sharder := route.NewSharder(request.Shard.Modulus, 0)
	var wg sync.WaitGroup
	var lock sync.Mutex
	var changes []*pfs.Change
	var loopErr error
	for shard := range shards {
		wg.Add(1)
		go func(shard uint64) {
			defer wg.Done()
			subChanges, err := a.driver.ListChange(request.File, request.From, shard)
			lock.Lock()
			defer lock.Unlock()
			if err != nil {
				if loopErr == nil {
					loopErr = err
				}
				return
			}
			for _, change := range subChanges {
				if sharder.GetShard(change.File) == request.Shard.Number {
					changes = append(changes, change)
				}
			}
		}(shard)
	}
	wg.Wait()
	if loopErr != nil {
		return nil, loopErr
	}
	return &pfs.Changes{
		Change: changes,
	}, nil
}

func (a *internalAPIServer) DeleteFile(ctx context.Context, request *pfs.DeleteFileRequest) (response *google_protobuf.Empty, err error) {
	defer func(start time.Time) { a.Log(request, response, err, time.Since(start)) }(time.Now())
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
	return google_protobuf.EmptyInstance, nil
}

func (a *internalAPIServer) PullDiff(request *pfs.PullDiffRequest, pullDiffServer pfs.ReplicaAPI_PullDiffServer) error {
	version, err := a.getVersion(pullDiffServer.Context())
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
	writer := protostream.NewStreamingBytesWriter(pullDiffServer)
	return a.driver.PullDiff(request.Commit, request.Shard, writer)
}

func (a *internalAPIServer) PushDiff(pushDiffServer pfs.ReplicaAPI_PushDiffServer) (retErr error) {
	version, err := a.getVersion(pushDiffServer.Context())
	if err != nil {
		return err
	}
	defer func() {
		if err := pushDiffServer.SendAndClose(google_protobuf.EmptyInstance); err != nil && retErr == nil {
			retErr = err
		}
	}()
	request, err := pushDiffServer.Recv()
	if err != nil {
		return err
	}
	ok, err := a.isLocalReplicaShard(request.Shard, version)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("pachyderm: illegal PushDiffRequest for unknown shard %d", request.Shard)
	}
	reader := &pushDiffReader{
		server: pushDiffServer,
	}
	_, err = reader.buffer.Write(request.Value)
	if err != nil {
		return err
	}
	return a.driver.PushDiff(request.Commit, request.Shard, reader)
}

func (a *internalAPIServer) AddShard(_shard uint64, version int64) error {
	if version == shard.InvalidVersion {
		return nil
	}
	ctx := versionToContext(version, context.Background())
	clientConn, err := a.router.GetMasterOrReplicaClientConn(_shard, version)
	if err != nil {
		return err
	}
	repoInfos, err := pfs.NewInternalAPIClient(clientConn).ListRepo(ctx, &pfs.ListRepoRequest{})
	if err != nil {
		return err
	}
	for _, repoInfo := range repoInfos.RepoInfo {
		if err := a.driver.CreateRepo(repoInfo.Repo); err != nil {
			return err
		}
		commitInfos, err := pfs.NewInternalAPIClient(clientConn).ListCommit(ctx, &pfs.ListCommitRequest{Repo: repoInfo.Repo})
		if err != nil {
			return err
		}
		for i := range commitInfos.CommitInfo {
			commit := commitInfos.CommitInfo[len(commitInfos.CommitInfo)-(i+1)].Commit
			commitInfo, err := a.driver.InspectCommit(commit, map[uint64]bool{_shard: true})
			if err != nil {
				return err
			}
			if commitInfo != nil {
				// we already have the commit so nothing to do
				continue
			}
			pullDiffRequest := &pfs.PullDiffRequest{
				Commit: commit,
				Shard:  _shard,
			}
			pullDiffClient, err := pfs.NewReplicaAPIClient(clientConn).PullDiff(ctx, pullDiffRequest)
			if err != nil {
				return err
			}
			diffReader := protostream.NewStreamingBytesReader(pullDiffClient)
			if err := a.driver.PushDiff(commit, _shard, diffReader); err != nil {
				return err
			}
		}
	}
	return nil
}

func (a *internalAPIServer) RemoveShard(shard uint64, version int64) error {
	return nil
}

func (a *internalAPIServer) LocalShards() (map[uint64]bool, error) {
	return nil, nil
}

func (a *internalAPIServer) getMasterShardForBlock(block *pfs.Block, version int64) (uint64, error) {
	shard := a.sharder.GetBlockShard(block)
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

func (a *internalAPIServer) getShardForBlock(block *pfs.Block, version int64) (uint64, error) {
	shard := a.sharder.GetBlockShard(block)
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

func (a *internalAPIServer) getMasterShardForFile(file *pfs.File, version int64) (uint64, error) {
	shard := a.sharder.GetShard(file)
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
	shard := a.sharder.GetShard(file)
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
	var loopErr error
	var wg sync.WaitGroup
	for shard := range shards {
		wg.Add(1)
		shard := shard
		go func() {
			defer wg.Done()
			clientConns, err := a.router.GetReplicaClientConns(shard, version)
			if err != nil && loopErr == nil {
				loopErr = err
				return
			}
			var writers []io.Writer
			for _, clientConn := range clientConns {
				pushDiffClient, err := pfs.NewReplicaAPIClient(clientConn).PushDiff(
					ctx,
				)
				if err != nil && loopErr == nil {
					loopErr = err
					return
				}
				request := pfs.PushDiffRequest{
					Commit: commit,
					Shard:  shard,
				}
				writers = append(writers,
					&pushDiffWriter{
						client:  pushDiffClient,
						request: &request,
					})
			}
			if err = a.driver.PullDiff(commit, shard, io.MultiWriter(writers...)); err != nil && loopErr == nil {
				loopErr = err
				return
			}
		}()
	}
	wg.Wait()
	return loopErr
}

type putFileReader struct {
	server pfs.InternalAPI_PutFileServer
	buffer bytes.Buffer
}

func (r *putFileReader) Read(p []byte) (int, error) {
	if r.buffer.Len() == 0 {
		request, err := r.server.Recv()
		if err != nil {
			return 0, err
		}
		_, err = r.buffer.Write(request.Value)
		if err != nil {
			return 0, err
		}
	}
	return r.buffer.Read(p)
}

type pushDiffReader struct {
	server pfs.ReplicaAPI_PushDiffServer
	buffer bytes.Buffer
}

func (r *pushDiffReader) Read(p []byte) (int, error) {
	if r.buffer.Len() == 0 {
		request, err := r.server.Recv()
		if err != nil {
			return 0, err
		}
		_, err = r.buffer.Write(request.Value)
		if err != nil {
			return 0, err
		}
	}
	return r.buffer.Read(p)
}

type pushDiffWriter struct {
	client  pfs.ReplicaAPI_PushDiffClient
	request *pfs.PushDiffRequest
}

func (w *pushDiffWriter) Write(value []byte) (int, error) {
	w.request.Value = value
	return len(value), w.client.Send(w.request)
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

// commitWait contains the values that describe which commits you're waiting for
type commitWait struct {
	//TODO don't use repo here, it's technically fine but using protobufs as map keys is fraught with peril
	repo       pfs.Repo
	commitType pfs.CommitType
}

func (a *internalAPIServer) registerCommitWaiter(request *pfs.ListCommitRequest, shards map[uint64]bool, outChan chan *pfs.CommitInfo) error {
	// This is a blocking request, which means we need to block until we
	// have at least one response.
	a.commitWaitersLock.Lock()
	defer a.commitWaitersLock.Unlock()
	protolog.Printf("registerCommitWaiter repo: %s", request.Repo.Name)
	// We need to redo the call to ListCommit because commits may have been
	// created between then and now.
	commitInfos, err := a.filteredListCommits(request.Repo, request.From, request.CommitType, shards)
	if err != nil {
		return err
	}
	if len(commitInfos) != 0 {
		go func() {
			for _, commitInfo := range commitInfos {
				outChan <- commitInfo
			}
			close(outChan)
		}()
	}
	key := commitWait{*request.Repo, request.CommitType}
	a.commitWaiters[key] =
		append(a.commitWaiters[key], outChan)
	return nil
}

func (a *internalAPIServer) pulseCommitWaiters(commit *pfs.Commit, commitType pfs.CommitType, shards map[uint64]bool) error {
	a.commitWaitersLock.Lock()
	defer a.commitWaitersLock.Unlock()
	protolog.Printf("pulseCommitWaiters repo: %s, commitID %d, commitType: %s", commit.Repo.Name, commit.Id, commitType.String())
	commitInfo, err := a.driver.InspectCommit(commit, shards)
	if err != nil {
		return err
	}
	key := commitWait{*commit.Repo, commitType}
	commitWaiters := a.commitWaiters[key]
	for _, commitWaiter := range commitWaiters {
		commitWaiter <- commitInfo
		close(commitWaiter)
	}
	delete(a.commitWaiters, key)

	if commitType != pfs.CommitType_COMMIT_TYPE_NONE {
		key.commitType = pfs.CommitType_COMMIT_TYPE_NONE
		commitWaiters := a.commitWaiters[key]
		for _, commitWaiter := range commitWaiters {
			commitWaiter <- commitInfo
			close(commitWaiter)
		}
		delete(a.commitWaiters, key)
	}
	return nil
}

func (a *internalAPIServer) filteredListCommits(repo *pfs.Repo, from *pfs.Commit, commitType pfs.CommitType, shards map[uint64]bool) ([]*pfs.CommitInfo, error) {
	commitInfos, err := a.driver.ListCommit(repo, from, shards)
	if err != nil {
		return nil, err
	}
	if commitType != pfs.CommitType_COMMIT_TYPE_NONE {
		var filtered []*pfs.CommitInfo
		for _, commitInfo := range commitInfos {
			if commitInfo.CommitType == commitType {
				filtered = append(filtered, commitInfo)
			}
		}
		commitInfos = filtered
	}
	return commitInfos, nil
}
