package server

import (
	"sync"

	"github.com/pachyderm/pachyderm/src/pkg/shard"
	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
	ppsserver "github.com/pachyderm/pachyderm/src/server/pps"
	"github.com/pachyderm/pachyderm/src/pps/persist"
	"go.pedge.io/proto/rpclog"
	kube "k8s.io/kubernetes/pkg/client/unversioned"
)

type APIServer interface {
	ppsclient.APIServer
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
		protorpclog.NewLogger("pachyderm.ppsclient.API"),
		hasher,
		router,
		pfsAddress,
		nil,
		sync.Once{},
		persistAPIServer,
		kubeClient,
		make(map[string]*jobState),
		sync.Mutex{},
		make(map[ppsclient.Pipeline]func()),
		sync.Mutex{},
		shard.InvalidVersion,
		sync.RWMutex{},
	}
}
