package server

import (
	"sync"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pkg/shard"
	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/metrics"
	ppsserver "github.com/pachyderm/pachyderm/src/server/pps"

	etcd "github.com/coreos/etcd/clientv3"
	"go.pedge.io/proto/rpclog"
	kube "k8s.io/kubernetes/pkg/client/unversioned"
)

// APIServer represents an api server.
type APIServer interface {
	ppsclient.APIServer
	shard.Frontend
	shard.Server
}

// NewAPIServer creates an APIServer.
func NewAPIServer(
	etcdAddress string,
	etcdPrefix string,
	hasher *ppsserver.Hasher,
	address string,
	kubeClient *kube.Client,
	namespace string,
	workerShimImage string,
	workerImagePullPolicy string,
	reporter *metrics.Reporter,
) (APIServer, error) {
	etcdClient, err := etcd.New(etcd.Config{
		Endpoints:   []string{etcdAddress},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	apiServer := &apiServer{
		Logger:                protorpclog.NewLogger("pps.API"),
		etcdPrefix:            etcdPrefix,
		hasher:                hasher,
		address:               address,
		etcdClient:            etcdClient,
		pfsAPIClient:          nil,
		pfsClientOnce:         sync.Once{},
		kubeClient:            kubeClient,
		version:               shard.InvalidVersion,
		versionLock:           sync.RWMutex{},
		shardCtxs:             make(map[uint64]*ctxAndCancel),
		namespace:             namespace,
		workerShimImage:       workerShimImage,
		workerImagePullPolicy: workerImagePullPolicy,
		reporter:              reporter,
	}
	go apiServer.pipelineWatcher()
	return apiServer, nil
}
