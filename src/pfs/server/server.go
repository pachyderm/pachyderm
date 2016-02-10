package server

import (
	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/drive"
	"github.com/pachyderm/pachyderm/src/pkg/obj"
	"github.com/pachyderm/pachyderm/src/pkg/shard"
)

var (
	blockSize = 8 * 1024 * 1024 // 8 Megabytes
)

type APIServer interface {
	pfs.APIServer
	shard.Frontend
}

type InternalAPIServer interface {
	pfs.InternalAPIServer
	shard.Server
}

func NewAPIServer(hasher *pfs.Hasher, router shard.Router) APIServer {
	return newAPIServer(hasher, router)
}

func NewInternalAPIServer(hasher *pfs.Hasher, router shard.Router, driver drive.Driver) InternalAPIServer {
	return newInternalAPIServer(hasher, router, driver)
}

func NewLocalBlockAPIServer(dir string) (pfs.BlockAPIServer, error) {
	return newLocalBlockAPIServer(dir)
}

func NewObjBlockAPIServer(dir string, objClient obj.Client) (pfs.BlockAPIServer, error) {
	return newObjBlockAPIServer(dir, objClient)
}
