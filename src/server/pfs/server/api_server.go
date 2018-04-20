package server

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/log"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/pachyderm/pachyderm/src/server/pkg/serviceenv"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

var (
	grpcErrorf = grpc.Errorf // needed to get passed govet
)

type apiServer struct {
	log.Logger
	// env contains connections to other services in the cluster
	env *serviceenv.ServiceEnv

	driver *driver
}

func newAPIServer(env *serviceenv.ServiceEnv, etcdPrefix string, cacheSize int64) (*apiServer, error) {
	d, err := newDriver(env, etcdPrefix, cacheSize)
	if err != nil {
		return nil, err
	}
	return &apiServer{
		Logger: log.NewLogger("pfs.API"),
		env:    env,
		driver: d,
	}, nil
}

func (a *apiServer) CreateRepo(ctx context.Context, request *pfs.CreateRepoRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	if err := a.driver.createRepo(
		a.env.GetPachClient(ctx),
		request.Repo, request.Description, request.Update,
	); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

func (a *apiServer) InspectRepo(ctx context.Context, request *pfs.InspectRepoRequest) (response *pfs.RepoInfo, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	return a.driver.inspectRepo(a.env.GetPachClient(ctx), request.Repo, true)
}

func (a *apiServer) ListRepo(ctx context.Context, request *pfs.ListRepoRequest) (response *pfs.ListRepoResponse, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	repoInfos, err := a.driver.listRepo(a.env.GetPachClient(ctx), true)
	return repoInfos, err
}

func (a *apiServer) DeleteRepo(ctx context.Context, request *pfs.DeleteRepoRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	pachClient := a.env.GetPachClient(ctx)

	if request.All {
		if err := a.driver.deleteAll(pachClient); err != nil {
			return nil, err
		}
	} else {
		if err := a.driver.deleteRepo(pachClient, request.Repo, request.Force); err != nil {
			return nil, err
		}
	}

	return &types.Empty{}, nil
}

func (a *apiServer) StartCommit(ctx context.Context, request *pfs.StartCommitRequest) (response *pfs.Commit, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	commit, err := a.driver.startCommit(a.env.GetPachClient(ctx),
		request.Parent, request.Branch, request.Provenance, request.Description)
	if err != nil {
		return nil, err
	}
	return commit, nil
}

func (a *apiServer) BuildCommit(ctx context.Context, request *pfs.BuildCommitRequest) (response *pfs.Commit, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	commit, err := a.driver.buildCommit(a.env.GetPachClient(ctx),
		request.ID, request.Parent, request.Branch, request.Provenance, request.Tree)
	if err != nil {
		return nil, err
	}
	return commit, nil
}

func (a *apiServer) FinishCommit(ctx context.Context, request *pfs.FinishCommitRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	if err := a.driver.finishCommit(
		a.env.GetPachClient(ctx), request.Commit, request.Tree, request.Empty, request.Description,
	); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

func (a *apiServer) InspectCommit(ctx context.Context, request *pfs.InspectCommitRequest) (response *pfs.CommitInfo, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	return a.driver.inspectCommit(a.env.GetPachClient(ctx), request.Commit, request.BlockState)
}

func (a *apiServer) ListCommit(ctx context.Context, request *pfs.ListCommitRequest) (response *pfs.CommitInfos, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	commitInfos, err := a.driver.listCommit(a.env.GetPachClient(ctx), request.Repo, request.To, request.From, request.Number)
	if err != nil {
		return nil, err
	}
	return &pfs.CommitInfos{
		CommitInfo: commitInfos,
	}, nil
}

func (a *apiServer) ListCommitStream(req *pfs.ListCommitRequest, respServer pfs.API_ListCommitStreamServer) (retErr error) {
	func() { a.Log(req, nil, nil, 0) }()
	sent := 0
	defer func(start time.Time) {
		a.Log(req, fmt.Sprintf("stream containing %d commits", sent), retErr, time.Since(start))
	}(time.Now())
	pachClient := a.env.GetPachClient(respServer.Context())

	commitInfos, err := a.driver.listCommit(pachClient, req.Repo, req.To, req.From, req.Number)
	if err != nil {
		return err
	}
	for _, ci := range commitInfos {
		if err := respServer.Send(ci); err != nil {
			return err
		}
		sent++
	}
	return nil
}

func (a *apiServer) CreateBranch(ctx context.Context, request *pfs.CreateBranchRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	if err := a.driver.createBranch(a.env.GetPachClient(ctx), request.Branch, request.Head, request.Provenance); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

func (a *apiServer) InspectBranch(ctx context.Context, request *pfs.InspectBranchRequest) (response *pfs.BranchInfo, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	return a.driver.inspectBranch(a.env.GetPachClient(ctx), request.Branch)
}

func (a *apiServer) ListBranch(ctx context.Context, request *pfs.ListBranchRequest) (response *pfs.BranchInfos, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	branches, err := a.driver.listBranch(a.env.GetPachClient(ctx), request.Repo)
	if err != nil {
		return nil, err
	}
	return &pfs.BranchInfos{BranchInfo: branches}, nil
}

func (a *apiServer) DeleteBranch(ctx context.Context, request *pfs.DeleteBranchRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	if err := a.driver.deleteBranch(a.env.GetPachClient(ctx), request.Branch, request.Force); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

func (a *apiServer) DeleteCommit(ctx context.Context, request *pfs.DeleteCommitRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	if err := a.driver.deleteCommit(a.env.GetPachClient(ctx), request.Commit); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

func (a *apiServer) FlushCommit(request *pfs.FlushCommitRequest, stream pfs.API_FlushCommitServer) (retErr error) {
	ctx := stream.Context()
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	return a.driver.flushCommit(a.env.GetPachClient(ctx), request.Commits, request.ToRepos, stream.Send)
}

func (a *apiServer) SubscribeCommit(request *pfs.SubscribeCommitRequest, stream pfs.API_SubscribeCommitServer) (retErr error) {
	ctx := stream.Context()
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	return a.driver.subscribeCommit(a.env.GetPachClient(ctx), request.Repo, request.Branch, request.From, request.State, stream.Send)
}

func (a *apiServer) PutFile(putFileServer pfs.API_PutFileServer) (retErr error) {
	pachClient := a.env.GetPachClient(putFileServer.Context())
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
		case "pfs":
			return a.putFilePfs(pachClient, request, url)
		default:
			url, err := obj.ParseURL(request.Url)
			if err != nil {
				return fmt.Errorf("error parsing url %v: %v", request.Url, err)
			}
			objClient, err := obj.NewClientFromURLAndSecret(putFileServer.Context(), url)
			if err != nil {
				return err
			}
			return a.putFileObj(pachClient, objClient, request, url.Object)
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
	return a.driver.putFile(pachClient, request.File, request.Delimiter, request.TargetFileDatums, request.TargetFileBytes, request.OverwriteIndex, r)
}

func (a *apiServer) putFilePfs(pachClient *client.APIClient, request *pfs.PutFileRequest, url *url.URL) error {
	put := func(outPath string, inRepo string, inCommit string, inFile string) (retErr error) {
		r, err := pachClient.GetFileReader(inRepo, inCommit, inFile, 0, 0)
		if err != nil {
			return err
		}
		return a.driver.putFile(pachClient, client.NewFile(request.File.Commit.Repo.Name, request.File.Commit.ID, outPath), request.Delimiter, request.TargetFileDatums, request.TargetFileBytes, request.OverwriteIndex, r)
	}
	splitPath := strings.Split(strings.TrimPrefix(url.Path, "/"), "/")
	if len(splitPath) < 2 {
		return fmt.Errorf("pfs put-file path must be of form repo/commit[/path/to/file] got: %s", url.Path)
	}
	repo := splitPath[0]
	commit := splitPath[1]
	file := ""
	if len(splitPath) >= 3 {
		file = filepath.Join(splitPath[2:]...)
	}
	if request.Recursive {
		var eg errgroup.Group
		if err := pachClient.Walk(splitPath[0], commit, file, func(fileInfo *pfs.FileInfo) error {
			if fileInfo.FileType != pfs.FileType_FILE {
				return nil
			}
			eg.Go(func() error {
				return put(filepath.Join(request.File.Path, strings.TrimPrefix(fileInfo.File.Path, file)), repo, commit, fileInfo.File.Path)
			})
			return nil
		}); err != nil {
			return err
		}
		return eg.Wait()
	}
	return put(request.File.Path, repo, commit, file)
}

func (a *apiServer) putFileObj(pachClient *client.APIClient, objClient obj.Client, request *pfs.PutFileRequest, object string) (retErr error) {
	put := func(pachClient *client.APIClient, filePath string, objPath string) error {
		logRequest := &pfs.PutFileRequest{
			Delimiter: request.Delimiter,
			Url:       objPath,
			File: &pfs.File{
				Path: filePath,
			},
			Recursive: request.Recursive,
		}
		a.Log(logRequest, nil, nil, 0)
		defer func(start time.Time) {
			a.Log(logRequest, nil, retErr, time.Since(start))
		}(time.Now())
		r, err := objClient.Reader(objPath, 0, 0)
		if err != nil {
			return err
		}
		defer func() {
			if err := r.Close(); err != nil && retErr == nil {
				retErr = err
			}
		}()
		return a.driver.putFile(pachClient, client.NewFile(request.File.Commit.Repo.Name, request.File.Commit.ID, filePath),
			request.Delimiter, request.TargetFileDatums, request.TargetFileBytes, request.OverwriteIndex, r)
	}
	if request.Recursive {
		eg, egContext := errgroup.WithContext(pachClient.Ctx()) // preserves pachClient metadata (incl. auth)
		pachClient = pachClient.WithCtx(egContext)
		path := strings.TrimPrefix(object, "/")
		sem := make(chan struct{}, client.DefaultMaxConcurrentStreams)
		objClient.Walk(path, func(name string) error {
			eg.Go(func() error {
				sem <- struct{}{}
				defer func() {
					<-sem
				}()

				if strings.HasSuffix(name, "/") {
					// Amazon S3 supports objs w keys that end in a '/'
					// PFS needs to treat such a key as a directory.
					// In this case, we rely on the driver PutFile to
					// construct the 'directory' diffs from the file prefix
					logrus.Warnf("ambiguous key %v, not creating a directory or putting this entry as a file", name)
					return nil
				}
				return put(pachClient, filepath.Join(request.File.Path, strings.TrimPrefix(name, path)), name)
			})
			return nil
		})
		return eg.Wait()
	}
	// Joining Host and Path to retrieve the full path after "scheme://"
	return put(pachClient, request.File.Path, object)
}

func (a *apiServer) CopyFile(ctx context.Context, request *pfs.CopyFileRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	pachClient := a.env.GetPachClient(ctx)
	if err := a.driver.copyFile(pachClient, request.Src, request.Dst, request.Overwrite); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

func (a *apiServer) GetFile(request *pfs.GetFileRequest, apiGetFileServer pfs.API_GetFileServer) (retErr error) {
	ctx := apiGetFileServer.Context()
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	pachClient := a.env.GetPachClient(ctx)

	file, err := a.driver.getFile(pachClient, request.File, request.OffsetBytes, request.SizeBytes)
	if err != nil {
		return err
	}
	return grpcutil.WriteToStreamingBytesServer(file, apiGetFileServer)
}

func (a *apiServer) InspectFile(ctx context.Context, request *pfs.InspectFileRequest) (response *pfs.FileInfo, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	pachClient := a.env.GetPachClient(ctx)

	return a.driver.inspectFile(pachClient, request.File)
}

func (a *apiServer) ListFile(ctx context.Context, request *pfs.ListFileRequest) (response *pfs.FileInfos, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) {
		if response != nil && len(response.FileInfo) > client.MaxListItemsLog {
			logrus.Infof("Response contains %d objects; logging the first %d", len(response.FileInfo), client.MaxListItemsLog)
			a.Log(request, &pfs.FileInfos{response.FileInfo[:client.MaxListItemsLog]}, retErr, time.Since(start))
		} else {
			a.Log(request, response, retErr, time.Since(start))
		}
	}(time.Now())

	fileInfos, err := a.driver.listFile(a.env.GetPachClient(ctx), request.File, request.Full)
	if err != nil {
		return nil, err
	}
	return &pfs.FileInfos{
		FileInfo: fileInfos,
	}, nil
}

func (a *apiServer) ListFileStream(request *pfs.ListFileRequest, respServer pfs.API_ListFileStreamServer) (retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	var sent int
	defer func(start time.Time) {
		a.Log(request, fmt.Sprintf("response stream with %d objects", sent), retErr, time.Since(start))
	}(time.Now())
	fileInfos, err := a.driver.listFile(a.env.GetPachClient(respServer.Context()), request.File, request.Full)
	if err != nil {
		return err
	}
	for i := 0; i < len(fileInfos); i++ {
		if err := respServer.Send(fileInfos[i]); err != nil {
			return err
		}
		sent++
	}
	return nil
}

func (a *apiServer) GlobFile(ctx context.Context, request *pfs.GlobFileRequest) (response *pfs.FileInfos, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) {
		if response != nil && len(response.FileInfo) > client.MaxListItemsLog {
			logrus.Infof("Response contains %d objects; logging the first %d", len(response.FileInfo), client.MaxListItemsLog)
			a.Log(request, &pfs.FileInfos{response.FileInfo[:client.MaxListItemsLog]}, retErr, time.Since(start))
		} else {
			a.Log(request, response, retErr, time.Since(start))
		}
	}(time.Now())

	fileInfos, err := a.driver.globFile(a.env.GetPachClient(ctx), request.Commit, request.Pattern)
	if err != nil {
		return nil, err
	}
	return &pfs.FileInfos{
		FileInfo: fileInfos,
	}, nil
}

func (a *apiServer) GlobFileStream(request *pfs.GlobFileRequest, respServer pfs.API_GlobFileStreamServer) (retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	var sent int
	defer func(start time.Time) {
		a.Log(request, fmt.Sprintf("response stream with %d objects", sent), retErr, time.Since(start))
	}(time.Now())
	fileInfos, err := a.driver.globFile(a.env.GetPachClient(respServer.Context()), request.Commit, request.Pattern)
	if err != nil {
		return err
	}
	for i := 0; i < len(fileInfos); i++ {
		if err := respServer.Send(fileInfos[i]); err != nil {
			return err
		}
		sent++
	}
	return nil
}

func (a *apiServer) DiffFile(ctx context.Context, request *pfs.DiffFileRequest) (response *pfs.DiffFileResponse, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) {
		if response != nil && (len(response.NewFiles) > client.MaxListItemsLog || len(response.OldFiles) > client.MaxListItemsLog) {
			logrus.Infof("Response contains too many objects; truncating.")
			a.Log(request, &pfs.DiffFileResponse{
				NewFiles: truncateFiles(response.NewFiles),
				OldFiles: truncateFiles(response.OldFiles),
			}, retErr, time.Since(start))
		} else {
			a.Log(request, response, retErr, time.Since(start))
		}
	}(time.Now())
	newFileInfos, oldFileInfos, err := a.driver.diffFile(
		a.env.GetPachClient(ctx), request.NewFile, request.OldFile, request.Shallow)
	if err != nil {
		return nil, err
	}
	return &pfs.DiffFileResponse{
		NewFiles: newFileInfos,
		OldFiles: oldFileInfos,
	}, nil
}

func (a *apiServer) DeleteFile(ctx context.Context, request *pfs.DeleteFileRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	err := a.driver.deleteFile(a.env.GetPachClient(ctx), request.File)
	if err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

func (a *apiServer) DeleteAll(ctx context.Context, request *types.Empty) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())

	if err := a.driver.deleteAll(a.env.GetPachClient(ctx)); err != nil {
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
		if err != nil {
			return 0, err
		}
		//buffer.Write cannot error
		r.buffer.Write(request.Value)
	}
	return r.buffer.Read(p)
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

func truncateFiles(fileInfos []*pfs.FileInfo) []*pfs.FileInfo {
	if len(fileInfos) > client.MaxListItemsLog {
		return fileInfos[:client.MaxListItemsLog]
	}
	return fileInfos
}
