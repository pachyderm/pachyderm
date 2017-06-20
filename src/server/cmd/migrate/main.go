package main

import (
	"fmt"

	"github.com/pachyderm/pachyderm/src/client/pkg/discovery"
	"github.com/pachyderm/pachyderm/src/server/pkg/cmdutil"
)

type appEnv struct {
	EtcdAddress   string `env:"ETCD_PORT_2379_TCP_ADDR,required"`
	PPSEtcdPrefix string `env:"PPS_ETCD_PREFIX,default=pachyderm_pps"`
	PFSEtcdPrefix string `env:"PFS_ETCD_PREFIX,default=pachyderm_pfs"`
}

func main() {
	cmdutil.Main(do, &appEnv{})
}

func do(appEnvObj interface{}) error {
	appEnv := appEnvObj.(*appEnv)
	etcdClient := discovery.NewEtcdClient(fmt.Sprintf("http://%s:2379", appEnv.EtcdAddress))

	return nil
}
