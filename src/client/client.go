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

	"golang.org/x/net/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/health/grpc_health_v1"

	// Import registers the grpc GZIP encoder
	_ "google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"

	types "github.com/gogo/protobuf/types"
	log "github.com/sirupsen/logrus"

	"github.com/pachyderm/pachyderm/v2/src/admin"
	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/client/limit"
	"github.com/pachyderm/pachyderm/v2/src/debug"
	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/identity"
	"github.com/pachyderm/pachyderm/v2/src/internal/config"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/grpcutil"
	"github.com/pachyderm/pachyderm/v2/src/internal/tls"
	"github.com/pachyderm/pachyderm/v2/src/internal/tracing"
	"github.com/pachyderm/pachyderm/v2/src/license"
	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"
	"github.com/pachyderm/pachyderm/v2/src/transaction"
	"github.com/pachyderm/pachyderm/v2/src/version/versionpb"
)

const (
	// MaxListItemsLog specifies the maximum number of items we log in response to a List* API
	MaxListItemsLog = 10
	// StorageSecretName is the name of the Kubernetes secret in which
	// storage credentials are stored.
	// TODO: The value "pachyderm-storage-secret" is hardcoded in the obj package to avoid a
	// obj -> client dependency, so any changes to this variable need to be applied there.
	// The obj package should eventually get refactored so that it does not have this dependency.
	StorageSecretName = "pachyderm-storage-secret"
	// PachctlSecretName is the name of the Kubernetes secret in which
	// pachctl credentials are stored.
	PachctlSecretName = "pachyderm-pachctl-secret"
)

// PfsAPIClient is an alias for pfs.APIClient.
type PfsAPIClient pfs.APIClient

// PpsAPIClient is an alias for pps.APIClient.
type PpsAPIClient pps.APIClient

// AuthAPIClient is an alias of auth.APIClient
type AuthAPIClient auth.APIClient

// IdentityAPIClient is an alias of identity.APIClient
type IdentityAPIClient identity.APIClient

// VersionAPIClient is an alias of versionpb.APIClient
type VersionAPIClient versionpb.APIClient

// AdminAPIClient is an alias of admin.APIClient
type AdminAPIClient admin.APIClient

// TransactionAPIClient is an alias of transaction.APIClient
type TransactionAPIClient transaction.APIClient

// DebugClient is an alias of debug.DebugClient
type DebugClient debug.DebugClient

// An APIClient is a wrapper around pfs, pps and block APIClients.
type APIClient struct {
	PfsAPIClient
	PpsAPIClient
	AuthAPIClient
	IdentityAPIClient
	VersionAPIClient
	AdminAPIClient
	TransactionAPIClient
	DebugClient
	Enterprise enterprise.APIClient // not embedded--method name conflicts with AuthAPIClient
	License    license.APIClient

	// addr is the parsed address used to connect to this server
	addr *grpcutil.PachdAddress

	// The trusted CAs, for authenticating a pachd server over TLS
	caCerts *x509.CertPool

	// gzipCompress configures whether to enable compression by default for all calls
	gzipCompress bool

	// clientConn is a cached grpc connection to 'addr'
	clientConn *grpc.ClientConn

	// healthClient is a cached healthcheck client connected to 'addr'
	healthClient grpc_health_v1.HealthClient

	// streamSemaphore limits the number of concurrent message streams between
	// this client and pachd
	limiter limit.ConcurrencyLimiter

	// metricsUserID is an identifier that is included in usage metrics sent to
	// Pachyderm Inc. and is used to count the number of unique Pachyderm users.
	// If unset, no usage metrics are sent back to Pachyderm Inc.
	metricsUserID string

	// metricsPrefix is used to send information from this client to Pachyderm Inc
	// for usage metrics
	metricsPrefix string

	// authenticationToken is an identifier that authenticates the caller in case
	// they want to access privileged data
	authenticationToken string

	// The context used in requests, can be set with WithCtx
	ctx context.Context

	portForwarder *PortForwarder

	// The client context name this client was created from, if it was created by
	// NewOnUserMachine
	clientContextName string
}

// GetAddress returns the pachd host:port with which 'c' is communicating. If
// 'c' was created using NewInCluster or NewOnUserMachine then this is how the
// address may be retrieved from the environment.
func (c *APIClient) GetAddress() *grpcutil.PachdAddress {
	return c.addr
}

// DefaultMaxConcurrentStreams defines the max number of Putfiles or Getfiles happening simultaneously
const DefaultMaxConcurrentStreams = 100

// DefaultDialTimeout is the max amount of time APIClient.connect() will wait
// for a connection to be established unless overridden by WithDialTimeout()
const DefaultDialTimeout = 30 * time.Second

type clientSettings struct {
	maxConcurrentStreams int
	gzipCompress         bool
	dialTimeout          time.Duration
	caCerts              *x509.CertPool
	unaryInterceptors    []grpc.UnaryClientInterceptor
	streamInterceptors   []grpc.StreamClientInterceptor
}

// NewFromURI creates a new client given a GRPC URI ex. grpc://test.example.com.
// If no scheme is specified `grpc://` is assumed. A scheme of `grpcs://` enables TLS.
func NewFromURI(uri string, options ...Option) (*APIClient, error) {
	pachdAddress, err := grpcutil.ParsePachdAddress(uri)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse the pachd address")
	}
	return NewFromPachdAddress(pachdAddress, options...)
}

// NewFromPachdAddress creates a new client given a parsed GRPC address
func NewFromPachdAddress(pachdAddress *grpcutil.PachdAddress, options ...Option) (*APIClient, error) {
	// By default, use the system CAs for secure connections
	// if no others are specified.
	if pachdAddress.Secured {
		options = append(options, WithSystemCAs)
	}

	// Apply creation options
	settings := clientSettings{
		maxConcurrentStreams: DefaultMaxConcurrentStreams,
		dialTimeout:          DefaultDialTimeout,
	}
	for _, option := range options {
		if err := option(&settings); err != nil {
			return nil, err
		}
	}
	if tracing.IsActive() {
		settings.unaryInterceptors = append(settings.unaryInterceptors, tracing.UnaryClientInterceptor())
		settings.streamInterceptors = append(settings.streamInterceptors, tracing.StreamClientInterceptor())
	}
	c := &APIClient{
		addr:         pachdAddress,
		caCerts:      settings.caCerts,
		limiter:      limit.New(settings.maxConcurrentStreams),
		gzipCompress: settings.gzipCompress,
	}
	if err := c.connect(settings.dialTimeout, settings.unaryInterceptors, settings.streamInterceptors); err != nil {
		return nil, err
	}
	return c, nil
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

func portForwarder(context *config.Context) (*PortForwarder, uint16, error) {
	fw, err := NewPortForwarder(context, "")
	if err != nil {
		return nil, 0, errors.Wrap(err, "failed to initialize port forwarder")
	}

	var port uint16
	if context.EnterpriseServer {
		port, err = fw.RunForEnterpriseServer(0, 650)
	} else {
		port, err = fw.RunForDaemon(0, 650)
	}

	if err != nil {
		return nil, 0, err
	}
	log.Debugf("Implicit port forwarder listening on port %d", port)

	return fw, port, nil
}

// NewForTest constructs a new APIClient for tests.
// TODO(actgardner): this should probably live in testutils and accept a testing.TB
func NewForTest() (*APIClient, error) {
	cfg, err := config.Read(false, false)
	if err != nil {
		return nil, errors.Wrap(err, "could not read config")
	}
	_, context, err := cfg.ActiveContext(true)
	if err != nil {
		return nil, errors.Wrap(err, "could not get active context")
	}

	// create new pachctl client
	pachdAddress, cfgOptions, err := getUserMachineAddrAndOpts(context)
	if err != nil {
		return nil, err
	}

	if pachdAddress == nil {
		pachdAddress = &grpcutil.DefaultPachdAddress
	}

	client, err := NewFromPachdAddress(pachdAddress, cfgOptions...)
	if err != nil {
		return nil, errors.Wrapf(err, "could not connect to pachd at %s", pachdAddress.Qualified())
	}
	return client, nil
}

// NewEnterpriseClientForTest constructs a new APIClient for tests.
// TODO(actgardner): this should probably live in testutils and accept a testing.TB
func NewEnterpriseClientForTest() (*APIClient, error) {
	cfg, err := config.Read(false, false)
	if err != nil {
		return nil, errors.Wrap(err, "could not read config")
	}
	_, context, err := cfg.ActiveEnterpriseContext(true)
	if err != nil {
		return nil, errors.Wrap(err, "could not get active context")
	}

	// create new pachctl client
	pachdAddress, cfgOptions, err := getUserMachineAddrAndOpts(context)
	if err != nil {
		return nil, err
	}

	if pachdAddress == nil {
		return nil, errors.New("no enterprise server configured")
	}

	client, err := NewFromPachdAddress(pachdAddress, cfgOptions...)
	if err != nil {
		return nil, errors.Wrapf(err, "could not connect to pachd at %s", pachdAddress.Qualified())
	}
	return client, nil
}

// NewOnUserMachine constructs a new APIClient using $HOME/.pachyderm/config
// if it exists. This is intended to be used in the pachctl binary.
func NewOnUserMachine(prefix string, options ...Option) (*APIClient, error) {
	cfg, err := config.Read(false, false)
	if err != nil {
		return nil, errors.Wrap(err, "could not read config")
	}
	name, context, err := cfg.ActiveContext(true)
	if err != nil {
		return nil, errors.Wrap(err, "could not get active context")
	}
	return newOnUserMachine(cfg, context, name, prefix, options...)
}

// NewEnterpriseClientOnUserMachine constructs a new APIClient using $HOME/.pachyderm/config
// if it exists. This is intended to be used in the pachctl binary to communicate with the
// enterprise server.
func NewEnterpriseClientOnUserMachine(prefix string, options ...Option) (*APIClient, error) {
	cfg, err := config.Read(false, false)
	if err != nil {
		return nil, errors.Wrap(err, "could not read config")
	}
	name, context, err := cfg.ActiveEnterpriseContext(true)
	if err != nil {
		return nil, errors.Wrap(err, "could not get active context")
	}
	return newOnUserMachine(cfg, context, name, prefix, options...)
}

// TODO(msteffen) this logic is fairly linux/unix specific, and makes the
// pachyderm client library incompatible with Windows. We may want to move this
// (and similar) logic into src/server and have it call a NewFromOptions()
// constructor.
func newOnUserMachine(cfg *config.Config, context *config.Context, contextName, prefix string, options ...Option) (*APIClient, error) {
	// create new pachctl client
	pachdAddress, cfgOptions, err := getUserMachineAddrAndOpts(context)
	if err != nil {
		return nil, err
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
			return nil, err
		}
		pachdAddress = &grpcutil.PachdAddress{
			Secured: false,
			Host:    "localhost",
			Port:    pachdLocalPort,
		}
	}

	client, err := NewFromPachdAddress(pachdAddress, append(options, cfgOptions...)...)
	if err != nil {
		return nil, errors.Wrapf(err, "could not connect to pachd at %q", pachdAddress.Qualified())
	}

	// Add metrics info & authentication token
	client.metricsPrefix = prefix
	if cfg.UserID != "" && cfg.V2.Metrics {
		client.metricsUserID = cfg.UserID
	}
	if context.SessionToken != "" {
		client.authenticationToken = context.SessionToken
	}

	// Verify cluster deployment ID
	clusterInfo, err := client.InspectCluster()
	if err != nil {
		return nil, errors.Wrap(err, "could not get cluster ID")
	}
	if context.ClusterDeploymentID != clusterInfo.DeploymentID {
		if context.ClusterDeploymentID == "" {
			context.ClusterDeploymentID = clusterInfo.DeploymentID
			if err = cfg.Write(); err != nil {
				return nil, errors.Wrap(err, "could not write config to save cluster deployment ID")
			}
		} else {
			return nil, errors.Errorf("connected to the wrong cluster (context cluster deployment ID = %q vs reported cluster deployment ID = %q)", context.ClusterDeploymentID, clusterInfo.DeploymentID)
		}
	}

	// Add port forwarding. This will set it to nil if port forwarding is
	// disabled, or an address is explicitly set.
	client.portForwarder = fw

	client.clientContextName = contextName

	return client, nil
}

// NewInCluster constructs a new APIClient using env vars that Kubernetes creates.
// This should be used to access Pachyderm from within a Kubernetes cluster
// with Pachyderm running on it.
func NewInCluster(options ...Option) (*APIClient, error) {
	// first try the pachd peer service (only supported on pachyderm >= 1.10),
	// which will work when TLS is enabled
	internalHost := os.Getenv("PACHD_PEER_SERVICE_HOST")
	internalPort := os.Getenv("PACHD_PEER_SERVICE_PORT")
	if internalHost != "" && internalPort != "" {
		return NewFromURI(fmt.Sprintf("%s:%s", internalHost, internalPort), options...)
	}

	host, ok := os.LookupEnv("PACHD_SERVICE_HOST")
	if !ok {
		return nil, errors.Errorf("PACHD_SERVICE_HOST not set")
	}
	port, ok := os.LookupEnv("PACHD_SERVICE_PORT")
	if !ok {
		return nil, errors.Errorf("PACHD_SERVICE_PORT not set")
	}
	// create new pachctl client
	return NewFromURI(fmt.Sprintf("%s:%s", host, port), options...)
}

// NewInWorker constructs a new APIClient intended to be used from a worker
// to talk to the sidecar pachd container
func NewInWorker(options ...Option) (*APIClient, error) {
	cfg, err := config.Read(false, true)
	if err != nil {
		return nil, errors.Wrap(err, "could not read config")
	}
	_, context, err := cfg.ActiveContext(true)
	if err != nil {
		return nil, errors.Wrap(err, "could not get active context")
	}

	if localPort, ok := os.LookupEnv("PEER_PORT"); ok {
		client, err := NewFromURI(fmt.Sprintf("127.0.0.1:%s", localPort), options...)
		if err != nil {
			return nil, errors.Wrap(err, "could not create client")
		}
		if context.SessionToken != "" {
			client.authenticationToken = context.SessionToken
		}
		return client, nil
	}
	return nil, errors.New("PEER_PORT not set")
}

// Close the connection to gRPC
func (c *APIClient) Close() error {
	if err := c.clientConn.Close(); err != nil {
		return err
	}

	if c.portForwarder != nil {
		c.portForwarder.Close()
	}

	return nil
}

// DeleteAll deletes everything in the cluster.
// Use with caution, there is no undo.
// TODO: rewrite this to use transactions
func (c APIClient) DeleteAll() error {
	if _, err := c.IdentityAPIClient.DeleteAll(
		c.Ctx(),
		&identity.DeleteAllRequest{},
	); err != nil && !auth.IsErrNotActivated(err) {
		return grpcutil.ScrubGRPC(err)
	}
	if _, err := c.AuthAPIClient.Deactivate(
		c.Ctx(),
		&auth.DeactivateRequest{},
	); err != nil && !auth.IsErrNotActivated(err) {
		return grpcutil.ScrubGRPC(err)
	}
	if _, err := c.License.DeleteAll(
		c.Ctx(),
		&license.DeleteAllRequest{},
	); err != nil && !auth.IsErrNotActivated(err) {
		return grpcutil.ScrubGRPC(err)
	}
	if _, err := c.PpsAPIClient.DeleteAll(
		c.Ctx(),
		&types.Empty{},
	); err != nil {
		return grpcutil.ScrubGRPC(err)
	}
	if _, err := c.PfsAPIClient.DeleteAll(
		c.Ctx(),
		&types.Empty{},
	); err != nil {
		return grpcutil.ScrubGRPC(err)
	}
	if _, err := c.TransactionAPIClient.DeleteAll(
		c.Ctx(),
		&transaction.DeleteAllRequest{},
	); err != nil {
		return grpcutil.ScrubGRPC(err)
	}
	return nil
}

// DeleteAllEnterprise deletes everything in the enterprise server.
// Use with caution, there is no undo.
// TODO: rewrite this to use transactions
func (c APIClient) DeleteAllEnterprise() error {
	if _, err := c.IdentityAPIClient.DeleteAll(
		c.Ctx(),
		&identity.DeleteAllRequest{},
	); err != nil && !auth.IsErrNotActivated(err) {
		return grpcutil.ScrubGRPC(err)
	}
	if _, err := c.AuthAPIClient.Deactivate(
		c.Ctx(),
		&auth.DeactivateRequest{},
	); err != nil && !auth.IsErrNotActivated(err) {
		return grpcutil.ScrubGRPC(err)
	}
	if _, err := c.License.DeleteAll(
		c.Ctx(),
		&license.DeleteAllRequest{},
	); err != nil && !auth.IsErrNotActivated(err) {
		return grpcutil.ScrubGRPC(err)
	}
	return nil
}

// SetMaxConcurrentStreams Sets the maximum number of concurrent streams the
// client can have. It is not safe to call this operations while operations are
// outstanding.
func (c APIClient) SetMaxConcurrentStreams(n int) {
	c.limiter = limit.New(n)
}

// DefaultDialOptions is a helper returning a slice of grpc.Dial options
// such that grpc.Dial() is synchronous: the call doesn't return until
// the connection has been established and it's safe to send RPCs
func DefaultDialOptions() []grpc.DialOption {
	return []grpc.DialOption{
		// Don't return from Dial() until the connection has been established.
		grpc.WithBlock(),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                20 * time.Second,
			Timeout:             20 * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(grpcutil.MaxMsgSize),
			grpc.MaxCallSendMsgSize(grpcutil.MaxMsgSize),
		),
	}
}

func (c *APIClient) connect(timeout time.Duration, unaryInterceptors []grpc.UnaryClientInterceptor, streamInterceptors []grpc.StreamClientInterceptor) error {
	dialOptions := DefaultDialOptions()
	if c.caCerts == nil {
		dialOptions = append(dialOptions, grpc.WithInsecure())
	} else {
		tlsCreds := credentials.NewClientTLSFromCert(c.caCerts, "")
		dialOptions = append(dialOptions, grpc.WithTransportCredentials(tlsCreds))
	}
	if c.gzipCompress {
		dialOptions = append(dialOptions, grpc.WithDefaultCallOptions(grpc.UseCompressor("gzip")))
	}
	if len(unaryInterceptors) > 0 {
		dialOptions = append(dialOptions, grpc.WithChainUnaryInterceptor(unaryInterceptors...))
	}
	if len(streamInterceptors) > 0 {
		dialOptions = append(dialOptions, grpc.WithChainStreamInterceptor(streamInterceptors...))
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// By default GRPC will attempt to get service config from a TXT record when
	// the `dns:///` scheme is used. Some DNS servers return the wrong type of error
	// when the TXT record doesn't exist, and it causes GRPC to back off and retry
	// service discovery forever.
	dialOptions = append(dialOptions, grpc.WithDisableServiceConfig())

	clientConn, err := grpc.DialContext(ctx, c.addr.Target(), dialOptions...)
	if err != nil {
		return err
	}
	c.PfsAPIClient = pfs.NewAPIClient(clientConn)
	c.PpsAPIClient = pps.NewAPIClient(clientConn)
	c.AuthAPIClient = auth.NewAPIClient(clientConn)
	c.IdentityAPIClient = identity.NewAPIClient(clientConn)
	c.Enterprise = enterprise.NewAPIClient(clientConn)
	c.License = license.NewAPIClient(clientConn)
	c.VersionAPIClient = versionpb.NewAPIClient(clientConn)
	c.AdminAPIClient = admin.NewAPIClient(clientConn)
	c.TransactionAPIClient = transaction.NewAPIClient(clientConn)
	c.DebugClient = debug.NewDebugClient(clientConn)
	c.clientConn = clientConn
	c.healthClient = grpc_health_v1.NewHealthClient(clientConn)
	return nil
}

// AddMetadata adds necessary metadata (including authentication credentials)
// to the context 'ctx', preserving any metadata that is present in either the
// incoming or outgoing metadata of 'ctx'.
func (c *APIClient) AddMetadata(ctx context.Context) context.Context {
	// TODO(msteffen): There are several places in this client where it's possible
	// to set per-request metadata (specifically auth tokens): client.WithCtx(),
	// client.SetAuthToken(), etc. These should be consolidated, as this API
	// doesn't make it obvious how these settings are resolved when they conflict.
	clientData := make(map[string]string)
	if c.authenticationToken != "" {
		clientData[auth.ContextTokenKey] = c.authenticationToken
	}
	// metadata API downcases all the key names
	if c.metricsUserID != "" {
		clientData["userid"] = c.metricsUserID
		clientData["prefix"] = c.metricsPrefix
	}

	// Rescue any metadata pairs already in 'ctx' (otherwise
	// metadata.NewOutgoingContext() would drop them). Note that this is similar
	// to metadata.Join(), but distinct because it discards conflicting k/v pairs
	// instead of merging them)
	incomingMD, _ := metadata.FromIncomingContext(ctx)
	outgoingMD, _ := metadata.FromOutgoingContext(ctx)
	clientMD := metadata.New(clientData)
	finalMD := make(metadata.MD) // Collect k/v pairs
	for _, md := range []metadata.MD{incomingMD, outgoingMD, clientMD} {
		for k, v := range md {
			finalMD[k] = v
		}
	}
	return metadata.NewOutgoingContext(ctx, finalMD)
}

// Ctx is a convenience function that returns adds Pachyderm authn metadata
// to context.Background().
func (c *APIClient) Ctx() context.Context {
	if c.ctx == nil {
		return c.AddMetadata(context.Background())
	}
	return c.AddMetadata(c.ctx)
}

// WithCtx returns a new APIClient that uses ctx for requests it sends. Note
// that the new APIClient will still use the authentication token and metrics
// metadata of this client, so this is only useful for propagating other
// context-associated metadata.
func (c *APIClient) WithCtx(ctx context.Context) *APIClient {
	result := *c // copy c
	result.ctx = ctx
	return &result
}

// AuthToken gets the authentication token that is set for this client.
func (c *APIClient) AuthToken() string {
	return c.authenticationToken
}

// SetAuthToken sets the authentication token that will be used for all
// API calls for this client.
func (c *APIClient) SetAuthToken(token string) {
	c.authenticationToken = token
}

// ClientContextName returns the name of the context in the client config
// that produced this client, or an empty string if the client was not
// produced from a configured client context.
func (c *APIClient) ClientContextName() string {
	return c.clientContextName
}
