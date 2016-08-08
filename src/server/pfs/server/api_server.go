package server

import (
	"fmt"
	"io"
	"math/rand"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/shard"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	pfsserver "github.com/pachyderm/pachyderm/src/server/pfs"
	"github.com/pachyderm/pachyderm/src/server/pkg/metrics"

	"go.pedge.io/pb/go/google/protobuf"
	"go.pedge.io/proto/rpclog"
	"go.pedge.io/proto/stream"
	"go.pedge.io/proto/time"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type apiServer struct {
	protorpclog.Logger
	hasher *pfsserver.Hasher
	router shard.Router

	// versionLock protects the version field.
	// versionLock must be held BEFORE reading from version and UNTIL all
	// requests using version have returned
	versionLock sync.RWMutex
	version     int64

	versionChanLock sync.RWMutex
	versionChans    map[int64]chan struct{}
}

func newAPIServer(
	hasher *pfsserver.Hasher,
	router shard.Router,
) *apiServer {
	return &apiServer{
		Logger:          protorpclog.NewLogger("pachyderm.pfsserver.API"),
		hasher:          hasher,
		router:          router,
		versionLock:     sync.RWMutex{},
		version:         shard.InvalidVersion,
		versionChanLock: sync.RWMutex{},
		versionChans:    make(map[int64]chan struct{}),
	}
}

func (a *apiServer) CreateRepo(ctx context.Context, request *pfs.CreateRepoRequest) (response *google_protobuf.Empty, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	defer func() {
		if retErr == nil {
			metrics.AddRepos(1)
		}
	}()

	a.versionLock.RLock()
	defer a.versionLock.RUnlock()

	ctx, done := a.getVersionContext(ctx)
	defer close(done)

	clientConns, err := a.router.GetAllClientConns(a.version)
	if err != nil {
		return nil, err
	}

	request.Created = prototime.TimeToTimestamp(time.Now())
	for _, clientConn := range clientConns {
		if _, err := pfs.NewInternalAPIClient(clientConn).CreateRepo(ctx, request); err != nil {
			return nil, err
		}
	}
	return google_protobuf.EmptyInstance, nil
}

func (a *apiServer) InspectRepo(ctx context.Context, request *pfs.InspectRepoRequest) (response *pfs.RepoInfo, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	a.versionLock.RLock()
	defer a.versionLock.RUnlock()

	ctx, done := a.getVersionContext(ctx)
	defer close(done)

	clientConns, err := a.router.GetAllClientConns(a.version)
	if err != nil {
		return nil, err
	}

	var lock sync.Mutex
	var wg sync.WaitGroup
	var repoInfos []*pfs.RepoInfo
	errCh := make(chan error, 1)
	for _, clientConn := range clientConns {
		wg.Add(1)
		go func(clientConn *grpc.ClientConn) {
			defer wg.Done()
			repoInfo, err := pfs.NewInternalAPIClient(clientConn).InspectRepo(ctx, request)
			if err != nil {
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
			repoInfos = append(repoInfos, repoInfo)
		}(clientConn)
	}
	wg.Wait()
	select {
	case err := <-errCh:
		return nil, err
	default:
	}

	repoInfos = pfsserver.ReduceRepoInfos(repoInfos)

	if len(repoInfos) != 1 || repoInfos[0].Repo.Name != request.Repo.Name {
		return nil, fmt.Errorf("incorrect repo returned (this is likely a bug)")
	}

	return repoInfos[0], nil
}

func (a *apiServer) ListRepo(ctx context.Context, request *pfs.ListRepoRequest) (response *pfs.RepoInfos, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	a.versionLock.RLock()
	defer a.versionLock.RUnlock()

	ctx, done := a.getVersionContext(ctx)
	defer close(done)

	var wg sync.WaitGroup
	clientConns, err := a.router.GetAllClientConns(a.version)
	if err != nil {
		return nil, err
	}
	var lock sync.Mutex
	var repoInfos []*pfs.RepoInfo
	errCh := make(chan error, 1)
	for _, clientConn := range clientConns {
		clientConn := clientConn
		wg.Add(1)
		go func() {
			defer wg.Done()
			response, err := pfs.NewInternalAPIClient(clientConn).ListRepo(ctx, request)
			if err != nil {
				select {
				default:
				case errCh <- err:
				}
				return
			}
			lock.Lock()
			repoInfos = append(repoInfos, response.RepoInfo...)
			lock.Unlock()
		}()
	}
	wg.Wait()
	select {
	default:
	case err := <-errCh:
		return nil, err
	}
	return &pfs.RepoInfos{RepoInfo: pfsserver.ReduceRepoInfos(repoInfos)}, nil
}

func (a *apiServer) DeleteRepo(ctx context.Context, request *pfs.DeleteRepoRequest) (response *google_protobuf.Empty, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	a.versionLock.RLock()
	defer a.versionLock.RUnlock()

	ctx, done := a.getVersionContext(ctx)
	defer close(done)

	clientConns, err := a.router.GetAllClientConns(a.version)
	if err != nil {
		return nil, err
	}
	for _, clientConn := range clientConns {
		if _, err := pfs.NewInternalAPIClient(clientConn).DeleteRepo(ctx, request); err != nil {
			return nil, err
		}
	}
	return google_protobuf.EmptyInstance, nil
}

func (a *apiServer) StartCommit(ctx context.Context, request *pfs.StartCommitRequest) (response *pfs.Commit, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	a.versionLock.RLock()
	defer a.versionLock.RUnlock()

	ctx, done := a.getVersionContext(ctx)
	defer close(done)

	defer func() {
		if retErr == nil {
			metrics.AddCommits(1)
		}
	}()
	clientConns, err := a.router.GetAllClientConns(a.version)
	if err != nil {
		return nil, err
	}
	if request.ID != "" {
		return nil, fmt.Errorf("request.ID should be empty")
	}
	request.ID = uuid.NewWithoutDashes()
	request.Started = prototime.TimeToTimestamp(time.Now())
	for _, clientConn := range clientConns {
		if _, err := pfs.NewInternalAPIClient(clientConn).StartCommit(ctx, request); err != nil {
			return nil, err
		}
	}
	return client.NewCommit(request.Repo.Name, request.ID), nil
}

func (a *apiServer) FinishCommit(ctx context.Context, request *pfs.FinishCommitRequest) (response *google_protobuf.Empty, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	a.versionLock.RLock()
	defer a.versionLock.RUnlock()

	ctx, done := a.getVersionContext(ctx)
	defer close(done)

	clientConns, err := a.router.GetAllClientConns(a.version)
	if err != nil {
		return nil, err
	}
	request.Finished = prototime.TimeToTimestamp(time.Now())
	for _, clientConn := range clientConns {
		if _, err := pfs.NewInternalAPIClient(clientConn).FinishCommit(ctx, request); err != nil {
			return nil, err
		}
	}
	return google_protobuf.EmptyInstance, nil
}

func (a *apiServer) InspectCommit(ctx context.Context, request *pfs.InspectCommitRequest) (response *pfs.CommitInfo, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	a.versionLock.RLock()
	defer a.versionLock.RUnlock()

	ctx, done := a.getVersionContext(ctx)
	defer close(done)

	clientConns, err := a.router.GetAllClientConns(a.version)
	if err != nil {
		return nil, err
	}

	var lock sync.Mutex
	var wg sync.WaitGroup
	var commitInfos []*pfs.CommitInfo
	errCh := make(chan error, 1)
	for _, clientConn := range clientConns {
		wg.Add(1)
		go func(clientConn *grpc.ClientConn) {
			defer wg.Done()
			commitInfo, err := pfs.NewInternalAPIClient(clientConn).InspectCommit(ctx, request)
			if err != nil {
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
			commitInfos = append(commitInfos, commitInfo)
		}(clientConn)
	}
	wg.Wait()
	select {
	case err := <-errCh:
		return nil, err
	default:
	}

	commitInfos = pfsserver.ReduceCommitInfos(commitInfos)

	if len(commitInfos) != 1 ||
		(commitInfos[0].Commit.ID != request.Commit.ID &&
			commitInfos[0].Branch != request.Commit.ID) {
		return nil, fmt.Errorf("incorrect commit returned (this is likely a bug)")
	}

	return commitInfos[0], nil
}

func (a *apiServer) ListCommit(ctx context.Context, request *pfs.ListCommitRequest) (response *pfs.CommitInfos, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	a.versionLock.RLock()
	defer a.versionLock.RUnlock()

	ctx, done := a.getVersionContext(ctx)
	defer close(done)

	clientConns, err := a.router.GetAllClientConns(a.version)
	if err != nil {
		return nil, err
	}
	var wg sync.WaitGroup
	var lock sync.Mutex
	var commitInfos []*pfs.CommitInfo
	errCh := make(chan error, 1)
	for _, clientConn := range clientConns {
		wg.Add(1)
		go func(clientConn *grpc.ClientConn) {
			defer wg.Done()
			subCommitInfos, err := pfs.NewInternalAPIClient(clientConn).ListCommit(ctx, request)
			if err != nil {
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
			commitInfos = append(commitInfos, subCommitInfos.CommitInfo...)
		}(clientConn)
	}
	wg.Wait()
	select {
	case err := <-errCh:
		return nil, err
	default:
	}
	return &pfs.CommitInfos{CommitInfo: pfsserver.ReduceCommitInfos(commitInfos)}, nil
}

func (a *apiServer) ListBranch(ctx context.Context, request *pfs.ListBranchRequest) (response *pfs.CommitInfos, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	a.versionLock.RLock()
	defer a.versionLock.RUnlock()

	ctx, done := a.getVersionContext(ctx)
	defer close(done)

	clientConns, err := a.router.GetAllClientConns(a.version)
	if err != nil {
		return nil, err
	}
	var wg sync.WaitGroup
	var lock sync.Mutex
	var commitInfos []*pfs.CommitInfo
	errCh := make(chan error, 1)
	for _, clientConn := range clientConns {
		wg.Add(1)
		go func(clientConn *grpc.ClientConn) {
			defer wg.Done()
			subCommitInfos, err := pfs.NewInternalAPIClient(clientConn).ListBranch(ctx, request)
			if err != nil {
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
			commitInfos = append(commitInfos, subCommitInfos.CommitInfo...)
		}(clientConn)
	}
	wg.Wait()
	select {
	case err := <-errCh:
		return nil, err
	default:
	}
	return &pfs.CommitInfos{CommitInfo: pfsserver.ReduceCommitInfos(commitInfos)}, nil
}

func (a *apiServer) DeleteCommit(ctx context.Context, request *pfs.DeleteCommitRequest) (response *google_protobuf.Empty, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	a.versionLock.RLock()
	defer a.versionLock.RUnlock()

	ctx, done := a.getVersionContext(ctx)
	defer close(done)

	clientConns, err := a.router.GetAllClientConns(a.version)
	if err != nil {
		return nil, err
	}
	for _, clientConn := range clientConns {
		if _, err := pfs.NewInternalAPIClient(clientConn).DeleteCommit(ctx, request); err != nil {
			return nil, err
		}
	}
	return google_protobuf.EmptyInstance, nil
}

func (a *apiServer) FlushCommit(ctx context.Context, request *pfs.FlushCommitRequest) (response *pfs.CommitInfos, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	a.versionLock.RLock()
	defer a.versionLock.RUnlock()

	ctx, done := a.getVersionContext(ctx)
	defer close(done)

	clientConns, err := a.router.GetAllClientConns(a.version)
	if err != nil {
		return nil, err
	}
	var wg sync.WaitGroup
	var lock sync.Mutex
	var commitInfos []*pfs.CommitInfo
	errCh := make(chan error, 1)
	for _, clientConn := range clientConns {
		wg.Add(1)
		go func(clientConn *grpc.ClientConn) {
			defer wg.Done()
			subCommitInfos, err := pfs.NewInternalAPIClient(clientConn).FlushCommit(ctx, request)
			if err != nil {
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
			commitInfos = append(commitInfos, subCommitInfos.CommitInfo...)
		}(clientConn)
	}
	wg.Wait()
	select {
	case err := <-errCh:
		return nil, err
	default:
	}
	return &pfs.CommitInfos{CommitInfo: pfsserver.ReduceCommitInfos(commitInfos)}, nil
}

func (a *apiServer) PutFile(putFileServer pfs.API_PutFileServer) (retErr error) {
	var request *pfs.PutFileRequest
	var err error
	defer func(start time.Time) {
		if request != nil {
			request.Value = nil // we set the value to nil so as not to spam logs
		}
		a.Log(request, google_protobuf.EmptyInstance, retErr, time.Since(start))
	}(time.Now())
	defer drainFileServer(putFileServer)
	a.versionLock.RLock()
	defer a.versionLock.RUnlock()

	ctx, done := a.getVersionContext(putFileServer.Context())
	defer close(done)

	defer func() {
		if err := putFileServer.SendAndClose(google_protobuf.EmptyInstance); err != nil && retErr == nil {
			retErr = err
		}
	}()
	request, err = putFileServer.Recv()
	if err != nil && err != io.EOF {
		return err
	}
	if err == io.EOF {
		// tolerate people calling and immediately hanging up
		return nil
	}
	if strings.HasPrefix(request.File.Path, "/") {
		// This is a subtle error case, the paths foo and /foo will hash to
		// different shards but will produce the same change once they get to
		// those shards due to how path.Join. This can go wrong in a number of
		// ways so we forbid leading slashes.
		return fmt.Errorf("pachyderm: leading slash in path: %s", request.File.Path)
	}

	var wg sync.WaitGroup
	errCh := make(chan error, 1)
	requests := []*pfs.PutFileRequest{request}
	// For a file like foo/bar/buzz, we want to send two requests to create two
	// directories foo/ and foo/bar/
	dirs := dirs(request.File.Path)
	for _, dir := range dirs {
		dirReq := &pfs.PutFileRequest{
			File: &pfs.File{
				Path:   dir,
				Commit: request.File.Commit,
			},
			FileType: pfs.FileType_FILE_TYPE_DIR,
		}
		requests = append(requests, dirReq)
	}

	sendReq := func(request *pfs.PutFileRequest) (retErr error) {
		clientConn, err := a.getClientConnForFile(request.File, a.version)
		if err != nil {
			return err
		}
		putFileClient, err := pfs.NewInternalAPIClient(clientConn).PutFile(ctx)
		if err != nil {
			return err
		}
		if request.FileType == pfs.FileType_FILE_TYPE_DIR {
			if len(request.Value) > 0 {
				return fmt.Errorf("PutFileRequest shouldn't have type dir and a value")
			}
			if err := putFileClient.Send(request); err != nil {
				return err
			}
			if _, err := putFileClient.CloseAndRecv(); err != nil {
				return err
			}
			return nil
		}
		defer func() {
			if _, err := putFileClient.CloseAndRecv(); err != nil && retErr == nil {
				retErr = err
			}
		}()
		if err := putFileClient.Send(request); err != nil {
			return err
		}
		for {
			request, err := putFileServer.Recv()
			if err != nil {
				if err == io.EOF {
					break
				}
				return err
			}
			if err := putFileClient.Send(request); err != nil {
				return err
			}
		}
		return nil
	}

	for _, req := range requests {
		wg.Add(1)
		req := req
		go func() {
			defer wg.Done()
			if err := sendReq(req); err != nil {
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
		return err
	default:
		return nil
	}
}

func dirs(path string) []string {
	var ancestors []string
	for {
		path = filepath.Dir(path)
		if path == "." {
			break
		}
		ancestors = append(ancestors, path)
	}
	return ancestors
}

func (a *apiServer) GetFile(request *pfs.GetFileRequest, apiGetFileServer pfs.API_GetFileServer) (retErr error) {
	defer func(start time.Time) { a.Log(request, google_protobuf.EmptyInstance, retErr, time.Since(start)) }(time.Now())
	a.versionLock.RLock()
	defer a.versionLock.RUnlock()

	ctx, done := a.getVersionContext(apiGetFileServer.Context())
	defer close(done)

	clientConn, err := a.getClientConnForFile(request.File, a.version)
	if err != nil {
		return err
	}

	fileGetClient, err := pfs.NewInternalAPIClient(clientConn).GetFile(ctx, request)
	if err != nil {
		return err
	}

	return protostream.RelayFromStreamingBytesClient(fileGetClient, apiGetFileServer)
}

func (a *apiServer) InspectFile(ctx context.Context, request *pfs.InspectFileRequest) (response *pfs.FileInfo, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	a.versionLock.RLock()
	defer a.versionLock.RUnlock()

	ctx, done := a.getVersionContext(ctx)
	defer close(done)

	clientConn, err := a.getClientConnForFile(request.File, a.version)
	if err != nil {
		return nil, err
	}

	fileInfo, err := pfs.NewInternalAPIClient(clientConn).InspectFile(ctx, request)
	if err != nil {
		return nil, err
	}

	if fileInfo.FileType == pfs.FileType_FILE_TYPE_DIR {
		// If it's a directory, chances are that its children are scattered around
		// the cluster, so we can't get the complete list of children without
		// doing a multicast.  So instead of returning a partial list of children,
		// we'd rather return no children at all.  To get the complete list of
		// children, the caller is expected to use ListFile
		fileInfo.Children = nil
	}

	return fileInfo, nil
}

func (a *apiServer) ListFile(ctx context.Context, request *pfs.ListFileRequest) (response *pfs.FileInfos, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	a.versionLock.RLock()
	defer a.versionLock.RUnlock()

	ctx, done := a.getVersionContext(ctx)
	defer close(done)

	clientConns, err := a.router.GetAllClientConns(a.version)
	if err != nil {
		return nil, err
	}
	var wg sync.WaitGroup
	var lock sync.Mutex
	var fileInfos []*pfs.FileInfo
	seenDirectories := make(map[string]bool)
	errCh := make(chan error, 1)
	for _, clientConn := range clientConns {
		wg.Add(1)
		go func(clientConn *grpc.ClientConn) {
			defer wg.Done()
			subFileInfos, err := pfs.NewInternalAPIClient(clientConn).ListFile(ctx, request)
			lock.Lock()
			defer lock.Unlock()
			if err != nil {
				select {
				case errCh <- err:
					// error reported
				default:
					// not the first error
				}
				return
			}
			for _, fileInfo := range subFileInfos.FileInfo {
				if fileInfo.FileType == pfs.FileType_FILE_TYPE_DIR {
					if seenDirectories[fileInfo.File.Path] {
						continue
					}
					seenDirectories[fileInfo.File.Path] = true
				}
				fileInfos = append(fileInfos, fileInfo)
			}
		}(clientConn)
	}
	wg.Wait()
	select {
	case err := <-errCh:
		return nil, err
	default:
	}
	return &pfs.FileInfos{
		FileInfo: pfsserver.ReduceFileInfos(fileInfos),
	}, nil
}

func (a *apiServer) DeleteFile(ctx context.Context, request *pfs.DeleteFileRequest) (response *google_protobuf.Empty, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	a.versionLock.RLock()
	defer a.versionLock.RUnlock()

	ctx, done := a.getVersionContext(ctx)
	defer close(done)

	fileInfo, err := a.InspectFile(ctx, &pfs.InspectFileRequest{
		File:   request.File,
		Unsafe: request.Unsafe,
		Handle: request.Handle,
	})
	if err != nil {
		return nil, err
	}

	if fileInfo.FileType == pfs.FileType_FILE_TYPE_REGULAR {
		clientConn, err := a.getClientConnForFile(request.File, a.version)
		if err != nil {
			return nil, err
		}
		return pfs.NewInternalAPIClient(clientConn).DeleteFile(ctx, request)
	}

	// If we are deleting a directory, we multicast it to the entire cluster
	clientConns, err := a.router.GetAllClientConns(a.version)
	if err != nil {
		return nil, err
	}
	var wg sync.WaitGroup
	errCh := make(chan error, 1)
	for _, clientConn := range clientConns {
		wg.Add(1)
		clientConn := clientConn
		go func() {
			defer wg.Done()
			_, err := pfs.NewInternalAPIClient(clientConn).DeleteFile(ctx, request)
			if err != nil {
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

func (a *apiServer) DeleteAll(ctx context.Context, request *google_protobuf.Empty) (response *google_protobuf.Empty, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	a.versionLock.RLock()
	defer a.versionLock.RUnlock()

	ctx, done := a.getVersionContext(ctx)
	defer close(done)

	clientConns, err := a.router.GetAllClientConns(a.version)
	if err != nil {
		return nil, err
	}
	for _, clientConn := range clientConns {
		if _, err := pfs.NewInternalAPIClient(clientConn).DeleteAll(ctx, request); err != nil {
			return nil, err
		}
	}

	return google_protobuf.EmptyInstance, nil
}

func (a *apiServer) Version(version int64) error {
	func() {
		a.versionLock.RLock()
		defer a.versionLock.RUnlock()

		a.versionChanLock.Lock()
		defer a.versionChanLock.Unlock()

		// Note that a channel cannot be closed be twice.  So here
		// we are relying on the invariant that Version() only gets called
		// once per version.
		if vChan, ok := a.versionChans[a.version]; ok {
			close(vChan)
		}
		a.versionChans[version] = make(chan struct{})
	}()

	return func() error {
		a.versionLock.Lock()
		defer a.versionLock.Unlock()

		a.version = version
		return a.router.CloseClientConns()
	}()
}

func (a *apiServer) getClientConn(version int64) (*grpc.ClientConn, error) {
	shards, err := a.router.GetShards(a.version)
	if err != nil {
		return nil, err
	}
	for shard := range shards {
		return a.router.GetClientConn(shard, version)
	}
	return a.router.GetClientConn(uint64(rand.Int())%a.hasher.FileModulus, version)
}

func (a *apiServer) getClientConnForFile(file *pfs.File, version int64) (*grpc.ClientConn, error) {
	return a.router.GetClientConn(a.hasher.HashFile(file), version)
}

// getVersionContext returns a context that gets cancelled when one of the following
// events occur:
// 1. the given context gets cancelled
// 2. a new version comes in
// The caller is expected to close the returned channel when they are done with
// using this context.
// The caller is expected to be holding a read lock on a.versionLock
func (a *apiServer) getVersionContext(ctx context.Context) (context.Context, chan struct{}) {
	a.versionChanLock.RLock()
	defer a.versionChanLock.RUnlock()
	versionChan := a.versionChans[a.version]

	ctx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	go func() {
		select {
		case <-versionChan:
			cancel()
		case <-done:
		}
	}()
	ctx = metadata.NewContext(
		ctx,
		metadata.Pairs("version", fmt.Sprint(a.version)),
	)
	return ctx, done
}
