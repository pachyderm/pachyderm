package server

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.pedge.io/pb/go/google/protobuf"
	"go.pedge.io/proto/rpclog"
	"go.pedge.io/proto/stream"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"

	pfsclient "github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/shard"
	pfsserver "github.com/pachyderm/pachyderm/src/server/pfs"
	"github.com/pachyderm/pachyderm/src/server/pfs/drive"
)

var (
	grpcErrorf = grpc.Errorf // needed to get passed govet
)

type internalAPIServer struct {
	protorpclog.Logger
	hasher            *pfsserver.Hasher
	router            shard.Router
	driver            drive.Driver
	commitWaiters     []*commitWait
	commitWaitersLock sync.Mutex
}

func newInternalAPIServer(
	hasher *pfsserver.Hasher,
	router shard.Router,
	driver drive.Driver,
) *internalAPIServer {
	return &internalAPIServer{
		Logger:            protorpclog.NewLogger("pachyderm.pfsserver.InternalAPI"),
		hasher:            hasher,
		router:            router,
		driver:            driver,
		commitWaiters:     nil,
		commitWaitersLock: sync.Mutex{},
	}
}

func (a *internalAPIServer) CreateRepo(ctx context.Context, request *pfsclient.CreateRepoRequest) (response *google_protobuf.Empty, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shards, err := a.router.GetShards(version)
	if err != nil {
		return nil, err
	}
	if err := a.driver.CreateRepo(request.Repo, request.Created, shards); err != nil {
		return nil, err
	}
	return google_protobuf.EmptyInstance, nil
}

func (a *internalAPIServer) InspectRepo(ctx context.Context, request *pfsclient.InspectRepoRequest) (response *pfsclient.RepoInfo, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shards, err := a.router.GetShards(version)
	if err != nil {
		return nil, err
	}
	return a.driver.InspectRepo(request.Repo, shards)
}

func (a *internalAPIServer) ListRepo(ctx context.Context, request *pfsclient.ListRepoRequest) (response *pfsclient.RepoInfos, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shards, err := a.router.GetShards(version)
	if err != nil {
		return nil, err
	}
	repoInfos, err := a.driver.ListRepo(shards)
	return &pfsclient.RepoInfos{RepoInfo: repoInfos}, err
}

func (a *internalAPIServer) DeleteRepo(ctx context.Context, request *pfsclient.DeleteRepoRequest) (response *google_protobuf.Empty, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shards, err := a.router.GetShards(version)
	if err != nil {
		return nil, err
	}
	if err := a.driver.DeleteRepo(request.Repo, shards); err != nil && err != pfsserver.ErrRepoNotFound {
		return nil, err
	}
	return google_protobuf.EmptyInstance, nil
}

func (a *internalAPIServer) StartCommit(ctx context.Context, request *pfsclient.StartCommitRequest) (response *google_protobuf.Empty, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shards, err := a.router.GetShards(version)
	if err != nil {
		return nil, err
	}
	if err := a.driver.StartCommit(request.Repo, request.ID, request.ParentID,
		request.Branch, request.Started, shards); err != nil {
		return nil, err
	}
	if err := a.pulseCommitWaiters(pfsclient.NewCommit(request.Repo.Name, request.ID), pfsclient.CommitType_COMMIT_TYPE_WRITE, shards); err != nil {
		return nil, err
	}
	return google_protobuf.EmptyInstance, nil
}

func (a *internalAPIServer) FinishCommit(ctx context.Context, request *pfsclient.FinishCommitRequest) (response *google_protobuf.Empty, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shards, err := a.router.GetShards(version)
	if err != nil {
		return nil, err
	}
	if err := a.driver.FinishCommit(request.Commit, request.Finished, request.Cancel, shards); err != nil {
		return nil, err
	}
	if err := a.pulseCommitWaiters(request.Commit, pfsclient.CommitType_COMMIT_TYPE_READ, shards); err != nil {
		return nil, err
	}
	return google_protobuf.EmptyInstance, nil
}

func (a *internalAPIServer) InspectCommit(ctx context.Context, request *pfsclient.InspectCommitRequest) (response *pfsclient.CommitInfo, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shards, err := a.router.GetShards(version)
	if err != nil {
		return nil, err
	}
	return a.driver.InspectCommit(request.Commit, shards)
}

func (a *internalAPIServer) ListCommit(ctx context.Context, request *pfsclient.ListCommitRequest) (response *pfsclient.CommitInfos, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shards, err := a.router.GetShards(version)
	if err != nil {
		return nil, err
	}
	commitInfos, err := a.filteredListCommits(request.Repo, request.FromCommit, request.CommitType, request.All, shards)
	if err != nil {
		return nil, err
	}
	if len(commitInfos) == 0 && request.Block {
		commitChan := make(chan *pfsclient.CommitInfo)
		if err := a.registerCommitWaiter(request, shards, commitChan); err != nil {
			return nil, err
		}
		for commitInfo := range commitChan {
			commitInfos = append(commitInfos, commitInfo)
		}
	}
	return &pfsclient.CommitInfos{
		CommitInfo: commitInfos,
	}, nil
}

func (a *internalAPIServer) ListBranch(ctx context.Context, request *pfsclient.ListBranchRequest) (response *pfsclient.CommitInfos, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shards, err := a.router.GetShards(version)
	if err != nil {
		return nil, err
	}
	commitInfos, err := a.driver.ListBranch(request.Repo, shards)
	if err != nil {
		return nil, err
	}
	return &pfsclient.CommitInfos{
		CommitInfo: commitInfos,
	}, nil
}

func (a *internalAPIServer) DeleteCommit(ctx context.Context, request *pfsclient.DeleteCommitRequest) (response *google_protobuf.Empty, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shards, err := a.router.GetShards(version)
	if err != nil {
		return nil, err
	}
	if err := a.driver.DeleteCommit(request.Commit, shards); err != nil {
		return nil, err
	}
	// TODO push delete to replicas
	return google_protobuf.EmptyInstance, nil
}

func (a *internalAPIServer) PutFile(putFileServer pfsclient.InternalAPI_PutFileServer) (retErr error) {
	var request *pfsclient.PutFileRequest
	defer func(start time.Time) {
		if request != nil {
			request.Value = nil // we set the value to nil so as not to spam logs
		}
		a.Log(request, nil, retErr, time.Since(start))
	}(time.Now())
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
	shard, err := a.getMasterShardForFile(request.File, version)
	if err != nil {
		return err
	}
	if request.FileType == pfsclient.FileType_FILE_TYPE_DIR {
		if len(request.Value) > 0 {
			return fmt.Errorf("PutFileRequest shouldn't have type dir and a value")
		}
		if err := a.driver.MakeDirectory(request.File, shard); err != nil {
			return err
		}
	} else {
		reader := putFileReader{
			server: putFileServer,
		}
		_, err = reader.buffer.Write(request.Value)
		if err != nil {
			return err
		}
		if err := a.driver.PutFile(request.File, shard, &reader); err != nil {
			return err
		}
	}
	return nil
}

func (a *internalAPIServer) GetFile(request *pfsclient.GetFileRequest, apiGetFileServer pfsclient.InternalAPI_GetFileServer) (retErr error) {
	defer func(start time.Time) { a.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	version, err := a.getVersion(apiGetFileServer.Context())
	if err != nil {
		return err
	}
	shard, err := a.getShardForFile(request.File, version)
	if err != nil {
		return err
	}
	file, err := a.driver.GetFile(request.File, request.Shard, request.OffsetBytes, request.SizeBytes, request.FromCommit, shard)
	if err != nil {
		// TODO this should be done more consistently throughout
		if err == pfsserver.ErrFileNotFound {
			return grpcErrorf(codes.NotFound, "%v", err)
		}
		return err
	}
	defer func() {
		if err := file.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	return protostream.WriteToStreamingBytesServer(file, apiGetFileServer)
}

func (a *internalAPIServer) InspectFile(ctx context.Context, request *pfsclient.InspectFileRequest) (response *pfsclient.FileInfo, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shard, err := a.getShardForFile(request.File, version)
	if err != nil {
		return nil, err
	}
	return a.driver.InspectFile(request.File, request.Shard, request.FromCommit, shard)
}

func (a *internalAPIServer) ListFile(ctx context.Context, request *pfsclient.ListFileRequest) (response *pfsclient.FileInfos, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shards, err := a.router.GetShards(version)
	if err != nil {
		return nil, err
	}
	var wg sync.WaitGroup
	var lock sync.Mutex
	var fileInfos []*pfsclient.FileInfo
	errCh := make(chan error, 1)
	for shard := range shards {
		shard := shard
		wg.Add(1)
		go func() {
			defer wg.Done()
			subFileInfos, err := a.driver.ListFile(request.File, request.Shard, request.FromCommit, shard, request.Recurse)
			if err != nil && err != pfsserver.ErrFileNotFound {
				select {
				case errCh <- err:
					// error reported
				default:
					// not the first error
				}
				return
			}
			lock.Lock()
			defer lock.Unlock()
			fileInfos = append(fileInfos, subFileInfos...)
		}()
	}
	wg.Wait()
	select {
	case err := <-errCh:
		return nil, err
	default:
	}
	return &pfsclient.FileInfos{
		FileInfo: fileInfos,
	}, nil
}

func (a *internalAPIServer) DeleteFile(ctx context.Context, request *pfsclient.DeleteFileRequest) (response *google_protobuf.Empty, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shards, err := a.router.GetShards(version)
	if err != nil {
		return nil, err
	}
	var wg sync.WaitGroup
	errCh := make(chan error, 1)
	for shard := range shards {
		shard := shard
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := a.driver.DeleteFile(request.File, shard)
			if err != nil && err != pfsserver.ErrFileNotFound {
				select {
				case errCh <- err:
					// error reported
				default:
					// not the first error
				}
				return
			}
		}()
	}
	wg.Wait()
	select {
	case err := <-errCh:
		return nil, err
	default:
	}
	return google_protobuf.EmptyInstance, nil
}

func (a *internalAPIServer) AddShard(shard uint64) error {
	return a.driver.AddShard(shard)
}

func (a *internalAPIServer) DeleteShard(shard uint64) error {
	return a.driver.DeleteShard(shard)
}

func (a *internalAPIServer) getMasterShardForFile(file *pfsclient.File, version int64) (uint64, error) {
	shard := a.hasher.HashFile(file)
	shards, err := a.router.GetShards(version)
	if err != nil {
		return 0, err
	}
	_, ok := shards[shard]
	if !ok {
		return 0, fmt.Errorf("pachyderm: shard %d not found locally", shard)
	}
	return shard, nil
}

func (a *internalAPIServer) getShardForFile(file *pfsclient.File, version int64) (uint64, error) {
	shard := a.hasher.HashFile(file)
	shards, err := a.router.GetShards(version)
	if err != nil {
		return 0, err
	}
	_, ok := shards[shard]
	if !ok {
		return 0, fmt.Errorf("pachyderm: shard %d not found locally", shard)
	}
	return shard, nil
}

type putFileReader struct {
	server pfsclient.InternalAPI_PutFileServer
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
	repos          []*pfsclient.Repo
	commitType     pfsclient.CommitType
	all            bool
	commitInfoChan chan *pfsclient.CommitInfo
}

func newCommitWait(repos []*pfsclient.Repo, commitType pfsclient.CommitType, all bool, commitInfoChan chan *pfsclient.CommitInfo) *commitWait {
	return &commitWait{
		repos:          repos,
		commitType:     commitType,
		all:            all,
		commitInfoChan: commitInfoChan,
	}
}

func (a *internalAPIServer) registerCommitWaiter(request *pfsclient.ListCommitRequest, shards map[uint64]bool, outChan chan *pfsclient.CommitInfo) error {
	// This is a blocking request, which means we need to block until we
	// have at least one response.
	a.commitWaitersLock.Lock()
	defer a.commitWaitersLock.Unlock()
	// We need to redo the call to ListCommit because commits may have been
	// created between then and now.
	commitInfos, err := a.filteredListCommits(request.Repo, request.FromCommit, request.CommitType, request.All, shards)
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
	a.commitWaiters = append(a.commitWaiters, newCommitWait(request.Repo, request.CommitType, request.All, outChan))
	return nil
}

func (a *internalAPIServer) pulseCommitWaiters(commit *pfsclient.Commit, commitType pfsclient.CommitType, shards map[uint64]bool) error {
	a.commitWaitersLock.Lock()
	defer a.commitWaitersLock.Unlock()
	commitInfo, err := a.driver.InspectCommit(commit, shards)
	if err != nil {
		return err
	}
	var unpulsedWaiters []*commitWait
WaitersLoop:
	for _, commitWaiter := range a.commitWaiters {
		if (commitWaiter.commitType == pfsclient.CommitType_COMMIT_TYPE_NONE || commitType == commitWaiter.commitType) && (commitWaiter.all || !commitInfo.Cancelled) {
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

func (a *internalAPIServer) filteredListCommits(repos []*pfsclient.Repo, fromCommit []*pfsclient.Commit, commitType pfsclient.CommitType, all bool, shards map[uint64]bool) ([]*pfsclient.CommitInfo, error) {
	commitInfos, err := a.driver.ListCommit(repos, fromCommit, all, shards)
	if err != nil && err != pfsserver.ErrRepoNotFound {
		return nil, err
	}
	var filtered []*pfsclient.CommitInfo
	for _, commitInfo := range commitInfos {
		if commitType != pfsclient.CommitType_COMMIT_TYPE_NONE && commitInfo.CommitType != commitType {
			continue
		}
		filtered = append(filtered, commitInfo)
	}
	return filtered, nil
}
