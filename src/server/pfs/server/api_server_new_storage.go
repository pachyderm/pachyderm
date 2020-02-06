package server

import (
	"bytes"
	"fmt"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
)

func (a *apiServer) PutTar(server pfs.API_PutTarServer) (retErr error) {
	if !a.env.NewStorageLayer {
		return fmt.Errorf("new storage layer disabled")
	}
	request, err := server.Recv()
	if err != nil {
		return err
	}
	ptr := &putTarReader{
		server: server,
		r:      bytes.NewReader(request.Data),
	}
	request.Data = nil
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	repo := request.Commit.Repo.Name
	commit := request.Commit.ID
	if err := a.driver.putFilesNewStorageLayer(server.Context(), repo, commit, ptr); err != nil {
		return err
	}
	return server.SendAndClose(&types.Empty{})
}

// (bryce) this pattern might be more generally applicable (interpret the first message then stream bytes from the following)
type putTarReader struct {
	server pfs.API_PutTarServer
	r      *bytes.Reader
}

func (ptr *putTarReader) Read(data []byte) (int, error) {
	if ptr.r.Len() == 0 {
		req, err := ptr.server.Recv()
		if err != nil {
			return 0, err
		}
		ptr.r = bytes.NewReader(req.Data)
	}
	return ptr.r.Read(data)
}

func (a *apiServer) GetTar(request *pfs.GetTarRequest, server pfs.API_GetTarServer) (retErr error) {
	if !a.env.NewStorageLayer {
		return fmt.Errorf("new storage layer disabled")
	}
	func() { a.Log(request, nil, nil, 0) }()
	defer func(start time.Time) { a.Log(request, nil, retErr, time.Since(start)) }(time.Now())
	repo := request.Commit.Repo.Name
	commit := request.Commit.ID
	glob := request.Glob
	return a.driver.getFilesNewStorageLayer(server.Context(), repo, commit, glob, grpcutil.NewStreamingBytesWriter(server))
}
