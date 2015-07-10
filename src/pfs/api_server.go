package pfs

import (
	"bufio"
	"fmt"
	"path"

	"github.com/pachyderm/pfs/src/btrfs"
)

type apiServer struct{}

func newAPIServer() *apiServer {
	return &apiServer{}
}

func (a *apiServer) GetFile(getFileRequest *GetFileRequest, apiGetFileServer Api_GetFileServer) error {
	filePath := path.Join(
		getFileRequest.Repository,
		getFileRequest.Commit,
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
	_, err := bufio.NewReader(file).WriteTo(newStreamingBytesWriter(apiGetFileServer))
	return err
}
