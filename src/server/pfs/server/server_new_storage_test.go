package server

import (
	"bytes"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	authtesting "github.com/pachyderm/pachyderm/src/server/auth/testing"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	"github.com/pachyderm/pachyderm/src/server/pkg/hashtree"
	"github.com/pachyderm/pachyderm/src/server/pkg/serviceenv"
	tu "github.com/pachyderm/pachyderm/src/server/pkg/testutil"
	txnenv "github.com/pachyderm/pachyderm/src/server/pkg/transactionenv"
	"golang.org/x/net/context"
)

func TestMerge(t *testing.T) {
	config := GetBasicConfig()
	config.NewStorageLayer = true
	config.StorageMemoryThreshold = 20
	config.StorageShardThreshold = 20
	c := GetPachClientNew(t, config)
	repo := "test"
	branch := "master"
	require.NoError(t, c.CreateRepo(repo))
	commit, err := c.StartCommit(repo, branch)
	require.NoError(t, err)
	pfc, err := c.NewPutFileClient()
	require.NoError(t, err)
	for i := 0; i < 100; i++ {
		s := strconv.Itoa(i)
		_, err := pfc.PutFile(repo, commit.ID, "/file"+s, strings.NewReader(s))
		require.NoError(t, err)
	}
	require.NoError(t, pfc.Close())
	require.NoError(t, c.FinishCommit(repo, commit.ID))
	buf := &bytes.Buffer{}
	require.NoError(t, c.GetFile(repo, commit.ID, "/file0", 0, 0, buf))
	require.Equal(t, "0", string(buf.Bytes()))
	buf.Reset()
	require.NoError(t, c.GetFile(repo, commit.ID, "/file50", 0, 0, buf))
	require.Equal(t, "50", string(buf.Bytes()))
	buf.Reset()
	require.NoError(t, c.GetFile(repo, commit.ID, "/file99", 0, 0, buf))
	require.Equal(t, "99", string(buf.Bytes()))
}

func GetBasicConfig() *serviceenv.Configuration {
	config := serviceenv.NewConfiguration(&serviceenv.PachdFullConfiguration{})
	config.EtcdHost = etcdHost
	config.EtcdPort = etcdPort
	return config
}

// GetPachClientNew initializes a new PFSAPIServer and blockAPIServer and begins
// serving requests for them on a new port, and then returns a client connected
// to the new servers (allows PFS tests to run in parallel without conflict)
func GetPachClientNew(t testing.TB, config *serviceenv.Configuration) *client.APIClient {
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
