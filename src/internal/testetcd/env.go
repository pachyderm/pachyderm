package testetcd

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"path"
	"testing"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/embed"
	etcdwal "github.com/coreos/etcd/wal"
	"github.com/coreos/pkg/capnslog"
	"golang.org/x/sync/errgroup"

	"github.com/pachyderm/pachyderm/v2/src/client"
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
func NewEnv(t testing.TB) *Env {
	// Use an error group with a cancelable context to supervise every component
	// and cancel everything if one fails
	ctx, cancel := context.WithCancel(context.Background())
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
	etcdConfig.LogOutput = "default"
	etcdConfig.MaxTxnOps = 10000

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
	require.NoError(t, err)
	require.NoError(t, listener.Close())

	clientURL, err := url.Parse(fmt.Sprintf("http://%s", listener.Addr().String()))
	require.NoError(t, err)

	etcdConfig.LPUrls = []url.URL{}
	etcdConfig.LCUrls = []url.URL{*clientURL}

	// Throw away noisy messages from etcd - comment these out if you need to debug
	// a failed start
	capnslog.SetGlobalLogLevel(capnslog.CRITICAL)

	env.Etcd, err = embed.StartEtcd(etcdConfig)
	require.NoError(t, err)
	t.Cleanup(env.Etcd.Close)

	capnslog.SetGlobalLogLevel(capnslog.CRITICAL)

	eg.Go(func() error {
		return errorWait(ctx, env.Etcd.Err())
	})

	env.EtcdClient, err = etcd.New(etcd.Config{
		Context:     env.Context,
		Endpoints:   []string{clientURL.String()},
		DialOptions: client.DefaultDialOptions(),
	})
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
