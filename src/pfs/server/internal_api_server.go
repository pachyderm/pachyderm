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
	commitWaiters     []*commitWait
	commitWaitersLock sync.Mutex
}

func newInternalAPIServer(
	sharder route.Sharder,
	router route.Router,
	driver drive.Driver,
) *internalAPIServer {
	return &internalAPIServer{
		Logger:            protorpclog.NewLogger("pachyderm.pfs.InternalAPI"),
		sharder:           sharder,
		router:            router,
		driver:            driver,
		commitWaiters:     nil,
		commitWaitersLock: sync.Mutex{},
	}
}

func (a *internalAPIServer) CreateRepo(ctx context.Context, request *pfs.CreateRepoRequest) (response *google_protobuf.Empty, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shards, err := a.router.GetAllShards(version)
	if err != nil {
		return nil, err
	}
	if err := a.driver.CreateRepo(request.Repo, request.Created, shards); err != nil {
		return nil, err
	}
	return google_protobuf.EmptyInstance, nil
}

func (a *internalAPIServer) InspectRepo(ctx context.Context, request *pfs.InspectRepoRequest) (response *pfs.RepoInfo, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shards, err := a.router.GetAllShards(version)
	if err != nil {
		return nil, err
	}
	return a.driver.InspectRepo(request.Repo, shards)
}

func (a *internalAPIServer) ListRepo(ctx context.Context, request *pfs.ListRepoRequest) (response *pfs.RepoInfos, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shards, err := a.router.GetAllShards(version)
	if err != nil {
		return nil, err
	}
	repoInfos, err := a.driver.ListRepo(shards)
	return &pfs.RepoInfos{RepoInfo: repoInfos}, err
}

func (a *internalAPIServer) DeleteRepo(ctx context.Context, request *pfs.DeleteRepoRequest) (response *google_protobuf.Empty, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
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

func (a *internalAPIServer) StartCommit(ctx context.Context, request *pfs.StartCommitRequest) (response *google_protobuf.Empty, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shards, err := a.router.GetMasterShards(version)
	if err != nil {
		return nil, err
	}
	if err := a.driver.StartCommit(request.Parent, request.Commit, request.Started, shards); err != nil {
		return nil, err
	}
	if err := a.pulseCommitWaiters(request.Commit, pfs.CommitType_COMMIT_TYPE_WRITE, shards); err != nil {
		return nil, err
	}
	return google_protobuf.EmptyInstance, nil
}

func (a *internalAPIServer) FinishCommit(ctx context.Context, request *pfs.FinishCommitRequest) (response *google_protobuf.Empty, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shards, err := a.router.GetMasterShards(version)
	if err != nil {
		return nil, err
	}
	if err := a.driver.FinishCommit(request.Commit, request.Finished, shards); err != nil {
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

func (a *internalAPIServer) InspectCommit(ctx context.Context, request *pfs.InspectCommitRequest) (response *pfs.CommitInfo, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
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

func (a *internalAPIServer) ListCommit(ctx context.Context, request *pfs.ListCommitRequest) (response *pfs.CommitInfos, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shards, err := a.router.GetMasterShards(version)
	if err != nil {
		return nil, err
	}
	commitInfos, err := a.filteredListCommits(request.Repo, request.FromCommit, request.CommitType, shards)
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
		CommitInfo: pfs.ReduceCommitInfos(commitInfos),
	}, nil
}

func (a *internalAPIServer) DeleteCommit(ctx context.Context, request *pfs.DeleteCommitRequest) (response *google_protobuf.Empty, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
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
	file, err := a.driver.GetFile(request.File, request.OffsetBytes, request.SizeBytes, shard)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	return protostream.WriteToStreamingBytesServer(file, apiGetFileServer)
}

func (a *internalAPIServer) InspectFile(ctx context.Context, request *pfs.InspectFileRequest) (response *pfs.FileInfo, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
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

func (a *internalAPIServer) ListFile(ctx context.Context, request *pfs.ListFileRequest) (response *pfs.FileInfos, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shards, err := a.router.GetMasterShards(version)
	if err != nil {
		return nil, err
	}
	if request.Shard == nil {
		request.Shard = &pfs.Shard{FileNumber: 0, FileModulus: 1}
	}
	sharder := route.NewSharder(request.Shard.FileModulus, request.Shard.BlockModulus)
	var wg sync.WaitGroup
	var lock sync.Mutex
	var fileInfos []*pfs.FileInfo
	var loopErr error
	for shard := range shards {
		shard := shard
		wg.Add(1)
		go func() {
			defer wg.Done()
			subFileInfos, err := a.driver.ListFile(request.File, shard)
			lock.Lock()
			defer lock.Unlock()
			if err != nil && err != pfs.ErrFileNotFound {
				if loopErr == nil {
					loopErr = err
				}
				return
			}
			for _, fileInfo := range subFileInfos {
				if sharder.GetShard(fileInfo.File) == request.Shard.FileNumber ||
					fileInfo.FileType == pfs.FileType_FILE_TYPE_DIR {
					fileInfos = append(fileInfos, fileInfo)
				}
			}
		}()
	}
	wg.Wait()
	if loopErr != nil {
		return nil, loopErr
	}
	return &pfs.FileInfos{
		FileInfo: pfs.ReduceFileInfos(fileInfos),
	}, nil
}

func (a *internalAPIServer) ListChange(ctx context.Context, request *pfs.ListChangeRequest) (response *pfs.Changes, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shards, err := a.router.GetMasterShards(version)
	if err != nil {
		return nil, err
	}
	sharder := route.NewSharder(request.Shard.FileModulus, request.Shard.BlockModulus)
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
				if sharder.GetShard(change.File) == request.Shard.FileNumber {
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

func (a *internalAPIServer) DeleteFile(ctx context.Context, request *pfs.DeleteFileRequest) (response *google_protobuf.Empty, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
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
		if err := a.driver.CreateRepo(repoInfo.Repo, nil, map[uint64]bool{_shard: true}); err != nil {
			return err
		}
		commitInfos, err := pfs.NewInternalAPIClient(clientConn).ListCommit(ctx, &pfs.ListCommitRequest{Repo: []*pfs.Repo{repoInfo.Repo}})
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
	repos          []*pfs.Repo
	commitType     pfs.CommitType
	commitInfoChan chan *pfs.CommitInfo
}

func newCommitWait(repos []*pfs.Repo, commitType pfs.CommitType, commitInfoChan chan *pfs.CommitInfo) *commitWait {
	return &commitWait{
		repos:          repos,
		commitType:     commitType,
		commitInfoChan: commitInfoChan,
	}
}

func (a *internalAPIServer) registerCommitWaiter(request *pfs.ListCommitRequest, shards map[uint64]bool, outChan chan *pfs.CommitInfo) error {
	// This is a blocking request, which means we need to block until we
	// have at least one response.
	a.commitWaitersLock.Lock()
	defer a.commitWaitersLock.Unlock()
	// We need to redo the call to ListCommit because commits may have been
	// created between then and now.
	commitInfos, err := a.filteredListCommits(request.Repo, request.FromCommit, request.CommitType, shards)
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
	a.commitWaiters = append(a.commitWaiters, newCommitWait(request.Repo, request.CommitType, outChan))
	return nil
}

func (a *internalAPIServer) pulseCommitWaiters(commit *pfs.Commit, commitType pfs.CommitType, shards map[uint64]bool) error {
	a.commitWaitersLock.Lock()
	defer a.commitWaitersLock.Unlock()
	commitInfo, err := a.driver.InspectCommit(commit, shards)
	if err != nil {
		return err
	}
	var unpulsedWaiters []*commitWait
WaitersLoop:
	for _, commitWaiter := range a.commitWaiters {
		if commitWaiter.commitType == pfs.CommitType_COMMIT_TYPE_NONE || commitType == commitWaiter.commitType {
			for _, repo := range commitWaiter.repos {
				if repo.Name == commit.Repo.Name {
					commitWaiter.commitInfoChan <- commitInfo
					close(commitWaiter.commitInfoChan)
					continue WaitersLoop
				}
			}
		}
		unpulsedWaiters = append(unpulsedWaiters, commitWaiter)
	}
	a.commitWaiters = unpulsedWaiters
	return nil
}

func (a *internalAPIServer) filteredListCommits(repos []*pfs.Repo, fromCommit []*pfs.Commit, commitType pfs.CommitType, shards map[uint64]bool) ([]*pfs.CommitInfo, error) {
	commitInfos, err := a.driver.ListCommit(repos, fromCommit, shards)
	if err != nil {
		return nil, err
	}
	var filtered []*pfs.CommitInfo
	for _, commitInfo := range commitInfos {
		if commitType != pfs.CommitType_COMMIT_TYPE_NONE && commitInfo.CommitType != commitType {
			continue
		}
		filtered = append(filtered, commitInfo)
	}
	return filtered, nil
}
