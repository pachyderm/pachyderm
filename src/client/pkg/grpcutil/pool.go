package grpcutil

import (
	"context"

	"google.golang.org/grpc"
)

// Pool stores a pool of grpc connections to, it's useful in places where you
// would otherwise need to create several connections.
type Pool struct {
	address string
	opts    []grpc.DialOption
	conns   chan *grpc.ClientConn
}

// NewPool creates a new connection pool, size is the maximum number of
// connections that it will cache. There is no limit to he number of
// connections that it can provide.
func NewPool(address string, size int, opts ...grpc.DialOption) *Pool {
	return &Pool{
		address: address,
		opts:    opts,
		conns:   make(chan *grpc.ClientConn, size),
	}
}

// Get returns a new connection, unlike sync.Pool if it has a cached connection
// it will always return it. Otherwise it will create a new one. Get errors
// only when it needs to Dial a new connection and that process fails.
func (p *Pool) Get(ctx context.Context) (*grpc.ClientConn, error) {
	select {
	case conn := <-p.conns:
		return conn, nil
	default:
		return grpc.DialContext(ctx, p.address, p.opts...)
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
	var retErr error
	for conn := range p.conns {
		if err := conn.Close(); err != nil {
			retErr = err
		}
	}
	return retErr
}
