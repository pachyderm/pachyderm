package grpcutil

import (
	"context"
	"fmt"
	"math"
	"net"
	"time"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"

	"github.com/pachyderm/pachyderm/src/client/pkg/tls"
	"github.com/pachyderm/pachyderm/src/client/pkg/tracing"
	log "github.com/sirupsen/logrus"
)

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
func NewServer(ctx context.Context, publicPortTLSAllowed bool) (*Server, error) {
	opts := []grpc.ServerOption{
		grpc.MaxConcurrentStreams(math.MaxUint32),
		grpc.MaxRecvMsgSize(MaxMsgSize),
		grpc.MaxSendMsgSize(MaxMsgSize),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             5 * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.UnaryInterceptor(tracing.UnaryServerInterceptor()),
		grpc.StreamInterceptor(tracing.StreamServerInterceptor()),
	}

	if publicPortTLSAllowed {
		// Validate environment
		certPath, keyPath, err := tls.GetCertPaths()
		if err != nil {
			log.Warnf("TLS disabled: %v", err)
		} else {
			// Read TLS cert and key
			transportCreds, err := credentials.NewServerTLSFromFile(certPath, keyPath)
			if err != nil {
				return nil, fmt.Errorf("couldn't build transport creds: %v", err)
			}
			opts = append(opts, grpc.Creds(transportCreds))
		}
	}

	server := grpc.NewServer(opts...)
	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		<-ctx.Done()
		server.GracefulStop() // This also closes the listeners
		return nil
	})

	return &Server{
		Server: server,
		eg:     eg,
	}, nil
}

// ListenTCP causes the gRPC server to listen on a given TCP host and port
func (s *Server) ListenTCP(host string, port uint16) error {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return err
	}

	s.eg.Go(func() error {
		return s.Server.Serve(listener)
	})

	return nil
}

// ListenUDS causes the gRPC server to listen on a given Unix domain socket
// path
func (s *Server) ListenUDS(name string) error {
	listener, err := net.Listen("unix", name)
	if err != nil {
		return nil, err
	}

	s.eg.Go(func() error {
		return s.Server.Serve(listener)
	})

	return nil
}

// Wait causes the gRPC server to wait until it finishes, returning any errors
// that happened
func (s *Server) Wait() error {
	return s.eg.Wait()
}
