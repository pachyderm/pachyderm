// +build !windows

package server

import (
	"io"
	"os"
	"path/filepath"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"
)

func (d *driver) downloadTree(pachClient *client.APIClient, object *pfs.Object, prefix string) (r io.ReadCloser, retErr error) {
	objClient, err := obj.NewClientFromSecret(d.storageRoot)
	if err != nil {
		return nil, err
	}
	info, err := pachClient.InspectObject(object.Hash)
	if err != nil {
		return nil, err
	}
	path, err := BlockPathFromEnv(info.BlockRef.Block)
	if err != nil {
		return nil, err
	}
	offset, size, err := getTreeRange(pachClient.Ctx(), objClient, path, prefix)
	if err != nil {
		return nil, err
	}
	objR, err := objClient.Reader(pachClient.Ctx(), path, offset, size)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := objR.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	name := filepath.Join(d.storageRoot, uuid.NewWithoutDashes())
	f, err := os.Create(name)
	if err != nil {
		return nil, err
	}

	// Ensure we close the file if we've errored
	defer func() {
		if retErr != nil {
			f.Close()
		}
	}()

	// Mark the file for removal (Linux won't remove it until we close the file)
	if err := os.Remove(name); err != nil {
		return nil, err
	}
	buf := grpcutil.GetBuffer()
	defer grpcutil.PutBuffer(buf)
	if _, err := io.CopyBuffer(f, objR, buf); err != nil {
		return nil, err
	}
	if _, err := f.Seek(0, 0); err != nil {
		return nil, err
	}
	return f, nil
}
