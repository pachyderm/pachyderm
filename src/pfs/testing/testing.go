package testing

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/satori/go.uuid"

	"go.pedge.io/proto/test"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/drive"
	"github.com/pachyderm/pachyderm/src/pfs/drive/btrfs"
	"github.com/pachyderm/pachyderm/src/pfs/route"
	"github.com/pachyderm/pachyderm/src/pfs/server"
	"github.com/pachyderm/pachyderm/src/pkg/discovery"
	"github.com/pachyderm/pachyderm/src/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/pkg/require"
	"github.com/pachyderm/pachyderm/src/pkg/shard"
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

func RunTest(
	t *testing.T,
	f func(*testing.T, pfs.APIClient, pfs.InternalAPIClient, Cluster),
) {
	discoveryClient, err := getEtcdClient()
	require.NoError(t, err)
	var cluster *cluster
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
			go func() {
				require.Equal(t, cluster.realSharder.AssignRoles(cluster.cancel), shard.ErrCancelled)
			}()
			cluster.WaitForAvailability()
			f(
				t,
				pfs.NewAPIClient(
					clientConn,
				),
				pfs.NewInternalAPIClient(
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
	f func(*testing.B, pfs.APIClient),
) {
	discoveryClient, err := getEtcdClient()
	require.NoError(b, err)
	var cluster *cluster
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
			go func() {
				require.Equal(b, cluster.realSharder.AssignRoles(cluster.cancel), shard.ErrCancelled)
			}()
			cluster.WaitForAvailability()
			f(
				b,
				pfs.NewAPIClient(
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
	KillRoleAssigner()
	RestartRoleAssigner()
	Shutdown()
}

type cluster struct {
	addresses       []string
	servers         map[string]server.APIServer
	internalServers map[string]server.InternalAPIServer
	cancels         map[string]chan bool
	internalCancels map[string]chan bool
	cancel          chan bool
	realSharder     shard.TestSharder
	sharder         route.Sharder
	tb              testing.TB
}

func (c *cluster) WaitForAvailability() {
	// We use address as the id for servers
	var frontendIds []string
	var serverIds []string
	for _, address := range c.addresses {
		if _, ok := c.cancels[address]; ok {
			frontendIds = append(frontendIds, address)
		}
		if _, ok := c.internalCancels[address]; ok {
			serverIds = append(serverIds, address)
		}
	}
	require.NoError(c.tb, c.realSharder.WaitForAvailability(frontendIds, serverIds))
}

func (c *cluster) Kill(index int) {
	close(c.cancels[c.addresses[index]])
	delete(c.cancels, c.addresses[index])
	delete(c.internalServers, c.addresses[index])
}

func (c *cluster) Restart(index int) {
	address := c.addresses[index]
	c.cancels[address] = make(chan bool)
	internalAPIServer := server.NewInternalAPIServer(
		c.sharder,
		route.NewRouter(
			c.realSharder,
			grpcutil.NewDialer(
				grpc.WithInsecure(),
			),
			address,
		),
		getDriver(c.tb, address),
	)
	c.internalServers[address] = internalAPIServer
	go func() {
		require.Equal(c.tb, c.realSharder.Register(c.cancels[address], address, address, c.internalServers[address]), shard.ErrCancelled)
	}()
}

func (c *cluster) KillRoleAssigner() {
	close(c.cancel)
}

func (c *cluster) RestartRoleAssigner() {
	c.cancel = make(chan bool)
	go func() {
		require.Equal(c.tb, c.realSharder.AssignRoles(c.cancel), shard.ErrCancelled)
	}()
}

func (c *cluster) Shutdown() {
	close(c.cancel)
	for _, cancel := range c.cancels {
		close(cancel)
	}
}

func newCluster(tb testing.TB, discoveryClient discovery.Client, servers map[string]*grpc.Server) *cluster {
	realSharder := shard.NewTestSharder(
		discoveryClient,
		testShardsPerServer*testNumServers,
		testNumReplicas,
		testNamespace(),
	)
	sharder := route.NewSharder(
		testShardsPerServer*testNumServers,
		testNumReplicas,
	)
	cluster := cluster{
		servers:         make(map[string]server.APIServer),
		internalServers: make(map[string]server.InternalAPIServer),
		cancels:         make(map[string]chan bool),
		internalCancels: make(map[string]chan bool),
		cancel:          make(chan bool),
		realSharder:     realSharder,
		sharder:         sharder,
		tb:              tb,
	}
	for address, s := range servers {
		cluster.addresses = append(cluster.addresses, address)
		router := route.NewRouter(
			cluster.realSharder,
			grpcutil.NewDialer(
				grpc.WithInsecure(),
			),
			address,
		)
		apiServer := server.NewAPIServer(
			cluster.sharder,
			router,
		)
		cluster.servers[address] = apiServer
		cluster.cancels[address] = make(chan bool)
		go func(address string) {
			require.Equal(tb, cluster.realSharder.RegisterFrontend(cluster.cancels[address], address, cluster.servers[address]), shard.ErrCancelled)
		}(address)
		pfs.RegisterAPIServer(s, apiServer)
		internalAPIServer := server.NewInternalAPIServer(
			cluster.sharder,
			router,
			getDriver(tb, address),
		)
		pfs.RegisterInternalAPIServer(s, internalAPIServer)
		cluster.internalServers[address] = internalAPIServer
		cluster.internalCancels[address] = make(chan bool)
		go func(address string) {
			require.Equal(tb, cluster.realSharder.Register(cluster.internalCancels[address], address, address, cluster.internalServers[address]), shard.ErrCancelled)
		}(address)
	}
	return &cluster
}

func registerFunc(tb testing.TB, discoveryClient discovery.Client, servers map[string]*grpc.Server) *cluster {
	return newCluster(tb, discoveryClient, servers)
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
	return fmt.Sprintf("test-%s", uuid.NewV4().String())
}
