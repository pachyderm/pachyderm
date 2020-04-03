package client

import (
	"bytes"
	"io"

	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
)

// PutTar puts a tar stream into PFS.
// Note: this should only be used for testing the new storage layer.
func (c APIClient) PutTar(repo, commit string, r io.Reader) (retErr error) {
	defer func() {
		retErr = grpcutil.ScrubGRPC(retErr)
	}()
	ptc, err := c.PfsAPIClient.PutTar(c.Ctx())
	if err != nil {
		return err
	}
	defer func() {
		if _, err := ptc.CloseAndRecv(); retErr == nil {
			retErr = err
		}
	}()
	if err := ptc.Send(&pfs.PutTarRequest{Commit: NewCommit(repo, commit)}); err != nil {
		return err
	}
	_, err = grpcutil.ChunkReader(r, func(data []byte) error {
		return ptc.Send(&pfs.PutTarRequest{Data: data})
	})
	return err
}

// GetTar gets a tar stream out of PFS that contains files at the repo and commit that match the path.
// Note: this should only be used for testing the new storage layer.
func (c APIClient) GetTar(repo, commit, path string) (_ io.Reader, retErr error) {
	defer func() {
		retErr = grpcutil.ScrubGRPC(retErr)
	}()
	req := &pfs.GetTarRequest{
		File: NewFile(repo, commit, path),
	}
	client, err := c.PfsAPIClient.GetTar(c.Ctx(), req)
	if err != nil {
		return nil, err
	}
	return grpcutil.NewStreamingBytesReader(client, nil), nil
}

// GetTarConditional functions similarly to GetTar with the key difference being that each file's content can be conditionally downloaded.
// GetTarConditional takes a callback that will be called for each file that matched the path.
// The callback will receive the file information for the file and a reader that will lazily download a tar stream that contains the file.
// Note: this should only be used for testing the new storage layer.
func (c APIClient) GetTarConditional(repoName string, commitID string, path string, f func(fileInfo *pfs.FileInfoNewStorage, r io.Reader) error) (retErr error) {
	defer func() {
		retErr = grpcutil.ScrubGRPC(retErr)
	}()
	client, err := c.PfsAPIClient.GetTarConditional(c.Ctx())
	if err != nil {
		return err
	}
	if err := client.Send(&pfs.GetTarConditionalRequest{File: NewFile(repoName, commitID, path)}); err != nil {
		return err
	}
	for {
		resp, err := client.Recv()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		r := &getTarConditionalReader{
			client: client,
			first:  true,
		}
		if err := f(resp.FileInfo, r); err != nil {
			return err
		}
		if r.first {
			if err := client.Send(&pfs.GetTarConditionalRequest{Skip: true}); err != nil {
				return err
			}
			continue
		}
		if err := r.drain(); err != nil {
			return err
		}
	}
	return nil

}

type getTarConditionalReader struct {
	client     pfs.API_GetTarConditionalClient
	r          *bytes.Reader
	first, eof bool
}

func (r *getTarConditionalReader) Read(data []byte) (int, error) {
	if r.first {
		if err := r.client.Send(&pfs.GetTarConditionalRequest{Skip: false}); err != nil {
			return 0, err
		}
		r.first = false
	}
	if r.r == nil || r.r.Len() == 0 {
		if err := r.nextResponse(); err != nil {
			return 0, err
		}
	}
	return r.r.Read(data)
}

func (r *getTarConditionalReader) nextResponse() error {
	if r.eof {
		return io.EOF
	}
	resp, err := r.client.Recv()
	if err != nil {
		return err
	}
	if resp.Eof {
		r.eof = true
		return io.EOF
	}
	r.r = bytes.NewReader(resp.Data)
	return nil
}

func (r *getTarConditionalReader) drain() error {
	for {
		if err := r.nextResponse(); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
}
