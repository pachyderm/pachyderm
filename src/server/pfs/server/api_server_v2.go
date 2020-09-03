package server

import (
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

var errV1NotImplemented = errors.Errorf("V1 method not implemented")

type apiServerV2 struct {
	*apiServer
	driver *driverV2
}

func newAPIServerV2(env *serviceenv.ServiceEnv, txnEnv *txnenv.TransactionEnv, etcdPrefix string, treeCache *hashtree.Cache, storageRoot string, memoryRequest int64) (*apiServerV2, error) {
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

// DeleteRepoInTransaction is identical to DeleteRepo except that it can run
// inside an existing etcd STM transaction.  This is not an RPC.
func (a *apiServerV2) DeleteRepoInTransaction(txnCtx *txnenv.TransactionContext, request *pfs.DeleteRepoRequest) error {
	if request.All {
		return a.driver.deleteAll(txnCtx)
	}
	return a.driver.deleteRepo(txnCtx, request.Repo, request.Force)
}

// FinishCommitInTransaction is identical to FinishCommit except that it can run
// inside an existing etcd STM transaction.  This is not an RPC.
func (a *apiServerV2) FinishCommitInTransaction(txnCtx *txnenv.TransactionContext, request *pfs.FinishCommitRequest) error {
	return metrics.ReportRequest(func() error {
		return a.driver.finishCommitV2(txnCtx, request.Commit, request.Description)
	})
}

// DeleteCommitInTransaction is identical to DeleteCommit except that it can run
// inside an existing etcd STM transaction.  This is not an RPC.
func (a *apiServerV2) DeleteCommitInTransaction(txnCtx *txnenv.TransactionContext, request *pfs.DeleteCommitRequest) error {
	return a.driver.deleteCommit(txnCtx, request.Commit)
}

// BuildCommit is not implemented in V2.
func (a *apiServerV2) BuildCommit(_ context.Context, _ *pfs.BuildCommitRequest) (*pfs.Commit, error) {
	return nil, errV1NotImplemented
}

// PutFile is not implemented in V2.
func (a *apiServerV2) PutFile(_ pfs.API_PutFileServer) error {
	return errV1NotImplemented
}

// CopyFile implements the protobuf pfs.CopyFile RPC
func (a *apiServerV2) CopyFile(ctx context.Context, request *pfs.CopyFileRequest) (response *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	if err := a.driver.copyFile(a.env.GetPachClient(ctx), request.Src, request.Dst, request.Overwrite); err != nil {
		return nil, err
	}
	return &types.Empty{}, nil
}

// GetFile is not implemented in V2.
func (a *apiServerV2) GetFile(_ *pfs.GetFileRequest, _ pfs.API_GetFileServer) error {
	return errV1NotImplemented
}

// InspectFile implements the protobuf pfs.InspectFile RPC
func (a *apiServerV2) InspectFile(ctx context.Context, request *pfs.InspectFileRequest) (response *pfs.FileInfo, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, response, retErr, time.Since(start)) }(time.Now())
	return a.driver.inspectFile(a.env.GetPachClient(ctx), request.File)
}

// ListFile is not implemented in V2.
func (a *apiServerV2) ListFile(_ context.Context, _ *pfs.ListFileRequest) (*pfs.FileInfos, error) {
	return nil, errV1NotImplemented
}

// ListFileStream implements the protobuf pfs.ListFileStream RPC
func (a *apiServerV2) ListFileStream(request *pfs.ListFileRequest, server pfs.API_ListFileStreamServer) (retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	pachClient := a.env.GetPachClient(server.Context())
	return a.driver.listFileV2(pachClient, request.File, request.Full, request.History, func(fi *pfs.FileInfo) error {
		return server.Send(fi)
	})
}

// WalkFile implements the protobuf pfs.WalkFile RPC
func (a *apiServerV2) WalkFile(request *pfs.WalkFileRequest, server pfs.API_WalkFileServer) (retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	pachClient := a.env.GetPachClient(server.Context())
	return a.driver.walkFile(pachClient, request.File, func(fi *pfs.FileInfo) error {
		return server.Send(fi)
	})
}

// GlobFile is not implemented in V2.
func (a *apiServerV2) GlobFile(_ context.Context, _ *pfs.GlobFileRequest) (*pfs.FileInfos, error) {
	return nil, errV1NotImplemented
}

// GlobFileStream implements the protobuf pfs.GlobFileStream RPC
func (a *apiServerV2) GlobFileStream(request *pfs.GlobFileRequest, server pfs.API_GlobFileStreamServer) (retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	return a.driver.globFileV2(a.env.GetPachClient(server.Context()), request.Commit, request.Pattern, func(fi *pfs.FileInfo) error {
		return server.Send(fi)
	})
}

// DeleteFile is not implemented in V2.
func (a *apiServerV2) DeleteFile(ctx context.Context, request *pfs.DeleteFileRequest) (response *types.Empty, retErr error) {
	return nil, errV1NotImplemented
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

// Fsck is not implemented in V2.
func (a *apiServerV2) Fsck(_ *pfs.FsckRequest, _ pfs.API_FsckServer) error {
	return errV1NotImplemented
}

func (a *apiServerV2) FileOperationV2(server pfs.API_FileOperationV2Server) (retErr error) {
	request, err := server.Recv()
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	return metrics.ReportRequestWithThroughput(func() (int64, error) {
		if err != nil {
			return 0, err
		}
		var bytesRead int64
		if err := a.driver.fileOperation(a.env.GetPachClient(server.Context()), request.Commit, func(uw *fileset.UnorderedWriter) error {
			for {
				request, err := server.Recv()
				if err != nil {
					if errors.Is(err, io.EOF) {
						return nil
					}
					return err
				}
				// TODO Validation.
				switch op := request.Operation.(type) {
				case *pfs.FileOperationRequestV2_PutTar:
					n, err := putTar(uw, server, op.PutTar)
					bytesRead += n
					if err != nil {
						return err
					}
				case *pfs.FileOperationRequestV2_DeleteFiles:
					if err := deleteFiles(uw, op.DeleteFiles); err != nil {
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

func putTar(uw *fileset.UnorderedWriter, server pfs.API_FileOperationV2Server, request *pfs.PutTarRequestV2) (int64, error) {
	ptr := &putTarReader{
		server: server,
		r:      bytes.NewReader(request.Data),
	}
	err := uw.Put(ptr, request.Overwrite, request.Tag)
	return ptr.bytesRead, err
}

type putTarReader struct {
	server    pfs.API_FileOperationV2Server
	r         *bytes.Reader
	bytesRead int64
}

func (ptr *putTarReader) Read(data []byte) (int, error) {
	if ptr.r.Len() == 0 {
		request, err := ptr.server.Recv()
		if err != nil {
			return 0, err
		}
		op := request.Operation.(*pfs.FileOperationRequestV2_PutTar)
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

func deleteFiles(uw *fileset.UnorderedWriter, request *pfs.DeleteFilesRequestV2) error {
	for _, file := range request.Files {
		uw.Delete(file, request.Tag)
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
		err := a.driver.getTar(a.env.GetPachClient(server.Context()), commit, glob, gtw)
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

// DiffFileV2 returns the files only in new or only in old
func (a *apiServerV2) DiffFileV2(request *pfs.DiffFileRequest, server pfs.API_DiffFileV2Server) (retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	return a.driver.diffFileV2(a.env.GetPachClient(server.Context()), request.OldFile, request.NewFile, func(oldFi, newFi *pfs.FileInfo) error {
		return server.Send(&pfs.DiffFileResponseV2{
			OldFile: oldFi,
			NewFile: newFi,
		})
	})
}

// ClearCommitV2 deletes all data in the commit.
func (a *apiServerV2) ClearCommitV2(ctx context.Context, request *pfs.ClearCommitRequestV2) (_ *types.Empty, retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	return nil, a.driver.clearCommitV2(a.env.GetPachClient(ctx), request.Commit)
}
