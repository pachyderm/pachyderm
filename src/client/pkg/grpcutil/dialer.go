package grpcutil

import (
	"sync"

	"google.golang.org/grpc"
)

type dialer struct {
	opts []grpc.DialOption
	// A map from addresses to connections
	connMap map[string]*grpc.ClientConn
	lock    sync.Mutex
}

func newDialer(opts ...grpc.DialOption) *dialer {
	return &dialer{
		opts:    opts,
		connMap: make(map[string]*grpc.ClientConn),
	}
}

func (d *dialer) Dial(address string) (*grpc.ClientConn, error) {
	d.lock.Lock()
	defer d.lock.Unlock()
	if conn, ok := d.connMap[address]; ok {
		return conn, nil
	}
	conn, err := grpc.Dial(address, d.opts...)
	if err != nil {
		return nil, err
	}
	d.connMap[address] = conn
	return conn, nil
}

func (d *dialer) CloseConns() error {
	d.lock.Lock()
	defer d.lock.Unlock()
	for address, conn := range d.connMap {
		if err := conn.Close(); err != nil {
			return err
		}
		delete(d.connMap, address)
	}
	return nil
}
