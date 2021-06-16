package server

import (
	"github.com/pachyderm/pachyderm/src/server/pkg/log"
	"github.com/pachyderm/pachyderm/src/server/pkg/metrics"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsdb"
	"github.com/pachyderm/pachyderm/src/server/pkg/serviceenv"
	txnenv "github.com/pachyderm/pachyderm/src/server/pkg/transactionenv"
	pps_iface "github.com/pachyderm/pachyderm/src/server/pps"
)

// NewAPIServer creates an APIServer.
func NewAPIServer(
	env *serviceenv.ServiceEnv,
	txnEnv *txnenv.TransactionEnv,
	etcdPrefix string,
	namespace string,
	workerImage string,
	workerSidecarImage string,
	workerImagePullPolicy string,
	storageRoot string,
	storageBackend string,
	storageHostPath string,
	cacheRoot string,
	iamRole string,
	imagePullSecret string,
	noExposeDockerSocket bool,
	reporter *metrics.Reporter,
	workerUsesRoot bool,
	workerGrpcPort uint16,
	port uint16,
	httpPort uint16,
	peerPort uint16,
	gcPercent int,
) (pps_iface.APIServer, error) {
	apiServer := &apiServer{
		Logger:                log.NewLogger("pps.API"),
		env:                   env,
		txnEnv:                txnEnv,
		etcdPrefix:            etcdPrefix,
		namespace:             namespace,
		workerImage:           workerImage,
		workerSidecarImage:    workerSidecarImage,
		workerImagePullPolicy: workerImagePullPolicy,
		storageRoot:           storageRoot,
		storageBackend:        storageBackend,
		storageHostPath:       storageHostPath,
		cacheRoot:             cacheRoot,
		iamRole:               iamRole,
		imagePullSecret:       imagePullSecret,
		noExposeDockerSocket:  noExposeDockerSocket,
		reporter:              reporter,
		workerUsesRoot:        workerUsesRoot,
		pipelines:             ppsdb.Pipelines(env.GetEtcdClient(), etcdPrefix),
		jobs:                  ppsdb.Jobs(env.GetEtcdClient(), etcdPrefix),
		workerGrpcPort:        workerGrpcPort,
		port:                  port,
		httpPort:              httpPort,
		peerPort:              peerPort,
		gcPercent:             gcPercent,
	}
	apiServer.validateKube()
	go apiServer.master()
	return apiServer, nil
}

// NewSidecarAPIServer creates an APIServer that has limited functionalities
// and is meant to be run as a worker sidecar.  It cannot, for instance,
// create pipelines.
func NewSidecarAPIServer(
	env *serviceenv.ServiceEnv,
	txnEnv *txnenv.TransactionEnv,
	etcdPrefix string,
	namespace string,
	iamRole string,
	reporter *metrics.Reporter,
	workerGrpcPort uint16,
	httpPort uint16,
	peerPort uint16,
) (pps_iface.APIServer, error) {
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
	return apiServer, nil
}
