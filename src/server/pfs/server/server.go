package server

import (
	pfsclient "github.com/pachyderm/pachyderm/src/client/pfs"
	pfsserver "github.com/pachyderm/pachyderm/src/server/pfs"
	"github.com/pachyderm/pachyderm/src/server/pfs/drive"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/pachyderm/pachyderm/src/client/pkg/shard"
)

var (
	blockSize = 8 * 1024 * 1024 // 8 Megabytes
)

type APIServer interface {
	pfsclient.APIServer // SJ: WOW this is bad naming
	shard.Frontend
}

type InternalAPIServer interface {
	pfsclient.InternalAPIServer // SJ: also bad naming
	shard.Server
}

func NewAPIServer(hasher *pfsserver.Hasher, router shard.Router) APIServer {
	return newAPIServer(hasher, router)
}

func NewInternalAPIServer(hasher *pfsserver.Hasher, router shard.Router, driver drive.Driver) InternalAPIServer {
	return newInternalAPIServer(hasher, router, driver)
}

func NewLocalBlockAPIServer(dir string) (pfsclient.BlockAPIServer, error) { // SJ: also bad naming
	return newLocalBlockAPIServer(dir)
}

func NewObjBlockAPIServer(dir string, objClient obj.Client) (pfsclient.BlockAPIServer, error) { // SJ: Also bad naming
	return newObjBlockAPIServer(dir, objClient)
}
