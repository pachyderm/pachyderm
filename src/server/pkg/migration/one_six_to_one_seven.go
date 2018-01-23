package migration

import (
	"fmt"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/server/pkg/pfsdb"
)

func oneSixToOneSeven(etcdAddress string, etcdPrefix string) error {
	etcdClient, err := etcd.New(etcd.Config{
		Endpoints:   []string{fmt.Sprintf("%s:2379", etcdAddress)},
		DialOptions: client.EtcdDialOptions(),
	})
	if err != nil {
		return err
	}
	repos := pfsdb.Repos(etcdClient, etcdPrefix)
	return nil
}
