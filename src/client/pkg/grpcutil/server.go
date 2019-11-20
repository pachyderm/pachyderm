package grpcutil

import (
	"context"
	"errors"
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

var (
	// ErrMustSpecifyNetConfig is used when no TCP or UDS config is specified
	ErrMustSpecifyNetConfig = errors.New("must specify a TCP or UDS config, or both")

	// ErrServerAlreadyStarted is used when attempting to start a server that
	// is already running
	ErrServerAlreadyStarted = errors.New("server already started")
)

type TCPConfig struct {
	Host string
	Port uint16
}

type UDSConfig struct {
	Name string
}

type Server struct {
	Server      *grpc.Server
	TCPListener net.Listener
	UDSListener net.Listener

	eg  *errgroup.Group
	ctx context.Context

	tcpConfig *TCPConfig
	udsConfig *UDSConfig
}

// NewServer creates a new gRPC server, but does not start serving yet.
//
// If 'publicPortTLSAllowed' is set, grpcutil may enable TLS. This should be
// set for public ports that serve GRPC services to 3rd party clients. If set,
// the criterion for actually serving over TLS is: if a signed TLS cert and
// corresponding private key in 'TLSVolumePath', this will serve GRPC traffic
// over TLS. If either are missing this will serve GRPC traffic over
// unencrypted HTTP,
func NewServer(ctx context.Context, tcpConfig *TCPConfig, udsConfig *UDSConfig, maxMsgSize int, publicPortTLSAllowed bool) (*Server, error) {
	// TODO make the TLS cert and key path a parameter, as pachd will need
	// multiple certificates for multiple ports
	if tcpConfig == nil && udsConfig == nil {
		return nil, ErrMustSpecifyNetConfig
	}

	opts := []grpc.ServerOption{
		grpc.MaxConcurrentStreams(math.MaxUint32),
		grpc.MaxRecvMsgSize(maxMsgSize),
		grpc.MaxSendMsgSize(maxMsgSize),
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

	return &Server{
		Server: grpc.NewServer(opts...),
		eg:     nil,
		ctx:    ctx,

		tcpConfig: tcpConfig,
		udsConfig: udsConfig,
	}, nil
}

func (s *Server) Start() error {
	if s.eg != nil {
		return ErrServerAlreadyStarted
	}

	var ctx context.Context
	var err error
	s.eg, ctx = errgroup.WithContext(s.ctx)

	if s.tcpConfig != nil {
		s.TCPListener, err = net.Listen("tcp", fmt.Sprintf("%s:%d", s.tcpConfig.Host, s.tcpConfig.Port))
		if err != nil {
			return err
		}

		s.eg.Go(func() error {
			return s.Server.Serve(s.TCPListener)
		})
	}

	if s.udsConfig != nil {
		s.UDSListener, err = net.Listen("unix", s.udsConfig.Name)
		if err != nil {
			return err
		}

		s.eg.Go(func() error {
			return s.Server.Serve(s.UDSListener)
		})
	}

	s.eg.Go(func() error {
		<-ctx.Done()
		s.Server.GracefulStop() // This also closes the listeners
		return nil
	})

	return nil
}

func (s *Server) StartAndWait() error {
	if err := s.Start(); err != nil {
		return err
	}
	return s.eg.Wait()
}
