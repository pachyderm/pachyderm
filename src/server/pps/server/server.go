package server

import (
	"path"

	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/metrics"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	ppsiface "github.com/pachyderm/pachyderm/v2/src/server/pps"
)

// NewAPIServer creates an APIServer.
func NewAPIServer(
	env serviceenv.ServiceEnv,
	txnEnv *txnenv.TransactionEnv,
	reporter *metrics.Reporter,
) (ppsiface.APIServer, error) {
	etcdPrefix := path.Join(env.Config().EtcdPrefix, env.Config().PPSEtcdPrefix)
	apiServer := &apiServer{
		Logger:                log.NewLogger("pps.API", env.Logger()),
		env:                   env,
		txnEnv:                txnEnv,
		etcdPrefix:            etcdPrefix,
		namespace:             env.Config().Namespace,
		workerImage:           env.Config().WorkerImage,
		workerSidecarImage:    env.Config().WorkerSidecarImage,
		workerImagePullPolicy: env.Config().WorkerImagePullPolicy,
		storageRoot:           env.Config().StorageRoot,
		storageBackend:        env.Config().StorageBackend,
		storageHostPath:       env.Config().StorageHostPath,
		imagePullSecrets:      env.Config().ImagePullSecrets,
		reporter:              reporter,
		workerUsesRoot:        env.Config().WorkerUsesRoot,
		pipelines:             ppsdb.Pipelines(env.GetDBClient(), env.GetPostgresListener()),
		jobs:                  ppsdb.Jobs(env.GetDBClient(), env.GetPostgresListener()),
		workerGrpcPort:        env.Config().PPSWorkerPort,
		port:                  env.Config().Port,
		peerPort:              env.Config().PeerPort,
		gcPercent:             env.Config().GCPercent,
	}
	apiServer.validateKube()
	go apiServer.master()
	return apiServer, nil
}

// NewSidecarAPIServer creates an APIServer that has limited functionalities
// and is meant to be run as a worker sidecar.  It cannot, for instance,
// create pipelines.
func NewSidecarAPIServer(
	env serviceenv.ServiceEnv,
	txnEnv *txnenv.TransactionEnv,
	etcdPrefix string,
	namespace string,
	reporter *metrics.Reporter,
	workerGrpcPort uint16,
	peerPort uint16,
) (*apiServer, error) {
	apiServer := &apiServer{
		Logger:         log.NewLogger("pps.API", env.Logger()),
		env:            env,
		txnEnv:         txnEnv,
		etcdPrefix:     etcdPrefix,
		reporter:       reporter,
		namespace:      namespace,
		workerUsesRoot: true,
		pipelines:      ppsdb.Pipelines(env.GetDBClient(), env.GetPostgresListener()),
		jobs:           ppsdb.Jobs(env.GetDBClient(), env.GetPostgresListener()),
		workerGrpcPort: workerGrpcPort,
		peerPort:       peerPort,
	}
	go apiServer.ServeSidecarS3G()
	return apiServer, nil
}
