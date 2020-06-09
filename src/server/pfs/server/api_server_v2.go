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
var _ pfs.APIV2Server = &apiServerV2{}

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

func (a *apiServerV2) PutTar(server pfs.APIV2_PutTarServer) (retErr error) {
	req, err := server.Recv()
	func() { a.Log(req, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(req, nil, retErr, time.Since(start)) }(time.Now())
	return metrics.ReportRequestWithThroughput(func() (int64, error) {
		if err != nil {
			return 0, err
		}
		if !a.env.StorageV2 {
			return 0, errors.Errorf("new storage layer disabled")
		}
		repo := req.Commit.Repo.Name
		commit := req.Commit.ID
		var bytesRead int64
		if err := a.driver.withFileSet(server.Context(), repo, commit, func(fs *fileset.FileSet) error {
			for {
				req, err := server.Recv()
				if err != nil {
					if err == io.EOF {
						return nil
					}
					return err
				}
				ptr := &putTarReader{
					server: server,
					r:      bytes.NewReader(req.Data),
				}
				if err := func() error {
					defer func() {
						bytesRead += ptr.bytesRead
					}()
					return fs.Put(ptr, req.Tag)
				}(); err != nil {
					return err
				}
			}
		}); err != nil {
			return bytesRead, err
		}
		return bytesRead, server.SendAndClose(&types.Empty{})
	})
}

type putTarReader struct {
	server    pfs.APIV2_PutTarServer
	r         *bytes.Reader
	bytesRead int64
}

func (ptr *putTarReader) Read(data []byte) (int, error) {
	if ptr.r.Len() == 0 {
		req, err := ptr.server.Recv()
		if err != nil {
			return 0, err
		}
		if req.EOF {
			return 0, io.EOF
		}
		ptr.r = bytes.NewReader(req.Data)
	}
	n, err := ptr.r.Read(data)
	ptr.bytesRead += int64(n)
	return n, err
}

func (a *apiServer) GetTar(request *pfs.GetTarRequest, server pfs.APIV2_GetTarServer) (retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	return metrics.ReportRequestWithThroughput(func() (int64, error) {
		if !a.env.StorageV2 {
			return 0, errors.Errorf("new storage layer disabled")
		}
		repo := request.File.Commit.Repo.Name
		commit := request.File.Commit.ID
		glob := request.File.Path
		gtw := newGetTarWriter(grpcutil.NewStreamingBytesWriter(server))
		err := a.driver.getTar(server.Context(), repo, commit, glob, gtw)
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

func (a *apiServer) GetTarConditional(server pfs.APIV2_GetTarConditionalServer) (retErr error) {
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
			if err := server.Send(&pfs.GetTarConditionalResponse{FileInfo: fr.Info()}); err != nil {
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
			return server.Send(&pfs.GetTarConditionalResponse{EOF: true})
		})
		return bytesWritten, err
	})
}

type getTarConditionalWriter struct {
	server       pfs.APIV2_GetTarConditionalServer
	bytesWritten int64
}

func newGetTarConditionalWriter(server pfs.APIV2_GetTarConditionalServer) *getTarConditionalWriter {
	return &getTarConditionalWriter{
		server: server,
	}
}

func (w *getTarConditionalWriter) Write(data []byte) (int, error) {
	if err := w.server.Send(&pfs.GetTarConditionalResponse{Data: data}); err != nil {
		return 0, err
	}
	w.bytesWritten += int64(len(data))
	return len(data), nil
}

func (a *apiServerV2) ListFileV2(req *pfs.ListFileRequest, server pfs.APIV2_ListFileV2Server) error {
	pachClient := a.env.GetPachClient(server.Context())
	return a.driver.listFileV2(pachClient, req.File, req.Full, req.History, func(finfo *pfs.FileInfoV2) error {
		return server.Send(finfo)
	})
}
