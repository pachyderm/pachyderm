package client

import (
	"errors"
	"fmt"
	"os"

	"google.golang.org/grpc"

	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pps"
)

type PfsAPIClient pfs.APIClient
type PpsAPIClient pps.APIClient
type BlockAPIClient pfs.BlockAPIClient

type APIClient struct {
	PfsAPIClient
	PpsAPIClient
	BlockAPIClient
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

func sanitizeErr(err error) error {
	if err == nil {
		return nil
	}

	return errors.New(grpc.ErrorDesc(err))
}
