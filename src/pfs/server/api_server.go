package server

import (
	"fmt"
	"math/rand"
	"strings"

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

func (a *apiServer) RepoCreate(ctx context.Context, request *pfs.RepoCreateRequest) (*google_protobuf.Empty, error) {
	version, ctx, err := a.versionAndCtx(ctx)
	if err != nil {
		return nil, err
	}
	clientConns, err := a.router.GetAllClientConns(version)
	if err != nil {
		return nil, err
	}
	for _, clientConn := range clientConns {
		if _, err := pfs.NewInternalApiClient(clientConn).RepoCreate(ctx, request); err != nil {
			return nil, err
		}
	}
	// Create the initial commit
	if _, err = a.CommitStart(ctx, &pfs.CommitStartRequest{
		Parent: nil,
		Commit: &pfs.Commit{
			Repo: request.Repo,
			Id:   InitialCommitID,
		},
	}); err != nil {
		return nil, err
	}
	if _, err = a.CommitFinish(ctx, &pfs.CommitFinishRequest{
		Commit: &pfs.Commit{
			Repo: request.Repo,
			Id:   InitialCommitID,
		},
	}); err != nil {
		return nil, err
	}
	return emptyInstance, nil
}

func (a *apiServer) RepoInspect(ctx context.Context, request *pfs.RepoInspectRequest) (*pfs.RepoInfo, error) {
	version, ctx, err := a.versionAndCtx(ctx)
	if err != nil {
		return nil, err
	}
	clientConn, err := a.getClientConn(version)
	if err != nil {
		return nil, err
	}
	return pfs.NewInternalApiClient(clientConn).RepoInspect(ctx, request)
}

func (a *apiServer) RepoList(ctx context.Context, request *pfs.RepoListRequest) (*pfs.RepoInfos, error) {
	version, ctx, err := a.versionAndCtx(ctx)
	if err != nil {
		return nil, err
	}
	clientConn, err := a.getClientConn(version)
	if err != nil {
		return nil, err
	}
	return pfs.NewInternalApiClient(clientConn).RepoList(ctx, request)
}

func (a *apiServer) RepoDelete(ctx context.Context, request *pfs.RepoDeleteRequest) (*google_protobuf.Empty, error) {
	version, ctx, err := a.versionAndCtx(ctx)
	if err != nil {
		return nil, err
	}
	clientConns, err := a.router.GetAllClientConns(version)
	if err != nil {
		return nil, err
	}
	for _, clientConn := range clientConns {
		if _, err := pfs.NewInternalApiClient(clientConn).RepoDelete(ctx, request); err != nil {
			return nil, err
		}
	}
	return emptyInstance, nil

}

func (a *apiServer) CommitStart(ctx context.Context, request *pfs.CommitStartRequest) (*pfs.Commit, error) {
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
		if _, err := pfs.NewInternalApiClient(clientConn).CommitStart(ctx, request); err != nil {
			return nil, err
		}
	}
	return request.Commit, nil
}

func (a *apiServer) CommitFinish(ctx context.Context, request *pfs.CommitFinishRequest) (*google_protobuf.Empty, error) {
	version, ctx, err := a.versionAndCtx(ctx)
	if err != nil {
		return nil, err
	}
	clientConns, err := a.router.GetAllClientConns(version)
	if err != nil {
		return nil, err
	}
	for _, clientConn := range clientConns {
		if _, err := pfs.NewInternalApiClient(clientConn).CommitFinish(ctx, request); err != nil {
			return nil, err
		}
	}
	return emptyInstance, nil
}

// TODO(pedge): race on Branch
func (a *apiServer) CommitInspect(ctx context.Context, request *pfs.CommitInspectRequest) (*pfs.CommitInfo, error) {
	version, ctx, err := a.versionAndCtx(ctx)
	if err != nil {
		return nil, err
	}
	clientConn, err := a.getClientConn(version)
	if err != nil {
		return nil, err
	}
	return pfs.NewInternalApiClient(clientConn).CommitInspect(ctx, request)
}

func (a *apiServer) CommitList(ctx context.Context, request *pfs.CommitListRequest) (*pfs.CommitInfos, error) {
	version, ctx, err := a.versionAndCtx(ctx)
	if err != nil {
		return nil, err
	}
	clientConn, err := a.getClientConn(version)
	if err != nil {
		return nil, err
	}
	return pfs.NewInternalApiClient(clientConn).CommitList(ctx, request)
}

func (a *apiServer) CommitDelete(ctx context.Context, request *pfs.CommitDeleteRequest) (*google_protobuf.Empty, error) {
	version, ctx, err := a.versionAndCtx(ctx)
	if err != nil {
		return nil, err
	}
	clientConns, err := a.router.GetAllClientConns(version)
	if err != nil {
		return nil, err
	}
	for _, clientConn := range clientConns {
		if _, err := pfs.NewApiClient(clientConn).CommitDelete(ctx, request); err != nil {
			return nil, err
		}
	}
	return emptyInstance, nil
}

func (a *apiServer) FilePut(ctx context.Context, request *pfs.FilePutRequest) (*google_protobuf.Empty, error) {
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
			return emptyInstance, fmt.Errorf("FilePutRequest shouldn't have type dir and a value")
		}
		clientConns, err := a.router.GetAllClientConns(version)
		if err != nil {
			return nil, err
		}
		for _, clientConn := range clientConns {
			if _, err := pfs.NewInternalApiClient(clientConn).FilePut(ctx, request); err != nil {
				return nil, err
			}
		}
		return emptyInstance, nil
	}
	clientConn, err := a.getClientConnForFile(request.File, version)
	if err != nil {
		return nil, err
	}
	return pfs.NewInternalApiClient(clientConn).FilePut(ctx, request)
}

func (a *apiServer) FileGet(request *pfs.FileGetRequest, apiFileGetServer pfs.Api_FileGetServer) error {
	version, ctx, err := a.versionAndCtx(context.Background())
	if err != nil {
		return err
	}
	clientConn, err := a.getClientConnForFile(request.File, version)
	if err != nil {
		return err
	}
	fileGetClient, err := pfs.NewInternalApiClient(clientConn).FileGet(ctx, request)
	if err != nil {
		return err
	}
	return protostream.RelayFromStreamingBytesClient(fileGetClient, apiFileGetServer)
}

func (a *apiServer) FileInspect(ctx context.Context, request *pfs.FileInspectRequest) (*pfs.FileInfo, error) {
	version, ctx, err := a.versionAndCtx(ctx)
	if err != nil {
		return nil, err
	}
	clientConn, err := a.getClientConnForFile(request.File, version)
	if err != nil {
		return nil, err
	}
	return pfs.NewInternalApiClient(clientConn).FileInspect(ctx, request)
}

func (a *apiServer) FileList(ctx context.Context, request *pfs.FileListRequest) (*pfs.FileInfos, error) {
	version, ctx, err := a.versionAndCtx(ctx)
	if err != nil {
		return nil, err
	}
	clientConns, err := a.router.GetAllClientConns(version)
	if err != nil {
		return nil, err
	}
	var fileInfos []*pfs.FileInfo
	seenDirectories := make(map[string]bool)
	for _, clientConn := range clientConns {
		subFileInfos, err := pfs.NewInternalApiClient(clientConn).FileList(ctx, request)
		if err != nil {
			return nil, err
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
	}
	return &pfs.FileInfos{
		FileInfo: fileInfos,
	}, nil
}

func (a *apiServer) FileDelete(ctx context.Context, request *pfs.FileDeleteRequest) (*google_protobuf.Empty, error) {
	version, ctx, err := a.versionAndCtx(ctx)
	if err != nil {
		return nil, err
	}
	clientConn, err := a.getClientConnForFile(request.File, version)
	if err != nil {
		return nil, err
	}
	return pfs.NewInternalApiClient(clientConn).FileDelete(ctx, request)
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
	shard, err := a.sharder.GetShard(file)
	if err != nil {
		return nil, err
	}
	return a.router.GetMasterClientConn(shard, version)
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
