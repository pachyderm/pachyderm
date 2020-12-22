package client

import (
	"context"
	"io"
	"time"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/internal/errors"
	"github.com/pachyderm/pachyderm/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/src/internal/storage/renew"
)

// AppendFile appends a set of files.
func (c APIClient) AppendFile(repo, commit string, r io.Reader, overwrite bool, tag ...string) error {
	foc, err := c.NewFileOperationClient(repo, commit)
	if err != nil {
		return err
	}
	if err := foc.AppendFile(r, overwrite, tag...); err != nil {
		return err
	}
	return foc.Close()
}

// DeleteFile deletes a set of files.
// The optional tag field indicates specific tags in the files to delete.
func (c APIClient) deleteFile(repo, commit, path string, tag ...string) error {
	foc, err := c.NewFileOperationClient(repo, commit)
	if err != nil {
		return err
	}
	if err := foc.DeleteFile(path, tag...); err != nil {
		return err
	}
	return foc.Close()
}

// FileOperationClient is used for performing a stream of file operations.
// The operations are not persisted until the FileOperationClient is closed.
// FileOperationClient is not thread safe. Multiple FileOperationClients
// should be used for concurrent upload.
type FileOperationClient struct {
	client pfs.API_FileOperationClient
	fileOperationCore
}

// WithFileOperationClient creates a new FileOperationClient that is scoped to the passed in callback.
func (c APIClient) WithFileOperationClient(repo, commit string, cb func(*FileOperationClient) error) (retErr error) {
	foc, err := c.NewFileOperationClient(repo, commit)
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

// NewFileOperationClient creates a new FileOperationClient.
func (c APIClient) NewFileOperationClient(repo, commit string) (_ *FileOperationClient, retErr error) {
	defer func() {
		retErr = grpcutil.ScrubGRPC(retErr)
	}()
	client, err := c.PfsAPIClient.FileOperation(c.Ctx())
	if err != nil {
		return nil, err
	}
	if err := client.Send(&pfs.FileOperationRequest{
		Commit: NewCommit(repo, commit),
	}); err != nil {
		return nil, err
	}
	return &FileOperationClient{
		client: client,
		fileOperationCore: fileOperationCore{
			client: client,
		},
	}, nil
}

// Close closes the FileOperationClient.
func (foc *FileOperationClient) Close() error {
	return foc.maybeError(func() error {
		_, err := foc.client.CloseAndRecv()
		return err
	})
}

type fileOperationCore struct {
	client interface {
		Send(*pfs.FileOperationRequest) error
	}
	err error
}

// AppendFile appends a set of files.
func (foc *fileOperationCore) AppendFile(r io.Reader, overwrite bool, tag ...string) error {
	return foc.maybeError(func() error {
		ptr := &pfs.AppendFileRequest{Overwrite: overwrite}
		if len(tag) > 0 {
			if len(tag) > 1 {
				return errors.Errorf("AppendFile called with %v tags, expected 0 or 1", len(tag))
			}
			ptr.Tag = tag[0]
		}
		if err := foc.sendAppendFile(ptr); err != nil {
			return err
		}
		_, err := grpcutil.ChunkReader(r, func(data []byte) error {
			return foc.sendAppendFile(&pfs.AppendFileRequest{Data: data})
		})
		return err
	})
}

func (foc *fileOperationCore) maybeError(f func() error) (retErr error) {
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

func (foc *fileOperationCore) sendAppendFile(req *pfs.AppendFileRequest) error {
	return foc.client.Send(&pfs.FileOperationRequest{
		Operation: &pfs.FileOperationRequest_AppendFile{
			AppendFile: req,
		},
	})
}

// DeleteFile deletes a set of files.
// The optional tag field indicates specific tags in the files to delete.
func (foc *fileOperationCore) DeleteFile(path string, tag ...string) error {
	return foc.maybeError(func() error {
		req := &pfs.DeleteFileRequest{File: path}
		if len(tag) > 0 {
			if len(tag) > 1 {
				return errors.Errorf("DeleteFile called with %v tags, expected 0 or 1", len(tag))
			}
			req.Tag = tag[0]
		}
		return foc.sendDeleteFile(req)
	})
}

func (foc *fileOperationCore) sendDeleteFile(req *pfs.DeleteFileRequest) error {
	return foc.client.Send(&pfs.FileOperationRequest{
		Operation: &pfs.FileOperationRequest_DeleteFile{
			DeleteFile: req,
		},
	})
}

// TmpRepoName is a reserved repo name used for namespacing temporary filesets
const TmpRepoName = "__tmp__"

// TmpFileSetCommit creates a commit which can be used to access the temporary fileset fileSetID
func (c APIClient) TmpFileSetCommit(fileSetID string) *pfs.Commit {
	return &pfs.Commit{
		ID:   fileSetID,
		Repo: &pfs.Repo{Name: TmpRepoName},
	}
}

// DefaultTTL is the default time-to-live for a temporary fileset.
const DefaultTTL = 10 * time.Minute

// WithRenewer provides a scoped fileset renewer.
func (c APIClient) WithRenewer(cb func(context.Context, *renew.StringSet) error) error {
	rf := func(ctx context.Context, p string, ttl time.Duration) error {
		return c.WithCtx(ctx).RenewFileSet(p, ttl)
	}
	return renew.WithStringSet(c.Ctx(), DefaultTTL, rf, cb)
}

// WithCreateFilesetClient provides a scoped fileset client.
func (c APIClient) WithCreateFilesetClient(cb func(*CreateFilesetClient) error) (resp *pfs.CreateFilesetResponse, retErr error) {
	ctfsc, err := c.NewCreateFilesetClient()
	if err != nil {
		return nil, err
	}
	defer func() {
		if retErr == nil {
			resp, retErr = ctfsc.Close()
		}
	}()
	return nil, cb(ctfsc)
}

// CreateFilesetClient is used to create a temporary fileset.
type CreateFilesetClient struct {
	client pfs.API_CreateFilesetClient
	fileOperationCore
}

// NewCreateFilesetClient returns a CreateFilesetClient instance backed by this client
func (c APIClient) NewCreateFilesetClient() (_ *CreateFilesetClient, retErr error) {
	defer func() {
		retErr = grpcutil.ScrubGRPC(retErr)
	}()
	client, err := c.PfsAPIClient.CreateFileset(c.Ctx())
	if err != nil {
		return nil, err
	}
	return &CreateFilesetClient{
		client: client,
		fileOperationCore: fileOperationCore{
			client: client,
		},
	}, nil
}

// Close closes the CreateFilesetClient.
func (ctfsc *CreateFilesetClient) Close() (*pfs.CreateFilesetResponse, error) {
	var ret *pfs.CreateFilesetResponse
	if err := ctfsc.maybeError(func() error {
		resp, err := ctfsc.client.CloseAndRecv()
		if err != nil {
			return err
		}
		ret = resp
		return nil
	}); err != nil {
		return nil, err
	}
	return ret, nil
}

// RenewFileSet renews a fileset.
func (c APIClient) RenewFileSet(ID string, ttl time.Duration) (retErr error) {
	defer func() {
		retErr = grpcutil.ScrubGRPC(retErr)
	}()
	_, err := c.PfsAPIClient.RenewFileset(
		c.Ctx(),
		&pfs.RenewFilesetRequest{
			FilesetId:  ID,
			TtlSeconds: int64(ttl.Seconds()),
		},
	)
	return err
}
