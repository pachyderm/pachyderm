package server

import (
	pfsclient "github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/server/pkg/serviceenv"
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
// cacheSize is the number of commit trees which will be cached in the server.
func NewAPIServer(env *serviceenv.ServiceEnv, etcdPrefix string, cacheSize int64) (APIServer, error) {
	return newAPIServer(env, etcdPrefix, cacheSize)
}

// NewBlockAPIServer creates a BlockAPIServer using the credentials it finds in
// the environment
func NewBlockAPIServer(env *serviceenv.ServiceEnv, dir string, cacheBytes int64, backend string) (BlockAPIServer, error) {
	switch backend {
	case MinioBackendEnvVar:
		// S3 compatible doesn't like leading slashes
		if len(dir) > 0 && dir[0] == '/' {
			dir = dir[1:]
		}
		blockAPIServer, err := newMinioBlockAPIServer(env, dir, cacheBytes)
		if err != nil {
			return nil, err
		}
		return blockAPIServer, nil
	case AmazonBackendEnvVar:
		// amazon doesn't like leading slashes
		if len(dir) > 0 && dir[0] == '/' {
			dir = dir[1:]
		}
		blockAPIServer, err := newAmazonBlockAPIServer(env, dir, cacheBytes)
		if err != nil {
			return nil, err
		}
		return blockAPIServer, nil
	case GoogleBackendEnvVar:
		// TODO figure out if google likes leading slashses
		blockAPIServer, err := newGoogleBlockAPIServer(env, dir, cacheBytes)
		if err != nil {
			return nil, err
		}
		return blockAPIServer, nil
	case MicrosoftBackendEnvVar:
		blockAPIServer, err := newMicrosoftBlockAPIServer(env, dir, cacheBytes)
		if err != nil {
			return nil, err
		}
		return blockAPIServer, nil
	case LocalBackendEnvVar:
		fallthrough
	default:
		blockAPIServer, err := newLocalBlockAPIServer(env, dir, cacheBytes)
		if err != nil {
			return nil, err
		}
		return blockAPIServer, nil
	}
}
