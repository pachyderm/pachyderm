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
	// ErrMustSpecifyRegisterFunc is used when a register func is nil.
	ErrMustSpecifyRegisterFunc = errors.New("must specify registerFunc")

	// ErrMustSpecifyPort is used when a port is 0
	ErrMustSpecifyPort = errors.New("must specify port on which to serve")
)

// ServerOptions represent optional fields for serving.
type ServerOptions struct {
	Host         string
	Port         uint16
	AnyPort      bool
	MaxMsgSize   int
	RegisterFunc func(*grpc.Server) error

	// If set, grpcutil may enable TLS.  This should be set for public ports that
	// serve GRPC services to 3rd party clients.
	//
	// If set, the criterion for actually serving over TLS is:
	// if a signed TLS cert and corresponding private key in 'TLSVolumePath',
	// this will serve GRPC traffic over TLS. If either are missing this will
	// serve GRPC traffic over unencrypted HTTP,
	//
	// TODO make the TLS cert and key path a parameter, as pachd will need
	// multiple certificates for multiple ports
	PublicPortTLSAllowed bool
}

// ServerRun is the long-lived results of calling Serve and can be used to
// inspect or modify the servers directly. Specifically Listener can be used to
// obtain the listener port if AnyPort=true is used.
type ServerRun struct {
	GrpcServer *grpc.Server
	Listener   net.Listener
	Err        error
}

// Serve serves stuff.
func Serve(
	ctx context.Context,
	servers ...ServerOptions,
) ([]*ServerRun, *errgroup.Group) {
	serverRuns := []*ServerRun{}
	eg, ctx := errgroup.WithContext(ctx)
	ready := make(chan bool, len(servers))

	for _, server := range servers {
		serverRun := &ServerRun{}
		serverRuns = append(serverRuns, serverRun)

		eg.Go(func() (err error) {
			defer func() {
				serverRun.Err = err
			}()

			if server.RegisterFunc == nil {
				return ErrMustSpecifyRegisterFunc
			}
			if server.Port == 0 && !server.AnyPort {
				return ErrMustSpecifyPort
			}
			opts := []grpc.ServerOption{
				grpc.MaxConcurrentStreams(math.MaxUint32),
				grpc.MaxRecvMsgSize(server.MaxMsgSize),
				grpc.MaxSendMsgSize(server.MaxMsgSize),
				grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
					MinTime:             5 * time.Second,
					PermitWithoutStream: true,
				}),
				grpc.UnaryInterceptor(tracing.UnaryServerInterceptor()),
				grpc.StreamInterceptor(tracing.StreamServerInterceptor()),
			}
			if server.PublicPortTLSAllowed {
				// Validate environment
				certPath, keyPath, err := tls.GetCertPaths()
				if err != nil {
					log.Warnf("TLS disabled: %v", err)
				} else {
					// Read TLS cert and key
					transportCreds, err := credentials.NewServerTLSFromFile(certPath, keyPath)
					if err != nil {
						return fmt.Errorf("couldn't build transport creds: %v", err)
					}
					opts = append(opts, grpc.Creds(transportCreds))
				}
			}

			serverRun.GrpcServer = grpc.NewServer(opts...)
			if err = server.RegisterFunc(serverRun.GrpcServer); err != nil {
				return err
			}

			port := server.Port
			if server.AnyPort {
				port = 0
			}
			serverRun.Listener, err = net.Listen("tcp", fmt.Sprintf("%s:%d", server.Host, port))
			if err != nil {
				return err
			}

			eg.Go(func() error {
				<-ctx.Done()
				serverRun.GrpcServer.GracefulStop() // This also closes the TCP listener
				return nil
			})

			ready <- true

			return serverRun.GrpcServer.Serve(serverRun.Listener)
		})
	}

	for range servers {
		select {
		case <-ready:
			continue
		case <-ctx.Done():
			err := fmt.Errorf("Canceled before servers became ready: %v", ctx.Err())
			eg.Go(func() error {
				return err
			})
			for _, run := range serverRuns {
				if run.Err != nil {
					run.Err = err
				}
			}
			return serverRuns, eg
		}
	}

	return serverRuns, eg
}
