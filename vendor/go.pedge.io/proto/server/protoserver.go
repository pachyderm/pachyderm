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
	// ErrCannotSpecifyBothHTTPPortAndHTTPAddress is the error if both HTTPPort and HTTPAddress are specified in ServeOptions.
	ErrCannotSpecifyBothHTTPPortAndHTTPAddress = errors.New("cannot specify both HTTPPort and HTTPAddress")
)

// ServeOptions represent optional fields for serving.
type ServeOptions struct {
	HTTPPort         uint16
	DebugPort        uint16
	Version          *protoversion.Version
	HTTPRegisterFunc func(context.Context, *runtime.ServeMux, *grpc.ClientConn) error
	// either HTTPPort or HTTPAddress can be set, but not both
	HTTPAddress     string
	HTTPListener    net.Listener
	ServeMuxOptions []runtime.ServeMuxOption
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
	if opts.HTTPPort != 0 && opts.HTTPAddress != "" {
		return ErrCannotSpecifyBothHTTPPortAndHTTPAddress
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
	if opts.DebugPort != 0 {
		go func() { errC <- http.ListenAndServe(fmt.Sprintf(":%d", opts.DebugPort), nil) }()
	}
	if (opts.HTTPPort != 0 || opts.HTTPAddress != "") && (opts.Version != nil || opts.HTTPRegisterFunc != nil) {
		defer glog.Flush()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		var mux *runtime.ServeMux
		if len(opts.ServeMuxOptions) == 0 {
			mux = runtime.NewServeMux()
		} else {
			mux = runtime.NewServeMux(opts.ServeMuxOptions...)
		}
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
		httpAddress := fmt.Sprintf(":%d", opts.HTTPPort)
		if opts.HTTPAddress != "" {
			httpAddress = opts.HTTPAddress
		}
		httpServer := &http.Server{
			Addr:    httpAddress,
			Handler: mux,
		}
		go func() {
			if opts.HTTPListener != nil {
				errC <- httpServer.Serve(listener)
			} else {
				errC <- httpServer.ListenAndServe()
			}
		}()
	}
	protolog.Info(
		&ServerStarted{
			Port:        uint32(port),
			HttpPort:    uint32(opts.HTTPPort),
			DebugPort:   uint32(opts.DebugPort),
			HttpAddress: opts.HTTPAddress,
		},
	)
	return <-errC
}
