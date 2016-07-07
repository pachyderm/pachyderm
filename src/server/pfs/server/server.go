package server

import (
	pfsclient "github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/shard"
	pfsserver "github.com/pachyderm/pachyderm/src/server/pfs"
	"github.com/pachyderm/pachyderm/src/server/pfs/drive"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
)

var (
	blockSize = 8 * 1024 * 1024 // 8 Megabytes
)

// Valid backends
const (
	AmazonBackendEnvVar = "AMAZON"
	GoogleBackendEnvVar = "GOOGLE"
)

// APIServer represents an api server.
type APIServer interface {
	pfsclient.APIServer // SJ: WOW this is bad naming
	shard.Frontend
}

// InternalAPIServer represents and internal api server.
type InternalAPIServer interface {
	pfsclient.InternalAPIServer // SJ: also bad naming
	shard.Server
}

// NewAPIServer creates an APIServer.
func NewAPIServer(hasher *pfsserver.Hasher, router shard.Router) APIServer {
	return newAPIServer(hasher, router)
}

// NewInternalAPIServer creates an InternalAPIServer.
func NewInternalAPIServer(hasher *pfsserver.Hasher, router shard.Router, driver drive.Driver) InternalAPIServer {
	return newInternalAPIServer(hasher, router, driver)
}

// NewLocalBlockAPIServer creates a BlockAPIServer.
func NewLocalBlockAPIServer(dir string) (pfsclient.BlockAPIServer, error) { // SJ: also bad naming
	return newLocalBlockAPIServer(dir)
}

// NewObjBlockAPIServer create a BlockAPIServer from an obj.Client.
func NewObjBlockAPIServer(dir string, objClient obj.Client) (pfsclient.BlockAPIServer, error) { // SJ: Also bad naming
	return newObjBlockAPIServer(dir, objClient)
}

// NewBlockAPIServer creates a BlockAPIServer using the credentials it finds in
// the environment
func NewBlockAPIServer(dir string, backend string) (pfsclient.BlockAPIServer, error) {
	switch backend {
	case AmazonBackendEnvVar:
		blockAPIServer, err := newAmazonBlockAPIServer(dir)
		if err != nil {
			return nil, err
		}
		return blockAPIServer, nil
	case GoogleBackendEnvVar:
		blockAPIServer, err := newGoogleBlockAPIServer(dir)
		if err != nil {
			return nil, err
		}
		return blockAPIServer, nil
	default:
		return NewLocalBlockAPIServer(dir)
	}
}
