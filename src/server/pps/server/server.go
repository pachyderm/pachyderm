package server

import (
	"path"
	"sync"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pkg/shard"
	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
	col "github.com/pachyderm/pachyderm/src/server/pkg/collection"
	"github.com/pachyderm/pachyderm/src/server/pkg/metrics"
	ppsserver "github.com/pachyderm/pachyderm/src/server/pps"

	etcd "github.com/coreos/etcd/clientv3"
	"go.pedge.io/proto/rpclog"
	"golang.org/x/net/context"
	kube "k8s.io/kubernetes/pkg/client/unversioned"
)

// APIServer represents an api server.
type APIServer interface {
	ppsclient.APIServer
	shard.Frontend
	shard.Server
}

const (
	pipelinesPrefix   = "/pipelines"
	jobsPrefix        = "/jobs"
	jobsPipelineIndex = "Pipeline"
	jobsInputsIndex   = "Inputs"
	stoppedIndex      = "Stopped"
)

// NewAPIServer creates an APIServer.
func NewAPIServer(
	etcdAddress string,
	etcdPrefix string,
	hasher *ppsserver.Hasher,
	address string,
	kubeClient *kube.Client,
	namespace string,
	workerImage string,
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
		pachConnOnce:          sync.Once{},
		kubeClient:            kubeClient,
		version:               shard.InvalidVersion,
		shardCtxs:             make(map[uint64]*ctxAndCancel),
		pipelineCancels:       make(map[string]context.CancelFunc),
		jobCancels:            make(map[string]context.CancelFunc),
		workerPools:           make(map[string]WorkerPool),
		namespace:             namespace,
		workerImage:           workerImage,
		workerImagePullPolicy: workerImagePullPolicy,
		reporter:              reporter,
		pipelines: col.NewCollection(
			etcdClient,
			path.Join(etcdPrefix, pipelinesPrefix),
			[]col.Index{stoppedIndex},
		),
		jobs: col.NewCollection(
			etcdClient,
			path.Join(etcdPrefix, jobsPrefix),
			[]col.Index{jobsPipelineIndex, stoppedIndex, jobsInputsIndex},
		),
	}
	go apiServer.pipelineWatcher()
	go apiServer.jobWatcher()
	return apiServer, nil
}
