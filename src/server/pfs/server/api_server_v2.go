package server

import (
	"bufio"
	"bytes"
	"io"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/hashtree"
	"github.com/pachyderm/pachyderm/src/server/pkg/serviceenv"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/metrics"
	txnenv "github.com/pachyderm/pachyderm/src/server/pkg/transactionenv"
	"golang.org/x/net/context"
)

var _ pfs.APIServer = &apiServerV2{}

type apiServerV2 struct {
	*apiServer
	driver *driverV2
}

func newAPIServerV2(
	env *serviceenv.ServiceEnv,
	txnEnv *txnenv.TransactionEnv,
	etcdPrefix string,
	treeCache *hashtree.Cache,
	storageRoot string,
	memoryRequest int64,
) (*apiServerV2, error) {
	s1, err := newAPIServer(env, txnEnv, etcdPrefix, treeCache, storageRoot, memoryRequest)
	if err != nil {
		return nil, err
	}
	d, err := newDriverV2(env, txnEnv, etcdPrefix, treeCache, storageRoot, memoryRequest)
	if err != nil {
		return nil, err
	}
	s2 := &apiServerV2{
		apiServer: s1,
		driver:    d,
	}
	go func() { s2.env.GetPachClient(context.Background()) }() // Begin dialing connection on startup
	return s2, nil
}

func (a *apiServerV2) FileOperationV2(server pfs.API_FileOperationV2Server) (retErr error) {
	req, err := server.Recv()
	func() { a.Log(req, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(req, nil, retErr, time.Since(start)) }(time.Now())
	return metrics.ReportRequestWithThroughput(func() (int64, error) {
		if err != nil {
			return 0, err
		}
		repo := req.Commit.Repo.Name
		commit := req.Commit.ID
		var bytesRead int64
		if err := a.driver.withFileSet(server.Context(), repo, commit, func(fs *fileset.FileSet) error {
			for {
				req, err := server.Recv()
				if err != nil {
					if errors.Is(err, io.EOF) {
						return nil
					}
					return err
				}
				// TODO Validation.
				switch op := req.Operation.(type) {
				case *pfs.FileOperationRequestV2_PutTar:
					n, err := putTar(fs, server, op.PutTar)
					bytesRead += n
					if err != nil {
						return err
					}
				case *pfs.FileOperationRequestV2_DeleteFiles:
					if err := deleteFiles(fs, op.DeleteFiles); err != nil {
						return err
					}
				}
			}
		}); err != nil {
			return bytesRead, err
		}
		return bytesRead, server.SendAndClose(&types.Empty{})
	})
}

func putTar(fs *fileset.FileSet, server pfs.API_FileOperationV2Server, req *pfs.PutTarRequestV2) (int64, error) {
	ptr := &putTarReader{
		server: server,
		r:      bytes.NewReader(req.Data),
	}
	err := fs.Put(ptr, req.Tag)
	return ptr.bytesRead, err
}

type putTarReader struct {
	server    pfs.API_FileOperationV2Server
	r         *bytes.Reader
	bytesRead int64
}

func (ptr *putTarReader) Read(data []byte) (int, error) {
	if ptr.r.Len() == 0 {
		req, err := ptr.server.Recv()
		if err != nil {
			return 0, err
		}
		op := req.Operation.(*pfs.FileOperationRequestV2_PutTar)
		putTarReq := op.PutTar
		if putTarReq.EOF {
			return 0, io.EOF
		}
		ptr.r = bytes.NewReader(putTarReq.Data)
	}
	n, err := ptr.r.Read(data)
	ptr.bytesRead += int64(n)
	return n, err
}

func deleteFiles(fs *fileset.FileSet, req *pfs.DeleteFilesRequestV2) error {
	for _, file := range req.Files {
		fs.Delete(file, req.Tag)
	}
	return nil
}

func (a *apiServerV2) GetTarV2(request *pfs.GetTarRequestV2, server pfs.API_GetTarV2Server) (retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	return metrics.ReportRequestWithThroughput(func() (int64, error) {
		commit := request.File.Commit
		glob := request.File.Path
		gtw := newGetTarWriter(grpcutil.NewStreamingBytesWriter(server))
		err := a.driver.getTar(server.Context(), commit, glob, gtw)
		return gtw.bytesWritten, err
	})
}

type getTarWriter struct {
	w            io.Writer
	bytesWritten int64
}

func newGetTarWriter(w io.Writer) *getTarWriter {
	return &getTarWriter{w: w}
}

func (gtw *getTarWriter) Write(data []byte) (int, error) {
	n, err := gtw.w.Write(data)
	gtw.bytesWritten += int64(n)
	return n, err
}

func (a *apiServerV2) GetTarConditionalV2(server pfs.API_GetTarConditionalV2Server) (retErr error) {
	request, err := server.Recv()
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	return metrics.ReportRequestWithThroughput(func() (int64, error) {
		if err != nil {
			return 0, err
		}
		if !a.env.StorageV2 {
			return 0, errors.Errorf("new storage layer disabled")
		}
		repo := request.File.Commit.Repo.Name
		commit := request.File.Commit.ID
		glob := request.File.Path
		var bytesWritten int64
		err := a.driver.getTarConditional(server.Context(), repo, commit, glob, func(fr *FileReader) error {
			if err := server.Send(&pfs.GetTarConditionalResponseV2{FileInfo: fr.Info()}); err != nil {
				return err
			}
			req, err := server.Recv()
			if err != nil {
				return err
			}
			// Skip to the next file (client does not want the file content).
			if req.Skip {
				return nil
			}
			gtcw := newGetTarConditionalWriter(server)
			defer func() {
				bytesWritten += gtcw.bytesWritten
			}()
			w := bufio.NewWriterSize(gtcw, grpcutil.MaxMsgSize)
			if err := fr.Get(w); err != nil {
				return err
			}
			if err := w.Flush(); err != nil {
				return err
			}
			return server.Send(&pfs.GetTarConditionalResponseV2{EOF: true})
		})
		return bytesWritten, err
	})
}

type getTarConditionalWriter struct {
	server       pfs.API_GetTarConditionalV2Server
	bytesWritten int64
}

func newGetTarConditionalWriter(server pfs.API_GetTarConditionalV2Server) *getTarConditionalWriter {
	return &getTarConditionalWriter{
		server: server,
	}
}

func (w *getTarConditionalWriter) Write(data []byte) (int, error) {
	if err := w.server.Send(&pfs.GetTarConditionalResponseV2{Data: data}); err != nil {
		return 0, err
	}
	w.bytesWritten += int64(len(data))
	return len(data), nil
}

func (a *apiServerV2) ListFileV2(req *pfs.ListFileRequest, server pfs.API_ListFileV2Server) error {
	pachClient := a.env.GetPachClient(server.Context())
	return a.driver.listFileV2(pachClient, req.File, req.Full, req.History, func(finfo *pfs.FileInfoV2) error {
		return server.Send(finfo)
	})
}

// FinishCommitInTransaction is identical to FinishCommit except that it can run
// inside an existing etcd STM transaction.  This is not an RPC.
func (a *apiServerV2) FinishCommitInTransaction(
	txnCtx *txnenv.TransactionContext,
	request *pfs.FinishCommitRequest,
) error {
	return metrics.ReportRequest(func() error {
		return a.driver.finishCommitV2(txnCtx, request.Commit, request.Description)
	})
}

func (a *apiServerV2) GlobFileV2(request *pfs.GlobFileRequest, server pfs.API_GlobFileV2Server) (retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	return a.driver.globFileV2(a.env.GetPachClient(server.Context()), request.Commit, request.Pattern, func(fi *pfs.FileInfoV2) error {
		return server.Send(fi)
	})
}

// CopyFile implements the protobuf pfs.CopyFile RPC
func (a *apiServerV2) CopyFile(ctx context.Context, request *pfs.CopyFileRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	if err := a.driver.copyFile(a.env.GetPachClient(ctx), request.Src, request.Dst, request.Overwrite); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

// InspectFileV2 returns info about a file.
func (a *apiServerV2) InspectFileV2(ctx context.Context, req *pfs.InspectFileRequest) (*pfs.FileInfoV2, error) {
	return a.driver.inspectFile(a.env.GetPachClient(ctx), req.File)
}

// WalkFileV2 walks over all the files under a directory, including children of children.
func (a *apiServerV2) WalkFileV2(req *pfs.WalkFileRequest, server pfs.API_WalkFileV2Server) error {
	pachClient := a.env.GetPachClient(server.Context())
	return a.driver.walkFile(pachClient, req.File, func(fi *pfs.FileInfoV2) error {
		return server.Send(fi)
	})
}

// DeleteRepoInTransaction is identical to DeleteRepo except that it can run
// inside an existing etcd STM transaction.  This is not an RPC.
func (a *apiServerV2) DeleteRepoInTransaction(
	txnCtx *txnenv.TransactionContext,
	request *pfs.DeleteRepoRequest,
) error {
	if request.All {
		return a.driver.deleteAll(txnCtx)
	}
	return a.driver.deleteRepo(txnCtx, request.Repo, request.Force)
}

// DeleteCommitInTransaction is identical to DeleteCommit except that it can run
// inside an existing etcd STM transaction.  This is not an RPC.
func (a *apiServerV2) DeleteCommitInTransaction(
	txnCtx *txnenv.TransactionContext,
	request *pfs.DeleteCommitRequest,
) error {
	return a.driver.deleteCommit(txnCtx, request.Commit)
}

// DeleteAll implements the protobuf pfs.DeleteAll RPC
func (a *apiServerV2) DeleteAll(ctx context.Context, request *types.Empty) (response *types.Empty, retErr error) {
	err := a.txnEnv.WithWriteContext(ctx, func(txnCtx *txnenv.TransactionContext) error {
		return a.driver.deleteAll(txnCtx)
	})
	if err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

var errV1NotImplemented = errors.Errorf("V1 method not implemented")

// BuildCommit is not implemented in V2.
func (a *apiServerV2) BuildCommit(_ context.Context, _ *pfs.BuildCommitRequest) (*pfs.Commit, error) {
	return nil, errV1NotImplemented
}

// Fsck is not implemented in V2.
func (a *apiServerV2) Fsck(_ *pfs.FsckRequest, _ pfs.API_FsckServer) error {
	return errV1NotImplemented
}

// PutFile is not implemented in V2.
func (a *apiServerV2) PutFile(_ pfs.API_PutFileServer) error {
	return errV1NotImplemented
}

// ListFileStream is not implemented in V2.
func (a *apiServerV2) ListFileStream(_ *pfs.ListFileRequest, _ pfs.API_ListFileStreamServer) error {
	return errV1NotImplemented
}

// GlobFileStream is not implemented in V2.
func (a *apiServerV2) GlobFileStream(_ *pfs.GlobFileRequest, _ pfs.API_GlobFileStreamServer) error {
	return errV1NotImplemented
}
