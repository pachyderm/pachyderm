package server

import (
	"github.com/pachyderm/pachyderm/src/client"
	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/log"
	"github.com/pachyderm/pachyderm/src/server/pkg/metrics"
	"github.com/pachyderm/pachyderm/src/server/pkg/ppsdb"

	etcd "github.com/coreos/etcd/clientv3"
	kube "k8s.io/client-go/kubernetes"
)

// NewAPIServer creates an APIServer.
func NewAPIServer(
	etcdAddress string,
	etcdPrefix string,
	address string,
	kubeClient *kube.Clientset,
	namespace string,
	workerImage string,
	workerSidecarImage string,
	workerImagePullPolicy string,
	storageRoot string,
	storageBackend string,
	storageHostPath string,
	iamRole string,
	imagePullSecret string,
	reporter *metrics.Reporter,
) (ppsclient.APIServer, error) {
	etcdClient, err := etcd.New(etcd.Config{
		Endpoints:   []string{etcdAddress},
		DialOptions: client.EtcdDialOptions(),
	})
	if err != nil {
		return nil, err
	}

	apiServer := &apiServer{
		Logger:                log.NewLogger("pps.API"),
		etcdPrefix:            etcdPrefix,
		address:               address,
		etcdClient:            etcdClient,
		kubeClient:            kubeClient,
		namespace:             namespace,
		workerImage:           workerImage,
		workerSidecarImage:    workerSidecarImage,
		workerImagePullPolicy: workerImagePullPolicy,
		storageRoot:           storageRoot,
		storageBackend:        storageBackend,
		storageHostPath:       storageHostPath,
		iamRole:               iamRole,
		imagePullSecret:       imagePullSecret,
		reporter:              reporter,
		pipelines:             ppsdb.Pipelines(etcdClient, etcdPrefix),
		jobs:                  ppsdb.Jobs(etcdClient, etcdPrefix),
	}
	apiServer.validateKube()
	go apiServer.master()
	return apiServer, nil
}

// NewSidecarAPIServer creates an APIServer that has limited functionalities
// and is meant to be run as a worker sidecar.  It cannot, for instance,
// create pipelines.
func NewSidecarAPIServer(
	etcdAddress string,
	etcdPrefix string,
	address string,
	iamRole string,
	reporter *metrics.Reporter,
) (ppsclient.APIServer, error) {
	etcdClient, err := etcd.New(etcd.Config{
		Endpoints:   []string{etcdAddress},
		DialOptions: client.EtcdDialOptions(),
	})
	if err != nil {
		return nil, err
	}

	apiServer := &apiServer{
		Logger:     log.NewLogger("pps.API"),
		address:    address,
		etcdPrefix: etcdPrefix,
		etcdClient: etcdClient,
		iamRole:    iamRole,
		reporter:   reporter,
		pipelines:  ppsdb.Pipelines(etcdClient, etcdPrefix),
		jobs:       ppsdb.Jobs(etcdClient, etcdPrefix),
	}
	return apiServer, nil
}
