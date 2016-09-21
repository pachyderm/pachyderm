package client

import (
	"errors"
	"fmt"
	"os"
	"time"

	"golang.org/x/net/context"

	"google.golang.org/grpc"

	"github.com/pachyderm/pachyderm/src/client/health"
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pps"

	google_protobuf "go.pedge.io/pb/go/google/protobuf"
)

// PfsAPIClient is an alias for pfs.APIClient.
type PfsAPIClient pfs.APIClient

// PpsAPIClient is an alias for pps.APIClient.
type PpsAPIClient pps.APIClient

// BlockAPIClient is an alias for pfs.BlockAPIClient.
type BlockAPIClient pfs.BlockAPIClient

// An APIClient is a wrapper around pfs, pps and block APIClients.
type APIClient struct {
	PfsAPIClient
	PpsAPIClient
	BlockAPIClient
	addr         string
	clientConn   *grpc.ClientConn
	healthClient health.HealthClient
	_ctx         context.Context
	cancel       func()
}

// NewFromAddress constructs a new APIClient for the server at addr.
func NewFromAddress(addr string) (*APIClient, error) {
	c := &APIClient{
		addr: addr,
	}
	if err := c.connect(); err != nil {
		return nil, err
	}
	return c, nil
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
		default:
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()
			if _, err := c.healthClient.Health(ctx, google_protobuf.EmptyInstance); err != nil {
				c.cancel()
				c.connect()
			}
		}
	}
}

// DeleteAll deletes everything in the cluster.
// Use with caution, there is no undo.
func (c APIClient) DeleteAll() error {
	if _, err := c.PpsAPIClient.DeleteAll(
		c.ctx(),
		google_protobuf.EmptyInstance,
	); err != nil {
		return sanitizeErr(err)
	}
	if _, err := c.PfsAPIClient.DeleteAll(
		c.ctx(),
		google_protobuf.EmptyInstance,
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
	c.clientConn = clientConn
	c.healthClient = health.NewHealthClient(clientConn)
	c._ctx = ctx
	c.cancel = cancel
	return nil
}

// TODO this method only exists because we initialize some APIClient in such a
// way that ctx will be nil
func (c *APIClient) ctx() context.Context {
	if c._ctx == nil {
		return context.Background()
	}
	return c._ctx
}

func sanitizeErr(err error) error {
	if err == nil {
		return nil
	}

	return errors.New(grpc.ErrorDesc(err))
}
