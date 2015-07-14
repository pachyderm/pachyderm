package server

import (
	"fmt"
	"path"

	"github.com/pachyderm/pachyderm/src/btrfs"
	"github.com/pachyderm/pachyderm/src/pfs"
)

type apiServer struct{}

func newAPIServer() *apiServer {
	return &apiServer{}
}

func (a *apiServer) GetFile(getFileRequest *pfs.GetFileRequest, apiGetFileServer pfs.Api_GetFileServer) (retErr error) {
	filePath := path.Join(
		getFileRequest.Repository,
		getFileRequest.CommitId,
		getFileRequest.Path,
	)
	info, err := btrfs.Stat(filePath)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("%s is a directory", getFileRequest.Path)
	}
	file, err := btrfs.Open(filePath)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()
	return writeToStreamingBytesServer(file, apiGetFileServer)
}
