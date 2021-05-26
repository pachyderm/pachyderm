package server

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/v2/src/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/pfsload"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/metrics"
	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	"github.com/pachyderm/pachyderm/v2/src/internal/transactionenv/txncontext"
	"github.com/pachyderm/pachyderm/v2/src/internal/uuid"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs"
	"gopkg.in/yaml.v2"

	"golang.org/x/net/context"
)

// apiServer implements the public interface of the Pachyderm File System,
// including all RPCs defined in the protobuf spec.  Implementation details
// occur in the 'driver' code, and this layer serves to translate the protobuf
// request structures into normal function calls.
type apiServer struct {
	log.Logger
	driver *driver
	txnEnv *txnenv.TransactionEnv

	// env generates clients for pachyderm's downstream services
	env serviceenv.ServiceEnv
}

func newAPIServer(env serviceenv.ServiceEnv, txnEnv *txnenv.TransactionEnv, etcdPrefix string) (*apiServer, error) {
	d, err := newDriver(env, txnEnv, etcdPrefix)
	if err != nil {
		return nil, err
	}
	s := &apiServer{
		Logger: log.NewLogger("pfs.API", env.Logger()),
		driver: d,
		env:    env,
		txnEnv: txnEnv,
	}
	return s, nil
}

// ActivateAuth implements the protobuf pfs.ActivateAuth RPC
func (a *apiServer) ActivateAuth(ctx context.Context, request *pfs.ActivateAuthRequest) (response *pfs.ActivateAuthResponse, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	if err := a.txnEnv.WithWriteContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		return a.driver.activateAuth(txnCtx)
	}); err != nil {
		return nil, err
	}
	return &pfs.ActivateAuthResponse{}, nil
}

// CreateRepoInTransaction is identical to CreateRepo except that it can run
// inside an existing postgres transaction.  This is not an RPC.
func (a *apiServer) CreateRepoInTransaction(txnCtx *txncontext.TransactionContext, request *pfs.CreateRepoRequest) error {
	if repo := request.GetRepo(); repo != nil && repo.Name == fileSetsRepo {
		return errors.Errorf("%s is a reserved name", fileSetsRepo)
	}
	return a.driver.createRepo(txnCtx, request.Repo, request.Description, request.Update)
}

// CreateRepo implements the protobuf pfs.CreateRepo RPC
func (a *apiServer) CreateRepo(ctx context.Context, request *pfs.CreateRepoRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	if err := a.txnEnv.WithTransaction(ctx, func(txn txnenv.Transaction) error {
		return txn.CreateRepo(request)
	}); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

// InspectRepoInTransaction is identical to InspectRepo except that it can run
// inside an existing postgres transaction.  This is not an RPC.
func (a *apiServer) InspectRepoInTransaction(txnCtx *txncontext.TransactionContext, originalRequest *pfs.InspectRepoRequest) (*pfs.RepoInfo, error) {
	request := proto.Clone(originalRequest).(*pfs.InspectRepoRequest)
	return a.driver.inspectRepo(txnCtx, request.Repo, true)
}

// InspectRepo implements the protobuf pfs.InspectRepo RPC
func (a *apiServer) InspectRepo(ctx context.Context, request *pfs.InspectRepoRequest) (response *pfs.RepoInfo, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	var info *pfs.RepoInfo
	err := a.txnEnv.WithReadContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		var err error
		info, err = a.InspectRepoInTransaction(txnCtx, request)
		return err
	})
	if err != nil {
		return nil, err
	}
	size, err := a.driver.getRepoSize(ctx, info.Repo)
	if err != nil {
		return nil, err
	}
	info.SizeBytes = uint64(size)
	return info, nil
}

// ListRepo implements the protobuf pfs.ListRepo RPC
func (a *apiServer) ListRepo(ctx context.Context, request *pfs.ListRepoRequest) (response *pfs.ListRepoResponse, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	repoInfos, err := a.driver.listRepo(ctx, true, request.Type)
	return repoInfos, err
}

// DeleteRepoInTransaction is identical to DeleteRepo except that it can run
// inside an existing postgres transaction.  This is not an RPC.
func (a *apiServer) DeleteRepoInTransaction(txnCtx *txncontext.TransactionContext, request *pfs.DeleteRepoRequest) error {
	if request.All {
		return a.driver.deleteAll(txnCtx)
	}
	return a.driver.deleteRepo(txnCtx, request.Repo, request.Force)
}

// DeleteRepo implements the protobuf pfs.DeleteRepo RPC
func (a *apiServer) DeleteRepo(ctx context.Context, request *pfs.DeleteRepoRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	if err := a.txnEnv.WithTransaction(ctx, func(txn txnenv.Transaction) error {
		return txn.DeleteRepo(request)
	}); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

// StartCommitInTransaction is identical to StartCommit except that it can run
// inside an existing postgres transaction.  This is not an RPC.  The target
// commit can be specified but is optional.  This is so that the transaction can
// report the commit ID back to the client before the transaction has finished
// and it can be used in future commands inside the same transaction.
func (a *apiServer) StartCommitInTransaction(txnCtx *txncontext.TransactionContext, request *pfs.StartCommitRequest, commit *pfs.Commit) (*pfs.Commit, error) {
	id := ""
	if commit != nil {
		id = commit.ID
	}
	return a.driver.startCommit(txnCtx, id, request.Parent, request.Branch, request.Provenance, request.Description)
}

// StartCommit implements the protobuf pfs.StartCommit RPC
func (a *apiServer) StartCommit(ctx context.Context, request *pfs.StartCommitRequest) (response *pfs.Commit, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	var err error
	commit := &pfs.Commit{}
	if err = a.txnEnv.WithTransaction(ctx, func(txn txnenv.Transaction) error {
		commit, err = txn.StartCommit(request, nil)
		return err
	}); err != nil {
		return nil, err
	}
	return commit, nil
}

// FinishCommitInTransaction is identical to FinishCommit except that it can run
// inside an existing postgres transaction.  This is not an RPC.
func (a *apiServer) FinishCommitInTransaction(txnCtx *txncontext.TransactionContext, request *pfs.FinishCommitRequest) error {
	return metrics.ReportRequest(func() error {
		if request.Empty {
			request.Description += pfs.EmptyStr
		}
		return a.driver.finishCommit(txnCtx, request.Commit, request.Description)
	})
}

// FinishCommit implements the protobuf pfs.FinishCommit RPC
func (a *apiServer) FinishCommit(ctx context.Context, request *pfs.FinishCommitRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	if err := a.txnEnv.WithTransaction(ctx, func(txn txnenv.Transaction) error {
		return txn.FinishCommit(request)
	}); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

// InspectCommitInTransaction is identical to InspectCommit (some features excluded) except that it can run
// inside an existing postgres transaction.  This is not an RPC.
func (a *apiServer) InspectCommitInTransaction(txnCtx *txncontext.TransactionContext, request *pfs.InspectCommitRequest) (*pfs.CommitInfo, error) {
	return a.driver.resolveCommit(txnCtx.SqlTx, request.Commit)
}

// InspectCommit implements the protobuf pfs.InspectCommit RPC
func (a *apiServer) InspectCommit(ctx context.Context, request *pfs.InspectCommitRequest) (response *pfs.CommitInfo, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	return a.driver.inspectCommit(ctx, request.Commit, request.BlockState)
}

// ListCommit implements the protobuf pfs.ListCommit RPC
func (a *apiServer) ListCommit(request *pfs.ListCommitRequest, respServer pfs.API_ListCommitServer) (retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	sent := 0
	defer func(start time.Time) {
		a.Log(request, fmt.Sprintf("stream containing %d commits", sent), retErr, time.Since(start))
	}(time.Now())
	return a.driver.listCommit(respServer.Context(), request.Repo, request.To, request.From, request.Number, request.Reverse, func(ci *pfs.CommitInfo) error {
		sent++
		return respServer.Send(ci)
	})
}

// SquashCommitInTransaction is identical to SquashCommit except that it can run
// inside an existing postgres transaction.  This is not an RPC.
func (a *apiServer) SquashCommitInTransaction(txnCtx *txncontext.TransactionContext, request *pfs.SquashCommitRequest) error {
	return a.driver.squashCommit(txnCtx, request.Commit)
}

// SquashCommit implements the protobuf pfs.SquashCommit RPC
func (a *apiServer) SquashCommit(ctx context.Context, request *pfs.SquashCommitRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	if err := a.txnEnv.WithTransaction(ctx, func(txn txnenv.Transaction) error {
		return txn.SquashCommit(request)
	}); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

// FlushCommit implements the protobuf pfs.FlushCommit RPC
func (a *apiServer) FlushCommit(request *pfs.FlushCommitRequest, stream pfs.API_FlushCommitServer) (retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	return a.driver.flushCommit(stream.Context(), request.Commits, request.ToRepos, stream.Send)
}

// SubscribeCommit implements the protobuf pfs.SubscribeCommit RPC
func (a *apiServer) SubscribeCommit(request *pfs.SubscribeCommitRequest, stream pfs.API_SubscribeCommitServer) (retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	return a.driver.subscribeCommit(stream.Context(), request.Repo, request.Branch, request.Prov, request.From, request.State, stream.Send)
}

// ClearCommit deletes all data in the commit.
func (a *apiServer) ClearCommit(ctx context.Context, request *pfs.ClearCommitRequest) (_ *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	return &types.Empty{}, a.driver.clearCommit(ctx, request.Commit)
}

// CreateBranchInTransaction is identical to CreateBranch except that it can run
// inside an existing postgres transaction.  This is not an RPC.
func (a *apiServer) CreateBranchInTransaction(txnCtx *txncontext.TransactionContext, request *pfs.CreateBranchRequest) error {
	return a.driver.createBranch(txnCtx, request.Branch, request.Head, request.Provenance, request.Trigger)
}

// CreateBranch implements the protobuf pfs.CreateBranch RPC
func (a *apiServer) CreateBranch(ctx context.Context, request *pfs.CreateBranchRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	if err := a.txnEnv.WithTransaction(ctx, func(txn txnenv.Transaction) error {
		return txn.CreateBranch(request)
	}); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

// InspectBranch implements the protobuf pfs.InspectBranch RPC
func (a *apiServer) InspectBranch(ctx context.Context, request *pfs.InspectBranchRequest) (response *pfs.BranchInfo, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	branchInfo := &pfs.BranchInfo{}
	if err := a.txnEnv.WithReadContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		var err error
		branchInfo, err = a.driver.inspectBranch(txnCtx, request.Branch)
		return err
	}); err != nil {
		return nil, err
	}
	return branchInfo, nil
}

func (a *apiServer) InspectBranchInTransaction(txnCtx *txncontext.TransactionContext, request *pfs.InspectBranchRequest) (*pfs.BranchInfo, error) {
	return a.driver.inspectBranch(txnCtx, request.Branch)
}

// ListBranch implements the protobuf pfs.ListBranch RPC
func (a *apiServer) ListBranch(ctx context.Context, request *pfs.ListBranchRequest) (response *pfs.BranchInfos, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	branches, err := a.driver.listBranch(ctx, request.Repo, request.Reverse)
	if err != nil {
		return nil, err
	}
	return &pfs.BranchInfos{BranchInfo: branches}, nil
}

// DeleteBranchInTransaction is identical to DeleteBranch except that it can run
// inside an existing postgres transaction.  This is not an RPC.
func (a *apiServer) DeleteBranchInTransaction(txnCtx *txncontext.TransactionContext, request *pfs.DeleteBranchRequest) error {
	return a.driver.deleteBranch(txnCtx, request.Branch, request.Force)
}

// DeleteBranch implements the protobuf pfs.DeleteBranch RPC
func (a *apiServer) DeleteBranch(ctx context.Context, request *pfs.DeleteBranchRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	if err := a.txnEnv.WithTransaction(ctx, func(txn txnenv.Transaction) error {
		return txn.DeleteBranch(request)
	}); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

func (a *apiServer) ModifyFile(server pfs.API_ModifyFileServer) (retErr error) {
	request, err := server.Recv()
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	return metrics.ReportRequestWithThroughput(func() (int64, error) {
		if err != nil {
			return 0, err
		}
		var bytesRead int64
		if err := a.driver.modifyFile(server.Context(), request.Commit, func(uw *fileset.UnorderedWriter) error {
			var err error
			bytesRead, err = a.modifyFile(server.Context(), uw, server, request)
			return err
		}); err != nil {
			return bytesRead, err
		}
		return bytesRead, server.SendAndClose(&types.Empty{})
	})
}

type modifyFileSource interface {
	Recv() (*pfs.ModifyFileRequest, error)
}

func (a *apiServer) modifyFile(ctx context.Context, uw *fileset.UnorderedWriter, server modifyFileSource, req *pfs.ModifyFileRequest) (int64, error) {
	var bytesRead int64
	for {
		// TODO Validation.
		if req != nil && req.Modification != nil {
			switch mod := req.Modification.(type) {
			case *pfs.ModifyFileRequest_PutFile:
				var err error
				var n int64
				switch mod.PutFile.Source.(type) {
				case *pfs.PutFile_RawFileSource:
					n, err = putFileRaw(uw, server, mod.PutFile)
				case *pfs.PutFile_TarFileSource:
					n, err = putFileTar(uw, server, mod.PutFile)
				case *pfs.PutFile_UrlFileSource:
					n, err = putFileURL(ctx, uw, mod.PutFile)
				}
				if err != nil {
					return bytesRead, err
				}
				bytesRead += n
			case *pfs.ModifyFileRequest_DeleteFile:
				if err := deleteFile(uw, mod.DeleteFile); err != nil {
					return bytesRead, err
				}
			case *pfs.ModifyFileRequest_CopyFile:
				cf := mod.CopyFile
				if err := a.driver.copyFile(ctx, uw, cf.Dst, cf.Src, cf.Append, cf.Tag); err != nil {
					return bytesRead, err
				}
			}
		}
		var err error
		req, err = server.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return bytesRead, nil
			}
			return bytesRead, err
		}
	}
}

func putFileTar(uw *fileset.UnorderedWriter, server modifyFileSource, req *pfs.PutFile) (int64, error) {
	src := req.Source.(*pfs.PutFile_TarFileSource).TarFileSource
	tfsr := &tarFileSourceReader{
		server: server,
		r:      bytes.NewReader(src.Data),
	}
	tr := tar.NewReader(tfsr)
	for {
		hdr, err := tr.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return tfsr.bytesRead, nil
			}
			return tfsr.bytesRead, err
		}
		if hdr.Typeflag == tar.TypeDir {
			continue
		}
		if err := uw.Put(hdr.Name, req.Tag, req.Append, tr); err != nil {
			return tfsr.bytesRead, err
		}
	}
}

type tarFileSourceReader struct {
	server    modifyFileSource
	r         *bytes.Reader
	bytesRead int64
}

func (tfsr *tarFileSourceReader) Read(data []byte) (int, error) {
	for tfsr.r.Len() == 0 {
		req, err := tfsr.server.Recv()
		if err != nil {
			return 0, err
		}
		mod := req.Modification.(*pfs.ModifyFileRequest_PutFile).PutFile
		src := mod.Source.(*pfs.PutFile_TarFileSource).TarFileSource
		tfsr.r = bytes.NewReader(src.Data)
	}
	n, err := tfsr.r.Read(data)
	tfsr.bytesRead += int64(n)
	return n, err
}

// TODO: Collect and return bytes read and figure out parallel download (task chain in chunk package might be helpful).
func putFileURL(ctx context.Context, uw *fileset.UnorderedWriter, req *pfs.PutFile) (_ int64, retErr error) {
	src := req.Source.(*pfs.PutFile_UrlFileSource).UrlFileSource
	url, err := url.Parse(src.URL)
	if err != nil {
		return 0, err
	}
	switch url.Scheme {
	case "http":
		fallthrough
	case "https":
		resp, err := http.Get(src.URL)
		if err != nil {
			return 0, err
		} else if resp.StatusCode >= 400 {
			return 0, errors.Errorf("error retrieving content from %q: %s", src.URL, resp.Status)
		}
		defer func() {
			if err := resp.Body.Close(); retErr == nil {
				retErr = err
			}
		}()
		return 0, uw.Put(src.Path, req.Tag, req.Append, resp.Body)
	default:
		url, err := obj.ParseURL(src.URL)
		if err != nil {
			return 0, errors.Wrapf(err, "error parsing url %v", src.URL)
		}
		objClient, err := obj.NewClientFromURLAndSecret(url, false)
		if err != nil {
			return 0, err
		}
		if src.Recursive {
			path := strings.TrimPrefix(url.Object, "/")
			return 0, objClient.Walk(ctx, path, func(name string) error {
				return obj.WithPipe(func(w io.Writer) error {
					return objClient.Get(ctx, name, w)
				}, func(r io.Reader) error {
					return uw.Put(filepath.Join(src.Path, strings.TrimPrefix(name, path)), req.Tag, req.Append, r)
				})
			})
		}
		return 0, obj.WithPipe(func(w io.Writer) error {
			return objClient.Get(ctx, url.Object, w)
		}, func(r io.Reader) error {
			return uw.Put(src.Path, req.Tag, req.Append, r)
		})
	}
}

func putFileRaw(uw *fileset.UnorderedWriter, server modifyFileSource, req *pfs.PutFile) (int64, error) {
	src := req.Source.(*pfs.PutFile_RawFileSource).RawFileSource
	rfsr := &rawFileSourceReader{
		server: server,
		r:      bytes.NewReader(src.Data),
		done:   src.EOF,
	}
	err := uw.Put(src.Path, req.Tag, req.Append, rfsr)
	return rfsr.bytesRead, err
}

type rawFileSourceReader struct {
	server    modifyFileSource
	r         *bytes.Reader
	bytesRead int64
	done      bool
}

func (rfsr *rawFileSourceReader) Read(data []byte) (int, error) {
	for !rfsr.done && rfsr.r.Len() == 0 {
		req, err := rfsr.server.Recv()
		if err != nil {
			return 0, err
		}
		mod := req.Modification.(*pfs.ModifyFileRequest_PutFile).PutFile
		src := mod.Source.(*pfs.PutFile_RawFileSource).RawFileSource
		rfsr.r = bytes.NewReader(src.Data)
		rfsr.done = src.EOF
	}
	n, err := rfsr.r.Read(data)
	rfsr.bytesRead += int64(n)
	return n, err
}

func deleteFile(uw *fileset.UnorderedWriter, request *pfs.DeleteFile) error {
	uw.Delete(request.File, request.Tag)
	return nil
}

// GetTAR implements the protobuf pfs.GetFile RPC
func (a *apiServer) GetTAR(request *pfs.GetFileRequest, server pfs.API_GetTARServer) (retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	return metrics.ReportRequestWithThroughput(func() (int64, error) {
		ctx := server.Context()
		src, err := a.driver.getFile(ctx, request.File)
		if err != nil {
			return 0, err
		}
		if request.URL != "" {
			return getFileURL(ctx, request.URL, src)
		}
		var bytesWritten int64
		err = grpcutil.WithStreamingBytesWriter(server, func(w io.Writer) error {
			var err error
			bytesWritten, err = withGetFileWriter(w, func(w io.Writer) error {
				return getFileTar(ctx, w, src)
			})
			return err
		})
		return bytesWritten, err

	})
}

// TODO: Parallelize and decide on appropriate config.
func getFileURL(ctx context.Context, URL string, src Source) (int64, error) {
	parsedURL, err := obj.ParseURL(URL)
	if err != nil {
		return 0, err
	}
	objClient, err := obj.NewClientFromURLAndSecret(parsedURL, false)
	if err != nil {
		return 0, err
	}
	var bytesWritten int64
	err = src.Iterate(ctx, func(fi *pfs.FileInfo, file fileset.File) (retErr error) {
		if fi.FileType != pfs.FileType_FILE {
			return nil
		}
		if err := obj.WithPipe(func(w io.Writer) error {
			return file.Content(w)
		}, func(r io.Reader) error {
			return objClient.Put(ctx, filepath.Join(parsedURL.Object, fi.File.Path), r)
		}); err != nil {
			return err
		}
		bytesWritten += int64(fi.SizeBytes)
		return nil
	})
	return bytesWritten, err
}

func withGetFileWriter(w io.Writer, cb func(io.Writer) error) (int64, error) {
	gfw := &getFileWriter{w: w}
	err := cb(gfw)
	return gfw.bytesWritten, err
}

type getFileWriter struct {
	w            io.Writer
	bytesWritten int64
}

func (gfw *getFileWriter) Write(data []byte) (int, error) {
	n, err := gfw.w.Write(data)
	gfw.bytesWritten += int64(n)
	return n, err
}

func getFileTar(ctx context.Context, w io.Writer, src Source) error {
	// TODO: remove absolute paths on the way out?
	// nonAbsolute := &fileset.HeaderMapper{
	// 	R: filter,
	// 	F: func(th *tar.Header) *tar.Header {
	// 		th.Name = "." + th.Name
	// 		return th
	// 	},
	// }
	if err := src.Iterate(ctx, func(fi *pfs.FileInfo, file fileset.File) error {
		return fileset.WriteTarEntry(w, file)
	}); err != nil {
		return err
	}
	return tar.NewWriter(w).Close()
}

// InspectFile implements the protobuf pfs.InspectFile RPC
func (a *apiServer) InspectFile(ctx context.Context, request *pfs.InspectFileRequest) (response *pfs.FileInfo, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	return a.driver.inspectFile(ctx, request.File)
}

// ListFile implements the protobuf pfs.ListFile RPC
func (a *apiServer) ListFile(request *pfs.ListFileRequest, server pfs.API_ListFileServer) (retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	var sent int
	defer func(start time.Time) {
		a.Log(request, fmt.Sprintf("response stream with %d objects", sent), retErr, time.Since(start))
	}(time.Now())
	return a.driver.listFile(server.Context(), request.File, request.Full, func(fi *pfs.FileInfo) error {
		sent++
		return server.Send(fi)
	})
}

// WalkFile implements the protobuf pfs.WalkFile RPC
func (a *apiServer) WalkFile(request *pfs.WalkFileRequest, server pfs.API_WalkFileServer) (retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	var sent int
	defer func(start time.Time) {
		a.Log(request, fmt.Sprintf("response stream with %d objects", sent), retErr, time.Since(start))
	}(time.Now())
	return a.driver.walkFile(server.Context(), request.File, func(fi *pfs.FileInfo) error {
		sent++
		return server.Send(fi)
	})
}

// GlobFile implements the protobuf pfs.GlobFile RPC
func (a *apiServer) GlobFile(request *pfs.GlobFileRequest, respServer pfs.API_GlobFileServer) (retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	var sent int
	defer func(start time.Time) {
		a.Log(request, fmt.Sprintf("response stream with %d objects", sent), retErr, time.Since(start))
	}(time.Now())
	return a.driver.globFile(respServer.Context(), request.Commit, request.Pattern, func(fi *pfs.FileInfo) error {
		sent++
		return respServer.Send(fi)
	})
}

// DiffFile implements the protobuf pfs.DiffFile RPC
func (a *apiServer) DiffFile(request *pfs.DiffFileRequest, server pfs.API_DiffFileServer) (retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	var sent int
	defer func(start time.Time) {
		a.Log(request, fmt.Sprintf("response stream with %d objects", sent), retErr, time.Since(start))
	}(time.Now())
	return a.driver.diffFile(server.Context(), request.OldFile, request.NewFile, func(oldFi, newFi *pfs.FileInfo) error {
		sent++
		return server.Send(&pfs.DiffFileResponse{
			OldFile: oldFi,
			NewFile: newFi,
		})
	})
}

// DeleteAll implements the protobuf pfs.DeleteAll RPC
func (a *apiServer) DeleteAll(ctx context.Context, request *types.Empty) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	err := a.txnEnv.WithWriteContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		return a.driver.deleteAll(txnCtx)
	})
	if err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

// Fsckimplements the protobuf pfs.Fsck RPC
func (a *apiServer) Fsck(request *pfs.FsckRequest, fsckServer pfs.API_FsckServer) (retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	sent := 0
	defer func(start time.Time) {
		a.Log(request, fmt.Sprintf("stream containing %d messages", sent), retErr, time.Since(start))
	}(time.Now())
	if err := a.driver.fsck(fsckServer.Context(), request.Fix, func(resp *pfs.FsckResponse) error {
		sent++
		return fsckServer.Send(resp)
	}); err != nil {
		return err
	}
	return nil
}

// CreateFileset implements the pfs.CreateFileset RPC
func (a *apiServer) CreateFileset(server pfs.API_CreateFilesetServer) error {
	fsID, err := a.driver.createFileset(server.Context(), func(uw *fileset.UnorderedWriter) error {
		_, err := a.modifyFile(server.Context(), uw, server, nil)
		return err
	})
	if err != nil {
		return err
	}
	return server.SendAndClose(&pfs.CreateFilesetResponse{
		FilesetId: fsID.HexString(),
	})
}

func (a *apiServer) GetFileset(ctx context.Context, req *pfs.GetFilesetRequest) (*pfs.CreateFilesetResponse, error) {
	filesetID, err := a.driver.getFileset(ctx, req.Commit)
	if err != nil {
		return nil, err
	}
	return &pfs.CreateFilesetResponse{
		FilesetId: filesetID.HexString(),
	}, nil
}

func (a *apiServer) AddFileset(ctx context.Context, req *pfs.AddFilesetRequest) (*types.Empty, error) {
	if err := a.txnEnv.WithWriteContext(ctx, func(txnCtx *txncontext.TransactionContext) error {
		return a.AddFilesetInTransaction(txnCtx, req)
	}); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

func (a *apiServer) AddFilesetInTransaction(txnCtx *txncontext.TransactionContext, request *pfs.AddFilesetRequest) error {
	fsid, err := fileset.ParseID(request.FilesetId)
	if err != nil {
		return err
	}
	if err := a.driver.addFileset(txnCtx, request.Commit, *fsid); err != nil {
		return err
	}
	return nil
}

// RenewFileset implements the pfs.RenewFileset RPC
func (a *apiServer) RenewFileset(ctx context.Context, req *pfs.RenewFilesetRequest) (*types.Empty, error) {
	fsid, err := fileset.ParseID(req.FilesetId)
	if err != nil {
		return nil, err
	}
	if err := a.driver.renewFileset(ctx, *fsid, time.Duration(req.TtlSeconds)*time.Second); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

// RunLoadTest implements the pfs.RunLoadTest RPC
func (a *apiServer) RunLoadTest(ctx context.Context, req *pfs.RunLoadTestRequest) (_ *pfs.RunLoadTestResponse, retErr error) {
	func() { a.Log(req, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(req, nil, retErr, time.Since(start)) }(time.Now())
	pachClient := a.env.GetPachClient(ctx)
	repo := "load_test"
	if err := pachClient.CreateRepo(repo); err != nil && !pfsserver.IsRepoExistsErr(err) {
		return nil, err
	}
	branch := uuid.New()
	if err := pachClient.CreateBranch(repo, branch, "", "", nil); err != nil {
		return nil, err
	}
	seed := time.Now().UTC().UnixNano()
	if req.Seed > 0 {
		seed = req.Seed
	}
	resp := &pfs.RunLoadTestResponse{
		Branch: client.NewBranch(repo, branch),
		Seed:   seed,
	}
	if err := a.runLoadTest(pachClient, resp.Branch, req.Spec, seed); err != nil {
		resp.Error = err.Error()
	}
	return resp, nil
}

func (a *apiServer) runLoadTest(pachClient *client.APIClient, branch *pfs.Branch, specBytes []byte, seed int64) error {
	spec := &pfsload.CommitsSpec{}
	if err := yaml.UnmarshalStrict(specBytes, spec); err != nil {
		return err
	}
	return pfsload.Commits(pachClient, branch.Repo.Name, branch.Name, spec, seed)
}
