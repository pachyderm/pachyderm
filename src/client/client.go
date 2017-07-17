package client

import (
	"errors"
	"fmt"
	"os"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	types "github.com/gogo/protobuf/types"
	log "github.com/sirupsen/logrus"

	"github.com/pachyderm/pachyderm/src/client/health"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/config"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/client/pps"
)

// PfsAPIClient is an alias for pfs.APIClient.
type PfsAPIClient pfs.APIClient

// PpsAPIClient is an alias for pps.APIClient.
type PpsAPIClient pps.APIClient

// ObjectAPIClient is an alias for pfs.ObjectAPIClient
type ObjectAPIClient pfs.ObjectAPIClient

// An APIClient is a wrapper around pfs, pps and block APIClients.
type APIClient struct {
	PfsAPIClient
	PpsAPIClient
	ObjectAPIClient

	// addr is a "host:port" string pointing at a pachd endpoint
	addr string

	// context.Context includes context options for all RPCs, including deadline
	_ctx context.Context

	// clientConn is a cached grpc connection to 'addr'
	clientConn *grpc.ClientConn

	// healthClient is a cached healthcheck client connected to 'addr'
	healthClient health.HealthClient

	// cancel is the cancel function for 'clientConn' and is called if
	// 'healthClient' fails health checking
	cancel func()

	// streamSemaphore limits the number of concurrent message streams between
	// this client and pachd
	streamSemaphore chan struct{}

	// metricsUserID is an identifier that is included in usage metrics sent to
	// Pachyderm Inc. and is used to count the number of unique Pachyderm users.
	// If unset, no usage metrics are sent back to Pachyderm Inc.
	metricsUserID string

	// metricsPrefix is used to send information from this client to Pachyderm Inc
	// for usage metrics
	metricsPrefix string
}

// GetAddress returns the pachd host:post with which 'c' is communicating. If
// 'c' was created using NewInCluster or NewOnUserMachine then this is how the
// address may be retrieved from the environment.
func (c *APIClient) GetAddress() string {
	return c.addr
}

// DefaultMaxConcurrentStreams defines the max number of Putfiles or Getfiles happening simultaneously
const DefaultMaxConcurrentStreams uint = 100

// NewFromAddressWithConcurrency constructs a new APIClient and sets the max
// concurrency of streaming requests (GetFile / PutFile)
func NewFromAddressWithConcurrency(addr string, maxConcurrentStreams uint) (*APIClient, error) {
	c := &APIClient{
		addr:            addr,
		streamSemaphore: make(chan struct{}, maxConcurrentStreams),
	}
	if err := c.connect(); err != nil {
		return nil, err
	}
	return c, nil
}

// NewFromAddress constructs a new APIClient for the server at addr.
func NewFromAddress(addr string) (*APIClient, error) {
	return NewFromAddressWithConcurrency(addr, DefaultMaxConcurrentStreams)
}

// GetAddressFromUserMachine interprets the Pachyderm config in 'cfg' in the
// context of local environment variables and returns a "host:port" string
// pointing at a Pachd target.
func GetAddressFromUserMachine(cfg *config.Config) string {
	address := "0.0.0.0:30650"
	if cfg != nil && cfg.V1 != nil && cfg.V1.PachdAddress != "" {
		address = cfg.V1.PachdAddress
	}
	// ADDRESS environment variable (shell-local) overrides global config
	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		address = envAddr
	}
	return address
}

// NewOnUserMachine constructs a new APIClient using env vars that may be set
// on a user's machine (i.e. ADDRESS), as well as $HOME/.pachyderm/config if it
// exists. This is primarily intended to be used with the pachctl binary, but
// may also be useful in tests.
//
// TODO(msteffen) this logic is fairly linux/unix specific, and makes the
// pachyderm client library incompatible with Windows. We may want to move this
// (and similar) logic into src/server and have it call a NewFromOptions()
// constructor.
func NewOnUserMachine(reportMetrics bool, prefix string) (*APIClient, error) {
	return NewOnUserMachineWithConcurrency(reportMetrics, prefix, DefaultMaxConcurrentStreams)
}

// NewOnUserMachineWithConcurrency is identical to NewOnUserMachine, but
// explicitly sets a limit on the number of RPC streams that may be open
// simultaneously
func NewOnUserMachineWithConcurrency(reportMetrics bool, prefix string, maxConcurrentStreams uint) (*APIClient, error) {
	cfg, err := config.Read()
	if err != nil {
		// metrics errors are non fatal
		log.Warningf("error loading user config from ~/.pachderm/config: %v", err)
	}

	// create new pachctl client
	client, err := NewFromAddress(GetAddressFromUserMachine(cfg))
	if err != nil {
		return nil, err
	}

	// Add metrics info
	if cfg != nil && cfg.UserID != "" && reportMetrics {
		client.metricsUserID = cfg.UserID
	}
	return client, nil
}

// NewInCluster constructs a new APIClient using env vars that Kubernetes creates.
// This should be used to access Pachyderm from within a Kubernetes cluster
// with Pachyderm running on it.
func NewInCluster() (*APIClient, error) {
	if addr := os.Getenv("PACHD_PORT_650_TCP_ADDR"); addr != "" {
		return NewFromAddress(fmt.Sprintf("%v:650", addr))
	}
	return nil, fmt.Errorf("PACHD_PORT_650_TCP_ADDR not set")
}

// Close the connection to gRPC
func (c *APIClient) Close() error {
	return c.clientConn.Close()
}

// KeepConnected periodically health checks the connection and attempts to
// reconnect if it becomes unhealthy.
func (c *APIClient) KeepConnected(cancel chan bool) {
	for {
		select {
		case <-cancel:
			return
		case <-time.After(time.Second * 5):
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			if _, err := c.healthClient.Health(ctx, &types.Empty{}); err != nil {
				c.cancel()
				c.connect()
			}
			cancel()
		}
	}
}

// DeleteAll deletes everything in the cluster.
// Use with caution, there is no undo.
func (c APIClient) DeleteAll() error {
	if _, err := c.PpsAPIClient.DeleteAll(
		c.ctx(),
		&types.Empty{},
	); err != nil {
		return sanitizeErr(err)
	}
	if _, err := c.PfsAPIClient.DeleteAll(
		c.ctx(),
		&types.Empty{},
	); err != nil {
		return sanitizeErr(err)
	}
	return nil
}

// SetMaxConcurrentStreams Sets the maximum number of concurrent streams the
// client can have. It is not safe to call this operations while operations are
// outstanding.
func (c APIClient) SetMaxConcurrentStreams(n int) {
	c.streamSemaphore = make(chan struct{}, n)
}

// EtcdDialOptions is a helper returning a slice of grpc.Dial options
// such that grpc.Dial() is synchronous: the call doesn't return until
// the connection has been established and it's safe to send RPCs
func EtcdDialOptions() []grpc.DialOption {
	return []grpc.DialOption{
		// Don't return from Dial() until the connection has been established
		grpc.WithBlock(),

		// If no connection is established in 10s, fail the call
		grpc.WithTimeout(10 * time.Second),

		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(grpcutil.MaxMsgSize),
			grpc.MaxCallSendMsgSize(grpcutil.MaxMsgSize),
		),
	}
}

// PachDialOptions is a helper returning a slice of grpc.Dial options
// such that
// - TLS is disabled
// - Dial is synchronous: the call doesn't return until the connection has been
//                        established and it's safe to send RPCs
//
// This is primarily useful for Pachd and Worker clients
func PachDialOptions() []grpc.DialOption {
	return append(EtcdDialOptions(), grpc.WithInsecure())
}

func (c *APIClient) connect() error {
	clientConn, err := grpc.Dial(c.addr, PachDialOptions()...)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	c.PfsAPIClient = pfs.NewAPIClient(clientConn)
	c.PpsAPIClient = pps.NewAPIClient(clientConn)
	c.ObjectAPIClient = pfs.NewObjectAPIClient(clientConn)
	c.clientConn = clientConn
	c.healthClient = health.NewHealthClient(clientConn)
	c._ctx = ctx
	c.cancel = cancel
	return nil
}

func (c *APIClient) addMetadata(ctx context.Context) context.Context {
	if c.metricsUserID == "" {
		return ctx
	}
	// metadata API downcases all the key names
	return metadata.NewContext(
		ctx,
		metadata.Pairs(
			"userid", c.metricsUserID,
			"prefix", c.metricsPrefix,
		),
	)
}

// TODO this method only exists because we initialize some APIClient in such a
// way that ctx will be nil
func (c *APIClient) ctx() context.Context {
	if c._ctx == nil {
		return c.addMetadata(context.Background())
	}
	return c.addMetadata(c._ctx)
}

func sanitizeErr(err error) error {
	if err == nil {
		return nil
	}

	return errors.New(grpc.ErrorDesc(err))
}
