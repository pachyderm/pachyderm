package server

import (
	pfsclient "github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/server/pkg/hashtree"
)

// Valid object storage backends
const (
	MinioBackendEnvVar     = "MINIO"
	AmazonBackendEnvVar    = "AMAZON"
	GoogleBackendEnvVar    = "GOOGLE"
	MicrosoftBackendEnvVar = "MICROSOFT"
	LocalBackendEnvVar     = "LOCAL"
)

// APIServer represents and api server.
type APIServer interface {
	pfsclient.APIServer
}

// BlockAPIServer combines BlockAPIServer and ObjectAPIServer.
type BlockAPIServer interface {
	pfsclient.ObjectAPIServer
}

// NewAPIServer creates an APIServer.
func NewAPIServer(address string, etcdAddresses []string, etcdPrefix string, treeCache *hashtree.Cache, storageRoot string, memoryRequest int64) (APIServer, error) {
	return newAPIServer(address, etcdAddresses, etcdPrefix, treeCache, storageRoot, memoryRequest)
}

// NewBlockAPIServer creates a BlockAPIServer using the credentials it finds in
// the environment
func NewBlockAPIServer(dir string, cacheBytes int64, backend string, etcdAddress string) (BlockAPIServer, error) {
	switch backend {
	case MinioBackendEnvVar:
		// S3 compatible doesn't like leading slashes
		if len(dir) > 0 && dir[0] == '/' {
			dir = dir[1:]
		}
		blockAPIServer, err := newMinioBlockAPIServer(dir, cacheBytes, etcdAddress)
		if err != nil {
			return nil, err
		}
		return blockAPIServer, nil
	case AmazonBackendEnvVar:
		// amazon doesn't like leading slashes
		if len(dir) > 0 && dir[0] == '/' {
			dir = dir[1:]
		}
		blockAPIServer, err := newAmazonBlockAPIServer(dir, cacheBytes, etcdAddress)
		if err != nil {
			return nil, err
		}
		return blockAPIServer, nil
	case GoogleBackendEnvVar:
		// TODO figure out if google likes leading slashses
		blockAPIServer, err := newGoogleBlockAPIServer(dir, cacheBytes, etcdAddress)
		if err != nil {
			return nil, err
		}
		return blockAPIServer, nil
	case MicrosoftBackendEnvVar:
		blockAPIServer, err := newMicrosoftBlockAPIServer(dir, cacheBytes, etcdAddress)
		if err != nil {
			return nil, err
		}
		return blockAPIServer, nil
	case LocalBackendEnvVar:
		fallthrough
	default:
		blockAPIServer, err := newLocalBlockAPIServer(dir, cacheBytes, etcdAddress)
		if err != nil {
			return nil, err
		}
		return blockAPIServer, nil
	}
}
