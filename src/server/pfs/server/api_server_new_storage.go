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
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/metrics"
)

func (a *apiServer) PutTar(server pfs.API_PutTarServer) (retErr error) {
	req, err := server.Recv()
	func() { a.Log(req, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(req, nil, retErr, time.Since(start)) }(time.Now())
	return metrics.ReportRequestWithThroughput(func() (int64, error) {
		if err != nil {
			return 0, err
		}
		if !a.env.NewStorageLayer {
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
				defer func() {
					bytesRead += ptr.bytesRead
				}()
				if req.Tag != "" {
					if err := fs.Put(ptr, req.Tag); err != nil {
						return err
					}
					continue
				}
				if err := fs.Put(ptr); err != nil {
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
	server    pfs.API_PutTarServer
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

func (a *apiServer) GetTar(request *pfs.GetTarRequest, server pfs.API_GetTarServer) (retErr error) {
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	return metrics.ReportRequestWithThroughput(func() (int64, error) {
		if !a.env.NewStorageLayer {
			return 0, errors.Errorf("new storage layer disabled")
		}
		repo := request.File.Commit.Repo.Name
		commit := request.File.Commit.ID
		glob := request.File.Path
		gtw := newGetTarWriter(grpcutil.NewStreamingBytesWriter(server))
		err := a.driver.getFilesNewStorageLayer(server.Context(), repo, commit, glob, gtw)
		return gtw.bytesWritten, err
	})
}

// (bryce) this is probably unecessary.
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

func (a *apiServer) GetTarConditional(server pfs.API_GetTarConditionalServer) (retErr error) {
	request, err := server.Recv()
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	return metrics.ReportRequestWithThroughput(func() (int64, error) {
		if err != nil {
			return 0, err
		}
		if !a.env.NewStorageLayer {
			return 0, errors.Errorf("new storage layer disabled")
		}
		repo := request.File.Commit.Repo.Name
		commit := request.File.Commit.ID
		glob := request.File.Path
		var bytesWritten int64
		err := a.driver.getFilesConditional(server.Context(), repo, commit, glob, func(fr *FileReader) error {
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
	server       pfs.API_GetTarConditionalServer
	bytesWritten int64
}

func newGetTarConditionalWriter(server pfs.API_GetTarConditionalServer) *getTarConditionalWriter {
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
