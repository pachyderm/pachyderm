package server

import (
	"fmt"
	"math/rand"
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
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"
	"google.golang.org/grpc"
)

const (
	testingTreeCacheSize       = 8
	etcdAddress                = "localhost:32379" // etcd must already be serving at this address
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
func runServers(t testing.TB, port int32, apiServer pfs.APIServer,
	blockAPIServer BlockAPIServer) {
	ready := make(chan bool)
	go func() {
		err := grpcutil.Serve(
			grpcutil.ServerOptions{
				Port:       uint16(port),
				MaxMsgSize: grpcutil.MaxMsgSize,
				RegisterFunc: func(s *grpc.Server) error {
					defer close(ready)
					pfs.RegisterAPIServer(s, apiServer)
					pfs.RegisterObjectAPIServer(s, blockAPIServer)
					auth.RegisterAPIServer(s, &authtesting.InactiveAPIServer{}) // PFS server uses auth API
					versionpb.RegisterAPIServer(s,
						version.NewAPIServer(version.Version, version.APIServerOptions{}))
					return nil
				}},
		)
		require.NoError(t, err)
	}()
	<-ready
}

// GetPachClient initializes a new PFSAPIServer and blockAPIServer and begins
// serving requests for them on a new port, and then returns a client connected
// to the new servers (allows PFS tests to run in parallel without conflict)
func GetPachClient(t testing.TB) *client.APIClient {
	// src/server/pfs/server/driver.go expects an etcd server at "localhost:32379"
	// Try to establish a connection before proceeding with the test (which will
	// fail if the connection can't be established)
	checkEtcdOnce.Do(func() {
		require.NoError(t, backoff.Retry(func() error {
			_, err := etcd.New(etcd.Config{
				Endpoints:   []string{etcdAddress},
				DialOptions: client.DefaultDialOptions(),
			})
			if err != nil {
				return fmt.Errorf("could not connect to etcd: %s", err.Error())
			}
			return nil
		}, backoff.NewTestingBackOff()))
	})

	root := tu.UniqueString("/tmp/pach_test/run")
	t.Logf("root %s", root)
	servePort := atomic.AddInt32(&port, 1)
	serveAddress := fmt.Sprintf("localhost:%d", port)

	// initialize new BlockAPIServier
	blockAPIServer, err := newLocalBlockAPIServer(root, localBlockServerCacheBytes, etcdAddress)
	require.NoError(t, err)
	etcdPrefix := generateRandomString(32)
	treeCache, err := hashtree.NewCache(testingTreeCacheSize)
	if err != nil {
		panic(fmt.Sprintf("could not initialize treeCache: %v", err))
	}
	apiServer, err := newAPIServer(serveAddress, []string{"localhost:32379"}, etcdPrefix, treeCache, "/tmp", 64*1024*1024)
	require.NoError(t, err)
	runServers(t, servePort, apiServer, blockAPIServer)
	c, err := client.NewFromAddress(serveAddress)
	require.NoError(t, err)
	return c
}
