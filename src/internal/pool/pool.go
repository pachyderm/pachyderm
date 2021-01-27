package pool

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/pachyderm/pachyderm/src/client/pkg/errors"

	"google.golang.org/grpc"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	kube "k8s.io/client-go/kubernetes"
)

// connCount stores a connection and a count of the number of datums currently outstanding
// cc is left nil when connCount is first created so that the connection can be made in
type connCount struct {
	cc    *grpc.ClientConn
	count int64
}

// Pool stores a pool of grpc connections to a k8s service, it's useful in
// places where you would otherwise need to keep recreating connections.
type Pool struct {
	port           int
	conns          map[string]*connCount
	connsLock      sync.Mutex
	connsCond      *sync.Cond
	endpointsWatch watch.Interface
	opts           []grpc.DialOption
	done           chan struct{}
	queueSize      int64
}

// NewPool creates a new connection pool with connections to pods in the
// given service.
func NewPool(kubeClient *kube.Clientset, namespace string, serviceName string, port int, queueSize int64, opts ...grpc.DialOption) (*Pool, error) {
	endpointsInterface := kubeClient.CoreV1().Endpoints(namespace)

	watch, err := endpointsInterface.Watch(metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(metav1.SetAsLabelSelector(
			map[string]string{"app": serviceName},
		)),
		Watch: true,
	})
	if err != nil {
		return nil, err
	}

	pool := &Pool{
		port:           port,
		endpointsWatch: watch,
		opts:           opts,
		done:           make(chan struct{}),
		queueSize:      queueSize,
	}
	pool.connsCond = sync.NewCond(&pool.connsLock)
	go pool.watchEndpoints()
	return pool, nil
}

func (p *Pool) watchEndpoints() {
	for {
		select {
		case event, ok := <-p.endpointsWatch.ResultChan():
			if !ok {
				return
			}
			endpoints := event.Object.(*v1.Endpoints)
			p.updateAddresses(endpoints)
		case <-p.done:
			return
		}
	}
}

func (p *Pool) updateAddresses(endpoints *v1.Endpoints) {
	addresses := make(map[string]*connCount)
	p.connsLock.Lock()
	defer p.connsLock.Unlock()
	for _, subset := range endpoints.Subsets {
		// According the k8s docs, the full set of endpoints is the cross
		// product of (addresses x ports).
		for _, address := range subset.Addresses {
			addr := fmt.Sprintf("%s:%d", address.IP, p.port)
			if cc := p.conns[addr]; cc != nil {
				addresses[addr] = cc
			} else {
				// we don't actually connect here because there's no way to
				// return the error
				addresses[addr] = &connCount{}
			}
		}
	}
	p.conns = addresses
	p.connsCond.Broadcast()
}

// Do allows you to do something with a grpc.ClientConn.
// Errors returned from f will be returned by Do.
func (p *Pool) Do(ctx context.Context, f func(cc *grpc.ClientConn) error) error {
	var conn *connCount
	if err := func() error {
		p.connsLock.Lock()
		defer p.connsLock.Unlock()
		for {
			for addr, mapConn := range p.conns {
				if mapConn.cc == nil {
					cc, err := grpc.DialContext(ctx, addr, p.opts...)
					if err != nil {
						return errors.Wrapf(err, "failed to connect to %s", addr)
					}
					mapConn.cc = cc
					conn = mapConn
					// We break because this conn has a count of 0 which we know
					// we're not beating
					break
				} else {
					mapConnCount := atomic.LoadInt64(&mapConn.count)
					if mapConnCount < p.queueSize && (conn == nil || mapConnCount < atomic.LoadInt64(&conn.count)) {
						conn = mapConn
					}
				}
			}
			if conn == nil {
				p.connsCond.Wait()
			} else {
				atomic.AddInt64(&conn.count, 1)
				break
			}
		}
		return nil
	}(); err != nil {
		return err
	}
	defer p.connsCond.Broadcast()
	defer atomic.AddInt64(&conn.count, -1)
	return f(conn.cc)
}

// Close closes all connections stored in the pool, it returns an error if any
// of the calls to Close error.
func (p *Pool) Close() error {
	close(p.done)
	var retErr error
	for _, conn := range p.conns {
		if conn.cc != nil {
			if err := conn.cc.Close(); err != nil {
				retErr = err
			}
		}
	}
	return retErr
}
