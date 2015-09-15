package testing

import (
	"errors"
	"fmt"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"go.pedge.io/proto/test"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/drive"
	"github.com/pachyderm/pachyderm/src/pfs/drive/btrfs"
	"github.com/pachyderm/pachyderm/src/pfs/role"
	"github.com/pachyderm/pachyderm/src/pfs/route"
	"github.com/pachyderm/pachyderm/src/pfs/server"
	"github.com/pachyderm/pachyderm/src/pkg/discovery"
	"github.com/pachyderm/pachyderm/src/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/pkg/require"
	"google.golang.org/grpc"
)

const (
	// TODO(pedge): large numbers of shards takes forever because
	// we are doing tons of btrfs operations on init, is there anything
	// we can do about that?
	testShardsPerServer = 8
	testNumServers      = 8
	testNumReplicas     = 1
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
	f func(*testing.T, pfs.ApiClient, pfs.InternalApiClient, Cluster),
) {
	discoveryClient, err := getEtcdClient()
	require.NoError(t, err)
	var cluster Cluster
	prototest.RunT(
		t,
		testNumServers,
		func(servers map[string]*grpc.Server) {
			cluster = registerFunc(t, discoveryClient, servers)
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
				cluster,
			)
		},
	)
	cluster.Shutdown()
}

func RunBench(
	b *testing.B,
	f func(*testing.B, pfs.ApiClient),
) {
	discoveryClient, err := getEtcdClient()
	require.NoError(b, err)
	var cluster Cluster
	prototest.RunB(
		b,
		testNumServers,
		func(servers map[string]*grpc.Server) {
			cluster = registerFunc(b, discoveryClient, servers)
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
	cluster.Shutdown()
}

type Cluster interface {
	WaitForAvailability()
	Kill(server int)
	Restart(server int)
	Shutdown()
}

type cluster struct {
	addresses       []string
	rolers          map[string]role.Roler
	servers         map[string]server.APIServer
	internalServers map[string]server.InternalAPIServer
	addresser       route.Addresser
	sharder         route.Sharder
	tb              testing.TB
}

func (c *cluster) WaitForAvailability() {
	cancel := make(chan bool)
	time.AfterFunc(60*time.Second, func() { close(cancel) })
	var _shardToMasterAddress map[int]route.Address
	var _shardToReplicaAddress map[int]map[int]route.Address
	err := c.addresser.WatchShardToAddress(cancel, func(shardToMasterAddress map[int]route.Address, shardToReplicaAddress map[int]map[int]route.Address) (uint64, error) {
		_shardToMasterAddress = shardToMasterAddress
		_shardToReplicaAddress = shardToReplicaAddress
		if len(shardToMasterAddress) != testShardsPerServer*testNumServers {
			return 0, nil
		}
		if len(shardToReplicaAddress) != testShardsPerServer*testNumServers {
			return 0, nil
		}
		for _, addresses := range shardToReplicaAddress {
			if len(addresses) != testNumReplicas {
				return 0, nil
			}
		}
		for _, address := range shardToMasterAddress {
			if address.Backfilling {
				return 0, nil
			}
			if _, ok := c.rolers[address.Address]; !ok {
				return 0, nil
			}
		}
		for _, addresses := range shardToReplicaAddress {
			for _, address := range addresses {
				if address.Backfilling {
					return 0, nil
				}
				if _, ok := c.rolers[address.Address]; !ok {
					return 0, nil
				}
			}
		}
		return 0, fmt.Errorf("Complete")
	})
	require.Equal(c.tb, err.Error(), "Complete")
}

func (c *cluster) Kill(server int) {
	c.rolers[c.addresses[server]].Cancel()
	delete(c.rolers, c.addresses[server])
}

func (c *cluster) Restart(server int) {
	c.rolers[c.addresses[server]] = role.NewRoler(c.addresser, c.sharder, c.internalServers[c.addresses[server]], c.addresses[server], testNumReplicas)
	go func() { require.Equal(c.tb, c.rolers[c.addresses[server]].Run(), discovery.ErrCancelled) }()
}

func (c *cluster) Shutdown() {
	for _, roler := range c.rolers {
		roler.Cancel()
	}
}

func newCluster(tb testing.TB, discoveryClient discovery.Client, servers map[string]*grpc.Server) Cluster {
	cluster := cluster{
		rolers:          make(map[string]role.Roler),
		servers:         make(map[string]server.APIServer),
		internalServers: make(map[string]server.InternalAPIServer),
		addresser: route.NewDiscoveryAddresser(
			discoveryClient,
			testNamespace(),
		),
		sharder: route.NewSharder(
			testShardsPerServer * testNumServers,
		),
		tb: tb,
	}
	for address, s := range servers {
		apiServer := server.NewAPIServer(
			cluster.sharder,
			route.NewRouter(
				cluster.addresser,
				grpcutil.NewDialer(
					grpc.WithInsecure(),
				),
				address,
			),
		)
		pfs.RegisterApiServer(s, apiServer)
		internalAPIServer := server.NewInternalAPIServer(
			cluster.sharder,
			route.NewRouter(
				cluster.addresser,
				grpcutil.NewDialer(
					grpc.WithInsecure(),
				),
				address,
			),
			getDriver(tb, address),
		)
		pfs.RegisterInternalApiServer(s, internalAPIServer)
		roler := role.NewRoler(cluster.addresser, cluster.sharder, internalAPIServer, address, testNumReplicas)
		go func() { require.Equal(tb, roler.Run(), discovery.ErrCancelled) }()
		cluster.addresses = append(cluster.addresses, address)
		cluster.rolers[address] = roler
		cluster.servers[address] = apiServer
		cluster.internalServers[address] = internalAPIServer
	}
	return &cluster
}

func registerFunc(tb testing.TB, discoveryClient discovery.Client, servers map[string]*grpc.Server) Cluster {
	cluster := newCluster(tb, discoveryClient, servers)
	cluster.WaitForAvailability()
	return cluster
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
