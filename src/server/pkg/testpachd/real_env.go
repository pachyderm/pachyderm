package testpachd

import (
	"net"
	"net/url"
	"path"

	authserver "github.com/pachyderm/pachyderm/src/server/auth/server"
	authtesting "github.com/pachyderm/pachyderm/src/server/auth/testing"
	pfsserver "github.com/pachyderm/pachyderm/src/server/pfs/server"
	"github.com/pachyderm/pachyderm/src/server/pkg/hashtree"
	"github.com/pachyderm/pachyderm/src/server/pkg/serviceenv"
	txnenv "github.com/pachyderm/pachyderm/src/server/pkg/transactionenv"
	txnserver "github.com/pachyderm/pachyderm/src/server/transaction/server"
)

// RealEnv contains a setup for running end-to-end pachyderm tests locally.  It
// includes the base MockEnv struct as well as a real instance of the API server.
// These calls can still be mocked, but they default to calling into the real
// server endpoints.
type RealEnv struct {
	MockEnv

	treeCache *hashtree.Cache

	AuthServer        authserver.APIServer
	PFSBlockServer    pfsserver.BlockAPIServer
	PFSServer         pfsserver.APIServer
	TransactionServer txnserver.APIServer
}

const (
	testingTreeCacheSize       = 8
	localBlockServerCacheBytes = 256 * 1024 * 1024
)

// WithRealEnv constructs a MockEnv, then forwards all API calls to go to API
// server instances for supported operations. PPS requires a kubernetes
// environment in order to spin up pipelines, which is not yet supported by this
// package, but the other API servers work.
func WithRealEnv(cb func(*RealEnv) error) error {
	return WithMockEnv(func(mockEnv *MockEnv) (err error) {
		realEnv := &RealEnv{MockEnv: *mockEnv}

		defer func() {
			if realEnv.treeCache != nil {
				realEnv.treeCache.Close()
			}
		}()

		config := serviceenv.NewConfiguration(&serviceenv.PachdFullConfiguration{})
		etcdClientURL, err := url.Parse(realEnv.EtcdClient.Endpoints()[0])
		if err != nil {
			return err
		}

		config.EtcdHost = etcdClientURL.Hostname()
		config.EtcdPort = etcdClientURL.Port()
		config.PeerPort = uint16(realEnv.MockPachd.Addr.(*net.TCPAddr).Port)
		servEnv := serviceenv.InitServiceEnv(config)

		realEnv.PFSBlockServer, err = pfsserver.NewBlockAPIServer(
			path.Join(realEnv.Directory, "objects"),
			localBlockServerCacheBytes,
			pfsserver.LocalBackendEnvVar,
			net.JoinHostPort(config.EtcdHost, config.EtcdPort),
			true, // duplicate
		)
		if err != nil {
			return err
		}

		etcdPrefix := ""
		realEnv.treeCache, err = hashtree.NewCache(testingTreeCacheSize)
		if err != nil {
			return err
		}

		txnEnv := &txnenv.TransactionEnv{}

		realEnv.PFSServer, err = pfsserver.NewAPIServer(
			servEnv,
			txnEnv,
			etcdPrefix,
			realEnv.treeCache,
			path.Join(realEnv.Directory, "pfs"),
			64*1024*1024,
		)
		if err != nil {
			return err
		}

		realEnv.AuthServer = &authtesting.InactiveAPIServer{}

		realEnv.TransactionServer, err = txnserver.NewAPIServer(servEnv, txnEnv, etcdPrefix)
		if err != nil {
			return err
		}

		txnEnv.Initialize(servEnv, realEnv.TransactionServer, realEnv.AuthServer, realEnv.PFSServer, txnenv.NewMockPpsTransactionServer())

		linkServers(&realEnv.MockPachd.Object, realEnv.PFSBlockServer)
		linkServers(&realEnv.MockPachd.PFS, realEnv.PFSServer)
		linkServers(&realEnv.MockPachd.Auth, realEnv.AuthServer)
		linkServers(&realEnv.MockPachd.Transaction, realEnv.TransactionServer)

		return cb(realEnv)
	})
}
