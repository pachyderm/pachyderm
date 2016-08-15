package client

import (
	"errors"
	"fmt"
	"os"

	"golang.org/x/net/context"

	"google.golang.org/grpc"

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
	GRPCConn *grpc.ClientConn
}

// NewFromAddress constructs a new APIClient for the server at pachAddr.
func NewFromAddress(pachAddr string) (*APIClient, error) {
	clientConn, err := grpc.Dial(pachAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return &APIClient{
		pfs.NewAPIClient(clientConn),
		pps.NewAPIClient(clientConn),
		pfs.NewBlockAPIClient(clientConn),
		clientConn,
	}, nil
}

// NewInCluster constructs a new APIClient using env vars that Kubernetes creates.
// This should be used to access Pachyderm from within a Kubernetes cluster
// with Pachyderm running on it.
func NewInCluster() (*APIClient, error) {
	pachAddr := os.Getenv("PACHD_PORT_650_TCP_ADDR")

	if pachAddr == "" {
		return nil, fmt.Errorf("PACHD_PORT_650_TCP_ADDR not set")
	}

	return NewFromAddress(fmt.Sprintf("%v:650", pachAddr))
}

// Close the connection to gRPC
func (c *APIClient) Close() error {
	return c.GRPCConn.Close()
}

// DeleteAll deletes everything in the cluster.
// Use with caution, there is no undo.
func (c APIClient) DeleteAll() error {
	if _, err := c.PpsAPIClient.DeleteAll(
		context.Background(),
		google_protobuf.EmptyInstance,
	); err != nil {
		return sanitizeErr(err)
	}
	if _, err := c.PfsAPIClient.DeleteAll(
		context.Background(),
		google_protobuf.EmptyInstance,
	); err != nil {
		return sanitizeErr(err)
	}
	return nil
}

func sanitizeErr(err error) error {
	if err == nil {
		return nil
	}

	return errors.New(grpc.ErrorDesc(err))
}
