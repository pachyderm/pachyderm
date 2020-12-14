package client

import (
	"context"
	"testing"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"google.golang.org/grpc"
)

type countingInterceptor struct {
	s int
	u int
}

func (i *countingInterceptor) unary() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, call grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		i.u++
		return call(ctx, method, req, reply, cc, opts...)
	}
}

func (i *countingInterceptor) stream() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		i.s++
		return streamer(ctx, desc, cc, method, opts...)
	}
}

func TestInterceptors(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	server, err := grpcutil.NewServer(ctx, false)
	if err != nil {
		t.Fatalf("server: %v", err)
	}
	defer server.Wait()

	listener, err := server.ListenTCP("localhost", 0)
	if err != nil {
		t.Fatalf("listener: %v", err)
	}
	defer listener.Close()
	pfs.RegisterAPIServer(server.Server, new(pfs.UnimplementedAPIServer))

	interceptor := new(countingInterceptor)
	c, err := NewFromAddress(listener.Addr().String(), WithAdditionalUnaryClientInterceptors(interceptor.unary()), WithAdditionalStreamClientInterceptors(interceptor.stream()))
	if err != nil {
		t.Fatalf("create client: %v", err)
	}
	defer c.Close()

	// Unary call.
	if err := c.CreateRepo("foo"); err == nil {
		t.Fatal("create repo: expected error")
	}
	if got, want := interceptor.u, 1; got != want {
		t.Errorf("unary call count:\n  got: %v\n want: %v", got, want)
	}
	if got, want := interceptor.s, 0; got != want {
		t.Errorf("stream call count:\n  got: %v\n want: %v", got, want)
	}

	// Stream call.
	if err := c.Fsck(true, func(*pfs.FsckResponse) error { return nil }); err == nil {
		t.Fatal("fsck: expected error")
	}
	if got, want := interceptor.u, 1; got != want {
		t.Errorf("unary call count:\n  got: %v\n want: %v", got, want)
	}
	if got, want := interceptor.s, 1; got != want {
		t.Errorf("stream call count:\n  got: %v\n want: %v", got, want)
	}
}
