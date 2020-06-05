package server

import (
	pfsclient "github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/server/pkg/hashtree"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/pachyderm/pachyderm/src/server/pkg/serviceenv"
	txnenv "github.com/pachyderm/pachyderm/src/server/pkg/transactionenv"
)

// Valid object storage backends
const (
	MinioBackendEnvVar     = "MINIO"
	AmazonBackendEnvVar    = "AMAZON"
	GoogleBackendEnvVar    = "GOOGLE"
	MicrosoftBackendEnvVar = "MICROSOFT"
	LocalBackendEnvVar     = "LOCAL"
)

// APIServer represents an api server.
type APIServer interface {
	pfsclient.APIServer
	txnenv.PfsTransactionServer
}

// BlockAPIServer combines BlockAPIServer and ObjectAPIServer.
type BlockAPIServer interface {
	pfsclient.ObjectAPIServer
}

// NewAPIServer creates an APIServer.
func NewAPIServer(
	env *serviceenv.ServiceEnv,
	txnEnv *txnenv.TransactionEnv,
	etcdPrefix string,
	treeCache *hashtree.Cache,
	storageRoot string,
	memoryRequest int64,
) (APIServer, error) {
	if env.StorageV2 {
		a, err := newAPIServerV2(env, txnEnv, etcdPrefix, treeCache, storageRoot, memoryRequest)
		if err != nil {
			return nil, err
		}
		return newValidated(a, env), nil
	}
	return newAPIServer(env, txnEnv, etcdPrefix, treeCache, storageRoot, memoryRequest)
}

// NewBlockAPIServer creates a BlockAPIServer using the credentials it finds in
// the environment
// TODO(msteffen) accept serviceenv.ServiceEnv instead of 'dir', 'backend', and
// 'duplicate'?
func NewBlockAPIServer(dir string, cacheBytes int64, backend string, etcdAddress string, duplicate bool) (BlockAPIServer, error) {
	switch backend {
	case MinioBackendEnvVar:
		// S3 compatible doesn't like leading slashes
		if len(dir) > 0 && dir[0] == '/' {
			dir = dir[1:]
		}
		blockAPIServer, err := newMinioBlockAPIServer(dir, cacheBytes, etcdAddress, duplicate)
		if err != nil {
			return nil, err
		}
		return blockAPIServer, nil
	case AmazonBackendEnvVar:
		// amazon doesn't like leading slashes
		if len(dir) > 0 && dir[0] == '/' {
			dir = dir[1:]
		}
		blockAPIServer, err := newAmazonBlockAPIServer(dir, cacheBytes, etcdAddress, duplicate)
		if err != nil {
			return nil, err
		}
		return blockAPIServer, nil
	case GoogleBackendEnvVar:
		// TODO figure out if google likes leading slashses
		blockAPIServer, err := newGoogleBlockAPIServer(dir, cacheBytes, etcdAddress, duplicate)
		if err != nil {
			return nil, err
		}
		return blockAPIServer, nil
	case MicrosoftBackendEnvVar:
		blockAPIServer, err := newMicrosoftBlockAPIServer(dir, cacheBytes, etcdAddress, duplicate)
		if err != nil {
			return nil, err
		}
		return blockAPIServer, nil
	case LocalBackendEnvVar:
		fallthrough
	default:
		blockAPIServer, err := newLocalBlockAPIServer(dir, cacheBytes, etcdAddress, duplicate)
		if err != nil {
			return nil, err
		}
		return blockAPIServer, nil
	}
}

// NewObjClient creates an obj.Client by selecting a construcot from the obj package.
func NewObjClient(conf *serviceenv.Configuration) (obj.Client, error) {
	dir := conf.StorageRoot
	switch conf.StorageBackend {
	case MinioBackendEnvVar:
		// S3 compatible doesn't like leading slashes
		if len(dir) > 0 && dir[0] == '/' {
			dir = dir[1:]
		}
		return obj.NewMinioClientFromSecret(dir)

	case AmazonBackendEnvVar:
		// amazon doesn't like leading slashes
		if len(dir) > 0 && dir[0] == '/' {
			dir = dir[1:]
		}
		return obj.NewAmazonClientFromSecret(dir)

	case GoogleBackendEnvVar:
		// TODO figure out if google likes leading slashses
		return obj.NewGoogleClientFromSecret(dir)

	case MicrosoftBackendEnvVar:
		return obj.NewMicrosoftClientFromSecret(dir)

	case LocalBackendEnvVar:
		fallthrough

	default:
		return obj.NewLocalClient(dir)
	}
}
