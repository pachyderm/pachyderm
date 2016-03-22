package server

import (
	"fmt"
	"io"
	"math/rand"
	"strings"
	"sync"
	"time"

	pfsclient "github.com/pachyderm/pachyderm/src/client/pfs"
	pfsserver "github.com/pachyderm/pachyderm/src/server/pfs"
	"github.com/pachyderm/pachyderm/src/server/pkg/metrics"
	"github.com/pachyderm/pachyderm/src/server/pkg/shard"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
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
	hasher  *pfsserver.Hasher
	router  shard.Router
	version int64
	// versionLock protects the version field.
	// versionLock must be held BEFORE reading from version and UNTIL all
	// requests using version have returned
	versionLock sync.RWMutex
}

func newAPIServer(
	hasher *pfsserver.Hasher,
	router shard.Router,
) *apiServer {
	return &apiServer{
		protorpclog.NewLogger("pachyderm.pfsserver.API"),
		hasher,
		router,
		shard.InvalidVersion,
		sync.RWMutex{},
	}
}

func (a *apiServer) CreateRepo(ctx context.Context, request *pfsclient.CreateRepoRequest) (response *google_protobuf.Empty, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	a.versionLock.RLock()
	defer a.versionLock.RUnlock()
	ctx = versionToContext(a.version, ctx)
	defer func() {
		if retErr == nil {
			metrics.AddRepos(1)
		}
	}()
	if strings.Contains(request.Repo.Name, "/") {
		return nil, fmt.Errorf("repo names cannot contain /")
	}
	clientConns, err := a.router.GetAllClientConns(a.version)
	if err != nil {
		return nil, err
	}
	request.Created = prototime.TimeToTimestamp(time.Now())
	for _, clientConn := range clientConns {
		if _, err := pfsclient.NewInternalAPIClient(clientConn).CreateRepo(ctx, request); err != nil {
			return nil, err
		}
	}
	return google_protobuf.EmptyInstance, nil
}

func (a *apiServer) InspectRepo(ctx context.Context, request *pfsclient.InspectRepoRequest) (response *pfsserver.RepoInfo, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	a.versionLock.RLock()
	defer a.versionLock.RUnlock()
	ctx = versionToContext(a.version, ctx)
	clientConn, err := a.getClientConn(a.version)
	if err != nil {
		return nil, err
	}
	return pfsclient.NewInternalAPIClient(clientConn).InspectRepo(ctx, request)
}

func (a *apiServer) ListRepo(ctx context.Context, request *pfsclient.ListRepoRequest) (response *pfsserver.RepoInfos, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	a.versionLock.RLock()
	defer a.versionLock.RUnlock()
	ctx = versionToContext(a.version, ctx)
	clientConn, err := a.getClientConn(a.version)
	if err != nil {
		return nil, err
	}
	return pfsclient.NewInternalAPIClient(clientConn).ListRepo(ctx, request)
}

func (a *apiServer) DeleteRepo(ctx context.Context, request *pfsclient.DeleteRepoRequest) (response *google_protobuf.Empty, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	a.versionLock.RLock()
	defer a.versionLock.RUnlock()
	ctx = versionToContext(a.version, ctx)
	clientConns, err := a.router.GetAllClientConns(a.version)
	if err != nil {
		return nil, err
	}
	for _, clientConn := range clientConns {
		if _, err := pfsclient.NewInternalAPIClient(clientConn).DeleteRepo(ctx, request); err != nil {
			return nil, err
		}
	}
	return google_protobuf.EmptyInstance, nil

}

func (a *apiServer) StartCommit(ctx context.Context, request *pfsclient.StartCommitRequest) (response *pfsserver.Commit, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	a.versionLock.RLock()
	defer a.versionLock.RUnlock()
	ctx = versionToContext(a.version, ctx)
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
		if _, err := pfsclient.NewInternalAPIClient(clientConn).StartCommit(ctx, request); err != nil {
			return nil, err
		}
	}
	return pfsclient.NewCommit(request.Repo.Name, request.ID), nil
}

func (a *apiServer) FinishCommit(ctx context.Context, request *pfsclient.FinishCommitRequest) (response *google_protobuf.Empty, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	a.versionLock.RLock()
	defer a.versionLock.RUnlock()
	ctx = versionToContext(a.version, ctx)
	clientConns, err := a.router.GetAllClientConns(a.version)
	if err != nil {
		return nil, err
	}
	request.Finished = prototime.TimeToTimestamp(time.Now())
	for _, clientConn := range clientConns {
		if _, err := pfsclient.NewInternalAPIClient(clientConn).FinishCommit(ctx, request); err != nil {
			return nil, err
		}
	}
	return google_protobuf.EmptyInstance, nil
}

func (a *apiServer) InspectCommit(ctx context.Context, request *pfsclient.InspectCommitRequest) (response *pfsserver.CommitInfo, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	a.versionLock.RLock()
	defer a.versionLock.RUnlock()
	ctx = versionToContext(a.version, ctx)
	clientConn, err := a.getClientConn(a.version)
	if err != nil {
		return nil, err
	}
	return pfsclient.NewInternalAPIClient(clientConn).InspectCommit(ctx, request)
}

func (a *apiServer) ListCommit(ctx context.Context, request *pfsclient.ListCommitRequest) (response *pfsserver.CommitInfos, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	a.versionLock.RLock()
	defer a.versionLock.RUnlock()
	ctx = versionToContext(a.version, ctx)
	clientConns, err := a.router.GetAllClientConns(a.version)
	if err != nil {
		return nil, err
	}
	var wg sync.WaitGroup
	var lock sync.Mutex
	var commitInfos []*pfsserver.CommitInfo
	errCh := make(chan error, 1)
	for _, clientConn := range clientConns {
		wg.Add(1)
		go func(clientConn *grpc.ClientConn) {
			defer wg.Done()
			subCommitInfos, err := pfsclient.NewInternalAPIClient(clientConn).ListCommit(ctx, request)
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
	return &pfsserver.CommitInfos{CommitInfo: pfsserver.ReduceCommitInfos(commitInfos)}, nil
}

func (a *apiServer) ListBranch(ctx context.Context, request *pfsclient.ListBranchRequest) (response *pfsserver.CommitInfos, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	a.versionLock.RLock()
	defer a.versionLock.RUnlock()
	ctx = versionToContext(a.version, ctx)
	clientConns, err := a.router.GetAllClientConns(a.version)
	if err != nil {
		return nil, err
	}
	var wg sync.WaitGroup
	var lock sync.Mutex
	var commitInfos []*pfsserver.CommitInfo
	errCh := make(chan error, 1)
	for _, clientConn := range clientConns {
		wg.Add(1)
		go func(clientConn *grpc.ClientConn) {
			defer wg.Done()
			subCommitInfos, err := pfsclient.NewInternalAPIClient(clientConn).ListBranch(ctx, request)
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
	return &pfsserver.CommitInfos{CommitInfo: pfsserver.ReduceCommitInfos(commitInfos)}, nil
}

func (a *apiServer) DeleteCommit(ctx context.Context, request *pfsclient.DeleteCommitRequest) (response *google_protobuf.Empty, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	a.versionLock.RLock()
	defer a.versionLock.RUnlock()
	ctx = versionToContext(a.version, ctx)
	clientConns, err := a.router.GetAllClientConns(a.version)
	if err != nil {
		return nil, err
	}
	for _, clientConn := range clientConns {
		if _, err := pfsclient.NewInternalAPIClient(clientConn).DeleteCommit(ctx, request); err != nil {
			return nil, err
		}
	}
	return google_protobuf.EmptyInstance, nil
}

func (a *apiServer) PutFile(putFileServer pfsclient.API_PutFileServer) (retErr error) {
	var request *pfsclient.PutFileRequest
	var err error
	defer func(start time.Time) { a.Log(request, google_protobuf.EmptyInstance, retErr, time.Since(start)) }(time.Now())
	ctx := versionToContext(a.version, putFileServer.Context())
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
	if request.FileType == pfsserver.FileType_FILE_TYPE_DIR {
		if len(request.Value) > 0 {
			return fmt.Errorf("PutFileRequest shouldn't have type dir and a value")
		}
		clientConns, err := a.router.GetAllClientConns(a.version)
		if err != nil {
			return err
		}
		for _, clientConn := range clientConns {
			putFileClient, err := pfsclient.NewInternalAPIClient(clientConn).PutFile(ctx)
			if err != nil {
				return err
			}
			if err := putFileClient.Send(request); err != nil {
				return err
			}
			if _, err := putFileClient.CloseAndRecv(); err != nil {
				return err
			}
		}
		return nil
	}
	clientConn, err := a.getClientConnForFile(request.File, a.version)
	if err != nil {
		return err
	}
	putFileClient, err := pfsclient.NewInternalAPIClient(clientConn).PutFile(ctx)
	if err != nil {
		return err
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

func (a *apiServer) GetFile(request *pfsclient.GetFileRequest, apiGetFileServer pfsclient.API_GetFileServer) (retErr error) {
	defer func(start time.Time) { a.Log(request, google_protobuf.EmptyInstance, retErr, time.Since(start)) }(time.Now())
	ctx := versionToContext(a.version, apiGetFileServer.Context())
	clientConn, err := a.getClientConnForFile(request.File, a.version)
	if err != nil {
		return err
	}
	fileGetClient, err := pfsclient.NewInternalAPIClient(clientConn).GetFile(ctx, request)
	if err != nil {
		return err
	}
	return protostream.RelayFromStreamingBytesClient(fileGetClient, apiGetFileServer)
}

func (a *apiServer) InspectFile(ctx context.Context, request *pfsclient.InspectFileRequest) (response *pfsserver.FileInfo, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	a.versionLock.RLock()
	defer a.versionLock.RUnlock()
	ctx = versionToContext(a.version, ctx)
	clientConn, err := a.getClientConnForFile(request.File, a.version)
	if err != nil {
		return nil, err
	}
	return pfsclient.NewInternalAPIClient(clientConn).InspectFile(ctx, request)
}

func (a *apiServer) ListFile(ctx context.Context, request *pfsclient.ListFileRequest) (response *pfsserver.FileInfos, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	a.versionLock.RLock()
	defer a.versionLock.RUnlock()
	ctx = versionToContext(a.version, ctx)
	clientConns, err := a.router.GetAllClientConns(a.version)
	if err != nil {
		return nil, err
	}
	var wg sync.WaitGroup
	var lock sync.Mutex
	var fileInfos []*pfsserver.FileInfo
	seenDirectories := make(map[string]bool)
	errCh := make(chan error, 1)
	for _, clientConn := range clientConns {
		wg.Add(1)
		go func(clientConn *grpc.ClientConn) {
			defer wg.Done()
			subFileInfos, err := pfsclient.NewInternalAPIClient(clientConn).ListFile(ctx, request)
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
				if fileInfo.FileType == pfsserver.FileType_FILE_TYPE_DIR {
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
	return &pfsserver.FileInfos{
		FileInfo: fileInfos,
	}, nil
}

func (a *apiServer) DeleteFile(ctx context.Context, request *pfsclient.DeleteFileRequest) (response *google_protobuf.Empty, retErr error) {
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	a.versionLock.RLock()
	defer a.versionLock.RUnlock()
	ctx = versionToContext(a.version, ctx)
	clientConn, err := a.getClientConnForFile(request.File, a.version)
	if err != nil {
		return nil, err
	}
	return pfsclient.NewInternalAPIClient(clientConn).DeleteFile(ctx, request)
}

func (a *apiServer) Version(version int64) error {
	a.versionLock.Lock()
	defer a.versionLock.Unlock()
	a.version = version
	return nil
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

func (a *apiServer) getClientConnForFile(file *pfsserver.File, version int64) (*grpc.ClientConn, error) {
	return a.router.GetClientConn(a.hasher.HashFile(file), version)
}

func versionToContext(version int64, ctx context.Context) context.Context {
	return metadata.NewContext(
		ctx,
		metadata.Pairs("version", fmt.Sprint(version)),
	)
}
