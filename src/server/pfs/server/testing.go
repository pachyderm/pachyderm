package server

import (
	"fmt"
	"math/rand"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/version"
	"github.com/pachyderm/pachyderm/src/client/version/versionpb"
	authtesting "github.com/pachyderm/pachyderm/src/server/auth/testing"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	"github.com/pachyderm/pachyderm/src/server/pkg/hashtree"
	"github.com/pachyderm/pachyderm/src/server/pkg/serviceenv"
	txnenv "github.com/pachyderm/pachyderm/src/server/pkg/transactionenv"
	"github.com/pachyderm/pachyderm/src/server/pkg/uuid"

	"golang.org/x/net/context"
)

const (
	testingTreeCacheSize       = 8
	etcdHost                   = "localhost"
	etcdPort                   = "32379"
	localBlockServerCacheBytes = 256 * 1024 * 1024
)

var (
	port          int32     = 30655 // Initial port on which pachd server processes will serve
	checkEtcdOnce sync.Once         // ensure we only test the etcd connection once
)

// generateRandomString is a helper function for getPachClient
func generateRandomString(n int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, n)
	for i := range b {
		b[i] = byte('a' + rand.Intn(26))
	}
	return string(b)
}

// runServers starts serving requests for the given apiServer & blockAPIServer
// in a separate goroutine. Helper for getPachClient()
func runServers(
	t testing.TB,
	port int32,
	apiServer APIServer,
	blockAPIServer BlockAPIServer,
) {
	server, err := grpcutil.NewServer(context.Background(), false)
	require.NoError(t, err)

	pfs.RegisterAPIServer(server.Server, apiServer)
	pfs.RegisterObjectAPIServer(server.Server, blockAPIServer)
	auth.RegisterAPIServer(server.Server, &authtesting.InactiveAPIServer{}) // PFS server uses auth API
	versionpb.RegisterAPIServer(server.Server,
		version.NewAPIServer(version.Version, version.APIServerOptions{}))

	_, err = server.ListenTCP("", uint16(port))
	require.NoError(t, err)

	go func() {
		require.NoError(t, server.Wait())
	}()
}

// GetBasicConfig gets a basic service environment configuration for testing pachd.
func GetBasicConfig() *serviceenv.Configuration {
	config := serviceenv.NewConfiguration(&serviceenv.PachdFullConfiguration{})
	config.EtcdHost = etcdHost
	config.EtcdPort = etcdPort
	return config
}

// GetPachClient initializes a new PFSAPIServer and blockAPIServer and begins
// serving requests for them on a new port, and then returns a client connected
// to the new servers (allows PFS tests to run in parallel without conflict)
func GetPachClient(t testing.TB, config *serviceenv.Configuration) *client.APIClient {
	// src/server/pfs/server/driver.go expects an etcd server at "localhost:32379"
	// Try to establish a connection before proceeding with the test (which will
	// fail if the connection can't be established)
	checkEtcdOnce.Do(func() {
		require.NoError(t, backoff.Retry(func() error {
			_, err := etcd.New(etcd.Config{
				Endpoints:   []string{net.JoinHostPort(etcdHost, etcdPort)},
				DialOptions: client.DefaultDialOptions(),
			})
			if err != nil {
				return fmt.Errorf("could not connect to etcd: %s", err.Error())
			}
			return nil
		}, backoff.NewTestingBackOff()))
	})

	root := "/tmp/pach_test/run" + uuid.NewWithoutDashes()[0:12]
	t.Logf("root %s", root)

	pfsPort := atomic.AddInt32(&port, 1)
	config.PeerPort = uint16(pfsPort)

	// initialize new BlockAPIServier
	env := serviceenv.InitServiceEnv(config)
	blockAPIServer, err := newLocalBlockAPIServer(
		root,
		localBlockServerCacheBytes,
		net.JoinHostPort(etcdHost, etcdPort),
		true /* duplicate--see comment in newObjBlockAPIServer */)
	require.NoError(t, err)
	etcdPrefix := generateRandomString(32)
	treeCache, err := hashtree.NewCache(testingTreeCacheSize)
	if err != nil {
		panic(fmt.Sprintf("could not initialize treeCache: %v", err))
	}

	txnEnv := &txnenv.TransactionEnv{}

	apiServer, err := newAPIServer(env, txnEnv, etcdPrefix, treeCache, "/tmp", 64*1024*1024)
	require.NoError(t, err)

	txnEnv.Initialize(env, nil, &authtesting.InactiveAPIServer{}, apiServer)

	runServers(t, pfsPort, apiServer, blockAPIServer)
	return env.GetPachClient(context.Background())
}
