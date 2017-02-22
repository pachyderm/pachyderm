package worker

import (
	"golang.org/x/net/context"

	etcd "github.com/coreos/etcd/clientv3"
)

type ApiServer struct {
	// A client connected to the Etcd instance at 'EtcdAddr'
	EtcdClient *etcd.Client
}

func NewApiServer(port uint16, etcdAddr string) *ApiServer {
	return &ApiServer{}
}

func (a *ApiServer) Process(ctx context.Context, req *ProcessRequest) (*ProcessResponse, error) {
	return nil, nil
}
