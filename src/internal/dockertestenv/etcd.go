package dockertestenv

import (
	"testing"

	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	etcd "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/pkg/mock/mockserver"
)

func NewTestEtcd(t *testing.T) *etcd.Client {
	ms, err := mockserver.StartMockServers(1)
	require.NoError(t, err)
	mockEtcd := ms.Servers[0]
	mockEtcdClient, err := etcd.New(etcd.Config{
		Endpoints: []string{mockEtcd.ResolverAddress().Addr},
	})
	require.NoError(t, err)
	return mockEtcdClient
}
