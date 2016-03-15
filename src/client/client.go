package client

import(
	"fmt"
	"os"

	"google.golang.org/grpc"	

	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pps"
)


type PfsAPIClient pfs.APIClient
type PpsAPIClient pps.APIClient

type APIClient struct {
	PfsAPIClient
	PpsAPIClient
}

func NewFromAddress(pachAddr string) (*APIClient, error) {
	clientConn, err := grpc.Dial(pachAddr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return &APIClient{
		pfs.NewAPIClient(clientConn),
		pps.NewAPIClient(clientConn),
	}, nil
}

func New() (*APIClient, error) {
	pachAddr := os.Getenv("PACHD_PORT_650_TCP_ADDR")

	if pachAddr == "" {
		return nil, fmt.Errorf("PACHD_PORT_650_TCP_ADDR not set")
	}

	return NewFromAddress(pachAddr)
}

