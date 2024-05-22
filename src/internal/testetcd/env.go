package testetcd

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"path"
	"testing"
	"time"

	etcd "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/server/v3/embed"
	etcdwal "go.etcd.io/etcd/server/v3/wal"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/require"
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

// NewEnv constructs a default Env for testing, which will be destroyed at the
// end of the test.
func NewEnv(rctx context.Context, t testing.TB) *Env {
	// Use an error group with a cancelable context to supervise every component
	// and cancel everything if one fails
	ctx, cancel := pctx.WithCancel(rctx)
	eg, ctx := errgroup.WithContext(ctx)
	t.Cleanup(func() {
		require.NoError(t, eg.Wait())
	})
	t.Cleanup(cancel)

	env := &Env{Context: ctx, Directory: t.TempDir()}

	// NOTE: this is changing a GLOBAL variable in etcd. This function should not
	// be run in the same process as production code where this may affect
	// performance (but there should be little risk of that as this is only for
	// test code).
	etcdwal.SegmentSizeBytes = 1 * 1000 * 1000 // 1 MB

	etcdConfig := embed.NewConfig()
	etcdConfig.MaxTxnOps = 10000

	// Create test dirs for etcd data
	etcdConfig.Dir = path.Join(env.Directory, "etcd_data")
	etcdConfig.WalDir = path.Join(env.Directory, "etcd_wal")

	// Disable fsync, because it makes parallel tests perform poorly, and we don't need the data
	// to be durable.
	etcdConfig.UnsafeNoFsync = true

	// Speed up initial election, hopefully this has no other impact since there
	// is only one etcd instance
	etcdConfig.InitialElectionTickAdvance = false
	etcdConfig.TickMs = 10
	etcdConfig.ElectionMs = 50

	// Log to the test log.
	level := log.AddLoggerToEtcdServer(ctx, etcdConfig)
	// We want to assign a random unused port to etcd, but etcd doesn't give us a
	// way to read it back out later. We can work around this by creating our own
	// listener on a random port, find out which port was used, close that
	// listener, and pass that port down for etcd to use. There is a small race
	// condition here where someone else can steal the port between these steps,
	// but it should be fairly minimal and depends on how the OS assigns
	// unallocated ports.
	listener, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)
	require.NoError(t, listener.Close())

	clientURL, err := url.Parse(fmt.Sprintf("http://%s", listener.Addr().String()))
	require.NoError(t, err)

	etcdConfig.ListenPeerUrls = []url.URL{}
	etcdConfig.ListenClientUrls = []url.URL{*clientURL}

	// Throw away noisy messages from etcd - comment these out if you need to debug
	// a failed start
	level.SetLevel(zapcore.ErrorLevel)

	env.Etcd, err = embed.StartEtcd(etcdConfig)
	require.NoError(t, err)
	t.Cleanup(env.Etcd.Close)

	eg.Go(func() error {
		return errorWait(ctx, env.Etcd.Err())
	})

	// Wait for the server to become ready, then restore the log level.
	select {
	case <-env.Etcd.Server.ReadyNotify():
		// This used to be DebugLevel.  It's a lot of noise to sort through and I'm not sure
		// anyone has ever wanted to read them.  Feel free to change this back to DebugLevel
		// if it helps you, though.
		level.SetLevel(zapcore.InfoLevel)
	case <-time.After(30 * time.Second):
		t.Fatal("etcd did not start after 30 seconds")
	}

	cfg := log.GetEtcdClientConfig(env.Context)
	cfg.Endpoints = []string{clientURL.String()}
	cfg.DialOptions = client.DefaultDialOptions()
	env.EtcdClient, err = etcd.New(cfg)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, env.EtcdClient.Close())
	})

	// TODO: supervise the EtcdClient connection and error the errgroup if they
	// go down

	return env
}

func errorWait(ctx context.Context, errChan <-chan error) error {
	select {
	case <-ctx.Done():
		return nil
	case err := <-errChan:
		return err
	}
}
