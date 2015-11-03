package main

import (
	"errors"
	"fmt"
	"os"

	"go.pedge.io/env"

	"github.com/pachyderm/pachyderm/src/pkg/discovery"
	"github.com/pachyderm/pachyderm/src/pkg/shard"
)

type appEnv struct {
	NumShards   uint64 `env:"PFS_NUM_SHARDS,default=16"`
	NumReplicas uint64 `env:"PFS_NUM_REPLICAS"`
}

func main() {
	env.Main(do, &appEnv{})
}

func do(appEnvObj interface{}) error {
	appEnv := appEnvObj.(*appEnv)
	discoveryClient, err := getEtcdClient()
	if err != nil {
		return err
	}
	sharder := shard.NewSharder(
		discoveryClient,
		appEnv.NumShards,
		appEnv.NumReplicas,
		"namespace",
	)
	return sharder.AssignRoles(nil)
}

func getEtcdClient() (discovery.Client, error) {
	etcdAddress, err := getEtcdAddress()
	if err != nil {
		return nil, err
	}
	return discovery.NewEtcdClient(etcdAddress), nil
}

func getEtcdAddress() (string, error) {
	etcdAddr := os.Getenv("ETCD_PORT_2379_TCP_ADDR")
	if etcdAddr == "" {
		return "", errors.New("ETCD_PORT_2379_TCP_ADDR not set")
	}
	return fmt.Sprintf("http://%s:2379", etcdAddr), nil
}
