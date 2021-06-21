package testpachd

import (
	"net"
	"net/url"
	"path"
	"testing"

	units "github.com/docker/go-units"

	"github.com/pachyderm/pachyderm/v2/src/internal/clusterstate"
	"github.com/pachyderm/pachyderm/v2/src/internal/cmdutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
	"github.com/pachyderm/pachyderm/v2/src/internal/obj"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	authapi "github.com/pachyderm/pachyderm/v2/src/server/auth"
	authtesting "github.com/pachyderm/pachyderm/v2/src/server/auth/testing"
	pfsapi "github.com/pachyderm/pachyderm/v2/src/server/pfs"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs/server"
	proxyserver "github.com/pachyderm/pachyderm/v2/src/server/proxy/server"
	txnserver "github.com/pachyderm/pachyderm/v2/src/server/transaction/server"
)

// RealEnv contains a setup for running end-to-end pachyderm tests locally.  It
// includes the base MockEnv struct as well as a real instance of the API server.
// These calls can still be mocked, but they default to calling into the real
// server endpoints.
type RealEnv struct {
	MockEnv

	ServiceEnv               serviceenv.ServiceEnv
	AuthServer               authapi.APIServer
	PFSServer                pfsapi.APIServer
	TransactionServer        txnserver.APIServer
	ProxyServer              *proxyserver.APIServer
	MockPPSTransactionServer *MockPPSTransactionServer
}

// NewRealEnv constructs a MockEnv, then forwards all API calls to go to API
// server instances for supported operations. PPS requires a kubernetes
// environment in order to spin up pipelines, which is not yet supported by this
// package, but the other API servers work.
func NewRealEnv(t testing.TB, customOpts ...serviceenv.ConfigOption) *RealEnv {
	mockEnv := NewMockEnv(t)

	realEnv := &RealEnv{MockEnv: *mockEnv}
	etcdClientURL, err := url.Parse(realEnv.EtcdClient.Endpoints()[0])
	require.NoError(t, err)

	opts := []serviceenv.ConfigOption{
		func(config *serviceenv.Configuration) {
			require.NoError(t, cmdutil.PopulateDefaults(config))
			config.StorageBackend = obj.Local
			config.StorageRoot = path.Join(realEnv.Directory, "localStorage")
		},
		DefaultConfigOptions,
		serviceenv.WithEtcdHostPort(etcdClientURL.Hostname(), etcdClientURL.Port()),
		serviceenv.WithPachdPeerPort(uint16(realEnv.MockPachd.Addr.(*net.TCPAddr).Port)),
	}
	opts = append(opts, customOpts...) // Overwrite with any custom options
	realEnv.ServiceEnv = serviceenv.InitServiceEnv(serviceenv.ConfigFromOptions(opts...))

	// Overwrite the mock pach client with the ServiceEnv's client so it gets closed earlier
	realEnv.PachClient = realEnv.ServiceEnv.GetPachClient(realEnv.ServiceEnv.Context())

	t.Cleanup(func() {
		// There is a race condition here, although not too serious because this only
		// happens in tests and the errors should not propagate back to the clients -
		// ideally we would close the client connection first and wait for the server
		// RPCs to end before closing the underlying service connections (like
		// postgres and etcd), so we don't get spurious errors. Instead, some RPCs may
		// fail because of losing the database connection.
		// TODO: It appears the postgres db.Close() may return errors due to
		// background goroutines using a closed TCP session because we don't do an
		// orderly shutdown, so we don't check the error here.
		realEnv.ServiceEnv.Close()
	})

	err = migrations.ApplyMigrations(realEnv.ServiceEnv.Context(), realEnv.ServiceEnv.GetDBClient(), migrations.Env{}, clusterstate.DesiredClusterState)
	require.NoError(t, err)
	err = migrations.BlockUntil(realEnv.ServiceEnv.Context(), realEnv.ServiceEnv.GetDBClient(), clusterstate.DesiredClusterState)
	require.NoError(t, err)

	txnEnv := &txnenv.TransactionEnv{}

	etcdPrefix := ""
	realEnv.PFSServer, err = pfsserver.NewAPIServer(
		realEnv.ServiceEnv,
		txnEnv,
		etcdPrefix,
	)
	require.NoError(t, err)

	realEnv.AuthServer = &authtesting.InactiveAPIServer{}

	realEnv.TransactionServer, err = txnserver.NewAPIServer(realEnv.ServiceEnv, txnEnv)
	require.NoError(t, err)

	realEnv.ProxyServer = proxyserver.NewAPIServer(realEnv.ServiceEnv)

	realEnv.MockPPSTransactionServer = NewMockPPSTransactionServer()

	realEnv.ServiceEnv.(*serviceenv.NonblockingServiceEnv).SetAuthServer(realEnv.AuthServer)
	realEnv.ServiceEnv.(*serviceenv.NonblockingServiceEnv).SetPfsServer(realEnv.PFSServer)
	realEnv.ServiceEnv.(*serviceenv.NonblockingServiceEnv).SetPpsServer(&realEnv.MockPPSTransactionServer.api)

	txnEnv.Initialize(realEnv.ServiceEnv, realEnv.TransactionServer)

	linkServers(&realEnv.MockPachd.PFS, realEnv.PFSServer)
	linkServers(&realEnv.MockPachd.Auth, realEnv.AuthServer)
	linkServers(&realEnv.MockPachd.Transaction, realEnv.TransactionServer)
	linkServers(&realEnv.MockPachd.Proxy, realEnv.ProxyServer)

	return realEnv
}

// DefaultConfigOptions is a serviceenv config option with the defaults used for tests
func DefaultConfigOptions(config *serviceenv.Configuration) {
	config.StorageMemoryThreshold = units.GB
	config.StorageShardThreshold = units.GB
	config.StorageLevelFactor = 10
	config.StorageGCPolling = "30s"
	config.StorageCompactionMaxFanIn = 10
	config.StorageMemoryCacheSize = 20
}
