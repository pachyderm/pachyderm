package client

import (
	"context"
	"io"
	"io/fs"

	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/pachyderm/pachyderm/v2/src/storage"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

// FileSystemToFileset creates a new fileset containing fileSystem, returning the fileset ID.
func (c APIClient) FileSystemToFileset(ctx context.Context, fileSystem fs.FS) (string, error) {
	cc, err := c.FilesetClient.CreateFileset(ctx)
	if err != nil {
		return "", errors.Wrap(err, "could not create fileset")
	}
	if err := fs.WalkDir(fileSystem, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return errors.Wrapf(err, "could not walk %s", path)
		}
		if d.IsDir() {
			// PFS does not have a concept of directories
			return nil
		}
		f, err := fileSystem.Open(path)
		if err != nil {
			return errors.Wrapf(err, "open  %s", path)
		}
		data, err := io.ReadAll(f)
		if err != nil {
			return errors.Wrapf(err, "read all of %s", path)
		}
		if err := cc.Send(&storage.CreateFilesetRequest{
			Modification: &storage.CreateFilesetRequest_AppendFile{
				AppendFile: &storage.AppendFile{
					Path: path,
					Data: &wrapperspb.BytesValue{Value: data},
				},
			},
		}); err != nil {
			return errors.Wrapf(err, "upload %s", path)
		}
		return nil
	}); err != nil {
		return "", errors.Wrap(err, "walk filesystem")
	}
	resp, err := cc.CloseAndRecv()
	if err != nil {
		return "", errors.Wrap(err, "close fileset")
	}
	return resp.FilesetId, nil
}
