package server

import (
	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	pfsclient "github.com/pachyderm/pachyderm/v2/src/pfs"
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

// NewAPIServer creates an APIServer.
func NewAPIServer(env *serviceenv.ServiceEnv, txnEnv *txnenv.TransactionEnv, etcdPrefix string, db *sqlx.DB) (APIServer, error) {
	a, err := newAPIServer(env, txnEnv, etcdPrefix, db)
	if err != nil {
		return nil, err
	}
	return newValidatedAPIServer(a, env), nil
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
