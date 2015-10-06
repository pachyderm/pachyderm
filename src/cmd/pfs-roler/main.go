package main

import (
	"errors"
	"fmt"
	"os"

	"go.pedge.io/env"
	"go.pedge.io/protolog/logrus"

	"go.pachyderm.com/pachyderm/src/pfs/route"
	"go.pachyderm.com/pachyderm/src/pkg/discovery"
)

var (
	defaultEnv = map[string]string{
		"PFS_NUM_SHARDS":   "16",
		"PFS_NUM_REPLICAS": "0",
	}
)

type appEnv struct {
	NumShards   uint64 `env:"PFS_NUM_SHARDS"`
	NumReplicas uint64 `env:"PFS_NUM_REPLICAS"`
}

func main() {
	env.Main(do, &appEnv{}, defaultEnv)
}

func do(appEnvObj interface{}) error {
	appEnv := appEnvObj.(*appEnv)
	logrus.Register()
	discoveryClient, err := getEtcdClient()
	if err != nil {
		return err
	}
	sharder := route.NewSharder(appEnv.NumShards, appEnv.NumReplicas)
	addresser := route.NewDiscoveryAddresser(
		discoveryClient,
		sharder,
		"namespace",
	)
	return addresser.AssignRoles(nil)
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
