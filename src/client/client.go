package client

import (
	"errors"
	"fmt"
	"os"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	log "github.com/Sirupsen/logrus"
	types "github.com/gogo/protobuf/types"

	"github.com/pachyderm/pachyderm/src/client/health"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pkg/config"
	"github.com/pachyderm/pachyderm/src/client/pps"
)

// PfsAPIClient is an alias for pfs.APIClient.
type PfsAPIClient pfs.APIClient

// PpsAPIClient is an alias for pps.APIClient.
type PpsAPIClient pps.APIClient

// BlockAPIClient is an alias for pfs.BlockAPIClient.
type BlockAPIClient pfs.BlockAPIClient

// ObjectAPIClient is an alias for pfs.ObjectAPIClient
type ObjectAPIClient pfs.ObjectAPIClient

// An APIClient is a wrapper around pfs, pps and block APIClients.
type APIClient struct {
	PfsAPIClient
	PpsAPIClient
	BlockAPIClient
	ObjectAPIClient
	addr              string
	clientConn        *grpc.ClientConn
	healthClient      health.HealthClient
	_ctx              context.Context
	config            *config.Config
	cancel            func()
	reportUserMetrics bool
	metricsPrefix     string
	streamSemaphore   chan struct{}
}

// DefaultMaxConcurrentStreams defines the max number of Putfiles or Getfiles happening simultaneously
const DefaultMaxConcurrentStreams uint = 100

var (
	// MaxMsgSize is used to define the GRPC frame size
	MaxMsgSize = 10 * 1024 * 1024
)

// NewMetricsClientFromAddress Creates a client that will report a user's Metrics
func NewMetricsClientFromAddress(addr string, metrics bool, prefix string) (*APIClient, error) {
	return NewMetricsClientFromAddressWithConcurrency(addr, metrics, prefix,
		DefaultMaxConcurrentStreams)
}

// NewMetricsClientFromAddressWithConcurrency Creates a client that will report
// a user's Metrics, and sets the max concurrency of streaming requests (GetFile
// / PutFile)
func NewMetricsClientFromAddressWithConcurrency(addr string, metrics bool, prefix string, maxConcurrentStreams uint) (*APIClient, error) {
	c, err := NewFromAddress(addr)
	if err != nil {
		return nil, err
	}
	cfg, err := config.Read()
	if err != nil {
		// metrics errors are non fatal
		log.Errorf("error loading user config from ~/.pachderm/config: %v\n", err)
	} else {
		c.config = cfg
	}
	c.reportUserMetrics = metrics
	c.metricsPrefix = prefix
	return c, err
}

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

// NewInCluster constructs a new APIClient using env vars that Kubernetes creates.
// This should be used to access Pachyderm from within a Kubernetes cluster
// with Pachyderm running on it.
func NewInCluster() (*APIClient, error) {
	addr := os.Getenv("PACHD_PORT_650_TCP_ADDR")

	if addr == "" {
		return nil, fmt.Errorf("PACHD_PORT_650_TCP_ADDR not set")
	}

	return NewFromAddress(fmt.Sprintf("%v:650", addr))
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

func (c *APIClient) connect() error {
	clientConn, err := grpc.Dial(c.addr, grpc.WithInsecure())
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	c.PfsAPIClient = pfs.NewAPIClient(clientConn)
	c.PpsAPIClient = pps.NewAPIClient(clientConn)
	c.BlockAPIClient = pfs.NewBlockAPIClient(clientConn)
	c.ObjectAPIClient = pfs.NewObjectAPIClient(clientConn)
	c.clientConn = clientConn
	c.healthClient = health.NewHealthClient(clientConn)
	c._ctx = ctx
	c.cancel = cancel
	return nil
}

func (c *APIClient) addMetadata(ctx context.Context) context.Context {
	if !c.reportUserMetrics {
		return ctx
	}
	if c.config == nil {
		cfg, err := config.Read()
		if err != nil {
			// Don't report error if config fails to read
			// metrics errors are non fatal
			log.Errorf("Error loading config: %v\n", err)
			return ctx
		}
		c.config = cfg
	}
	// metadata API downcases all the key names
	return metadata.NewContext(
		ctx,
		metadata.Pairs(
			"userid", c.config.UserID,
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
