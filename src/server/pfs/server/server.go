package server

import (
	pfsclient "github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/shard"
	pfsserver "github.com/pachyderm/pachyderm/src/server/pfs"
	"github.com/pachyderm/pachyderm/src/server/pfs/drive"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"go.pedge.io/lion/proto"
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

// NewBlockAPIServer creates a BlockAPIServer using the credentials it finds in
// the environment
func NewBlockAPIServer(dir string) (pfsclient.BlockAPIServer, error) {
	if blockAPIServer, err := newAmazonBlockAPIServer(dir); err == nil {
		return blockAPIServer, nil
	} else {
		protolion.Errorf("error create Amazon block backend: %s", err.Error())
	}
	if blockAPIServer, err := newGoogleBlockAPIServer(dir); err == nil {
		return blockAPIServer, nil
	} else {
		protolion.Errorf("error create Google block backend: %s", err.Error())
	}
	protolion.Errorf("failed to create obj backend, falling back to local")
	return NewLocalBlockAPIServer(dir)
}
