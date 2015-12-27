package server

import (
	"github.com/pachyderm/pachyderm/src/pfs/drive"
)

func NewLocalAPIServer(dir string) (drive.APIServer, error) {
	return newLocalAPIServer(dir)
}
