package grpcutil

import (
	"net"
	"time"

	"google.golang.org/grpc"
)

type staticAddr struct{}

func (staticAddr) Network() string {
	return ""
}

func (staticAddr) String() string {
	return ""
}

type staticListener struct {
	conn net.Conn
}

func newStaticListener(conn net.Conn) *staticListener {
	return &staticListener{conn: conn}
}

func (s *staticListener) Accept() (c net.Conn, err error) {
	return s.conn, nil
}

func (s *staticListener) Close() error {
	return nil
}

func (s *staticListener) Addr() net.Addr {
	return staticAddr{}
}

type dialerFunc func(addr string, timeout time.Duration) (net.Conn, error)

func staticDialFunc(conn net.Conn) dialerFunc {
	return func(addr string, timeout time.Duration) (net.Conn, error) {
		return conn, nil
	}
}

func staticClienConn(conn net.Conn) *grpc.ClientConn {
	// ignoring err is ok, we know it's nil
	result, _ := grpc.Dial("", grpc.WithDialer(staticDialFunc(conn)))
	return result
}
