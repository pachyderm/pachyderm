package worker_shim

import (
	"context"
)

type AppEnv struct {
	Port        uint16 `env:"PORT,default=650"`
	EtcdAddress string `env:"ETCD_PORT_2379_TCP_ADDR,required"`
}

type apiServer struct {
	env AppEnv
}

func NewApiServer(env *AppEnv) *apiServer {
	return new(apiServer)
}

func getEtcdClient(env *AppEnv) *etc.Client {
	etcd.NewClient(fmt.Sprintf("http://%s:2379", env.EtcdAddress))
	etc
}

func (a *apiServer) Process(ctx context.Context, req *ProcessRequest) (*ProcessResponse, error) {
	return nil, nil
}
