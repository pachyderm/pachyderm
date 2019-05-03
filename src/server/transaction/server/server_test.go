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
	"github.com/pachyderm/pachyderm/src/client/transaction"
	authtesting "github.com/pachyderm/pachyderm/src/server/auth/testing"
	pfsserver "github.com/pachyderm/pachyderm/src/server/pfs/server"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	"github.com/pachyderm/pachyderm/src/server/pkg/hashtree"
	"github.com/pachyderm/pachyderm/src/server/pkg/serviceenv"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"
	txnenv "github.com/pachyderm/pachyderm/src/server/pkg/transactionenv"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
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

func runServers(
	t testing.TB,
	port int32,
	pfsServer pfs.APIServer,
	pfsBlockServer pfs.ObjectAPIServer,
	authServer auth.APIServer,
	txnServer APIServer,
) {
	ready := make(chan bool)
	go func() {
		err := grpcutil.Serve(
			grpcutil.ServerOptions{
				Port:       uint16(port),
				MaxMsgSize: grpcutil.MaxMsgSize,
				RegisterFunc: func(s *grpc.Server) error {
					defer close(ready)
					pfs.RegisterAPIServer(s, pfsServer)
					pfs.RegisterObjectAPIServer(s, pfsBlockServer)
					auth.RegisterAPIServer(s, authServer)
					transaction.RegisterAPIServer(s, txnServer)
					return nil
				}},
		)
		require.NoError(t, err)
	}()
	<-ready
}

// GetPachClient initializes a new PFS, Block, and Transaction servers and
// begins serving requests from them on a new port, then returns a client
// connected to the new servers.
func GetPachClient(t testing.TB) *client.APIClient {
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

	root := tu.UniqueString("/tmp/pach_test/run")
	t.Logf("root %s", root)
	testPort := atomic.AddInt32(&port, 1)

	config := serviceenv.NewConfiguration(&serviceenv.GlobalConfiguration{})
	config.EtcdHost = etcdHost
	config.EtcdPort = etcdPort
	config.PeerPort = uint16(testPort)
	env := serviceenv.InitServiceEnv(config)

	pfsBlockServer, err := pfsserver.NewBlockAPIServer(root, localBlockServerCacheBytes, pfsserver.LocalBackendEnvVar, net.JoinHostPort(etcdHost, etcdPort))
	require.NoError(t, err)

	etcdPrefix := generateRandomString(32)
	treeCache, err := hashtree.NewCache(testingTreeCacheSize)
	if err != nil {
		panic(fmt.Sprintf("could not initialize treeCache: %v", err))
	}

	txnEnv := &txnenv.TransactionEnv{}

	pfsServer, err := pfsserver.NewAPIServer(env, txnEnv, etcdPrefix, treeCache, "/tmp", 64*1024*1024)
	require.NoError(t, err)

	authServer := &authtesting.InactiveAPIServer{}

	txnServer, err := NewAPIServer(env, txnEnv, etcdPrefix)
	require.NoError(t, err)

	txnEnv.Initialize(env, txnServer, authServer, pfsServer)

	runServers(t, testPort, pfsServer, pfsBlockServer, authServer, txnServer)
	return env.GetPachClient(context.Background())
}

func TestEmptyTransaction(t *testing.T) {
	c := GetPachClient(t)

	txn, err := c.StartTransaction()
	require.NoError(t, err)

	info, err := c.InspectTransaction(txn)
	require.NoError(t, err)
	require.Equal(t, txn, info.Transaction)
	require.Equal(t, 0, len(info.Requests))
	require.Equal(t, 0, len(info.Responses))
	require.NotNil(t, info.Started)

	info, err = c.FinishTransaction(txn)
	require.NoError(t, err)
	require.Equal(t, txn, info.Transaction)
	require.Equal(t, 0, len(info.Requests))
	require.Equal(t, 0, len(info.Responses))
	require.NotNil(t, info.Started)

	info, err = c.InspectTransaction(txn)
	require.YesError(t, err)
	require.Nil(t, info)
}

func TestInvalidatedTransaction(t *testing.T) {
	c := GetPachClient(t)

	txn, err := c.StartTransaction()
	require.NoError(t, err)

	ct := c.WithTransaction(txn)
	createRepo := &pfs.CreateRepoRequest{
		Repo: client.NewRepo("foo"),
	}

	// Tell the transaction to create a repo
	_, err = ct.PfsAPIClient.CreateRepo(ct.Ctx(), createRepo)
	require.NoError(t, err)

	// Create the same repo outside of the transaction, so it can't run
	_, err = c.PfsAPIClient.CreateRepo(c.Ctx(), createRepo)
	require.NoError(t, err)

	// Finishing the transaction should fail
	info, err := c.FinishTransaction(txn)
	require.YesError(t, err)
	require.Nil(t, info)

	// Appending to the transaction should fail
	_, err = ct.PfsAPIClient.CreateRepo(ct.Ctx(), &pfs.CreateRepoRequest{Repo: client.NewRepo("bar")})
	require.YesError(t, err)
}

func TestFailedAppend(t *testing.T) {
	c := GetPachClient(t)

	txn, err := c.StartTransaction()
	require.NoError(t, err)

	ct := c.WithTransaction(txn)
	createRepo := &pfs.CreateRepoRequest{
		Repo: client.NewRepo("foo"),
	}

	// Create a repo outside of a transaction
	_, err = c.PfsAPIClient.CreateRepo(c.Ctx(), createRepo)
	require.NoError(t, err)

	// Tell the transaction to create the same repo, which should fail
	_, err = ct.PfsAPIClient.CreateRepo(ct.Ctx(), createRepo)
	require.YesError(t, err)

	info, err := c.InspectTransaction(txn)
	require.NoError(t, err)
	require.Equal(t, txn, info.Transaction)
	require.Equal(t, 0, len(info.Requests))
	require.Equal(t, 0, len(info.Responses))

	info, err = c.FinishTransaction(txn)
	require.NoError(t, err)
	require.Equal(t, txn, info.Transaction)
	require.Equal(t, 0, len(info.Requests))
	require.Equal(t, 0, len(info.Responses))
}

func requireEmptyResponse(t *testing.T, response *transaction.TransactionResponse) {
}

func requireCommitResponse(t *testing.T, response *transaction.TransactionResponse, commit *pfs.Commit) {
}

func TestDependency(t *testing.T) {
	c := GetPachClient(t)

	txn, err := c.StartTransaction()
	require.NoError(t, err)

	ct := c.WithTransaction(txn)

	// Create repo, branch, start commit, finish commit
	_, err = ct.PfsAPIClient.CreateRepo(ct.Ctx(), &pfs.CreateRepoRequest{
		Repo: client.NewRepo("foo"),
	})
	require.NoError(t, err)

	_, err = ct.PfsAPIClient.CreateBranch(ct.Ctx(), &pfs.CreateBranchRequest{
		Branch: client.NewBranch("foo", "master")},
	)
	require.NoError(t, err)

	commit, err := ct.PfsAPIClient.StartCommit(ct.Ctx(), &pfs.StartCommitRequest{
		Branch: "master",
		Parent: client.NewCommit("foo", ""),
	})
	require.NoError(t, err)

	_, err = ct.PfsAPIClient.FinishCommit(ct.Ctx(), &pfs.FinishCommitRequest{
		Commit: client.NewCommit("foo", "master"),
	})
	require.NoError(t, err)

	info, err := c.InspectTransaction(txn)
	require.NoError(t, err)
	require.Equal(t, txn, info.Transaction)
	require.Equal(t, 4, len(info.Requests))
	require.Equal(t, 4, len(info.Responses))

	// Check each response value
	requireEmptyResponse(t, info.Responses[0])
	requireEmptyResponse(t, info.Responses[1])
	requireCommitResponse(t, info.Responses[2], commit)
	requireEmptyResponse(t, info.Responses[3])

	info, err = c.FinishTransaction(txn)
	require.NoError(t, err)
	require.Equal(t, txn, info.Transaction)
	require.Equal(t, 4, len(info.Requests))
	require.Equal(t, 4, len(info.Responses))

	// Double-check each response value
	requireEmptyResponse(t, info.Responses[0])
	requireEmptyResponse(t, info.Responses[1])
	requireCommitResponse(t, info.Responses[2], commit)
	requireEmptyResponse(t, info.Responses[3])

	info, err = c.InspectTransaction(txn)
	require.YesError(t, err)
	require.Nil(t, info)
}

func TestDeleteAllTransactions(t *testing.T) {
	c := GetPachClient(t)

	_, err := c.StartTransaction()
	require.NoError(t, err)

	_, err = c.StartTransaction()
	require.NoError(t, err)

	txns, err := c.ListTransaction()
	require.NoError(t, err)
	require.Equal(t, 2, len(txns))

	_, err = c.TransactionAPIClient.DeleteAll(c.Ctx(), &transaction.DeleteAllRequest{})
	require.NoError(t, err)

	txns, err = c.ListTransaction()
	require.NoError(t, err)
	require.Equal(t, 0, len(txns))
}
