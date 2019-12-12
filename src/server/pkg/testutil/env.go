package testutil

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"path"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/embed"
	"github.com/coreos/pkg/capnslog"
	"golang.org/x/sync/errgroup"

	"github.com/pachyderm/pachyderm/src/client"
	authserver "github.com/pachyderm/pachyderm/src/server/auth/server"
	authtesting "github.com/pachyderm/pachyderm/src/server/auth/testing"
	pfsserver "github.com/pachyderm/pachyderm/src/server/pfs/server"
	"github.com/pachyderm/pachyderm/src/server/pkg/hashtree"
	"github.com/pachyderm/pachyderm/src/server/pkg/serviceenv"
	txnenv "github.com/pachyderm/pachyderm/src/server/pkg/transactionenv"
	txnserver "github.com/pachyderm/pachyderm/src/server/transaction/server"
)

// Env contains the basic setup for running end-to-end pachyderm tests entirely
// locally within the test process. It provides a temporary directory for
// storing data, and an embedded etcd server with a connected client.
type Env struct {
	Context    context.Context
	Directory  string
	Etcd       *embed.Etcd
	EtcdClient *etcd.Client
}

// WithEnv constructs a default Env for testing during the lifetime of the
// callback.
func WithEnv(cb func(*Env) error) (err error) {
	// Use an error group with a cancelable context to supervise every component
	// and cancel everything if one fails
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	eg, ctx := errgroup.WithContext(ctx)

	env := &Env{Context: ctx}

	// Cleanup any state when we return
	defer func() {
		// We return the first error that occurs during teardown, but still try to
		// close everything
		saveErr := func(e error) error {
			if e != nil && err == nil {
				err = e
			}
			return e
		}

		if env.EtcdClient != nil {
			saveErr(env.EtcdClient.Close())
		}

		if env.Etcd != nil {
			env.Etcd.Close()
		}

		saveErr(os.RemoveAll(env.Directory))
		cancel()
		saveErr(eg.Wait())
	}()

	dirBase := path.Join(os.TempDir(), "pachyderm_test")

	err = os.MkdirAll(dirBase, 0700)
	if err != nil {
		return err
	}

	env.Directory, err = ioutil.TempDir(dirBase, "")
	if err != nil {
		return err
	}

	etcdConfig := embed.NewConfig()
	etcdConfig.LogOutput = "default"

	// Create test dirs for etcd data
	etcdConfig.Dir = path.Join(env.Directory, "etcd_data")
	etcdConfig.WalDir = path.Join(env.Directory, "etcd_wal")

	// Speed up initial election, hopefully this has no other impact since there
	// is only one etcd instance
	etcdConfig.InitialElectionTickAdvance = false
	etcdConfig.TickMs = 10
	etcdConfig.ElectionMs = 50

	// We want to assign a random unused port to etcd, but etcd doesn't give us a
	// way to read it back out later. We can work around this by creating our own
	// listener on a random port, find out which port was used, close that
	// listener, and pass that port down for etcd to use. There is a small race
	// condition here where someone else can steal the port between these steps,
	// but it should be fairly minimal and depends on how the OS assigns
	// unallocated ports.
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return err
	}
	defer listener.Close()

	clientURL, err := url.Parse(fmt.Sprintf("http://%s", listener.Addr().String()))
	if err != nil {
		return err
	}

	etcdConfig.LPUrls = []url.URL{}
	etcdConfig.LCUrls = []url.URL{*clientURL}

	listener.Close()

	// Throw away noisy messages from etcd - comment these out if you need to debug
	// a failed start
	capnslog.SetGlobalLogLevel(capnslog.CRITICAL)

	env.Etcd, err = embed.StartEtcd(etcdConfig)
	if err != nil {
		return err
	}

	capnslog.SetGlobalLogLevel(capnslog.CRITICAL)

	eg.Go(func() error {
		return errorWait(ctx, env.Etcd.Err())
	})

	env.EtcdClient, err = etcd.New(etcd.Config{
		Context:     env.Context,
		Endpoints:   []string{clientURL.String()},
		DialOptions: client.DefaultDialOptions(),
	})
	if err != nil {
		return err
	}

	// TODO: supervise the EtcdClient connection and error the errgroup if they
	// go down

	return cb(env)
}

// MockEnv contains the basic setup for running end-to-end pachyderm tests
// entirely locally within the test process. It provides a temporary directory
// for storing data, an embedded etcd server with a connected client, as well as
// a local mock pachd instance which allows a test to hook into any pachd calls.
type MockEnv struct {
	Env
	MockPachd  *MockPachd
	PachClient *client.APIClient
}

// WithMockEnv sets up an Env structure, passes it to the provided callback,
// then cleans up everything in the environment, regardless of if an assertion
// fails.
func WithMockEnv(cb func(*MockEnv) error) error {
	return WithEnv(func(env *Env) (err error) {
		// Use an error group with a cancelable context to supervise every component
		// and cancel everything if one fails
		ctx, cancel := context.WithCancel(env.Context)
		defer cancel()
		eg, ctx := errgroup.WithContext(ctx)

		mockEnv := &MockEnv{Env: *env}
		mockEnv.Context = ctx

		// Cleanup any state when we return
		defer func() {
			saveErr := func(e error) error {
				if e != nil && err == nil {
					err = e
				}
				return e
			}

			if mockEnv.PachClient != nil {
				saveErr(mockEnv.PachClient.Close())
			}

			if mockEnv.MockPachd != nil {
				saveErr(mockEnv.MockPachd.Close())
			}

			cancel()
			saveErr(eg.Wait())
		}()

		mockEnv.MockPachd, err = NewMockPachd(mockEnv.Context)
		if err != nil {
			return err
		}

		eg.Go(func() error {
			return errorWait(ctx, mockEnv.MockPachd.Err())
		})

		mockEnv.PachClient, err = client.NewFromAddress(mockEnv.MockPachd.Addr.String())
		if err != nil {
			return err
		}

		// TODO: supervise the PachClient connection and error the errgroup if they
		// go down

		return cb(mockEnv)
	})
}

// RealEnv contains a setup for running end-to-end pachyderm tests locally.  It
// includes the base Env struct as well as a real instance of the API server.
// These calls can still be mocked, but they default to calling into the real
// server endpoints.
type RealEnv struct {
	MockEnv
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
			if realEnv.PFSServer != nil {
				realEnv.PFSServer.Close()
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
		treeCache, err := hashtree.NewCache(testingTreeCacheSize)
		if err != nil {
			return err
		}

		txnEnv := &txnenv.TransactionEnv{}

		realEnv.PFSServer, err = pfsserver.NewAPIServer(
			servEnv,
			txnEnv,
			etcdPrefix,
			treeCache,
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

func errorWait(ctx context.Context, errChan <-chan error) error {
	select {
	case <-ctx.Done():
		return nil
	case err := <-errChan:
		return err
	}
}
