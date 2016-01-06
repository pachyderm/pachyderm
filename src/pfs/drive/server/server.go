package server

import (
	"github.com/pachyderm/pachyderm/src/pfs/drive"
)

var (
	blockSize = 128 * 1024 * 1024 // 128 Megabytes
)

func NewLocalAPIServer(dir string) (drive.APIServer, error) {
	return newLocalAPIServer(dir)
}
