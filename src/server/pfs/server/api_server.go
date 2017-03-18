package server

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/server/pfs/drive"
	"github.com/pachyderm/pachyderm/src/server/pkg/metrics"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"

	"go.pedge.io/proto/rpclog"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var (
	grpcErrorf = grpc.Errorf // needed to get passed govet
)

type apiServer struct {
	protorpclog.Logger
	driver   drive.Driver
	reporter *metrics.Reporter
}

func newAPIServer(driver drive.Driver, reporter *metrics.Reporter) *apiServer {
	return &apiServer{
		Logger:   protorpclog.NewLogger("pfs.API"),
		driver:   driver,
		reporter: reporter,
	}
}

func (a *apiServer) CreateRepo(ctx context.Context, request *pfs.CreateRepoRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "CreateRepo")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	if err := a.driver.CreateRepo(ctx, request.Repo, request.Provenance); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

func (a *apiServer) InspectRepo(ctx context.Context, request *pfs.InspectRepoRequest) (response *pfs.RepoInfo, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "InspectRepo")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	return a.driver.InspectRepo(ctx, request.Repo)
}

func (a *apiServer) ListRepo(ctx context.Context, request *pfs.ListRepoRequest) (response *pfs.RepoInfos, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "ListRepo")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	repoInfos, err := a.driver.ListRepo(ctx, request.Provenance)
	return &pfs.RepoInfos{RepoInfo: repoInfos}, err
}

func (a *apiServer) DeleteRepo(ctx context.Context, request *pfs.DeleteRepoRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "DeleteRepo")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	err := a.driver.DeleteRepo(ctx, request.Repo, request.Force)
	if err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

func (a *apiServer) StartCommit(ctx context.Context, request *pfs.StartCommitRequest) (response *pfs.Commit, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "StartCommit")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	commit, err := a.driver.StartCommit(ctx, request.Parent, request.Provenance)
	if err != nil {
		return nil, err
	}
	return commit, nil
}

func (a *apiServer) BuildCommit(ctx context.Context, request *pfs.BuildCommitRequest) (response *pfs.Commit, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "StartCommit")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	commit, err := a.driver.BuildCommit(ctx, request.Parent, request.Provenance, request.Tree)
	if err != nil {
		return nil, err
	}
	return commit, nil
}

func (a *apiServer) FinishCommit(ctx context.Context, request *pfs.FinishCommitRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "FinishCommit")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	if err := a.driver.FinishCommit(ctx, request.Commit); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

func (a *apiServer) InspectCommit(ctx context.Context, request *pfs.InspectCommitRequest) (response *pfs.CommitInfo, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "InspectCommit")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	return a.driver.InspectCommit(ctx, request.Commit)
}

func (a *apiServer) ListCommit(ctx context.Context, request *pfs.ListCommitRequest) (response *pfs.CommitInfos, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "ListCommit")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	commitInfos, err := a.driver.ListCommit(ctx, request.Repo, request.To, request.From, request.Number)
	if err != nil {
		return nil, err
	}
	return &pfs.CommitInfos{
		CommitInfo: commitInfos,
	}, nil
}

func (a *apiServer) ListBranch(ctx context.Context, request *pfs.ListBranchRequest) (response *pfs.Branches, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "ListBranch")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	branches, err := a.driver.ListBranch(ctx, request.Repo)
	if err != nil {
		return nil, err
	}
	return &pfs.Branches{Branches: branches}, nil
}

func (a *apiServer) SetBranch(ctx context.Context, request *pfs.SetBranchRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "ListBranch")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	if err := a.driver.SetBranch(ctx, request.Commit, request.Branch); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

func (a *apiServer) DeleteBranch(ctx context.Context, request *pfs.DeleteBranchRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "ListBranch")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	if err := a.driver.DeleteBranch(ctx, request.Repo, request.Branch); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

func (a *apiServer) DeleteCommit(ctx context.Context, request *pfs.DeleteCommitRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "DeleteCommit")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	if err := a.driver.DeleteCommit(ctx, request.Commit); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

func (a *apiServer) FlushCommit(request *pfs.FlushCommitRequest, stream pfs.API_FlushCommitServer) (retErr error) {
	ctx := stream.Context()
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "GetFile")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	commitStream, err := a.driver.FlushCommit(ctx, request.Commits, request.ToRepos)
	if err != nil {
		return err
	}
	defer func() {
		commitStream.Close()
	}()

	for {
		ev, ok := <-commitStream.Stream()
		if !ok {
			return nil
		}
		if ev.Err != nil {
			return ev.Err
		}
		if err := stream.Send(ev.Value); err != nil {
			return err
		}
	}
}

func (a *apiServer) SubscribeCommit(request *pfs.SubscribeCommitRequest, stream pfs.API_SubscribeCommitServer) (retErr error) {
	ctx := stream.Context()
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "GetFile")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	commitStream, err := a.driver.SubscribeCommit(ctx, request.Repo, request.Branch, request.From)
	if err != nil {
		return err
	}
	defer func() {
		commitStream.Close()
	}()

	for {
		ev, ok := <-commitStream.Stream()
		if !ok {
			return nil
		}
		if ev.Err != nil {
			return ev.Err
		}
		if err := stream.Send(ev.Value); err != nil {
			return err
		}
	}
	return nil
}

func (a *apiServer) PutFile(putFileServer pfs.API_PutFileServer) (retErr error) {
	ctx := putFileServer.Context()
	defer drainFileServer(putFileServer)
	defer func() {
		if err := putFileServer.SendAndClose(&types.Empty{}); err != nil && retErr == nil {
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
	// We remove request.Value from the logs otherwise they would be too big.
	func() {
		requestValue := request.Value
		request.Value = nil
		a.Log(request, nil, nil, 0)
		request.Value = requestValue
	}()
	defer func(start time.Time) {
		requestValue := request.Value
		request.Value = nil
		a.Log(request, nil, retErr, time.Since(start))
		request.Value = requestValue
	}(time.Now())
	// not cleaning the path can result in weird effects like files called
	// ./foo which won't display correctly when the filesystem is mounted
	request.File.Path = path.Clean(request.File.Path)
	if request.FileType == pfs.FileType_DIR {
		if len(request.Value) > 0 {
			return fmt.Errorf("PutFileRequest shouldn't have type dir and a value")
		}
		if err := a.driver.MakeDirectory(ctx, request.File); err != nil {
			return err
		}
	} else {
		var r io.Reader
		if request.Url != "" {
			url, err := url.Parse(request.Url)
			if err != nil {
				return err
			}
			switch url.Scheme {
			case "http":
				fallthrough
			case "https":
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
			default:
				objClient, err := obj.NewClientFromURLAndSecret(putFileServer.Context(), request.Url)
				if err != nil {
					return err
				}
				return a.putFileObj(ctx, objClient, request, url)
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
		}
		if err := a.driver.PutFile(ctx, request.File, request.Delimiter, request.TargetFileDatums, request.TargetFileBytes, r); err != nil {
			return err
		}
	}
	return nil
}

func (a *apiServer) putFileObj(ctx context.Context, objClient obj.Client, request *pfs.PutFileRequest, url *url.URL) (retErr error) {
	put := func(filePath string, objPath string) error {
		r, err := objClient.Reader(objPath, 0, 0)
		if err != nil {
			return err
		}
		defer func() {
			if err := r.Close(); err != nil && retErr == nil {
				retErr = err
			}
		}()
		return a.driver.PutFile(ctx, client.NewFile(request.File.Commit.Repo.Name, request.File.Commit.ID, filePath),
			request.Delimiter, request.TargetFileDatums, request.TargetFileBytes, r)
	}
	if request.Recursive {
		var eg errgroup.Group
		path := strings.TrimPrefix(url.Path, "/")
		sem := make(chan struct{}, client.DefaultMaxConcurrentStreams)
		objClient.Walk(path, func(name string) error {
			eg.Go(func() error {
				sem <- struct{}{}
				defer func() {
					<-sem
				}()
				return put(filepath.Join(request.File.Path, strings.TrimPrefix(name, path)), name)
			})
			return nil
		})
		return eg.Wait()
	}
	return put(request.File.Path, url.Path)
}

func (a *apiServer) GetFile(request *pfs.GetFileRequest, apiGetFileServer pfs.API_GetFileServer) (retErr error) {
	ctx := apiGetFileServer.Context()
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(apiGetFileServer.Context(), a.reporter, "GetFile")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	file, err := a.driver.GetFile(ctx, request.File, request.OffsetBytes, request.SizeBytes)
	if err != nil {
		return err
	}
	return grpcutil.WriteToStreamingBytesServer(file, apiGetFileServer)
}

func (a *apiServer) InspectFile(ctx context.Context, request *pfs.InspectFileRequest) (response *pfs.FileInfo, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "InspectFile")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	return a.driver.InspectFile(ctx, request.File)
}

func (a *apiServer) ListFile(ctx context.Context, request *pfs.ListFileRequest) (response *pfs.FileInfos, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "ListFile")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	fileInfos, err := a.driver.ListFile(ctx, request.File)
	if err != nil {
		return nil, err
	}
	return &pfs.FileInfos{
		FileInfo: fileInfos,
	}, nil
}

func (a *apiServer) GlobFile(ctx context.Context, request *pfs.GlobFileRequest) (response *pfs.FileInfos, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "GlobFile")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	fileInfos, err := a.driver.GlobFile(ctx, request.Commit, request.Pattern)
	if err != nil {
		return nil, err
	}
	return &pfs.FileInfos{
		FileInfo: fileInfos,
	}, nil
}

func (a *apiServer) DeleteFile(ctx context.Context, request *pfs.DeleteFileRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "DeleteFile")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	err := a.driver.DeleteFile(ctx, request.File)
	if err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

func (a *apiServer) DeleteAll(ctx context.Context, request *types.Empty) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	metricsFn := metrics.ReportUserAction(ctx, a.reporter, "PFSDeleteAll")
	defer func(start time.Time) { metricsFn(start, retErr) }(time.Now())

	if err := a.driver.DeleteAll(ctx); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

type putFileReader struct {
	server pfs.API_PutFileServer
	buffer bytes.Buffer
}

func (r *putFileReader) Read(p []byte) (int, error) {
	if r.buffer.Len() == 0 {
		request, err := r.server.Recv()
		fmt.Printf("receving PutFile req for file: %v\n", request)
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
