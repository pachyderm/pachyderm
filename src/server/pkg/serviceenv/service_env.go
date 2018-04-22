package serviceenv

import (
	"errors"
	"fmt"
	"os"
	"time"

	etcd "github.com/coreos/etcd/clientv3"
	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/server/pkg/backoff"
	log "github.com/sirupsen/logrus"
	kube "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

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
	// pachReady is closed once the pachyderm client has been initialized
	pachReady chan struct{}

	// etcdClient is an etcd client that's shared by all users of this environment
	etcdClient *etcd.Client
	// etcdReady is closed once the etcd client has been initialized
	etcdReady chan struct{}

	// kubeClient is a kubernetes client that, if initialized, is shared by all
	// users of this environment
	kubeClient *kube.Clientset
	// kubeReady is closed once the kubernetes client has been initialized
	kubeReady chan struct{}
}

// InitServiceEnv initializes this service environment. This dials a GRPC
// connection to pachd and etcd (in a background goroutine), and creates the
// template pachClient used by future calls to GetPachClient.
//
// This call returns immediately, but GetPachClient and GetEtcdClient block
// until their respective clients are ready
func InitServiceEnv(pachAddress, etcdAddress string) *ServiceEnv {
	env := &ServiceEnv{
		etcdReady: make(chan struct{}),
		pachReady: make(chan struct{}),
		kubeReady: make(chan struct{}),
	}
	go func() {
		var eg errgroup.Group
		eg.Go(env.initPachClient(pachAddress))
		eg.Go(env.initEtcdClient(etcdAddress))
		close(env.kubeReady) // close immediately so GetKubeClient doesn't block
		if err := eg.Wait(); err != nil {
			panic(err) // If pachd can't connect, there's no way to recover
		}
	}()
	return env // env is not ready yet
}

// InitWithKube is like InitServiceEnv, but also assumes that it's run inside
// a kubernetes cluster and tries to connect to the kubernetes API server.
func InitWithKube(pachAddress, etcdAddress string) *ServiceEnv {
	env := &ServiceEnv{
		etcdReady: make(chan struct{}),
		pachReady: make(chan struct{}),
		kubeReady: make(chan struct{}),
	}
	go func() {
		var eg errgroup.Group
		eg.Go(env.initPachClient(pachAddress))
		eg.Go(env.initEtcdClient(etcdAddress))
		eg.Go(env.initKubeClient)
		if err := eg.Wait(); err != nil {
			panic(err) // If pachd can't connect, there's no way to recover
		}
	}()
	return env // env is not ready yet
}

func (env *ServiceEnv) initPachClient(pachAddress string) func() error {
	return func() error {
		// validate argument
		if pachAddress == "" {
			return errors.New("cannot initialize pach client with empty pach address")
		}
		// Initialize pach client
		defer close(env.pachReady)
		return backoff.Retry(func() error {
			var err error
			env.pachClient, err = client.NewFromAddress(pachAddress)
			if err != nil {
				return fmt.Errorf("failed to initialize pach client: %v", err)
			}
			return nil
		}, backoff.RetryEvery(time.Second).For(time.Minute))
	}
}

func (env *ServiceEnv) initEtcdClient(etcdAddress string) func() error {
	return func() error {
		// validate argument
		if etcdAddress == "" {
			return errors.New("cannot initialize pach client with empty pach address")
		}
		// Initialize etcd
		defer close(env.etcdReady)
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
	}
}

func (env *ServiceEnv) initKubeClient() error {
	defer close(env.kubeReady)
	return backoff.Retry(func() error {
		// Get secure in-cluster config
		var kubeAddr string
		var ok bool
		cfg, err := rest.InClusterConfig()
		if err == nil {
			goto connect
		}

		// InClusterConfig failed, fall back to insecure config
		log.Errorf("falling back to insecure kube client due to error from NewInCluster: %s", err)
		kubeAddr, ok = os.LookupEnv("KUBERNETES_PORT_443_TCP_ADDR")
		if !ok {
			return fmt.Errorf("can't fall back to insecure kube client due to missing env var (failed to retrieve in-cluster config: %v)", err)
		}
		cfg = &rest.Config{
			Host: fmt.Sprintf("%s:443", kubeAddr),
			TLSClientConfig: rest.TLSClientConfig{
				Insecure: true,
			},
		}

	connect:
		env.kubeClient, err = kube.NewForConfig(cfg)
		if err != nil {
			return fmt.Errorf("could initialize kube client: %v", err)
		}
		return nil
	}, backoff.RetryEvery(time.Second).For(time.Minute))
}

// GetPachClient returns a pachd client with the same authentication
// credentials and cancellation as 'ctx' (ensuring that auth credentials are
// propagated through downstream RPCs).
//
// Functions that receive RPCs should call this to convert their RPC context to
// a Pachyderm client, and internal Pachyderm calls should accept clients
// returned by this call.
func (env *ServiceEnv) GetPachClient(ctx context.Context) *client.APIClient {
	<-env.pachReady // wait until pach client is connected
	return env.pachClient.WithCtx(ctx)
}

// GetEtcdClient returns the already connected etcd client without modification
func (env *ServiceEnv) GetEtcdClient() *etcd.Client {
	<-env.etcdReady // wait until etcd client is connected
	return env.etcdClient
}

// GetKubeClient returns the already connected Kubernetes API client without
// modification
func (env *ServiceEnv) GetKubeClient() *kube.Clientset {
	<-env.kubeReady
	if env.kubeClient == nil {
		panic("service env never connected to kubernetes")
	}
	return env.kubeClient
}
