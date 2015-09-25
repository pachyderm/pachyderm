package server

import (
	"fmt"
	"math/rand"
	"strings"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"golang.org/x/net/context"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/route"
	"github.com/satori/go.uuid"
	"go.pedge.io/google-protobuf"
	"go.pedge.io/proto/stream"
)

type apiServer struct {
	sharder route.Sharder
	router  route.Router
}

func newAPIServer(
	sharder route.Sharder,
	router route.Router,
) *apiServer {
	return &apiServer{
		sharder,
		router,
	}
}

func (a *apiServer) CreateRepo(ctx context.Context, request *pfs.CreateRepoRequest) (*google_protobuf.Empty, error) {
	version, ctx, err := a.versionAndCtx(ctx)
	if err != nil {
		return nil, err
	}
	clientConns, err := a.router.GetAllClientConns(version)
	if err != nil {
		return nil, err
	}
	for _, clientConn := range clientConns {
		if _, err := pfs.NewInternalApiClient(clientConn).CreateRepo(ctx, request); err != nil {
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
	version, ctx, err := a.versionAndCtx(ctx)
	if err != nil {
		return nil, err
	}
	clientConn, err := a.getClientConn(version)
	if err != nil {
		return nil, err
	}
	return pfs.NewInternalApiClient(clientConn).InspectRepo(ctx, request)
}

func (a *apiServer) ListRepo(ctx context.Context, request *pfs.ListRepoRequest) (*pfs.RepoInfos, error) {
	version, ctx, err := a.versionAndCtx(ctx)
	if err != nil {
		return nil, err
	}
	clientConn, err := a.getClientConn(version)
	if err != nil {
		return nil, err
	}
	return pfs.NewInternalApiClient(clientConn).ListRepo(ctx, request)
}

func (a *apiServer) DeleteRepo(ctx context.Context, request *pfs.DeleteRepoRequest) (*google_protobuf.Empty, error) {
	version, ctx, err := a.versionAndCtx(ctx)
	if err != nil {
		return nil, err
	}
	clientConns, err := a.router.GetAllClientConns(version)
	if err != nil {
		return nil, err
	}
	for _, clientConn := range clientConns {
		if _, err := pfs.NewInternalApiClient(clientConn).DeleteRepo(ctx, request); err != nil {
			return nil, err
		}
	}
	return emptyInstance, nil

}

func (a *apiServer) StartCommit(ctx context.Context, request *pfs.StartCommitRequest) (*pfs.Commit, error) {
	version, ctx, err := a.versionAndCtx(ctx)
	if err != nil {
		return nil, err
	}
	clientConns, err := a.router.GetAllClientConns(version)
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
		if _, err := pfs.NewInternalApiClient(clientConn).StartCommit(ctx, request); err != nil {
			return nil, err
		}
	}
	return request.Commit, nil
}

func (a *apiServer) FinishCommit(ctx context.Context, request *pfs.FinishCommitRequest) (*google_protobuf.Empty, error) {
	version, ctx, err := a.versionAndCtx(ctx)
	if err != nil {
		return nil, err
	}
	clientConns, err := a.router.GetAllClientConns(version)
	if err != nil {
		return nil, err
	}
	for _, clientConn := range clientConns {
		if _, err := pfs.NewInternalApiClient(clientConn).FinishCommit(ctx, request); err != nil {
			return nil, err
		}
	}
	return emptyInstance, nil
}

// TODO(pedge): race on Branch
func (a *apiServer) InspectCommit(ctx context.Context, request *pfs.InspectCommitRequest) (*pfs.CommitInfo, error) {
	version, ctx, err := a.versionAndCtx(ctx)
	if err != nil {
		return nil, err
	}
	clientConn, err := a.getClientConn(version)
	if err != nil {
		return nil, err
	}
	return pfs.NewInternalApiClient(clientConn).InspectCommit(ctx, request)
}

func (a *apiServer) ListCommit(ctx context.Context, request *pfs.ListCommitRequest) (*pfs.CommitInfos, error) {
	version, ctx, err := a.versionAndCtx(ctx)
	if err != nil {
		return nil, err
	}
	clientConn, err := a.getClientConn(version)
	if err != nil {
		return nil, err
	}
	return pfs.NewInternalApiClient(clientConn).ListCommit(ctx, request)
}

func (a *apiServer) DeleteCommit(ctx context.Context, request *pfs.DeleteCommitRequest) (*google_protobuf.Empty, error) {
	version, ctx, err := a.versionAndCtx(ctx)
	if err != nil {
		return nil, err
	}
	clientConns, err := a.router.GetAllClientConns(version)
	if err != nil {
		return nil, err
	}
	for _, clientConn := range clientConns {
		if _, err := pfs.NewApiClient(clientConn).DeleteCommit(ctx, request); err != nil {
			return nil, err
		}
	}
	return emptyInstance, nil
}

func (a *apiServer) PutFile(ctx context.Context, request *pfs.PutFileRequest) (*google_protobuf.Empty, error) {
	version, ctx, err := a.versionAndCtx(ctx)
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
		clientConns, err := a.router.GetAllClientConns(version)
		if err != nil {
			return nil, err
		}
		for _, clientConn := range clientConns {
			if _, err := pfs.NewInternalApiClient(clientConn).PutFile(ctx, request); err != nil {
				return nil, err
			}
		}
		return emptyInstance, nil
	}
	clientConn, err := a.getClientConnForFile(request.File, version)
	if err != nil {
		return nil, err
	}
	return pfs.NewInternalApiClient(clientConn).PutFile(ctx, request)
}

func (a *apiServer) GetFile(request *pfs.GetFileRequest, apiGetFileServer pfs.Api_GetFileServer) error {
	version, ctx, err := a.versionAndCtx(context.Background())
	if err != nil {
		return err
	}
	clientConn, err := a.getClientConnForFile(request.File, version)
	if err != nil {
		return err
	}
	fileGetClient, err := pfs.NewInternalApiClient(clientConn).GetFile(ctx, request)
	if err != nil {
		return err
	}
	return protostream.RelayFromStreamingBytesClient(fileGetClient, apiGetFileServer)
}

func (a *apiServer) InspectFile(ctx context.Context, request *pfs.InspectFileRequest) (*pfs.FileInfo, error) {
	version, ctx, err := a.versionAndCtx(ctx)
	if err != nil {
		return nil, err
	}
	clientConn, err := a.getClientConnForFile(request.File, version)
	if err != nil {
		return nil, err
	}
	return pfs.NewInternalApiClient(clientConn).InspectFile(ctx, request)
}

func (a *apiServer) ListFile(ctx context.Context, request *pfs.ListFileRequest) (*pfs.FileInfos, error) {
	version, ctx, err := a.versionAndCtx(ctx)
	if err != nil {
		return nil, err
	}
	clientConns, err := a.router.GetAllClientConns(version)
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
			subFileInfos, err := pfs.NewInternalApiClient(clientConn).ListFile(ctx, request)
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
	version, ctx, err := a.versionAndCtx(ctx)
	if err != nil {
		return nil, err
	}
	clientConns, err := a.router.GetAllClientConns(version)
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
			subChanges, err := pfs.NewInternalApiClient(clientConn).ListChange(ctx, request)
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
	version, ctx, err := a.versionAndCtx(ctx)
	if err != nil {
		return nil, err
	}
	clientConn, err := a.getClientConnForFile(request.File, version)
	if err != nil {
		return nil, err
	}
	return pfs.NewInternalApiClient(clientConn).DeleteFile(ctx, request)
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

func (a *apiServer) getClientConn(version int64) (*grpc.ClientConn, error) {
	shards, err := a.router.GetMasterShards(version)
	if err != nil {
		return nil, err
	}
	for shard := range shards {
		return a.router.GetMasterClientConn(shard, version)
	}
	return a.router.GetMasterClientConn(uint64(rand.Int())%a.sharder.NumShards(), version)
}

func (a *apiServer) getClientConnForFile(file *pfs.File, version int64) (*grpc.ClientConn, error) {
	return a.router.GetMasterClientConn(a.sharder.GetShard(file), version)
}

func (a *apiServer) versionAndCtx(ctx context.Context) (int64, context.Context, error) {
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
