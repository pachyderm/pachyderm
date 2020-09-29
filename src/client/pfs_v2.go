package client

import (
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
func (c APIClient) PutTarV2(repo, commit string, r io.Reader, overwrite bool, tag ...string) error {
	foc, err := c.NewFileOperationClientV2(repo, commit)
	if err != nil {
		return err
	}
	if err := foc.PutTar(r, overwrite, tag...); err != nil {
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

func (c APIClient) WithFileOperationClientV2(repo, commit string, cb func(*FileOperationClient) error) (retErr error) {
	foc, err := c.NewFileOperationClientV2(repo, commit)
	if err != nil {
		return err
	}
	defer func() {
		if retErr == nil {
			retErr = foc.Close()
		}
	}()
	return cb(foc)
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
func (foc *FileOperationClient) PutTar(r io.Reader, overwrite bool, tag ...string) error {
	return foc.maybeError(func() error {
		ptr := &pfs.PutTarRequestV2{Overwrite: overwrite}
		if len(tag) > 0 {
			if len(tag) > 1 {
				return errors.Errorf("PutTar called with %v tags, expected 0 or 1", len(tag))
			}
			ptr.Tag = tag[0]
		}
		if err := foc.sendPutTar(ptr); err != nil {
			return err
		}
		_, err := grpcutil.ChunkReader(r, func(data []byte) error {
			return foc.sendPutTar(&pfs.PutTarRequestV2{Data: data})
		})
		return err
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

// DiffFileV2 returns the differences between 2 paths at 2 commits.
// It streams back one file at a time which is either from the new path, or the old path
func (c APIClient) DiffFileV2(newRepo, newCommit, newPath, oldRepo,
	oldCommit, oldPath string, shallow bool, cb func(*pfs.FileInfo, *pfs.FileInfo) error) (retErr error) {
	defer func() {
		retErr = grpcutil.ScrubGRPC(retErr)
	}()
	ctx, cancel := context.WithCancel(c.Ctx())
	defer cancel()
	var oldFile *pfs.File
	if oldRepo != "" {
		oldFile = NewFile(oldRepo, oldCommit, oldPath)
	}
	req := &pfs.DiffFileRequest{
		NewFile: NewFile(newRepo, newCommit, newPath),
		OldFile: oldFile,
		Shallow: shallow,
	}
	client, err := c.PfsAPIClient.DiffFileV2(ctx, req)
	if err != nil {
		return err
	}
	for {
		resp, err := client.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return err
		}
		if err := cb(resp.NewFile, resp.OldFile); err != nil {
			return err
		}
	}
	return nil
}

// ClearCommitV2 clears the state of an open commit.
func (c APIClient) ClearCommitV2(repo, commit string) (retErr error) {
	defer func() {
		retErr = grpcutil.ScrubGRPC(retErr)
	}()
	_, err := c.PfsAPIClient.ClearCommitV2(
		c.Ctx(),
		&pfs.ClearCommitRequestV2{
			Commit: NewCommit(repo, commit),
		},
	)
	return err
}

// PutFileV2 puts a file into PFS.
// TODO: Change this to not buffer the file locally.
// We will want to move to a model where we buffer in chunk storage.
func (c APIClient) PutFileV2(repo string, commit string, path string, r io.Reader, overwrite bool) error {
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
		return c.PutTarV2(repo, commit, tarF, overwrite)
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

// GetFileV2 gets a file out of PFS.
func (c APIClient) GetFileV2(repo string, commit string, path string, w io.Writer) error {
	r, err := c.GetTarV2(repo, commit, path)
	if err != nil {
		return err
	}
	return tarutil.Iterate(r, func(f tarutil.File) error {
		return f.Content(w)
	}, true)
}

var errV1NotImplemented = errors.Errorf("V1 method not implemented")

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
	return 0, pfc.c.PutFileV2(repo, commit, path, r, false)
}

func (pfc *putFileClientV2) PutFileOverwrite(repo, commit, path string, r io.Reader, overwriteIndex int64) (int, error) {
	return 0, pfc.c.PutFileV2(repo, commit, path, r, true)
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
