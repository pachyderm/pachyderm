package client

import (
	"crypto/x509"
	"encoding/base64"
	"errors"
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
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"

	types "github.com/gogo/protobuf/types"
	log "github.com/sirupsen/logrus"

	"github.com/pachyderm/pachyderm/src/client/admin"
	"github.com/pachyderm/pachyderm/src/client/auth"
	"github.com/pachyderm/pachyderm/src/client/debug"
	"github.com/pachyderm/pachyderm/src/client/enterprise"
	"github.com/pachyderm/pachyderm/src/client/health"
	"github.com/pachyderm/pachyderm/src/client/limit"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/config"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/client/pkg/tls"
	"github.com/pachyderm/pachyderm/src/client/pkg/tracing"
	"github.com/pachyderm/pachyderm/src/client/pps"
	"github.com/pachyderm/pachyderm/src/client/transaction"
	"github.com/pachyderm/pachyderm/src/client/version/versionpb"
)

const (
	// MaxListItemsLog specifies the maximum number of items we log in response to a List* API
	MaxListItemsLog = 10
	// StorageSecretName is the name of the Kubernetes secret in which
	// storage credentials are stored.
	StorageSecretName = "pachyderm-storage-secret"
)

// PfsAPIClient is an alias for pfs.APIClient.
type PfsAPIClient pfs.APIClient

// PpsAPIClient is an alias for pps.APIClient.
type PpsAPIClient pps.APIClient

// ObjectAPIClient is an alias for pfs.ObjectAPIClient
type ObjectAPIClient pfs.ObjectAPIClient

// AuthAPIClient is an alias of auth.APIClient
type AuthAPIClient auth.APIClient

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
	ObjectAPIClient
	AuthAPIClient
	VersionAPIClient
	AdminAPIClient
	TransactionAPIClient
	DebugClient
	Enterprise enterprise.APIClient // not embedded--method name conflicts with AuthAPIClient

	// addr is a "host:port" string pointing at a pachd endpoint
	addr string

	// The trusted CAs, for authenticating a pachd server over TLS
	caCerts *x509.CertPool

	// clientConn is a cached grpc connection to 'addr'
	clientConn *grpc.ClientConn

	// healthClient is a cached healthcheck client connected to 'addr'
	healthClient health.HealthClient

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
}

// GetAddress returns the pachd host:port with which 'c' is communicating. If
// 'c' was created using NewInCluster or NewOnUserMachine then this is how the
// address may be retrieved from the environment.
func (c *APIClient) GetAddress() string {
	return c.addr
}

// DefaultMaxConcurrentStreams defines the max number of Putfiles or Getfiles happening simultaneously
const DefaultMaxConcurrentStreams = 100

// DefaultDialTimeout is the max amount of time APIClient.connect() will wait
// for a connection to be established unless overridden by WithDialTimeout()
const DefaultDialTimeout = 30 * time.Second

type clientSettings struct {
	maxConcurrentStreams int
	dialTimeout          time.Duration
	caCerts              *x509.CertPool
}

// NewFromAddress constructs a new APIClient for the server at addr.
func NewFromAddress(addr string, options ...Option) (*APIClient, error) {
	// Validate address
	if strings.Contains(addr, "://") {
		return nil, fmt.Errorf("address shouldn't contain protocol (\"://\"), but is: %q", addr)
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
	c := &APIClient{
		addr:    addr,
		caCerts: settings.caCerts,
		limiter: limit.New(settings.maxConcurrentStreams),
	}
	if err := c.connect(settings.dialTimeout); err != nil {
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
		return fmt.Errorf("could not read x509 cert from \"%s\": %v", path, err)
	}
	if ok := pool.AppendCertsFromPEM(bytes); !ok {
		return fmt.Errorf("could not add %s to cert pool as PEM", path)
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
			return fmt.Errorf("server CA certs are present in Pachyderm config, but could not be added to client")
		}
		return nil
	}
}

// WithSystemCAs uses the system certs for client creatin.
func WithSystemCAs(settings *clientSettings) error {
	certs, err := x509.SystemCertPool()
	if err != nil {
		return fmt.Errorf("could not retrieve system cert pool: %v", err)
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
			return nil, fmt.Errorf("could not parse 'PACHD_ADDRESS': %v", err)
		}

		if !pachdAddress.Secured {
			return nil, fmt.Errorf("cannot set 'PACH_CA_CERTS' if 'PACHD_ADDRESS' is not using grpcs")
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
			return nil, nil, fmt.Errorf("could not parse 'PACHD_ADDRESS': %v", err)
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
			return nil, nil, fmt.Errorf("could not parse the active context's pachd address: %v", err)
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
				return nil, nil, fmt.Errorf("could not decode server CA certs in config: %v", err)
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
		return nil, 0, fmt.Errorf("failed to initialize port forwarder: %v", err)
	}

	port, err := fw.RunForDaemon(0, 650)
	if err != nil {
		return nil, 0, err
	}

	log.Debugf("Implicit port forwarder listening on port %d", port)

	return fw, port, nil
}

// NewForTest constructs a new APIClient for tests.
func NewForTest() (*APIClient, error) {
	cfg, err := config.Read(false)
	if err != nil {
		return nil, fmt.Errorf("could not read config: %v", err)
	}
	_, context, err := cfg.ActiveContext()
	if err != nil {
		return nil, fmt.Errorf("could not get active context: %v", err)
	}

	// create new pachctl client
	pachdAddress, cfgOptions, err := getUserMachineAddrAndOpts(context)
	if err != nil {
		return nil, err
	}

	if pachdAddress == nil {
		pachdAddress = &grpcutil.DefaultPachdAddress
	}

	client, err := NewFromAddress(pachdAddress.Hostname(), cfgOptions...)
	if err != nil {
		return nil, fmt.Errorf("could not connect to pachd at %s: %v", pachdAddress.Qualified(), err)
	}
	return client, nil
}

// NewOnUserMachine constructs a new APIClient using $HOME/.pachyderm/config
// if it exists. This is intended to be used in the pachctl binary.
//
// TODO(msteffen) this logic is fairly linux/unix specific, and makes the
// pachyderm client library incompatible with Windows. We may want to move this
// (and similar) logic into src/server and have it call a NewFromOptions()
// constructor.
func NewOnUserMachine(prefix string, options ...Option) (*APIClient, error) {
	cfg, err := config.Read(false)
	if err != nil {
		return nil, fmt.Errorf("could not read config: %v", err)
	}
	_, context, err := cfg.ActiveContext()
	if err != nil {
		return nil, fmt.Errorf("could not get active context: %v", err)
	}

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

	client, err := NewFromAddress(pachdAddress.Hostname(), append(options, cfgOptions...)...)
	if err != nil {
		return nil, fmt.Errorf("could not connect to pachd at %q: %v", pachdAddress.Qualified(), err)
	}

	// Add metrics info & authentication token
	client.metricsPrefix = prefix
	if cfg.UserID != "" && cfg.V2.Metrics {
		client.metricsUserID = cfg.UserID
	}
	if context.SessionToken != "" {
		client.authenticationToken = context.SessionToken
	}

	// Verify cluster ID
	clusterID, err := client.ClusterID()
	if err != nil && (context.ClusterID != "" || !strings.Contains(err.Error(), "unknown method")) {
		return nil, fmt.Errorf("could not get cluster ID: %v", err)
	}
	if context.ClusterID != clusterID {
		if context.ClusterID == "" {
			context.ClusterID = clusterID
			if err = cfg.Write(); err != nil {
				return nil, fmt.Errorf("could not write config to save cluster ID: %v", err)
			}
		} else {
			return nil, fmt.Errorf("connected to the wrong cluster (context cluster ID = %q vs reported cluster ID = %q)", context.ClusterID, clusterID)
		}
	}

	// Add port forwarding. This will set it to nil if port forwarding is
	// disabled, or an address is explicitly set.
	client.portForwarder = fw

	return client, nil
}

// NewInCluster constructs a new APIClient using env vars that Kubernetes creates.
// This should be used to access Pachyderm from within a Kubernetes cluster
// with Pachyderm running on it.
func NewInCluster(options ...Option) (*APIClient, error) {
	host, ok := os.LookupEnv("PACHD_SERVICE_HOST")
	if !ok {
		return nil, fmt.Errorf("PACHD_SERVICE_HOST not set")
	}
	port, ok := os.LookupEnv("PACHD_SERVICE_PORT")
	if !ok {
		return nil, fmt.Errorf("PACHD_SERVICE_PORT not set")
	}
	// create new pachctl client
	return NewFromAddress(fmt.Sprintf("%s:%s", host, port), options...)
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
	if _, err := c.AuthAPIClient.Deactivate(
		c.Ctx(),
		&auth.DeactivateRequest{},
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
		// Don't return from Dial() until the connection has been established
		grpc.WithBlock(),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(grpcutil.MaxMsgSize),
			grpc.MaxCallSendMsgSize(grpcutil.MaxMsgSize),
		),
	}
}

func (c *APIClient) connect(timeout time.Duration) error {
	keepaliveOpt := grpc.WithKeepaliveParams(keepalive.ClientParameters{
		Time:                20 * time.Second, // if 20s since last msg (any kind), ping
		Timeout:             20 * time.Second, // if no response to ping for 20s, reset
		PermitWithoutStream: true,             // send ping even if no active RPCs
	})
	dialOptions := append(DefaultDialOptions(), keepaliveOpt)
	if c.caCerts == nil {
		dialOptions = append(dialOptions, grpc.WithInsecure())
	} else {
		tlsCreds := credentials.NewClientTLSFromCert(c.caCerts, "")
		dialOptions = append(dialOptions, grpc.WithTransportCredentials(tlsCreds))
	}
	if tracing.IsActive() {
		dialOptions = append(dialOptions,
			grpc.WithUnaryInterceptor(tracing.UnaryClientInterceptor()),
			grpc.WithStreamInterceptor(tracing.StreamClientInterceptor()),
		)
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	clientConn, err := grpc.DialContext(ctx, c.addr, dialOptions...)
	if err != nil {
		return err
	}
	c.PfsAPIClient = pfs.NewAPIClient(clientConn)
	c.PpsAPIClient = pps.NewAPIClient(clientConn)
	c.ObjectAPIClient = pfs.NewObjectAPIClient(clientConn)
	c.AuthAPIClient = auth.NewAPIClient(clientConn)
	c.Enterprise = enterprise.NewAPIClient(clientConn)
	c.VersionAPIClient = versionpb.NewAPIClient(clientConn)
	c.AdminAPIClient = admin.NewAPIClient(clientConn)
	c.TransactionAPIClient = transaction.NewAPIClient(clientConn)
	c.DebugClient = debug.NewDebugClient(clientConn)
	c.clientConn = clientConn
	c.healthClient = health.NewHealthClient(clientConn)
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

// SetAuthToken sets the authentication token that will be used for all
// API calls for this client.
func (c *APIClient) SetAuthToken(token string) {
	c.authenticationToken = token
}
