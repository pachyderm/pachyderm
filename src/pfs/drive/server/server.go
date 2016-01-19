package server

import (
	"github.com/pachyderm/pachyderm/src/pfs"
)

var (
	blockSize = 8 * 1024 * 1024 // 8 Megabytes
)

func NewLocalAPIServer(dir string) (pfs.BlockAPIServer, error) {
	return newLocalAPIServer(dir)
}
