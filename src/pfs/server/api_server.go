package server

import (
	"fmt"
	"io"
	"math/rand"
	"strings"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"golang.org/x/net/context"

	"github.com/satori/go.uuid"
	"go.pachyderm.com/pachyderm/src/pfs"
	"go.pachyderm.com/pachyderm/src/pfs/route"
	"go.pedge.io/google-protobuf"
	"go.pedge.io/proto/stream"
	"go.pedge.io/proto/time"
	"go.pedge.io/protolog"
)

type apiServer struct {
	sharder route.Sharder
	router  route.Router
	version int64
	// versionLock protects the version field.
	// versionLock must be held BEFORE reading from version and UNTIL all
	// requests using version have returned
	versionLock sync.RWMutex
}

func newAPIServer(
	sharder route.Sharder,
	router route.Router,
) *apiServer {
	return &apiServer{
		sharder,
		router,
		route.InvalidVersion,
		sync.RWMutex{},
	}
}

func (a *apiServer) CreateRepo(ctx context.Context, request *pfs.CreateRepoRequest) (*google_protobuf.Empty, error) {
	a.versionLock.RLock()
	defer a.versionLock.RUnlock()
	ctx = versionToContext(a.version, ctx)
	clientConns, err := a.router.GetAllClientConns(a.version)
	if err != nil {
		return nil, err
	}
	for _, clientConn := range clientConns {
		if _, err := pfs.NewInternalAPIClient(clientConn).CreateRepo(ctx, request); err != nil {
			return nil, err
		}
	}
	// Create the initial commit
	if _, err = a.StartCommit(ctx, &pfs.StartCommitRequest{
		Parent: nil,
		Commit: &pfs.Commit{
			Repo: request.Repo,
			Id:   InitialCommitID,
		},
	}); err != nil {
		return nil, err
	}
	if _, err = a.FinishCommit(ctx, &pfs.FinishCommitRequest{
		Commit: &pfs.Commit{
			Repo: request.Repo,
			Id:   InitialCommitID,
		},
	}); err != nil {
		return nil, err
	}
	return emptyInstance, nil
}

func (a *apiServer) InspectRepo(ctx context.Context, request *pfs.InspectRepoRequest) (*pfs.RepoInfo, error) {
	a.versionLock.RLock()
	defer a.versionLock.RUnlock()
	ctx = versionToContext(a.version, ctx)
	clientConn, err := a.getClientConn(a.version)
	if err != nil {
		return nil, err
	}
	return pfs.NewInternalAPIClient(clientConn).InspectRepo(ctx, request)
}

func (a *apiServer) ListRepo(ctx context.Context, request *pfs.ListRepoRequest) (*pfs.RepoInfos, error) {
	a.versionLock.RLock()
	defer a.versionLock.RUnlock()
	ctx = versionToContext(a.version, ctx)
	clientConn, err := a.getClientConn(a.version)
	if err != nil {
		return nil, err
	}
	return pfs.NewInternalAPIClient(clientConn).ListRepo(ctx, request)
}

func (a *apiServer) DeleteRepo(ctx context.Context, request *pfs.DeleteRepoRequest) (*google_protobuf.Empty, error) {
	a.versionLock.RLock()
	defer a.versionLock.RUnlock()
	ctx = versionToContext(a.version, ctx)
	clientConns, err := a.router.GetAllClientConns(a.version)
	if err != nil {
		return nil, err
	}
	for _, clientConn := range clientConns {
		if _, err := pfs.NewInternalAPIClient(clientConn).DeleteRepo(ctx, request); err != nil {
			return nil, err
		}
	}
	return emptyInstance, nil

}

func (a *apiServer) StartCommit(ctx context.Context, request *pfs.StartCommitRequest) (*pfs.Commit, error) {
	a.versionLock.RLock()
	defer a.versionLock.RUnlock()
	ctx = versionToContext(a.version, ctx)
	clientConns, err := a.router.GetAllClientConns(a.version)
	if err != nil {
		return nil, err
	}
	if request.Commit == nil {
		request.Commit = &pfs.Commit{
			Repo: request.Parent.Repo,
			Id:   strings.Replace(uuid.NewV4().String(), "-", "", -1),
		}
	}
	for _, clientConn := range clientConns {
		if _, err := pfs.NewInternalAPIClient(clientConn).StartCommit(ctx, request); err != nil {
			return nil, err
		}
	}
	return request.Commit, nil
}

func (a *apiServer) FinishCommit(ctx context.Context, request *pfs.FinishCommitRequest) (*google_protobuf.Empty, error) {
	a.versionLock.RLock()
	defer a.versionLock.RUnlock()
	ctx = versionToContext(a.version, ctx)
	clientConns, err := a.router.GetAllClientConns(a.version)
	if err != nil {
		return nil, err
	}
	for _, clientConn := range clientConns {
		if _, err := pfs.NewInternalAPIClient(clientConn).FinishCommit(ctx, request); err != nil {
			return nil, err
		}
	}
	return emptyInstance, nil
}

// TODO(pedge): race on Branch
func (a *apiServer) InspectCommit(ctx context.Context, request *pfs.InspectCommitRequest) (*pfs.CommitInfo, error) {
	a.versionLock.RLock()
	defer a.versionLock.RUnlock()
	ctx = versionToContext(a.version, ctx)
	clientConn, err := a.getClientConn(a.version)
	if err != nil {
		return nil, err
	}
	return pfs.NewInternalAPIClient(clientConn).InspectCommit(ctx, request)
}

func (a *apiServer) ListCommit(ctx context.Context, request *pfs.ListCommitRequest) (*pfs.CommitInfos, error) {
	a.versionLock.RLock()
	defer a.versionLock.RUnlock()
	ctx = versionToContext(a.version, ctx)
	clientConns, err := a.router.GetAllClientConns(a.version)
	if err != nil {
		return nil, err
	}
	var wg sync.WaitGroup
	var lock sync.Mutex
	var commitInfos [][]*pfs.CommitInfo
	var loopErr error
	for _, clientConn := range clientConns {
		wg.Add(1)
		go func(clientConn *grpc.ClientConn) {
			defer wg.Done()
			subCommitInfos, err := pfs.NewInternalAPIClient(clientConn).ListCommit(ctx, request)
			lock.Lock()
			defer lock.Unlock()
			if err != nil {
				loopErr = err
				return
			}
			commitInfos = append(commitInfos, subCommitInfos.CommitInfo)
		}(clientConn)
	}
	wg.Wait()
	if loopErr != nil {
		return nil, loopErr
	}
	idToCommitInfo := make(map[string]*pfs.CommitInfo)
	idToSeenCount := make(map[string]int)
	for _, subCommitInfos := range commitInfos {
		for _, subCommitInfo := range subCommitInfos {
			idToSeenCount[subCommitInfo.Commit.Id] = idToSeenCount[subCommitInfo.Commit.Id] + 1
			commitInfo, ok := idToCommitInfo[subCommitInfo.Commit.Id]
			if !ok {
				idToCommitInfo[subCommitInfo.Commit.Id] = subCommitInfo
				continue
			}
			if commitInfo.CommitType != pfs.CommitType_COMMIT_TYPE_READ {
				commitInfo.CommitType = subCommitInfo.CommitType
			}
			if prototime.TimestampToTime(subCommitInfo.Started).Before(prototime.TimestampToTime(commitInfo.Started)) {
				commitInfo.Started = subCommitInfo.Started
			}
			if commitInfo.Finished != nil {
				if subCommitInfo.Finished == nil {
					commitInfo.Finished = nil
				} else if prototime.TimestampToTime(subCommitInfo.Finished).After(prototime.TimestampToTime(commitInfo.Finished)) {
					commitInfo.Finished = subCommitInfo.Finished
				}
			}
			commitInfo.CommitBytes += subCommitInfo.CommitBytes
			commitInfo.TotalBytes += subCommitInfo.TotalBytes
		}
	}
	var result []*pfs.CommitInfo
	for _, subCommitInfos := range commitInfos {
		for _, commitInfo := range subCommitInfos {
			if idToSeenCount[commitInfo.Commit.Id] == len(commitInfos) {
				result = append(result, idToCommitInfo[commitInfo.Commit.Id])
			}
		}
		break
	}
	return &pfs.CommitInfos{CommitInfo: result}, nil
}

func (a *apiServer) DeleteCommit(ctx context.Context, request *pfs.DeleteCommitRequest) (*google_protobuf.Empty, error) {
	a.versionLock.RLock()
	defer a.versionLock.RUnlock()
	ctx = versionToContext(a.version, ctx)
	clientConns, err := a.router.GetAllClientConns(a.version)
	if err != nil {
		return nil, err
	}
	for _, clientConn := range clientConns {
		if _, err := pfs.NewAPIClient(clientConn).DeleteCommit(ctx, request); err != nil {
			return nil, err
		}
	}
	return emptyInstance, nil
}

func (a *apiServer) PutBlock(ctx context.Context, request *pfs.PutBlockRequest) (*pfs.Block, error) {
	a.versionLock.RLock()
	defer a.versionLock.RUnlock()
	ctx = versionToContext(a.version, ctx)
	block := a.sharder.GetBlock(request.Value)
	clientConn, err := a.getClientConnForBlock(block, a.version)
	if err != nil {
		return nil, err
	}
	if _, err := pfs.NewInternalAPIClient(clientConn).PutBlock(ctx, request); err != nil {
		return nil, err
	}
	return block, nil
}

func (a *apiServer) GetBlock(request *pfs.GetBlockRequest, apiGetBlockServer pfs.API_GetBlockServer) (retErr error) {
	ctx := versionToContext(a.version, apiGetBlockServer.Context())
	clientConn, err := a.getClientConnForBlock(request.Block, a.version)
	if err != nil {
		return err
	}
	blockGetClient, err := pfs.NewInternalAPIClient(clientConn).GetBlock(ctx, request)
	if err != nil {
		return err
	}
	return protostream.RelayFromStreamingBytesClient(blockGetClient, apiGetBlockServer)
}

func (a *apiServer) InspectBlock(ctx context.Context, request *pfs.InspectBlockRequest) (*pfs.BlockInfo, error) {
	a.versionLock.RLock()
	defer a.versionLock.RUnlock()
	ctx = versionToContext(a.version, ctx)
	clientConn, err := a.getClientConnForBlock(request.Block, a.version)
	if err != nil {
		return nil, err
	}
	return pfs.NewInternalAPIClient(clientConn).InspectBlock(ctx, request)
}

func (a *apiServer) ListBlock(ctx context.Context, request *pfs.ListBlockRequest) (*pfs.BlockInfos, error) {
	a.versionLock.RLock()
	defer a.versionLock.RUnlock()
	ctx = versionToContext(a.version, ctx)
	clientConns, err := a.router.GetAllClientConns(a.version)
	if err != nil {
		return nil, err
	}
	var wg sync.WaitGroup
	var lock sync.Mutex
	var blockInfos []*pfs.BlockInfo
	var loopErr error
	for _, clientConn := range clientConns {
		wg.Add(1)
		go func(clientConn *grpc.ClientConn) {
			defer wg.Done()
			subBlockInfos, err := pfs.NewInternalAPIClient(clientConn).ListBlock(ctx, request)
			lock.Lock()
			defer lock.Unlock()
			if err != nil {
				if loopErr == nil {
					loopErr = err
				}
				return
			}
			for _, blockInfo := range subBlockInfos.BlockInfo {
				blockInfos = append(blockInfos, blockInfo)
			}
		}(clientConn)
	}
	wg.Wait()
	if loopErr != nil {
		return nil, loopErr
	}
	return &pfs.BlockInfos{
		BlockInfo: blockInfos,
	}, nil
}

func (a *apiServer) PutFile(putFileServer pfs.API_PutFileServer) (retErr error) {
	ctx := versionToContext(a.version, putFileServer.Context())
	defer func() {
		if err := putFileServer.SendAndClose(emptyInstance); err != nil && retErr == nil {
			retErr = err
		}
	}()
	request, err := putFileServer.Recv()
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
		clientConns, err := a.router.GetAllClientConns(a.version)
		if err != nil {
			return err
		}
		for _, clientConn := range clientConns {
			putFileClient, err := pfs.NewInternalAPIClient(clientConn).PutFile(ctx)
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
	putFileClient, err := pfs.NewInternalAPIClient(clientConn).PutFile(ctx)
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
	if _, err := putFileClient.CloseAndRecv(); err != nil {
		return err
	}
	return nil
}

func (a *apiServer) GetFile(request *pfs.GetFileRequest, apiGetFileServer pfs.API_GetFileServer) error {
	ctx := versionToContext(a.version, apiGetFileServer.Context())
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

func (a *apiServer) InspectFile(ctx context.Context, request *pfs.InspectFileRequest) (*pfs.FileInfo, error) {
	a.versionLock.RLock()
	defer a.versionLock.RUnlock()
	ctx = versionToContext(a.version, ctx)
	clientConn, err := a.getClientConnForFile(request.File, a.version)
	if err != nil {
		return nil, err
	}
	return pfs.NewInternalAPIClient(clientConn).InspectFile(ctx, request)
}

func (a *apiServer) ListFile(ctx context.Context, request *pfs.ListFileRequest) (*pfs.FileInfos, error) {
	a.versionLock.RLock()
	defer a.versionLock.RUnlock()
	ctx = versionToContext(a.version, ctx)
	clientConns, err := a.router.GetAllClientConns(a.version)
	if err != nil {
		return nil, err
	}
	var wg sync.WaitGroup
	var lock sync.Mutex
	var fileInfos []*pfs.FileInfo
	seenDirectories := make(map[string]bool)
	var loopErr error
	for _, clientConn := range clientConns {
		wg.Add(1)
		go func(clientConn *grpc.ClientConn) {
			defer wg.Done()
			subFileInfos, err := pfs.NewInternalAPIClient(clientConn).ListFile(ctx, request)
			lock.Lock()
			defer lock.Unlock()
			if err != nil {
				if loopErr == nil {
					loopErr = err
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
	if loopErr != nil {
		return nil, loopErr
	}
	return &pfs.FileInfos{
		FileInfo: fileInfos,
	}, nil
}

func (a *apiServer) ListChange(ctx context.Context, request *pfs.ListChangeRequest) (*pfs.Changes, error) {
	a.versionLock.RLock()
	defer a.versionLock.RUnlock()
	ctx = versionToContext(a.version, ctx)
	clientConns, err := a.router.GetAllClientConns(a.version)
	if err != nil {
		return nil, err
	}
	var wg sync.WaitGroup
	var lock sync.Mutex
	var changes []*pfs.Change
	var loopErr error
	for _, clientConn := range clientConns {
		wg.Add(1)
		go func(clientConn *grpc.ClientConn) {
			defer wg.Done()
			subChanges, err := pfs.NewInternalAPIClient(clientConn).ListChange(ctx, request)
			lock.Lock()
			defer lock.Unlock()
			if err != nil {
				if loopErr == nil {
					loopErr = err
				}
				return
			}
			for _, change := range subChanges.Change {
				changes = append(changes, change)
			}
		}(clientConn)
	}
	wg.Wait()
	if loopErr != nil {
		return nil, loopErr
	}
	return &pfs.Changes{
		Change: changes,
	}, nil
}

func (a *apiServer) DeleteFile(ctx context.Context, request *pfs.DeleteFileRequest) (*google_protobuf.Empty, error) {
	a.versionLock.RLock()
	defer a.versionLock.RUnlock()
	ctx = versionToContext(a.version, ctx)
	clientConn, err := a.getClientConnForFile(request.File, a.version)
	if err != nil {
		return nil, err
	}
	return pfs.NewInternalAPIClient(clientConn).DeleteFile(ctx, request)
}

func (a *apiServer) InspectServer(ctx context.Context, request *pfs.InspectServerRequest) (*pfs.ServerInfo, error) {
	return a.router.InspectServer(request.Server)
}

func (a *apiServer) ListServer(ctx context.Context, request *pfs.ListServerRequest) (*pfs.ServerInfos, error) {
	serverInfo, err := a.router.ListServer()
	if err != nil {
		return nil, err
	}
	return &pfs.ServerInfos{
		ServerInfo: serverInfo,
	}, nil
}

func (a *apiServer) Version(version int64) error {
	protolog.Printf("apiServer.Version %d", version)
	a.versionLock.Lock()
	defer a.versionLock.Unlock()
	protolog.Printf("got lock")
	a.version = version
	return nil
}

func (a *apiServer) getClientConn(version int64) (*grpc.ClientConn, error) {
	shards, err := a.router.GetMasterShards(a.version)
	if err != nil {
		return nil, err
	}
	for shard := range shards {
		return a.router.GetMasterClientConn(shard, version)
	}
	return a.router.GetMasterClientConn(uint64(rand.Int())%a.sharder.NumShards(), version)
}

func (a *apiServer) getClientConnForBlock(block *pfs.Block, version int64) (*grpc.ClientConn, error) {
	return a.router.GetMasterClientConn(a.sharder.GetBlockShard(block), version)
}

func (a *apiServer) getClientConnForFile(file *pfs.File, version int64) (*grpc.ClientConn, error) {
	return a.router.GetMasterClientConn(a.sharder.GetShard(file), version)
}

func versionToContext(version int64, ctx context.Context) context.Context {
	return metadata.NewContext(
		ctx,
		metadata.Pairs("version", fmt.Sprint(version)),
	)
}
