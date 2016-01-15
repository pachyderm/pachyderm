package server

import (
	"github.com/pachyderm/pachyderm/src/pfs/drive"
)

var (
	blockSize = 8 * 1024 * 1024 // 8 Megabytes
)

func NewLocalAPIServer(dir string) (drive.APIServer, error) {
	return newLocalAPIServer(dir)
}
