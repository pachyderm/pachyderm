package testing

import (
	"errors"
	"fmt"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/drive"
	"github.com/pachyderm/pachyderm/src/pfs/drive/btrfs"
	"github.com/pachyderm/pachyderm/src/pfs/role"
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
	grpctest.Run(
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
	grpctest.RunB(
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
	addresses []string
	rolers    []role.Roler
	servers   []server.CombinedAPIServer
	addresser route.Addresser
	sharder   route.Sharder
	tb        testing.TB
}

func (c *cluster) WaitForAvailability() {
	cancel := make(chan bool)
	time.AfterFunc(15*time.Second, func() { close(cancel) })
	err := c.addresser.WatchShardToAddress(cancel, func(shardToMasterAddress map[int]string, shardToReplicaAddress map[int]map[int]string) (uint64, error) {
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
		return 0, fmt.Errorf("Complete")
	})
	require.Equal(c.tb, err, fmt.Errorf("Complete"))
}

func (c *cluster) Kill(server int) {
	c.rolers[server].Cancel()
}

func (c *cluster) Restart(server int) {
	c.rolers[server] = role.NewRoler(c.addresser, c.sharder, c.servers[server], c.addresses[server], testNumReplicas)
	go func() { require.NoError(c.tb, c.rolers[server].Run()) }()
}

func (c *cluster) Shutdown() {
	for _, roler := range c.rolers {
		roler.Cancel()
	}
}

func newCluster(tb testing.TB, discoveryClient discovery.Client, servers map[string]*grpc.Server) Cluster {
	cluster := cluster{
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
		combinedAPIServer := server.NewCombinedAPIServer(
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
		pfs.RegisterApiServer(s, combinedAPIServer)
		pfs.RegisterInternalApiServer(s, combinedAPIServer)
		roler := role.NewRoler(cluster.addresser, cluster.sharder, combinedAPIServer, address, testNumReplicas)
		go func() { require.NoError(tb, roler.Run()) }()
		cluster.addresses = append(cluster.addresses, address)
		cluster.rolers = append(cluster.rolers, roler)
		cluster.servers = append(cluster.servers, combinedAPIServer)
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
