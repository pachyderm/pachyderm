package server

import (
	"github.com/pachyderm/pachyderm/src/pfs/drive"
)

func NewLocalAPIServer(dir string) drive.APIServer {
	return newLocalAPIServer(dir)
}
