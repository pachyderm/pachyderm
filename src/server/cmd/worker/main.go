package main

import (
	"path"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
	wshim "github.com/pachyderm/pachyderm/src/server/pkg/worker"
)

type AppEnv struct {
	Port        uint16 `env:"PORT,default=650"`
	EtcdAddress string `env:"ETCD_PORT_2379_TCP_ADDR,required"`
	PpsPrefix   string `env:"PPS_PREFIX,required"`
	PpsWorkerIp string `env:"PPS_WORKER_IP,required"`
}

func main() {
	cmdutil.Main(do, &wshim.AppEnv{})
}

func do(appEnvObj interface{}) error {
	appEnv := appEnvObj.(*wshim.AppEnv)
	etcdClient := etcd.New(etcd.Config{
		Endpoints:   []string{fmt.Sprintf("%s:2379", EtcdAddress)},
		DialTimeout: 15 * time.Second,
	})
	apiServer := wshim.ApiServer{
		EtcdClient: etcdClient,
	}
	err := grpcutil.Serve(
		func(s *grpc.Server) {
			wshim.RegisterWorkerShimServer(s, workerShimServer)
		},
		grpcutil.ServeOptions{
			Version:    version.Version,
			MaxMsgSize: client.MaxMsgSize,
		},
		grpcutil.ServeEnv{
			GRPCPort: appEnv.Port,
		},
	)
	if err != nil {
		return err
	}
	etcdClient.Put(
		context.WithTimeout(context.Background(), 30*time.Second),
		path.Join(PpsPrefix, "workers", PpsWorkerIp), "")
}
