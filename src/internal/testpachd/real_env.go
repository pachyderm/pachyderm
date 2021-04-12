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
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
	txnenv "github.com/pachyderm/pachyderm/v2/src/internal/transactionenv"
	authserver "github.com/pachyderm/pachyderm/v2/src/server/auth/server"
	authtesting "github.com/pachyderm/pachyderm/v2/src/server/auth/testing"
	pfsserver "github.com/pachyderm/pachyderm/v2/src/server/pfs/server"
	txnserver "github.com/pachyderm/pachyderm/v2/src/server/transaction/server"
)

// RealEnv contains a setup for running end-to-end pachyderm tests locally.  It
// includes the base MockEnv struct as well as a real instance of the API server.
// These calls can still be mocked, but they default to calling into the real
// server endpoints.
type RealEnv struct {
	MockEnv

	LocalStorageDirectory    string
	AuthServer               authserver.APIServer
	PFSServer                pfsserver.APIServer
	TransactionServer        txnserver.APIServer
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
		DefaultConfigOptions,
		serviceenv.WithEtcdHostPort(etcdClientURL.Hostname(), etcdClientURL.Port()),
		serviceenv.WithPachdPeerPort(uint16(realEnv.MockPachd.Addr.(*net.TCPAddr).Port)),
	}
	opts = append(opts, customOpts...) // Overwrite with any custom options
	config := serviceenv.ConfigFromOptions(opts...)
	require.NoError(t, cmdutil.Populate(config)) // Overwrite with any environment variables

	servEnv := serviceenv.InitServiceEnv(config)

	// Overwrite the mock pach client with the ServiceEnv's client so it gets closed earlier
	realEnv.PachClient = servEnv.GetPachClient(servEnv.Context())

	t.Cleanup(func() {
		require.NoError(t, servEnv.Close())
	})

	err = migrations.ApplyMigrations(servEnv.Context(), servEnv.GetDBClient(), migrations.Env{}, clusterstate.DesiredClusterState)
	require.NoError(t, err)
	err = migrations.BlockUntil(servEnv.Context(), servEnv.GetDBClient(), clusterstate.DesiredClusterState)
	require.NoError(t, err)

	realEnv.LocalStorageDirectory = path.Join(realEnv.Directory, "localStorage")
	config.StorageRoot = realEnv.LocalStorageDirectory

	txnEnv := &txnenv.TransactionEnv{}

	etcdPrefix := ""
	realEnv.PFSServer, err = pfsserver.NewAPIServer(
		servEnv,
		txnEnv,
		etcdPrefix,
	)
	require.NoError(t, err)

	realEnv.AuthServer = &authtesting.InactiveAPIServer{}

	realEnv.TransactionServer, err = txnserver.NewAPIServer(servEnv, txnEnv, etcdPrefix)
	require.NoError(t, err)

	realEnv.MockPPSTransactionServer = NewMockPPSTransactionServer()

	txnEnv.Initialize(servEnv, realEnv.TransactionServer, realEnv.AuthServer, realEnv.PFSServer, &realEnv.MockPPSTransactionServer.api)

	linkServers(&realEnv.MockPachd.PFS, realEnv.PFSServer)
	linkServers(&realEnv.MockPachd.Auth, realEnv.AuthServer)
	linkServers(&realEnv.MockPachd.Transaction, realEnv.TransactionServer)

	return realEnv
}

// DefaultConfigOptions is a serviceenv config option with the defaults used for tests
func DefaultConfigOptions(config *serviceenv.Configuration) {
	config.StorageMemoryThreshold = units.GB
	config.StorageShardThreshold = units.GB
	config.StorageLevelFactor = 10
	config.StorageGCPolling = "30s"
	config.StorageCompactionMaxFanIn = 50
}
