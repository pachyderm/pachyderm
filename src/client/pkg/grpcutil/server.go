package grpcutil

import (
	"errors"
	"fmt"
	"math"
	"net"
	"os"
	"path"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"

	log "github.com/sirupsen/logrus"
)

const (
	// TLSVolumePath is the path at which the tls cert and private key (if any)
	// will be mounted in the pachd pod
	TLSVolumePath = "/pachd-tls-cert"

	// TLSCertFile is the name of the mounted file containing a TLS certificate
	// that identifies pachd
	TLSCertFile = "tls.crt"

	// TLSKeyFile is the name of the mounted file containing a private key
	// corresponding to the public certificate in TLSCertFile
	TLSKeyFile = "tls.key"
)

var (
	// ErrMustSpecifyRegisterFunc is used when a register func is nil.
	ErrMustSpecifyRegisterFunc = errors.New("must specify registerFunc")

	// ErrMustSpecifyPort is used when a port is 0
	ErrMustSpecifyPort = errors.New("must specify port on which to serve")
)

// ServerOptions represent optional fields for serving.
type ServerOptions struct {
	Port         uint16
	MaxMsgSize   int
	Cancel       chan struct{}
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

// Serve serves stuff.
func Serve(
	servers ...ServerOptions,
) (retErr error) {
	for _, server := range servers {
		if server.RegisterFunc == nil {
			return ErrMustSpecifyRegisterFunc
		}
		if server.Port == 0 {
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
		}
		if server.PublicPortTLSAllowed {
			// Validate environment
			certPath := path.Join(TLSVolumePath, TLSCertFile)
			keyPath := path.Join(TLSVolumePath, TLSKeyFile)
			_, certPathStatErr := os.Stat(certPath)
			_, keyPathStatErr := os.Stat(keyPath)
			if certPathStatErr != nil {
				log.Warnf("TLS disabled: could not stat public cert at %s: %v", certPath, certPathStatErr)
			}
			if keyPathStatErr != nil {
				log.Warnf("TLS disabled: could not stat private key at %s: %v", keyPath, keyPathStatErr)
			}
			if certPathStatErr == nil && keyPathStatErr == nil {
				// Read TLS cert and key
				transportCreds, err := credentials.NewServerTLSFromFile(certPath, keyPath)
				if err != nil {
					return fmt.Errorf("couldn't build transport creds: %v", err)
				}
				opts = append(opts, grpc.Creds(transportCreds))
			}
		}

		grpcServer := grpc.NewServer(opts...)
		if err := server.RegisterFunc(grpcServer); err != nil {
			return err
		}
		listener, err := net.Listen("tcp", fmt.Sprintf(":%d", server.Port))
		if err != nil {
			return err
		}
		if server.Cancel != nil {
			go func() {
				<-server.Cancel
				if err := listener.Close(); err != nil {
					fmt.Printf("listener.Close(): %v\n", err)
				}
			}()
		}
		if err := grpcServer.Serve(listener); err != nil {
			return err
		}
	}
	return nil
}
