package client

import (
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/config"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/tls"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type clientSettings struct {
	pachdAddress         *grpcutil.PachdAddress
	maxConcurrentStreams int
	gzipCompress         bool
	dialTimeout          time.Duration
	caCerts              *x509.CertPool
	unaryInterceptors    []grpc.UnaryClientInterceptor
	streamInterceptors   []grpc.StreamClientInterceptor
	contextName          string
	portForwarder        *PortForwarder
}

// Option is a client creation option that may be passed to NewOnUserMachine(), or NewInCluster()
type Option func(*clientSettings) error

// WithMaxConcurrentStreams instructs the New* functions to create client that
// can have at most 'streams' concurrent streams open with pachd at a time
func WithMaxConcurrentStreams(streams int) Option {
	return func(settings *clientSettings) error {
		settings.maxConcurrentStreams = streams
		return nil
	}
}

func addCertFromFile(pool *x509.CertPool, path string) error {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return errors.Wrapf(err, "could not read x509 cert from \"%s\"", path)
	}
	if ok := pool.AppendCertsFromPEM(bytes); !ok {
		return errors.Errorf("could not add %s to cert pool as PEM", path)
	}
	return nil
}

// WithRootCAs instructs the New* functions to create client that uses the
// given signed x509 certificates as the trusted root certificates (instead of
// the system certs). Introduced to pass certs provided via command-line flags
func WithRootCAs(path string) Option {
	return func(settings *clientSettings) error {
		settings.caCerts = x509.NewCertPool()
		return addCertFromFile(settings.caCerts, path)
	}
}

// WithAdditionalRootCAs instructs the New* functions to additionally trust the
// given base64-encoded, signed x509 certificates as root certificates.
// Introduced to pass certs in the Pachyderm config
func WithAdditionalRootCAs(pemBytes []byte) Option {
	return func(settings *clientSettings) error {
		// append certs from config
		if settings.caCerts == nil {
			settings.caCerts = x509.NewCertPool()
		}
		if ok := settings.caCerts.AppendCertsFromPEM(pemBytes); !ok {
			return errors.Errorf("server CA certs are present in Pachyderm config, but could not be added to client")
		}
		return nil
	}
}

// WithSystemCAs uses the system certs for client creation, if no others are provided.
// This is the default behaviour when the scheme is `https` or `grpcs`.
func WithSystemCAs(settings *clientSettings) error {
	if settings.caCerts != nil {
		return nil
	}

	certs, err := x509.SystemCertPool()
	if err != nil {
		return errors.Wrap(err, "could not retrieve system cert pool")
	}
	settings.caCerts = certs
	return nil
}

// WithDialTimeout instructs the New* functions to use 't' as the deadline to
// connect to pachd
func WithDialTimeout(t time.Duration) Option {
	return func(settings *clientSettings) error {
		settings.dialTimeout = t
		return nil
	}
}

// WithGZIPCompression enabled GZIP compression for data on the wire
func WithGZIPCompression() Option {
	return func(settings *clientSettings) error {
		settings.gzipCompress = true
		return nil
	}
}

// WithAdditionalPachdCert instructs the New* functions to additionally trust
// the signed cert mounted in Pachd's cert volume. This is used by Pachd
// when connecting to itself (if no cert is present, the clients cert pool
// will not be modified, so that if no other options have been passed, pachd
// will connect to itself over an insecure connection)
func WithAdditionalPachdCert() Option {
	return func(settings *clientSettings) error {
		if _, err := os.Stat(tls.VolumePath); err == nil {
			if settings.caCerts == nil {
				settings.caCerts = x509.NewCertPool()
			}
			return addCertFromFile(settings.caCerts, path.Join(tls.VolumePath, tls.CertFile))
		}
		return nil
	}
}

// WithAdditionalUnaryClientInterceptors instructs the New* functions to add the provided
// UnaryClientInterceptors to the gRPC dial options when opening a client connection.  Internally,
// all of the provided options are coalesced into one chain, so it is safe to provide this option
// more than once.
//
// This client creates both Unary and Stream client connections, so you will probably want to supply
// a corresponding WithAdditionalStreamClientInterceptors call.
func WithAdditionalUnaryClientInterceptors(interceptors ...grpc.UnaryClientInterceptor) Option {
	return func(settings *clientSettings) error {
		settings.unaryInterceptors = append(settings.unaryInterceptors, interceptors...)
		return nil
	}
}

// WithAdditionalStreamClientInterceptors instructs the New* functions to add the provided
// StreamClientInterceptors to the gRPC dial options when opening a client connection.  Internally,
// all of the provided options are coalesced into one chain, so it is safe to provide this option
// more than once.
//
// This client creates both Unary and Stream client connections, so you will probably want to supply
// a corresponding WithAdditionalUnaryClientInterceptors option.
func WithAdditionalStreamClientInterceptors(interceptors ...grpc.StreamClientInterceptor) Option {
	return func(settings *clientSettings) error {
		settings.streamInterceptors = append(settings.streamInterceptors, interceptors...)
		return nil
	}
}

// WithAddrFromURI will configure the client to use the address from the uri
func WithAddrFromURI(uri string) Option {
	return func(s *clientSettings) error {
		pachdAddress, err := grpcutil.ParsePachdAddress(uri)
		if err != nil {
			return errors.Wrap(err, "could not parse the pachd address")
		}
		s.pachdAddress = pachdAddress
		return nil
	}
}

// WithAddrFromEnv will configure the client using the options from the env
func WithAddrFromEnv() Option {
	return func(s *clientSettings) error {
		pachURI, ok := os.LookupEnv("PACHD_ADDRESS")
		if !ok {
			return nil
		}
		addr, err := grpcutil.ParsePachdAddress(pachURI)
		if err != nil {
			return err
		}
		s.pachdAddress = addr
		return nil
	}
}

// WithInClusterAddr will check the preset kubernetes environment variables and use them for the pachyderm API endpoint
func WithInClusterAddr() Option {
	return func(s *clientSettings) error {
		// first try the pachd peer service (only supported on pachyderm >= 1.10),
		// which will work when TLS is enabled
		internalHost := os.Getenv("PACHD_PEER_SERVICE_HOST")
		internalPort := os.Getenv("PACHD_PEER_SERVICE_PORT")
		if internalHost != "" && internalPort != "" {
			return WithAddrFromURI(fmt.Sprintf("%s:%s", internalHost, internalPort))(s)
		}

		host, ok := os.LookupEnv("PACHD_SERVICE_HOST")
		if !ok {
			return errors.Errorf("PACHD_SERVICE_HOST not set")
		}
		port, ok := os.LookupEnv("PACHD_SERVICE_PORT")
		if !ok {
			return errors.Errorf("PACHD_SERVICE_PORT not set")
		}
		return WithAddrFromURI(fmt.Sprintf("%s:%s", host, port))(s)
	}
}

// WithPachdAddress configures the client to use the provided address to connect to pachd
// If the address is secure, then the system CA will be used.
func WithPachdAddress(pachdAddress *grpcutil.PachdAddress) Option {
	// By default, use the system CAs for secure connections
	// if no others are specified.
	return func(s *clientSettings) error {
		s.pachdAddress = pachdAddress
		if pachdAddress.Secured {
			if err := WithSystemCAs(s); err != nil {
				return err
			}
		}
		return nil
	}
}

// WithUserMachineDefaults sets the client base $HOME/.pachyderm/config
// if it exists. This is intended to be used in the pachctl binary.
func WithUserMachineDefaults() Option {
	return func(s *clientSettings) error {
		cfg, err := config.Read(false, false)
		if err != nil {
			return errors.Wrap(err, "could not read config")
		}
		name, context, err := cfg.ActiveContext(true)
		if err != nil {
			return errors.Wrap(err, "could not get active context")
		}
		pachdAddress, cfgOptions, err := getUserMachineAddrAndOpts(context)
		if err != nil {
			return err
		}
		var fw *PortForwarder
		if pachdAddress == nil && context.PortForwarders != nil {
			pachdLocalPort, ok := context.PortForwarders["pachd"]
			if ok {
				log.Debugf("Connecting to explicitly port forwarded pachd instance on port %d", pachdLocalPort)
				pachdAddress = &grpcutil.PachdAddress{
					Secured: false,
					Host:    "localhost",
					Port:    uint16(pachdLocalPort),
				}
			}
		}
		if pachdAddress == nil {
			var pachdLocalPort uint16
			fw, pachdLocalPort, err = portForwarder(context)
			if err != nil {
				return err
			}
			pachdAddress = &grpcutil.PachdAddress{
				Secured: false,
				Host:    "localhost",
				Port:    pachdLocalPort,
			}
		}
		s.pachdAddress = pachdAddress
		s.contextName = name
		s.portForwarder = fw
		for _, o := range cfgOptions {
			if err := o(s); err != nil {
				return err
			}
		}
		return nil
	}
}

func getCertOptionsFromEnv() ([]Option, error) {
	var options []Option
	if certPaths, ok := os.LookupEnv("PACH_CA_CERTS"); ok {
		fmt.Fprintln(os.Stderr, "WARNING: 'PACH_CA_CERTS' is deprecated and will be removed in a future release, use Pachyderm contexts instead.")

		pachdAddressStr, ok := os.LookupEnv("PACHD_ADDRESS")
		if !ok {
			return nil, errors.New("cannot set 'PACH_CA_CERTS' without setting 'PACHD_ADDRESS'")
		}

		pachdAddress, err := grpcutil.ParsePachdAddress(pachdAddressStr)
		if err != nil {
			return nil, errors.Wrapf(err, "could not parse 'PACHD_ADDRESS'")
		}

		if !pachdAddress.Secured {
			return nil, errors.Errorf("cannot set 'PACH_CA_CERTS' if 'PACHD_ADDRESS' is not using grpcs")
		}

		paths := strings.Split(certPaths, ",")
		for _, p := range paths {
			// Try to read all certs under 'p'--skip any that we can't read/stat
			if err := filepath.Walk(p, func(p string, info os.FileInfo, err error) error {
				if err != nil {
					log.Warnf("skipping \"%s\", could not stat path: %v", p, err)
					return nil // Don't try and fix any errors encountered by Walk() itself
				}
				if info.IsDir() {
					return nil // We'll just read the children of any directories when we traverse them
				}
				pemBytes, err := ioutil.ReadFile(p)
				if err != nil {
					log.Warnf("could not read server CA certs at %s: %v", p, err)
					return nil
				}
				options = append(options, WithAdditionalRootCAs(pemBytes))
				return nil
			}); err != nil {
				return nil, err
			}
		}
	}
	return options, nil
}

// getUserMachineAddrAndOpts is a helper for NewOnUserMachine that uses
// environment variables, config files, etc to figure out which address a user
// running a command should connect to.
func getUserMachineAddrAndOpts(context *config.Context) (*grpcutil.PachdAddress, []Option, error) {
	var options []Option

	// 1) PACHD_ADDRESS environment variable (shell-local) overrides global config
	if envAddrStr, ok := os.LookupEnv("PACHD_ADDRESS"); ok {
		fmt.Fprintln(os.Stderr, "WARNING: 'PACHD_ADDRESS' is deprecated and will be removed in a future release, use Pachyderm contexts instead.")

		envAddr, err := grpcutil.ParsePachdAddress(envAddrStr)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "could not parse 'PACHD_ADDRESS'")
		}
		options, err := getCertOptionsFromEnv()
		if err != nil {
			return nil, nil, err
		}

		return envAddr, options, nil
	}

	// 2) Get target address from global config if possible
	if context != nil && (context.ServerCAs != "" || context.PachdAddress != "") {
		pachdAddress, err := grpcutil.ParsePachdAddress(context.PachdAddress)
		if err != nil {
			return nil, nil, errors.Wrap(err, "could not parse the active context's pachd address")
		}

		// Proactively return an error in this case, instead of falling back to the default address below
		if context.ServerCAs != "" && !pachdAddress.Secured {
			return nil, nil, errors.New("must set pachd_address to grpcs://... if server_cas is set")
		}

		if pachdAddress.Secured {
			options = append(options, WithSystemCAs)
		}
		// Also get cert info from config (if set)
		if context.ServerCAs != "" {
			pemBytes, err := base64.StdEncoding.DecodeString(context.ServerCAs)
			if err != nil {
				return nil, nil, errors.Wrap(err, "could not decode server CA certs in config")
			}
			return pachdAddress, []Option{WithAdditionalRootCAs(pemBytes)}, nil
		}
		return pachdAddress, options, nil
	}

	// 3) Use default address (broadcast) if nothing else works
	options, err := getCertOptionsFromEnv() // error if PACH_CA_CERTS is set
	if err != nil {
		return nil, nil, err
	}
	return nil, options, nil
}

// prepend prepends opt to opts and returns the result i.e. opt ++ opts
func prepend(opts []Option, opt Option) []Option {
	return append([]Option{opt}, opts...)
}
