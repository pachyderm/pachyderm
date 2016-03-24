package server

import (
	"sync"

	ppsclient "github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/server/pkg/shard"
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
