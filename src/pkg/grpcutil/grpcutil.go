package grpcutil

import (
	"fmt"
	"math"
	"net"
	"net/http"

	"golang.org/x/net/context"

	"github.com/gengo/grpc-gateway/runtime"
	"github.com/golang/glog"
	"github.com/pachyderm/pachyderm/src/pkg/discovery"
	"github.com/pachyderm/pachyderm/src/pkg/protoversion"
	"google.golang.org/grpc"
)

type Dialer interface {
	Dial(address string) (*grpc.ClientConn, error)
	Clean() error
}

func NewDialer(opts ...grpc.DialOption) Dialer {
	return newDialer(opts...)
}

type Registry interface {
	RegisterAddress(address string) <-chan error
}

type Provider interface {
	GetClientConn() (*grpc.ClientConn, error)
}

func NewRegistry(discoveryRegistry discovery.Registry) Registry {
	return newRegistry(discoveryRegistry, nil)
}

func NewProvider(discoveryRegistry discovery.Registry, dialer Dialer) Provider {
	return newRegistry(discoveryRegistry, dialer)
}

func GrpcDo(
	port int,
	httpPort int,
	tracePort int,
	version *protoversion.Version,
	registerFunc func(*grpc.Server),
	httpRegisterFunc func(context.Context, *runtime.ServeMux, *grpc.ClientConn) error,
) error {
	s := grpc.NewServer(grpc.MaxConcurrentStreams(math.MaxUint32))
	registerFunc(s)
	protoversion.RegisterApiServer(s, protoversion.NewAPIServer(version))
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	errC := make(chan error)
	go func() { errC <- s.Serve(listener) }()
	if tracePort != 0 {
		go func() { errC <- http.ListenAndServe(fmt.Sprintf(":%d", tracePort), nil) }()
	}
	if httpPort != 0 && httpRegisterFunc != nil {
		defer glog.Flush()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		mux := runtime.NewServeMux()
		conn, err := grpc.Dial(fmt.Sprintf("0.0.0.0:%d", port), grpc.WithInsecure())
		if err != nil {
			return err
		}
		go func() {
			<-ctx.Done()
			_ = conn.Close()
		}()
		if err := protoversion.RegisterApiHandler(ctx, mux, conn); err != nil {
			_ = conn.Close()
			return err
		}
		if err := httpRegisterFunc(ctx, mux, conn); err != nil {
			_ = conn.Close()
			return err
		}
		go func() { errC <- http.ListenAndServe(fmt.Sprintf(":%d", httpPort), mux) }()
	}
	return <-errC
}
