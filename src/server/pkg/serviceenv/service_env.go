package serviceenv

import (
	"fmt"
	"sync"
	"time"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"

	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
)

// ServiceEnv is a struct containing connections to other services in the
// cluster. In pachd, there is only one instance of this struct, but tests may
// create more, if they want to create multiple pachyderm "clusters" served in
// separate goroutines
type ServiceEnv struct {
	// pachClient is the "template" client that contains the original GRPC
	// connection and that other clients returned by this library are based on. It
	// has no ctx and therefore no auth credentials or cancellation
	pachClient *client.APIClient

	// etcdClient is an etcd client that's shared by all users of this environment
	etcdClient *etcd.Client

	// connectOnce ensures that InitServiceEnv only connects pachClient and
	// etcdClient once
	connectOnce sync.Once

	// ready is closed once the clients have been initialized
	ready chan struct{}
}

// InitServiceEnv initializes this service environment. This dials a GRPC
// connection to pachd and etcd, and creates the template pachClient used by
// future calls to GetPachClient.
func InitServiceEnv(pachAddress, etcdAddress string) (env *ServiceEnv) {
	// validate arguments
	if pachAddress == "" {
		panic("cannot initialize pach client with empty pach address")
	}
	if etcdAddress == "" {
		panic("cannot initialize etcd client with empty etcd address")
	}

	// Create env, initialize it in a separate goroutine, and return the
	// uninitialized env
	env = &ServiceEnv{}
	go env.init(pachAddress, etcdAddress)
	return env // env is not ready yet
}

// init actually dials all GRPC connections used by clients in ServiceEnv
func (env *ServiceEnv) init(pachAddress, etcdAddress string) {
	env.connectOnce.Do(func() {
		var eg errgroup.Group

		// Initialize etcd
		eg.Go(func() error {
			return backoff.Retry(func() error {
				var err error
				env.etcdClient, err = etcd.New(etcd.Config{
					Endpoints:   []string{etcdAddress},
					DialOptions: client.EtcdDialOptions(),
				})
				if err != nil {
					return fmt.Errorf("failed to initialize etcd client: %v", err)
				}
				return nil
			}, backoff.RetryEvery(time.Second).For(time.Minute))
		})

		// Initialize pachd
		eg.Go(func() error {
			return backoff.Retry(func() error {
				var err error
				env.pachClient, err = client.NewFromAddress(pachAddress)
				if err != nil {
					return fmt.Errorf("failed to initialize pach client: %v", err)
				}
				return nil
			}, backoff.RetryEvery(time.Second).For(time.Minute))
		})

		// Wait for connections to dial
		if err := eg.Wait(); err != nil {
			// don't bother returning an error. If pachd can't connect to other services in
			// the cluster, there's no way to recover
			panic(err)
		}
		close(env.ready)
	})
}

// GetPachClient returns a pachd client with the same authentication
// credentials and cancellation as 'ctx' (ensuring that auth credentials are
// propagated through downstream RPCs).
//
// Functions that receive RPCs should call this to convert their RPC context to
// a Pachyderm client, and internal Pachyderm calls should accept clients
// returned by this call.
func (env *ServiceEnv) GetPachClient(ctx context.Context) *client.APIClient {
	<-env.ready // wait until InitServiceEnv is finished
	return env.pachClient.WithCtx(ctx)
}

// GetEtcdClient returns the already connected etcd client without modification
func (env *ServiceEnv) GetEtcdClient() *etcd.Client {
	<-env.ready // wait until InitServiceEnv is finished
	return env.etcdClient
}
