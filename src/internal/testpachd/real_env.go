package testpachd

import (
	"context"
	"net"
	"net/url"
	"path"

	units "github.com/docker/go-units"
	"github.com/jmoiron/sqlx"
	"github.com/pachyderm/pachyderm/v2/src/internal/clusterstate"
	"github.com/pachyderm/pachyderm/v2/src/internal/migrations"
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

// WithRealEnv constructs a MockEnv, then forwards all API calls to go to API
// server instances for supported operations. PPS requires a kubernetes
// environment in order to spin up pipelines, which is not yet supported by this
// package, but the other API servers work.
func WithRealEnv(db *sqlx.DB, cb func(*RealEnv) error, customConfig ...*serviceenv.PachdFullConfiguration) error {
	return WithMockEnv(func(mockEnv *MockEnv) (err error) {
		realEnv := &RealEnv{MockEnv: *mockEnv}
		config := serviceenv.NewConfiguration(NewDefaultConfig())
		if len(customConfig) > 0 {
			config = serviceenv.NewConfiguration(customConfig[0])
		}

		etcdClientURL, err := url.Parse(realEnv.EtcdClient.Endpoints()[0])
		if err != nil {
			return err
		}
		config.EtcdHost = etcdClientURL.Hostname()
		config.EtcdPort = etcdClientURL.Port()
		config.PeerPort = uint16(realEnv.MockPachd.Addr.(*net.TCPAddr).Port)
		servEnv := serviceenv.InitServiceEnv(config)

		if err := migrations.ApplyMigrations(context.Background(), db, migrations.Env{}, clusterstate.DesiredClusterState); err != nil {
			return err
		}
		if err := migrations.BlockUntil(context.Background(), db, clusterstate.DesiredClusterState); err != nil {
			return err
		}

		realEnv.LocalStorageDirectory = path.Join(realEnv.Directory, "localStorage")
		config.StorageRoot = realEnv.LocalStorageDirectory

		etcdPrefix := ""

		txnEnv := &txnenv.TransactionEnv{}

		realEnv.PFSServer, err = pfsserver.NewAPIServer(
			servEnv,
			txnEnv,
			etcdPrefix,
			db,
		)
		if err != nil {
			return err
		}

		realEnv.AuthServer = &authtesting.InactiveAPIServer{}

		realEnv.TransactionServer, err = txnserver.NewAPIServer(servEnv, txnEnv, etcdPrefix)
		if err != nil {
			return err
		}

		realEnv.MockPPSTransactionServer = NewMockPPSTransactionServer()

		txnEnv.Initialize(servEnv, realEnv.TransactionServer, realEnv.AuthServer, realEnv.PFSServer, &realEnv.MockPPSTransactionServer.api)

		linkServers(&realEnv.MockPachd.PFS, realEnv.PFSServer)
		linkServers(&realEnv.MockPachd.Auth, realEnv.AuthServer)
		linkServers(&realEnv.MockPachd.Transaction, realEnv.TransactionServer)

		return cb(realEnv)
	})
}

// NewDefaultConfig creates a new default pachd configuration.
func NewDefaultConfig() *serviceenv.PachdFullConfiguration {
	config := &serviceenv.PachdFullConfiguration{}
	config.StorageMemoryThreshold = units.GB
	config.StorageShardThreshold = units.GB
	config.StorageLevelFactor = 10
	config.StorageGCPolling = "30s"
	config.StorageCompactionMaxFanIn = 50
	config.StorageMemoryCacheSize = 10
	return config
}
