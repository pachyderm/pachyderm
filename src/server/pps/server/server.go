package server

import (
	"path"

	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/metrics"
	"github.com/pachyderm/pachyderm/v2/src/internal/ppsdb"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	ppsclient "github.com/pachyderm/pachyderm/v2/src/pps"
)

// APIServer represents a PPS API server
type APIServer interface {
	ppsclient.APIServer
	txnenv.PpsTransactionServer
}

// NewAPIServer creates an APIServer.
func NewAPIServer(
	env serviceenv.ServiceEnv,
	txnEnv *txnenv.TransactionEnv,
	reporter *metrics.Reporter,
) (APIServer, error) {
	etcdPrefix := path.Join(env.Config().EtcdPrefix, env.Config().PPSEtcdPrefix)
	apiServer := &apiServer{
		Logger:                log.NewLogger("pps.API"),
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
		cacheRoot:             env.Config().CacheRoot,
		iamRole:               env.Config().IAMRole,
		imagePullSecret:       env.Config().ImagePullSecret,
		noExposeDockerSocket:  env.Config().NoExposeDockerSocket,
		reporter:              reporter,
		workerUsesRoot:        env.Config().WorkerUsesRoot,
		pipelines:             ppsdb.Pipelines(env.GetEtcdClient(), etcdPrefix),
		jobs:                  ppsdb.Jobs(env.GetEtcdClient(), etcdPrefix),
		workerGrpcPort:        env.Config().PPSWorkerPort,
		port:                  env.Config().Port,
		httpPort:              env.Config().HTTPPort,
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
	iamRole string,
	reporter *metrics.Reporter,
	workerGrpcPort uint16,
	httpPort uint16,
	peerPort uint16,
) (APIServer, error) {
	apiServer := &apiServer{
		Logger:         log.NewLogger("pps.API"),
		env:            env,
		txnEnv:         txnEnv,
		etcdPrefix:     etcdPrefix,
		iamRole:        iamRole,
		reporter:       reporter,
		namespace:      namespace,
		workerUsesRoot: true,
		pipelines:      ppsdb.Pipelines(env.GetEtcdClient(), etcdPrefix),
		jobs:           ppsdb.Jobs(env.GetEtcdClient(), etcdPrefix),
		workerGrpcPort: workerGrpcPort,
		httpPort:       httpPort,
		peerPort:       peerPort,
	}
	go apiServer.ServeSidecarS3G()
	return newSidecarAPIServer(apiServer, env), nil
}
