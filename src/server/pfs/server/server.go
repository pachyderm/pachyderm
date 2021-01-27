package server

import (
	"github.com/jmoiron/sqlx"
	pfsclient "github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/internal/obj"
	"github.com/pachyderm/pachyderm/src/internal/serviceenv"
	txnenv "github.com/pachyderm/pachyderm/src/internal/transactionenv"
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
func NewAPIServer(env *serviceenv.ServiceEnv, txnEnv *txnenv.TransactionEnv, etcdPrefix string, db *sqlx.DB) (APIServer, error) {
	a, err := newAPIServer(env, txnEnv, etcdPrefix, db)
	if err != nil {
		return nil, err
	}
	return newValidatedAPIServer(a, env), nil
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

// NewObjClient creates an obj.Client by selecting a constructor from the obj package.
// TODO: Not sure if we want to keep the storage root configuration for non-local deployments.
// If so, we will need to connect it to the object path prefix for chunks.
func NewObjClient(conf *serviceenv.Configuration) (obj.Client, error) {
	switch conf.StorageBackend {
	case MinioBackendEnvVar:
		// S3 compatible doesn't like leading slashes
		// TODO: Readd?
		//if len(dir) > 0 && dir[0] == '/' {
		//	dir = dir[1:]
		//}
		return obj.NewMinioClientFromSecret("")

	case AmazonBackendEnvVar:
		// amazon doesn't like leading slashes
		// TODO: Readd?
		//if len(dir) > 0 && dir[0] == '/' {
		//	dir = dir[1:]
		//}
		return obj.NewAmazonClientFromSecret("")

	case GoogleBackendEnvVar:
		// TODO figure out if google likes leading slashses
		return obj.NewGoogleClientFromSecret("")

	case MicrosoftBackendEnvVar:
		return obj.NewMicrosoftClientFromSecret("")

	case LocalBackendEnvVar:
		fallthrough

	default:
		return obj.NewLocalClient(conf.StorageRoot)
	}
}
