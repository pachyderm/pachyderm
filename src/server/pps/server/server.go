package server

import (
	"sync"

	"github.com/pachyderm/pachyderm/src/client/pkg/shard"
	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
	ppsserver "github.com/pachyderm/pachyderm/src/server/pps"
	"github.com/pachyderm/pachyderm/src/server/pps/persist"
	"go.pedge.io/proto/rpclog"
	kube "k8s.io/kubernetes/pkg/client/unversioned"
)

type APIServer interface {
	ppsclient.APIServer
	ppsserver.InternalJobAPIServer
	shard.Frontend
}

func NewAPIServer(
	hasher *ppsserver.Hasher,
	router shard.Router,
	pfsAddress string,
	persistAPIServer persist.APIServer,
	kubeClient *kube.Client,
) APIServer {
	return &apiServer{
		Logger:               protorpclog.NewLogger("pachyderm.ppsclient.API"),
		hasher:               hasher,
		router:               router,
		pfsAddress:           pfsAddress,
		pfsAPIClient:         nil,
		pfsClientOnce:        sync.Once{},
		persistAPIServer:     persistAPIServer,
		kubeClient:           kubeClient,
		cancelFuncs:          make(map[ppsclient.Pipeline]func()),
		cancelFuncsLock:      sync.Mutex{},
		shardCancelFuncs:     make(map[uint64]func()),
		shardCancelFuncsLock: sync.Mutex{},
		version:              shard.InvalidVersion,
		versionLock:          sync.RWMutex{},
	}
}
