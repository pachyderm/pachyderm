package server

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.pedge.io/pb/go/google/protobuf"
	"go.pedge.io/proto/rpclog"
	"go.pedge.io/proto/stream"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
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
	commitWaiters     map[*commitWait]bool
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
		commitWaiters:     make(map[*commitWait]bool),
		commitWaitersLock: sync.Mutex{},
	}
}

func (a *internalAPIServer) CreateRepo(ctx context.Context, request *pfs.CreateRepoRequest) (response *google_protobuf.Empty, retErr error) {
	func() {
		a.Log(request, nil, nil, 0)
	}()

	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shards, err := a.router.GetShards(version)
	if err != nil {
		return nil, err
	}
	if err := a.driver.CreateRepo(request.Repo, request.Created, request.Provenance, shards); err != nil {
		return nil, err
	}
	return google_protobuf.EmptyInstance, nil
}

func (a *internalAPIServer) InspectRepo(ctx context.Context, request *pfs.InspectRepoRequest) (response *pfs.RepoInfo, retErr error) {
	func() {
		a.Log(request, nil, nil, 0)
	}()

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

func (a *internalAPIServer) ListRepo(ctx context.Context, request *pfs.ListRepoRequest) (response *pfs.RepoInfos, retErr error) {
	func() {
		a.Log(request, nil, nil, 0)
	}()

	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shards, err := a.router.GetShards(version)
	if err != nil {
		return nil, err
	}
	repoInfos, err := a.driver.ListRepo(request.Provenance, shards)
	return &pfs.RepoInfos{RepoInfo: repoInfos}, err
}

func (a *internalAPIServer) DeleteRepo(ctx context.Context, request *pfs.DeleteRepoRequest) (response *google_protobuf.Empty, retErr error) {
	func() {
		a.Log(request, nil, nil, 0)
	}()

	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shards, err := a.router.GetShards(version)
	if err != nil {
		return nil, err
	}

	err = a.driver.DeleteRepo(request.Repo, shards, request.Force)
	_, ok := err.(*pfsserver.ErrRepoNotFound)

	if err != nil && !ok {
		return nil, err
	}
	return google_protobuf.EmptyInstance, nil
}

func (a *internalAPIServer) StartCommit(ctx context.Context, request *pfs.StartCommitRequest) (response *google_protobuf.Empty, retErr error) {
	func() {
		a.Log(request, nil, nil, 0)
	}()

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
		request.Branch, request.Started, request.Provenance, shards); err != nil {
		return nil, err
	}
	if err := a.pulseCommitWaiters(client.NewCommit(request.Repo.Name, request.ID), pfs.CommitType_COMMIT_TYPE_WRITE, shards); err != nil {
		return nil, err
	}
	return google_protobuf.EmptyInstance, nil
}

func (a *internalAPIServer) FinishCommit(ctx context.Context, request *pfs.FinishCommitRequest) (response *google_protobuf.Empty, retErr error) {
	func() {
		a.Log(request, nil, nil, 0)
	}()

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
	if err := a.pulseCommitWaiters(request.Commit, pfs.CommitType_COMMIT_TYPE_READ, shards); err != nil {
		return nil, err
	}
	return google_protobuf.EmptyInstance, nil
}

func (a *internalAPIServer) ArchiveCommit(ctx context.Context, request *pfs.ArchiveCommitRequest) (response *google_protobuf.Empty, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shards, err := a.router.GetShards(version)
	if err != nil {
		return nil, err
	}
	if err := a.driver.ArchiveCommit(request.Commit, shards); err != nil {
		return nil, err
	}
	return google_protobuf.EmptyInstance, nil
}

func (a *internalAPIServer) InspectCommit(ctx context.Context, request *pfs.InspectCommitRequest) (response *pfs.CommitInfo, retErr error) {
	func() {
		a.Log(request, nil, nil, 0)
	}()

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

func (a *internalAPIServer) ListCommit(ctx context.Context, request *pfs.ListCommitRequest) (response *pfs.CommitInfos, retErr error) {
	func() {
		a.Log(request, nil, nil, 0)
	}()

	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shards, err := a.router.GetShards(version)
	if err != nil {
		return nil, err
	}
	commitInfos, err := a.driver.ListCommit(request.Repo, request.CommitType,
		request.FromCommit, request.Provenance, request.Status, shards)
	_, ok := err.(*pfsserver.ErrRepoNotFound)
	if err != nil && (!request.Block || !ok) {
		return nil, err
	}
	if len(commitInfos) == 0 && request.Block {
		commitWait, err := a.newCommitWait(request, shards)
		if err != nil {
			return nil, err
		}
		select {
		case <-ctx.Done():
			a.cancelCommitWait(commitWait)
			return nil, ctx.Err()
		case commitInfos = <-commitWait.commitInfoChan:
		}
	}
	return &pfs.CommitInfos{
		CommitInfo: commitInfos,
	}, nil
}

func (a *internalAPIServer) ListBranch(ctx context.Context, request *pfs.ListBranchRequest) (response *pfs.CommitInfos, retErr error) {
	func() {
		a.Log(request, nil, nil, 0)
	}()

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
	return &pfs.CommitInfos{
		CommitInfo: commitInfos,
	}, nil
}

func (a *internalAPIServer) DeleteCommit(ctx context.Context, request *pfs.DeleteCommitRequest) (response *google_protobuf.Empty, retErr error) {
	func() {
		a.Log(request, nil, nil, 0)
	}()

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

func (a *internalAPIServer) FlushCommit(ctx context.Context, request *pfs.FlushCommitRequest) (response *pfs.CommitInfos, retErr error) {
	func() {
		a.Log(request, nil, nil, 0)
	}()

	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shards, err := a.router.GetShards(version)
	if err != nil {
		return nil, err
	}
	var provenanceRepos []*pfs.Repo
	for _, commit := range request.Commit {
		_, err := a.driver.InspectCommit(commit, shards)
		if err != nil {
			return nil, err
		}
		provenanceRepos = append(provenanceRepos, commit.Repo)
	}
	repoInfos, err := a.driver.ListRepo(provenanceRepos, shards)
	if err != nil {
		return nil, err
	}
	// repoWhiteList is the set of repos we're interested in, empty means we're
	// interested in all repos
	repoWhiteList := make(map[string]bool)
	for _, toRepo := range request.ToRepo {
		repoWhiteList[toRepo.Name] = true
		repoInfo, err := a.driver.InspectRepo(toRepo, shards)
		if err != nil {
			return nil, err
		}
		for _, repo := range repoInfo.Provenance {
			repoWhiteList[repo.Name] = true
		}
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	var wg sync.WaitGroup
	var result []*pfs.CommitInfo
	var lock sync.Mutex
	errCh := make(chan error, 1)
	for _, repoInfo := range repoInfos {
		if len(repoWhiteList) > 0 && !repoWhiteList[repoInfo.Repo.Name] {
			continue
		}
		repoInfo := repoInfo
		wg.Add(1)
		go func() {
			defer wg.Done()
			commitInfos, err := a.ListCommit(ctx, &pfs.ListCommitRequest{
				Repo:       []*pfs.Repo{repoInfo.Repo},
				CommitType: pfs.CommitType_COMMIT_TYPE_READ,
				Provenance: request.Commit,
				Block:      true,
				Status:     pfs.CommitStatus_CANCELLED,
			})
			if err != nil {
				select {
				case errCh <- err:
				default:
				}
				return
			}
			lock.Lock()
			defer lock.Unlock()
			for _, commitInfo := range commitInfos.CommitInfo {
				if commitInfo.Archived {
					// we don't consider archived commits for provenance
					continue
				}
				if commitInfo.Cancelled {
					// one of the commits was cancelled so downstream commits might not show up
					// cancel everything
					cancel()
					select {
					case errCh <- fmt.Errorf("commit %s/%s was cancelled", commitInfo.Commit.Repo.Name, commitInfo.Commit.ID):
					default:
					}
					return
				}
				result = append(result, commitInfo)
			}
		}()
	}
	wg.Wait()
	select {
	case err := <-errCh:
		return nil, err
	default:
	}
	return &pfs.CommitInfos{CommitInfo: result}, nil
}

func (a *internalAPIServer) PutFile(putFileServer pfs.InternalAPI_PutFileServer) (retErr error) {
	var request *pfs.PutFileRequest
	func() {
		a.Log(request, nil, nil, 0)
	}()

	defer func(start time.Time) {
		if request != nil {
			request.Value = nil // we set the value to nil so as not to spam logs
		}
		a.Log(request, nil, retErr, time.Since(start))
	}(time.Now())
	defer drainFileServer(putFileServer)
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
	if request.FileType == pfs.FileType_FILE_TYPE_DIR {
		if len(request.Value) > 0 {
			return fmt.Errorf("PutFileRequest shouldn't have type dir and a value")
		}
		if err := a.driver.MakeDirectory(request.File, shard); err != nil {
			return err
		}
	} else {
		var r io.Reader
		var delimiter pfs.Delimiter
		if request.Url != "" {
			resp, err := http.Get(request.Url)
			if err != nil {
				return err
			}
			defer func() {
				if err := resp.Body.Close(); err != nil && retErr == nil {
					retErr = err
				}
			}()
			r = resp.Body
			switch resp.Header.Get("Content-Type") {
			case "application/json":
				delimiter = pfs.Delimiter_JSON
			case "application/text":
				delimiter = pfs.Delimiter_LINE
			default:
				delimiter = pfs.Delimiter_NONE
			}
		} else {
			reader := putFileReader{
				server: putFileServer,
			}
			_, err = reader.buffer.Write(request.Value)
			if err != nil {
				return err
			}
			r = &reader
			delimiter = request.Delimiter
		}
		if err := a.driver.PutFile(request.File, request.Handle, delimiter, shard, r); err != nil {
			return err
		}
	}
	return nil
}

func (a *internalAPIServer) GetFile(request *pfs.GetFileRequest, apiGetFileServer pfs.InternalAPI_GetFileServer) (retErr error) {
	func() {
		a.Log(request, nil, nil, 0)
	}()

	defer func(start time.Time) { a.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	version, err := a.getVersion(apiGetFileServer.Context())
	if err != nil {
		return err
	}
	shard, err := a.getShardForFile(request.File, version)
	if err != nil {
		return err
	}
	file, err := a.driver.GetFile(request.File, request.Shard, request.OffsetBytes, request.SizeBytes,
		request.DiffMethod, shard, request.Unsafe, request.Handle)
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
	func() {
		a.Log(request, nil, nil, 0)
	}()

	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shard, err := a.getShardForFile(request.File, version)
	if err != nil {
		return nil, err
	}
	return a.driver.InspectFile(request.File, request.Shard, request.DiffMethod, shard, request.Unsafe, request.Handle)
}

func (a *internalAPIServer) ListFile(ctx context.Context, request *pfs.ListFileRequest) (response *pfs.FileInfos, retErr error) {
	func() {
		a.Log(request, nil, nil, 0)
	}()

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
	var fileInfos []*pfs.FileInfo
	errCh := make(chan error, 1)
	for shard := range shards {
		shard := shard
		wg.Add(1)
		go func() {
			defer wg.Done()
			subFileInfos, err := a.driver.ListFile(request.File, request.Shard,
				request.DiffMethod, shard, request.Recurse, request.Unsafe, request.Handle)
			_, ok := err.(*pfsserver.ErrFileNotFound)
			if err != nil && !ok {
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
	return &pfs.FileInfos{
		FileInfo: fileInfos,
	}, nil
}

func (a *internalAPIServer) DeleteFile(ctx context.Context, request *pfs.DeleteFileRequest) (response *google_protobuf.Empty, retErr error) {
	func() {
		a.Log(request, nil, nil, 0)
	}()

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
			err := a.driver.DeleteFile(request.File, shard, request.Unsafe, request.Handle)
			// We are ignoring ErrFileNotFound because the file being
			// deleted can be a directory, and directory is scattered
			// across many DiffInfos across many shards.  Yet not all
			// shards necessarily contain DiffInfos that contain the
			// directory, so some of them will report FileNotFound
			_, ok := err.(*pfsserver.ErrFileNotFound)
			if err != nil && !ok {
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

func (a *internalAPIServer) DeleteAll(ctx context.Context, request *google_protobuf.Empty) (response *google_protobuf.Empty, retErr error) {
	func() {
		a.Log(request, nil, nil, 0)
	}()

	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	version, err := a.getVersion(ctx)
	if err != nil {
		return nil, err
	}
	shards, err := a.router.GetShards(version)
	if err != nil {
		return nil, err
	}
	if err := a.driver.DeleteAll(shards); err != nil {
		return nil, err
	}
	return google_protobuf.EmptyInstance, nil
}

func (a *internalAPIServer) AddShard(shard uint64) error {
	return a.driver.AddShard(shard)
}

func (a *internalAPIServer) DeleteShard(shard uint64) error {
	return a.driver.DeleteShard(shard)
}

func (a *internalAPIServer) getMasterShardForFile(file *pfs.File, version int64) (uint64, error) {
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

func (a *internalAPIServer) getShardForFile(file *pfs.File, version int64) (uint64, error) {
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
	server pfs.InternalAPI_PutFileServer
	buffer bytes.Buffer
}

func (r *putFileReader) Read(p []byte) (int, error) {
	if r.buffer.Len() == 0 {
		request, err := r.server.Recv()
		if err != nil {
			return 0, err
		}
		//buffer.Write cannot error
		r.buffer.Write(request.Value)
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
	repos          []*pfs.Repo
	provenance     []*pfs.Commit
	commitType     pfs.CommitType
	status         pfs.CommitStatus
	commitInfoChan chan []*pfs.CommitInfo
}

func (a *internalAPIServer) newCommitWait(request *pfs.ListCommitRequest, shards map[uint64]bool) (*commitWait, error) {
	result := &commitWait{
		repos:          request.Repo,
		provenance:     request.Provenance,
		commitType:     request.CommitType,
		status:         request.Status,
		commitInfoChan: make(chan []*pfs.CommitInfo, 1),
	}
	a.commitWaitersLock.Lock()
	defer a.commitWaitersLock.Unlock()
	// We need to redo the call to ListCommit because commits may have been
	// created between then and now.
	commitInfos, err := a.driver.ListCommit(request.Repo, request.CommitType,
		request.FromCommit, request.Provenance, request.Status, shards)
	_, ok := err.(*pfsserver.ErrRepoNotFound)
	if err != nil && !ok {
		return nil, err
	}
	if len(commitInfos) != 0 {
		result.commitInfoChan <- commitInfos
	} else {
		a.commitWaiters[result] = true
	}
	return result, nil
}

func (a *internalAPIServer) cancelCommitWait(cw *commitWait) {
	a.commitWaitersLock.Lock()
	defer a.commitWaitersLock.Unlock()
	delete(a.commitWaiters, cw)
}

func (a *internalAPIServer) pulseCommitWaiters(commit *pfs.Commit, commitType pfs.CommitType, shards map[uint64]bool) error {
	a.commitWaitersLock.Lock()
	defer a.commitWaitersLock.Unlock()
	commitInfo, err := a.driver.InspectCommit(commit, shards)
	if err != nil {
		return err
	}
	var pulsedWaiters []*commitWait
WaitersLoop:
	for commitWaiter := range a.commitWaiters {
		if (commitWaiter.commitType == pfs.CommitType_COMMIT_TYPE_NONE || commitType == commitWaiter.commitType) &&
			(commitWaiter.status == pfs.CommitStatus_ALL ||
				(commitInfo.Cancelled && commitWaiter.status == pfs.CommitStatus_CANCELLED) ||
				(commitInfo.Archived && commitWaiter.status == pfs.CommitStatus_ARCHIVED) ||
				(!commitInfo.Archived && !commitInfo.Cancelled)) &&
			drive.MatchProvenance(commitWaiter.provenance, commitInfo.Provenance) {
			for _, repo := range commitWaiter.repos {
				if repo.Name == commit.Repo.Name {
					commitWaiter.commitInfoChan <- []*pfs.CommitInfo{commitInfo}
					pulsedWaiters = append(pulsedWaiters, commitWaiter)
					continue WaitersLoop
				}
			}
		}
	}
	for _, commitWaiter := range pulsedWaiters {
		delete(a.commitWaiters, commitWaiter)
	}
	return nil
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
