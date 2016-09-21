package server

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/server/pfs/drive"

	"go.pedge.io/pb/go/google/protobuf"
	"go.pedge.io/proto/rpclog"
	"go.pedge.io/proto/stream"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	grpcErrorf = grpc.Errorf // needed to get passed govet
)

type apiServer struct {
	protorpclog.Logger
	driver drive.Driver
}

func newAPIServer(driver drive.Driver) *apiServer {
	return &apiServer{
		Logger: protorpclog.NewLogger("pachyderm.pfsserver.API"),
		driver: driver,
	}
}

func (a *apiServer) CreateRepo(ctx context.Context, request *pfs.CreateRepoRequest) (response *google_protobuf.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	if err := a.driver.CreateRepo(request.Repo, request.Provenance); err != nil {
		return nil, err
	}
	return google_protobuf.EmptyInstance, nil
}

func (a *apiServer) InspectRepo(ctx context.Context, request *pfs.InspectRepoRequest) (response *pfs.RepoInfo, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	return a.driver.InspectRepo(request.Repo)
}

func (a *apiServer) ListRepo(ctx context.Context, request *pfs.ListRepoRequest) (response *pfs.RepoInfos, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	repoInfos, err := a.driver.ListRepo(request.Provenance)
	return &pfs.RepoInfos{RepoInfo: repoInfos}, err
}

func (a *apiServer) DeleteRepo(ctx context.Context, request *pfs.DeleteRepoRequest) (response *google_protobuf.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	err := a.driver.DeleteRepo(request.Repo, request.Force)
	if err != nil {
		return nil, err
	}
	return google_protobuf.EmptyInstance, nil
}

func (a *apiServer) ForkCommit(ctx context.Context, request *pfs.ForkCommitRequest) (response *pfs.Commit, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	commit, err := a.driver.ForkCommit(request.Parent, request.Branch, request.Provenance)
	if err != nil {
		return nil, err
	}
	return commit, nil
}

func (a *apiServer) StartCommit(ctx context.Context, request *pfs.StartCommitRequest) (response *pfs.Commit, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	commit, err := a.driver.StartCommit(request.Parent, request.Provenance)
	if err != nil {
		return nil, err
	}
	return commit, nil
}

func (a *apiServer) FinishCommit(ctx context.Context, request *pfs.FinishCommitRequest) (response *google_protobuf.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	if err := a.driver.FinishCommit(request.Commit, request.Cancel); err != nil {
		return nil, err
	}
	return google_protobuf.EmptyInstance, nil
}

func (a *apiServer) ArchiveCommit(ctx context.Context, request *pfs.ArchiveCommitRequest) (response *google_protobuf.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	if err := a.driver.ArchiveCommit(request.Commits); err != nil {
		return nil, err
	}
	return google_protobuf.EmptyInstance, nil
}

func (a *apiServer) InspectCommit(ctx context.Context, request *pfs.InspectCommitRequest) (response *pfs.CommitInfo, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	return a.driver.InspectCommit(request.Commit)
}

func (a *apiServer) ListCommit(ctx context.Context, request *pfs.ListCommitRequest) (response *pfs.CommitInfos, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	commitInfos, err := a.driver.ListCommit(request.FromCommits, request.Provenance, request.CommitType, request.Status, request.Block)
	if err != nil {
		return nil, err
	}
	return &pfs.CommitInfos{
		CommitInfo: commitInfos,
	}, nil
}

func (a *apiServer) SquashCommit(ctx context.Context, request *pfs.SquashCommitRequest) (response *google_protobuf.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	return google_protobuf.EmptyInstance, a.driver.SquashCommit(request.FromCommits, request.ToCommit)
}

func (a *apiServer) ReplayCommit(ctx context.Context, request *pfs.ReplayCommitRequest) (response *pfs.Commits, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	commits, err := a.driver.ReplayCommit(request.FromCommits, request.ToBranch)
	if err != nil {
		return nil, err
	}
	return &pfs.Commits{commits}, nil
}

func (a *apiServer) ListBranch(ctx context.Context, request *pfs.ListBranchRequest) (response *pfs.Branches, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	branches, err := a.driver.ListBranch(request.Repo, request.Status)
	if err != nil {
		return nil, err
	}
	return &pfs.Branches{branches}, nil
}

func (a *apiServer) DeleteCommit(ctx context.Context, request *pfs.DeleteCommitRequest) (response *google_protobuf.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	if err := a.driver.DeleteCommit(request.Commit); err != nil {
		return nil, err
	}
	return google_protobuf.EmptyInstance, nil
}

func (a *apiServer) FlushCommit(ctx context.Context, request *pfs.FlushCommitRequest) (response *pfs.CommitInfos, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	commitInfos, err := a.driver.FlushCommit(request.Commit, request.ToRepo)
	if err != nil {
		return nil, err
	}
	return &pfs.CommitInfos{
		CommitInfo: commitInfos,
	}, nil
}

func (a *apiServer) PutFile(putFileServer pfs.API_PutFileServer) (retErr error) {
	var request *pfs.PutFileRequest
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) {
		if request != nil {
			request.Value = nil // we set the value to nil so as not to spam logs
		}
		a.Log(request, nil, retErr, time.Since(start))
	}(time.Now())
	defer drainFileServer(putFileServer)
	defer func() {
		if err := putFileServer.SendAndClose(google_protobuf.EmptyInstance); err != nil && retErr == nil {
			retErr = err
		}
	}()
	request, err := putFileServer.Recv()
	if err != nil && err != io.EOF {
		return err
	}
	if err == io.EOF {
		// tolerate people calling and immediately hanging up
		return nil
	}
	if request.FileType == pfs.FileType_FILE_TYPE_DIR {
		if len(request.Value) > 0 {
			return fmt.Errorf("PutFileRequest shouldn't have type dir and a value")
		}
		if err := a.driver.MakeDirectory(request.File); err != nil {
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
		if err := a.driver.PutFile(request.File, delimiter, r); err != nil {
			return err
		}
	}
	return nil
}

func (a *apiServer) GetFile(request *pfs.GetFileRequest, apiGetFileServer pfs.API_GetFileServer) (retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	file, err := a.driver.GetFile(request.File, request.Shard, request.OffsetBytes, request.SizeBytes, request.DiffMethod)
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

func (a *apiServer) InspectFile(ctx context.Context, request *pfs.InspectFileRequest) (response *pfs.FileInfo, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	return a.driver.InspectFile(request.File, request.Shard, request.DiffMethod)
}

func (a *apiServer) ListFile(ctx context.Context, request *pfs.ListFileRequest) (response *pfs.FileInfos, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	fileInfos, err := a.driver.ListFile(request.File, request.Shard,
		request.DiffMethod, request.Recurse)
	if err != nil {
		return nil, err
	}
	return &pfs.FileInfos{
		FileInfo: fileInfos,
	}, nil
}

func (a *apiServer) DeleteFile(ctx context.Context, request *pfs.DeleteFileRequest) (response *google_protobuf.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	err := a.driver.DeleteFile(request.File)
	if err != nil {
		return nil, err
	}
	return google_protobuf.EmptyInstance, nil
}

func (a *apiServer) DeleteAll(ctx context.Context, request *google_protobuf.Empty) (response *google_protobuf.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	if err := a.driver.DeleteAll(); err != nil {
		return nil, err
	}
	return google_protobuf.EmptyInstance, nil
}

func (a *apiServer) ArchiveAll(ctx context.Context, request *google_protobuf.Empty) (response *google_protobuf.Empty, retErr error) {
	a.Log(request, nil, nil, 0)
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	if err := a.driver.ArchiveAll(); err != nil {
		return nil, err
	}
	return google_protobuf.EmptyInstance, nil
}

type putFileReader struct {
	server pfs.API_PutFileServer
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

func (a *apiServer) getVersion(ctx context.Context) (int64, error) {
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

func drainFileServer(putFileServer interface {
	Recv() (*pfs.PutFileRequest, error)
}) {
	for {
		if _, err := putFileServer.Recv(); err != nil {
			break
		}
	}
}
