package grpcutil

import (
	"fmt"
	"sync"

	"google.golang.org/grpc"
)

type dialer struct {
	opts []grpc.DialOption
	// TODO(pedge): this is insane for so many reasons
	addressToClientConn map[string]*grpc.ClientConn
	lock                *sync.RWMutex
}

func newDialer(opts ...grpc.DialOption) *dialer {
	return &dialer{opts, make(map[string]*grpc.ClientConn), &sync.RWMutex{}}
}

func (d *dialer) Dial(address string) (*grpc.ClientConn, error) {
	d.lock.RLock()
	clientConn := d.addressToClientConn[address]
	d.lock.RUnlock()
	if clientConn != nil {
		return clientConn, nil
	}
	d.lock.Lock()
	defer d.lock.Unlock()
	clientConn = d.addressToClientConn[address]
	if clientConn != nil {
		return clientConn, nil
	}
	var err error
	clientConn, err = grpc.Dial(address, d.opts...)
	if err != nil {
		return nil, err
	}
	d.addressToClientConn[address] = clientConn
	return clientConn, nil
}

func (d *dialer) Clean() error {
	d.lock.Lock()
	defer d.lock.Unlock()
	var errs []error
	for _, clientConn := range d.addressToClientConn {
		if err := clientConn.Close(); err != nil && err != grpc.ErrClientConnClosing {
			errs = append(errs, err)
		}
	}
	d.addressToClientConn = make(map[string]*grpc.ClientConn)
	if len(errs) > 0 {
		return fmt.Errorf("%v", errs)
	}
	return nil
}
