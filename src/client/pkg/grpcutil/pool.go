package grpcutil

import (
	"context"
	"fmt"
	"sync"

	"google.golang.org/grpc"
	"k8s.io/kubernetes/pkg/api"
	kube "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/watch"
)

// Pool stores a pool of grpc connections to, it's useful in places where you
// would otherwise need to create several connections.
type Pool struct {
	// The addresses of the pods that we currently know about
	addresses []string
	// The index to the next pod that we will dial
	addressesIndex int
	addressesLock  sync.Mutex
	// addressesCond is used to notify goros that addresses have been obtained
	addressesCond  *sync.Cond
	endpointsWatch watch.Interface
	opts           []grpc.DialOption
	conns          chan *grpc.ClientConn
}

// NewPool creates a new connection pool with connections to pods in the
// given service.
func NewPool(kubeClient *kube.Client, namespace string, serviceName string, numWorkers int, opts ...grpc.DialOption) (*Pool, error) {
	endpointsInterface := kubeClient.Endpoints(namespace)

	watch, err := endpointsInterface.Watch(api.ListOptions{
		LabelSelector: labels.SelectorFromSet(
			map[string]string{"app": serviceName},
		),
		Watch: true,
	})
	if err != nil {
		return nil, err
	}

	pool := &Pool{
		endpointsWatch: watch,
		opts:           opts,
		conns:          make(chan *grpc.ClientConn, numWorkers),
	}
	pool.addressesCond = sync.NewCond(&pool.addressesLock)
	go pool.watchEndpoints()
	return pool, nil
}

func (p *Pool) watchEndpoints() {
	for {
		event, ok := <-p.endpointsWatch.ResultChan()
		if !ok {
			return
		}

		endpoints := event.Object.(*api.Endpoints)
		p.updateAddresses(endpoints)
	}
}

func (p *Pool) updateAddresses(endpoints *api.Endpoints) {
	var addresses []string
	for _, subset := range endpoints.Subsets {
		// According the k8s docs, the full set of endpoints is the cross
		// product of (addresses x ports).
		for _, address := range subset.Addresses {
			for _, port := range subset.Ports {
				addresses = append(addresses, fmt.Sprintf("%s:%d", address.IP, port.Port))
			}
		}
	}
	p.addressesLock.Lock()
	defer p.addressesLock.Unlock()
	p.addresses = addresses
	p.addressesCond.Broadcast()
}

// Get returns a new connection, unlike sync.Pool if it has a cached connection
// it will always return it. Otherwise it will create a new one. Get errors
// only when it needs to Dial a new connection and that process fails.
func (p *Pool) Get(ctx context.Context) (*grpc.ClientConn, error) {
	select {
	case conn := <-p.conns:
		return conn, nil
	default:
		p.addressesLock.Lock()
		defer p.addressesLock.Unlock()
		for len(p.addresses) == 0 {
			p.addressesCond.Wait()
		}
		oldIndex := p.addressesIndex
		p.addressesIndex++
		if p.addressesIndex >= len(p.addresses) {
			p.addressesIndex = 0
		}
		return grpc.DialContext(ctx, p.addresses[oldIndex], p.opts...)
	}
}

// Put returns the connection to the pool. If there are more than `size`
// connections already cached in the pool the connection will be closed. Put
// errors only when it Closes a connection and that call errors.
func (p *Pool) Put(conn *grpc.ClientConn) error {
	select {
	case p.conns <- conn:
		return nil
	default:
		return conn.Close()
	}
}

// Close closes all connections stored in the pool, it returns an error if any
// of the calls to Close error.
func (p *Pool) Close() error {
	p.endpointsWatch.Stop()
	var retErr error
	for conn := range p.conns {
		if err := conn.Close(); err != nil {
			retErr = err
		}
	}
	return retErr
}
