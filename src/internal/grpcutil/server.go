package grpcutil

import (
	"context"
	gotls "crypto/tls"
	"fmt"
	"math"
	"net"
	"time"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	// Import registers the grpc GZIP decoder
	_ "google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/keepalive"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	log "github.com/sirupsen/logrus"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/tls"
)

// Interceptor can be used to configure Unary and Stream interceptors
type Interceptor struct {
	UnaryServerInterceptor  grpc.UnaryServerInterceptor
	StreamServerInterceptor grpc.StreamServerInterceptor
}

// Server is a convenience wrapper to gRPC servers that simplifies their
// setup and execution
type Server struct {
	Server *grpc.Server
	eg     *errgroup.Group
}

// NewServer creates a new gRPC server, but does not start serving yet.
//
// If 'publicPortTLSAllowed' is set, grpcutil may enable TLS. This should be
// set for public ports that serve GRPC services to 3rd party clients. If set,
// the criterion for actually serving over TLS is: if a signed TLS cert and
// corresponding private key in 'TLSVolumePath', this will serve GRPC traffic
// over TLS. If either are missing this will serve GRPC traffic over
// unencrypted HTTP,
func NewServer(ctx context.Context, publicPortTLSAllowed bool, options ...grpc.ServerOption) (*Server, error) {
	opts := append([]grpc.ServerOption{
		grpc.MaxConcurrentStreams(math.MaxUint32),
		grpc.MaxRecvMsgSize(MaxMsgSize),
		grpc.MaxSendMsgSize(MaxMsgSize),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             5 * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.ChainUnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
		grpc.ChainStreamInterceptor(grpc_prometheus.StreamServerInterceptor),
	}, options...)

	var cLoader *tls.CertLoader
	if publicPortTLSAllowed {
		// Validate environment
		certPath, keyPath, err := tls.GetCertPaths()
		if err != nil {
			log.Warnf("TLS disabled: %v", err)
		} else {
			cLoader = tls.NewCertLoader(certPath, keyPath, tls.CertCheckFrequency)
			// Read TLS cert and key
			err := cLoader.LoadAndStart()
			if err != nil {
				return nil, errors.Wrapf(err, "couldn't build transport creds: %v", err)
			}
			transportCreds := credentials.NewTLS(&gotls.Config{GetCertificate: cLoader.GetCertificate})
			opts = append(opts, grpc.Creds(transportCreds))
		}
	}

	server := grpc.NewServer(opts...)
	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		<-ctx.Done()
		server.GracefulStop() // This also closes the listeners
		if cLoader != nil {
			cLoader.Stop()
		}
		return nil
	})

	return &Server{
		Server: server,
		eg:     eg,
	}, nil
}

func (s *Server) ListenSocket(path string) error {
	listener, err := net.Listen("unix", path)
	if err != nil {
		return errors.EnsureStack(err)
	}
	s.eg.Go(func() error {
		return errors.EnsureStack(s.Server.Serve(listener))
	})
	return nil
}

// ListenTCP causes the gRPC server to listen on a given TCP host and port
func (s *Server) ListenTCP(host string, port uint16) (net.Listener, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return nil, errors.EnsureStack(err)
	}

	s.eg.Go(func() error {
		return errors.EnsureStack(s.Server.Serve(listener))
	})

	return listener, nil
}

// Wait causes the gRPC server to wait until it finishes, returning any errors
// that happened
func (s *Server) Wait() error {
	return errors.EnsureStack(s.eg.Wait())
}
