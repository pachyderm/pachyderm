package client

import (
	"io"

	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
)

func (c APIClient) PutTar(repo, commit string, r io.Reader) (retErr error) {
	ptc, err := c.PfsAPIClient.PutTar(c.Ctx())
	if err != nil {
		return grpcutil.ScrubGRPC(err)
	}
	defer func() {
		if _, err := ptc.CloseAndRecv(); err != nil && retErr == nil {
			retErr = grpcutil.ScrubGRPC(err)
		}
	}()
	firstReq := &pfs.PutTarRequest{
		Commit: NewCommit(repo, commit),
	}
	_, err = grpcutil.ChunkReader(r, func(data []byte) error {
		req := &pfs.PutTarRequest{}
		if firstReq != nil {
			req = firstReq
			firstReq = nil
		}
		req.Data = data
		return grpcutil.ScrubGRPC(ptc.Send(req))
	})
	return err
}

func (c APIClient) GetTar(repo, commit, glob string, w io.Writer) error {
	req := &pfs.GetTarRequest{
		Commit: NewCommit(repo, commit),
		Glob:   glob,
	}
	gtc, err := c.PfsAPIClient.GetTar(c.Ctx(), req)
	if err != nil {
		return grpcutil.ScrubGRPC(err)
	}
	err = grpcutil.WriteFromStreamingBytesClient(gtc, w)
	return grpcutil.ScrubGRPC(err)
}
