package main

import (
	"context"
	"fmt"
	"path"
	"time"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/gogo/protobuf/proto"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/client/version"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/worker"
	"google.golang.org/grpc"
)

type AppEnv struct {
	Port            uint16 `env:"PORT,default=650"`
	EtcdAddress     string `env:"ETCD_PORT_2379_TCP_ADDR,required"`
	PachdAddress    string `env:"PACHD_PORT_650_TCP_ADDR,required"`
	PPSPipelineName string `env:"PPS_PIPELINE_NAME,required"`
	PPSPrefix       string `env:"PPS_ETCD_PREFIX,required"`
	PPSWorkerIP     string `env:"PPS_WORKER_IP,required"`
}

func main() {
	cmdutil.Main(do, &AppEnv{})
}

func putAddress(appEnv *AppEnv, etcdClient *etcd.Client) error {
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	_, err := etcdClient.Put(ctx, path.Join(appEnv.PPSPrefix, "workers", appEnv.PPSPipelineName, appEnv.PPSWorkerIP), "")
	return err
}

func getPipelineInfo(appEnv *AppEnv, etcdClient *etcd.Client) (*pps.PipelineInfo, error) {
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	resp, err := etcdClient.Get(ctx, path.Join(appEnv.PPSPrefix, "pipelines", appEnv.PPSPipelineName))
	if err != nil {
		return nil, err
	}
	if len(resp.Kvs) != 1 {
		return nil, fmt.Errorf("expected to find 1 pipeline, got %d: %v", len(resp.Kvs), resp)
	}
	pipelineInfo := new(pps.PipelineInfo)
	if err := proto.UnmarshalText(string(resp.Kvs[0].Value), pipelineInfo); err != nil {
		return nil, err
	}
	return pipelineInfo, nil
}

func do(appEnvObj interface{}) error {
	appEnv := appEnvObj.(*AppEnv)
	pachClient, err := client.NewFromAddress(fmt.Sprintf("%v:650", appEnv.PachdAddress))
	if err != nil {
		return err
	}
	etcdClient, err := etcd.New(etcd.Config{
		Endpoints:   []string{fmt.Sprintf("%s:2379", appEnv.EtcdAddress)},
		DialTimeout: 15 * time.Second,
	})
	if err != nil {
		return err
	}
	pipelineInfo, err := getPipelineInfo(appEnv, etcdClient)
	if err != nil {
		return err
	}
	if err := putAddress(appEnv, etcdClient); err != nil {
		return err
	}
	apiServer := worker.NewAPIServer(pachClient, etcdClient, pipelineInfo)
	return grpcutil.Serve(
		func(s *grpc.Server) {
			worker.RegisterWorkerServer(s, apiServer)
		},
		grpcutil.ServeOptions{
			Version:    version.Version,
			MaxMsgSize: client.MaxMsgSize,
		},
		grpcutil.ServeEnv{
			GRPCPort: appEnv.Port,
		},
	)
}
