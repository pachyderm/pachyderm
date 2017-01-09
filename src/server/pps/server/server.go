package server

import (
	"sync"

	"github.com/pachyderm/pachyderm/src/client/pkg/shard"
	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/metrics"
	ppsserver "github.com/pachyderm/pachyderm/src/server/pps"

	kube "k8s.io/kubernetes/pkg/client/unversioned"
)

// APIServer represents an api server.
type APIServer interface {
	ppsclient.APIServer
	ppsserver.InternalPodAPIServer
	shard.Frontend
	shard.Server
}

// NewAPIServer creates an APIServer.
func NewAPIServer(
	hasher *ppsserver.Hasher,
	address string,
	kubeClient *kube.Client,
	namespace string,
	jobShimImage string,
	jobImagePullPolicy string,
	reporter *metrics.Reporter,
) APIServer {
	return &apiServer{
		hasher:                  hasher,
		address:                 address,
		pfsAPIClient:            nil,
		pfsClientOnce:           sync.Once{},
		persistAPIClient:        nil,
		persistClientOnce:       sync.Once{},
		kubeClient:              kubeClient,
		shardCancelFuncs:        make(map[uint64]func()),
		shardCancelFuncsLock:    sync.Mutex{},
		pipelineCancelFuncs:     make(map[string]func()),
		pipelineCancelFuncsLock: sync.Mutex{},
		jobCancelFuncs:          make(map[string]func()),
		jobCancelFuncsLock:      sync.Mutex{},
		version:                 shard.InvalidVersion,
		versionLock:             sync.RWMutex{},
		namespace:               namespace,
		jobShimImage:            jobShimImage,
		jobImagePullPolicy:      jobImagePullPolicy,
		reporter:                reporter,
	}
}
