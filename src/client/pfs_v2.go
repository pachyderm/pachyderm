package client

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"os"

	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/tar"
	"github.com/pachyderm/pachyderm/src/server/pkg/tarutil"
)

// PutTarV2 puts a tar stream into PFS.
func (c APIClient) PutTarV2(repo, commit string, r io.Reader, tag ...string) error {
	foc, err := c.NewFileOperationClientV2(repo, commit)
	if err != nil {
		return err
	}
	if err := foc.PutTar(r, tag...); err != nil {
		return err
	}
	return foc.Close()
}

// DeleteFilesV2 deletes a set of files.
// The optional tag field indicates specific tags in the files to delete.
func (c APIClient) DeleteFilesV2(repo, commit string, files []string, tag ...string) error {
	foc, err := c.NewFileOperationClientV2(repo, commit)
	if err != nil {
		return err
	}
	if err := foc.DeleteFiles(files, tag...); err != nil {
		return err
	}
	return foc.Close()
}

// FileOperationClient is used for performing a stream of file operations.
// The operations are not persisted until the FileOperationClient is closed.
// FileOperationClient is not thread safe. Multiple FileOperationClients
// should be used for concurrent upload.
type FileOperationClient struct {
	client pfs.API_FileOperationV2Client
	err    error
}

// NewFileOperationClientV2 creates a new FileOperationClient.
func (c APIClient) NewFileOperationClientV2(repo, commit string) (_ *FileOperationClient, retErr error) {
	defer func() {
		retErr = grpcutil.ScrubGRPC(retErr)
	}()
	client, err := c.PfsAPIClient.FileOperationV2(c.Ctx())
	if err != nil {
		return nil, err
	}
	if err := client.Send(&pfs.FileOperationRequestV2{
		Commit: NewCommit(repo, commit),
	}); err != nil {
		return nil, err
	}
	return &FileOperationClient{client: client}, nil
}

// PutTar puts a tar stream into PFS.
func (foc *FileOperationClient) PutTar(r io.Reader, tag ...string) error {
	return foc.maybeError(func() error {
		if len(tag) > 0 {
			if len(tag) > 1 {
				return errors.Errorf("PutTar called with %v tags, expected 0 or 1", len(tag))
			}
			if err := foc.sendPutTar(&pfs.PutTarRequestV2{Tag: tag[0]}); err != nil {
				return err
			}
		}
		if _, err := grpcutil.ChunkReader(r, func(data []byte) error {
			return foc.sendPutTar(&pfs.PutTarRequestV2{Data: data})
		}); err != nil {
			return err
		}
		return foc.sendPutTar(&pfs.PutTarRequestV2{EOF: true})
	})
}

func (foc *FileOperationClient) maybeError(f func() error) (retErr error) {
	if foc.err != nil {
		return foc.err
	}
	defer func() {
		retErr = grpcutil.ScrubGRPC(retErr)
		if retErr != nil {
			foc.err = retErr
		}
	}()
	return f()
}

func (foc *FileOperationClient) sendPutTar(req *pfs.PutTarRequestV2) error {
	return foc.client.Send(&pfs.FileOperationRequestV2{
		Operation: &pfs.FileOperationRequestV2_PutTar{
			PutTar: req,
		},
	})
}

// DeleteFiles deletes a set of files.
// The optional tag field indicates specific tags in the files to delete.
func (foc *FileOperationClient) DeleteFiles(files []string, tag ...string) error {
	return foc.maybeError(func() error {
		req := &pfs.DeleteFilesRequestV2{Files: files}
		if len(tag) > 0 {
			if len(tag) > 1 {
				return errors.Errorf("DeleteFiles called with %v tags, expected 0 or 1", len(tag))
			}
			req.Tag = tag[0]
		}
		return foc.sendDeleteFiles(req)
	})
}

func (foc *FileOperationClient) sendDeleteFiles(req *pfs.DeleteFilesRequestV2) error {
	return foc.client.Send(&pfs.FileOperationRequestV2{
		Operation: &pfs.FileOperationRequestV2_DeleteFiles{
			DeleteFiles: req,
		},
	})
}

// Close closes the FileOperationClient.
func (foc *FileOperationClient) Close() error {
	return foc.maybeError(func() error {
		_, err := foc.client.CloseAndRecv()
		return err
	})
}

// GetTarV2 gets a tar stream out of PFS that contains files at the repo and commit that match the path.
func (c APIClient) GetTarV2(repo, commit, path string) (_ io.Reader, retErr error) {
	defer func() {
		retErr = grpcutil.ScrubGRPC(retErr)
	}()
	req := &pfs.GetTarRequestV2{
		File: NewFile(repo, commit, path),
	}
	client, err := c.PfsAPIClient.GetTarV2(c.Ctx(), req)
	if err != nil {
		return nil, err
	}
	return grpcutil.NewStreamingBytesReader(client, nil), nil
}

// GetTarConditionalV2 functions similarly to GetTar with the key difference being that each file's content can be conditionally downloaded.
// GetTarConditional takes a callback that will be called for each file that matched the path.
// The callback will receive the file information for the file and a reader that will lazily download a tar stream that contains the file.
func (c APIClient) GetTarConditionalV2(repoName string, commitID string, path string, f func(fileInfo *pfs.FileInfo, r io.Reader) error) (retErr error) {
	defer func() {
		retErr = grpcutil.ScrubGRPC(retErr)
	}()
	client, err := c.PfsAPIClient.GetTarConditionalV2(c.Ctx())
	if err != nil {
		return err
	}
	if err := client.Send(&pfs.GetTarConditionalRequestV2{File: NewFile(repoName, commitID, path)}); err != nil {
		return err
	}
	for {
		resp, err := client.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
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
			if err := client.Send(&pfs.GetTarConditionalRequestV2{Skip: true}); err != nil {
				return err
			}
			continue
		}
		if err := r.drain(); err != nil {
			return err
		}
	}
}

// ListFileV2 returns info about all files in a Commit under path, calling f with each FileInfo.
func (c APIClient) ListFileV2(repoName string, commitID string, path string, f func(fileInfo *pfs.FileInfo) error) (retErr error) {
	defer func() {
		retErr = grpcutil.ScrubGRPC(retErr)
	}()
	ctx, cancel := context.WithCancel(c.Ctx())
	defer cancel()
	req := &pfs.ListFileRequest{
		File: NewFile(repoName, commitID, path),
		Full: true,
	}
	client, err := c.PfsAPIClient.ListFileV2(ctx, req)
	if err != nil {
		return err
	}
	for {
		finfo, err := client.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}

		if err := f(finfo); err != nil {
			return err
		}
	}
	return nil
}

type getTarConditionalReader struct {
	client     pfs.API_GetTarConditionalV2Client
	r          *bytes.Reader
	first, EOF bool
}

func (r *getTarConditionalReader) Read(data []byte) (int, error) {
	if r.first {
		if err := r.client.Send(&pfs.GetTarConditionalRequestV2{Skip: false}); err != nil {
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
	if r.EOF {
		return io.EOF
	}
	resp, err := r.client.Recv()
	if err != nil {
		return err
	}
	if resp.EOF {
		r.EOF = true
		return io.EOF
	}
	r.r = bytes.NewReader(resp.Data)
	return nil
}

func (r *getTarConditionalReader) drain() error {
	for {
		if err := r.nextResponse(); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
	}
}

// TODO: Change this to not buffer the file locally.
// We will want to move to a model where we buffer in chunk storage.
func (c APIClient) PutFileV2(repo string, commit string, path string, r io.Reader) error {
	return withTmpFile(func(tarF *os.File) error {
		if err := withTmpFile(func(f *os.File) error {
			size, err := io.Copy(f, r)
			if err != nil {
				return err
			}
			_, err = f.Seek(0, 0)
			if err != nil {
				return err
			}
			return tarutil.WithWriter(tarF, func(tw *tar.Writer) error {
				return tarutil.WriteFile(tw, tarutil.NewStreamFile(path, size, f))
			})
		}); err != nil {
			return err
		}
		_, err := tarF.Seek(0, 0)
		if err != nil {
			return err
		}
		return c.PutTarV2(repo, commit, tarF)
	})
}

// TODO: refactor into utility package, also exists in debug util.
func withTmpFile(cb func(*os.File) error) (retErr error) {
	if err := os.MkdirAll(os.TempDir(), 0700); err != nil {
		return err
	}
	f, err := ioutil.TempFile(os.TempDir(), "pachyderm_put_file")
	if err != nil {
		return err
	}
	defer func() {
		if err := os.Remove(f.Name()); retErr == nil {
			retErr = err
		}
		if err := f.Close(); retErr == nil {
			retErr = err
		}
	}()
	return cb(f)
}

func (c APIClient) GetFileV2(repo string, commit string, path string, w io.Writer) error {
	r, err := c.GetTarV2(repo, commit, path)
	if err != nil {
		return err
	}
	return tarutil.Iterate(r, func(f tarutil.File) error {
		return f.Content(w)
	}, true)
}

var errV1NotImplemented = errors.Errorf("v1 method not implemented")

type putFileClientV2 struct {
	c APIClient
}

func (c APIClient) newPutFileClientV2() PutFileClient {
	return &putFileClientV2{c: c}
}

func (pfc *putFileClientV2) PutFileWriter(repo, commit, path string) (io.WriteCloser, error) {
	return nil, errV1NotImplemented
}

func (pfc *putFileClientV2) PutFileSplitWriter(repo, commit, path string, delimiter pfs.Delimiter, targetFileDatums int64, targetFileBytes int64, headerRecords int64, overwrite bool) (io.WriteCloser, error) {
	return nil, errV1NotImplemented
}

func (pfc *putFileClientV2) PutFile(repo, commit, path string, r io.Reader) (int, error) {
	return 0, pfc.c.PutFileV2(repo, commit, path, r)
}

func (pfc *putFileClientV2) PutFileOverwrite(repo, commit, path string, reader io.Reader, overwriteIndex int64) (int, error) {
	return 0, errV1NotImplemented
}

func (pfc *putFileClientV2) PutFileSplit(repo, commit, path string, delimiter pfs.Delimiter, targetFileDatums int64, targetFileBytes int64, headerRecords int64, overwrite bool, r io.Reader) (int, error) {
	// TODO: Add split support.
	return 0, errV1NotImplemented
}

func (pfc *putFileClientV2) PutFileURL(repo, commit, path, url string, recursive bool, overwrite bool) error {
	// TODO: Add URL support.
	return errV1NotImplemented
}

func (pfc *putFileClientV2) DeleteFile(repo, commit, path string) error {
	return pfc.c.DeleteFilesV2(repo, commit, []string{path})
}

func (pfc *putFileClientV2) Close() error {
	return nil
}
