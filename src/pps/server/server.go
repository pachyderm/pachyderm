package server

import (
	"sync"

	"github.com/pachyderm/pachyderm/src/pkg/shard"
	"github.com/pachyderm/pachyderm/src/pps"
	"github.com/pachyderm/pachyderm/src/pps/persist"
	"go.pedge.io/proto/rpclog"
	kube "k8s.io/kubernetes/pkg/client/unversioned"
)

type APIServer interface {
	pps.APIServer
	shard.Frontend
}

func NewAPIServer(
	hasher *pps.Hasher,
	router shard.Router,
	pfsAddress string,
	persistAPIServer persist.APIServer,
	kubeClient *kube.Client,
) APIServer {
	return &apiServer{
		protorpclog.NewLogger("pachyderm.pps.API"),
		hasher,
		router,
		pfsAddress,
		nil,
		sync.Once{},
		persistAPIServer,
		kubeClient,
		make(map[string]*jobState),
		sync.Mutex{},
		make(map[pps.Pipeline]func()),
		sync.Mutex{},
		shard.InvalidVersion,
		sync.RWMutex{},
	}
}
