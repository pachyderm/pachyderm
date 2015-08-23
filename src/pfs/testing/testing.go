package testing

import (
	"errors"
	"fmt"
	"os"
	"sync/atomic"
	"testing"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/drive"
	"github.com/pachyderm/pachyderm/src/pfs/drive/btrfs"
	"github.com/pachyderm/pachyderm/src/pfs/route"
	"github.com/pachyderm/pachyderm/src/pfs/server"
	"github.com/pachyderm/pachyderm/src/pkg/discovery"
	"github.com/pachyderm/pachyderm/src/pkg/grpctest"
	"github.com/pachyderm/pachyderm/src/pkg/grpcutil"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

const (
	// TODO(pedge): large numbers of shards takes forever because
	// we are doing tons of btrfs operations on init, is there anything
	// we can do about that?
	testShardsPerServer = 8
	testNumServers      = 8
)

var (
	counter int32
)

func TestRepositoryName() string {
	// TODO could be nice to add callee to this string to make it easy to
	// recover results for debugging
	return fmt.Sprintf("test-%d", atomic.AddInt32(&counter, 1))
}

func RunTest(
	t *testing.T,
	f func(t *testing.T, apiClient pfs.ApiClient, internalAPIClient pfs.InternalApiClient),
) {
	discoveryClient, err := getEtcdClient()
	require.NoError(t, err)
	grpctest.Run(
		t,
		testNumServers,
		func(servers map[string]*grpc.Server) {
			registerFunc(t, discoveryClient, servers)
		},
		func(t *testing.T, clientConns map[string]*grpc.ClientConn) {
			var clientConn *grpc.ClientConn
			for _, c := range clientConns {
				clientConn = c
				break
			}
			f(
				t,
				pfs.NewApiClient(
					clientConn,
				),
				pfs.NewInternalApiClient(
					clientConn,
				),
			)
		},
	)
}

func RunBench(
	b *testing.B,
	f func(b *testing.B, apiClient pfs.ApiClient),
) {
	discoveryClient, err := getEtcdClient()
	require.NoError(b, err)
	grpctest.RunB(
		b,
		testNumServers,
		func(servers map[string]*grpc.Server) {
			registerFunc(b, discoveryClient, servers)
		},
		func(b *testing.B, clientConns map[string]*grpc.ClientConn) {
			var clientConn *grpc.ClientConn
			for _, c := range clientConns {
				clientConn = c
				break
			}
			f(
				b,
				pfs.NewApiClient(
					clientConn,
				),
			)
		},
	)
}

func registerFunc(tb testing.TB, discoveryClient discovery.Client, servers map[string]*grpc.Server) error {
	addresser := route.NewDiscoveryAddresser(
		discoveryClient,
		testNamespace(),
	)
	i := 0
	for address := range servers {
		for j := 0; j < testShardsPerServer; j++ {
			if _, err := addresser.SetMasterAddress((i*testShardsPerServer)+j, address, 0); err != nil {
				return err
			}
			if _, err := addresser.SetReplicaAddress((((i+1)%len(servers))*testShardsPerServer)+j, address, 0); err != nil {
				return err
			}
			if _, err := addresser.SetReplicaAddress((((i+2)%len(servers))*testShardsPerServer)+j, address, 0); err != nil {
				return err
			}
		}
		i++
	}
	for address, s := range servers {
		combinedAPIServer := server.NewCombinedAPIServer(
			route.NewSharder(
				testShardsPerServer*testNumServers,
			),
			route.NewRouter(
				addresser,
				grpcutil.NewDialer(),
				address,
			),
			getDriver(tb, address),
		)
		pfs.RegisterApiServer(s, combinedAPIServer)
		pfs.RegisterInternalApiServer(s, combinedAPIServer)
	}
	return nil
}

func getDriver(tb testing.TB, namespace string) drive.Driver {
	driver, err := btrfs.NewDriver(getBtrfsRootDir(tb), namespace)
	require.NoError(tb, err)
	return driver
}

func getBtrfsRootDir(tb testing.TB) string {
	// TODO(pedge)
	rootDir := os.Getenv("PFS_DRIVER_ROOT")
	if rootDir == "" {
		tb.Fatal("PFS_DRIVER_ROOT not set")
	}
	return rootDir
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

func testNamespace() string {
	return fmt.Sprintf("test-%d", atomic.AddInt32(&counter, 1))
}
