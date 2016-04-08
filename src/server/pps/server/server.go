package server

import (
	"sync"

	"github.com/pachyderm/pachyderm/src/client/pkg/shard"
	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
	ppsserver "github.com/pachyderm/pachyderm/src/server/pps"
	"go.pedge.io/proto/rpclog"
	kube "k8s.io/kubernetes/pkg/client/unversioned"
)

type APIServer interface {
	ppsclient.APIServer
	ppsserver.InternalJobAPIServer
	shard.Frontend
	shard.Server
}

func NewAPIServer(
	hasher *ppsserver.Hasher,
	address string,
	kubeClient *kube.Client,
) APIServer {
	return &apiServer{
		Logger:                 protorpclog.NewLogger("pachyderm.ppsclient.API"),
		hasher:                 hasher,
		address:                address,
		pfsAPIClient:           nil,
		pfsClientOnce:          sync.Once{},
		persistAPIClient:       nil,
		persistClientOnce:      sync.Once{},
		kubeClient:             kubeClient,
		cancelFuncs:            make(map[ppsclient.Pipeline]func()),
		cancelFuncsLock:        sync.Mutex{},
		shardPipelineChans:     make(map[uint64]chan *ppsclient.PipelineInfo),
		shardPipelineChansLock: sync.Mutex{},
		pipelineSubscribed:     false,
		pipelineSubscribedLock: sync.Mutex{},
		version:                shard.InvalidVersion,
		versionLock:            sync.RWMutex{},
	}
}
