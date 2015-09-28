package protoserver

import (
	"errors"
	"fmt"
	"math"
	"net"
	"net/http"
	"time"

	"go.pedge.io/proto/time"
	"go.pedge.io/proto/version"
	"go.pedge.io/protolog"

	"golang.org/x/net/context"

	"github.com/gengo/grpc-gateway/runtime"
	"github.com/golang/glog"
	"google.golang.org/grpc"
)

var (
	// ErrMustSpecifyPort is the error if no port is specified to Serve.
	ErrMustSpecifyPort = errors.New("must specify port")
	// ErrMustSpecifyRegisterFunc is the error if no registerFunc is specifed to Serve.
	ErrMustSpecifyRegisterFunc = errors.New("must specify registerFunc")
)

// ServeOptions represent optional fields for serving.
type ServeOptions struct {
	HTTPPort         uint16
	TracePort        uint16
	Version          *protoversion.Version
	HTTPRegisterFunc func(context.Context, *runtime.ServeMux, *grpc.ClientConn) error
	ConnState        func(net.Conn, http.ConnState)
}

// Serve serves stuff.
func Serve(
	port uint16,
	registerFunc func(*grpc.Server),
	opts ServeOptions,
) (retErr error) {
	start := time.Now()
	defer func() {
		if retErr != nil {
			protolog.Error(
				&ServerFinished{
					Error:    retErr.Error(),
					Duration: prototime.DurationToProto(time.Since(start)),
				},
			)
		} else {
			protolog.Info(
				&ServerFinished{
					Duration: prototime.DurationToProto(time.Since(start)),
				},
			)
		}
	}()
	if port == 0 {
		return ErrMustSpecifyPort
	}
	if registerFunc == nil {
		return ErrMustSpecifyRegisterFunc
	}
	s := grpc.NewServer(grpc.MaxConcurrentStreams(math.MaxUint32))
	registerFunc(s)
	if opts.Version != nil {
		protoversion.RegisterAPIServer(s, protoversion.NewAPIServer(opts.Version))
	}
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return err
	}
	errC := make(chan error)
	go func() { errC <- s.Serve(listener) }()
	if opts.TracePort != 0 {
		go func() { errC <- http.ListenAndServe(fmt.Sprintf(":%d", opts.TracePort), nil) }()
	}
	if opts.HTTPPort != 0 && (opts.Version != nil || opts.HTTPRegisterFunc != nil) {
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
		if opts.Version != nil {
			if err := protoversion.RegisterAPIHandler(ctx, mux, conn); err != nil {
				_ = conn.Close()
				return err
			}
		}
		if opts.HTTPRegisterFunc != nil {
			if err := opts.HTTPRegisterFunc(ctx, mux, conn); err != nil {
				_ = conn.Close()
				return err
			}
		}
		httpServer := &http.Server{
			Addr:      fmt.Sprintf(":%d", opts.HTTPPort),
			Handler:   mux,
			ConnState: opts.ConnState,
		}
		go func() { errC <- httpServer.ListenAndServe() }()
	}
	protolog.Info(
		&ServerStarted{
			Port:      uint32(port),
			HttpPort:  uint32(opts.HTTPPort),
			TracePort: uint32(opts.TracePort),
		},
	)
	return <-errC
}
