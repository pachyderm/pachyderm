package server

import (
	etcd "github.com/coreos/etcd/clientv3"

	client "github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"
)

func NewAPIServer(etcdAddress string) (auth.APIServer, error) {
	etcdClient, err := etcd.New(etcd.Config{
		Endpoints:   []string{etcdAddress},
		DialOptions: client.EtcdDialOptions(),
	})
	if err != nil {
		return nil, err
	}
	return &apiServer{
		etcdClient: etcdClient,
	}, nil
}
