package migration

import (
	"context"
	"fmt"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/server/pkg/migration/one_six"
)

func oneSixToOneSeven(etcdAddress string, etcdPrefix string) error {
	etcdClient, err := etcd.New(etcd.Config{
		Endpoints:   []string{fmt.Sprintf("%s:2379", etcdAddress)},
		DialOptions: client.EtcdDialOptions(),
	})
	if err != nil {
		return err
	}
	ctx := context.Background()
	if err := one_six.Repos(ctx, etcdClient, etcdPrefix, func(repoInfo *one_six.RepoInfo) error {
		return nil
	}); err != nil {
		return err
	}
	return nil
}
