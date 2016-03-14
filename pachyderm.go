package pachyderm

import (
	"fmt"
	"os"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pps"
	"go.pedge.io/proto/version"
	"google.golang.org/grpc"
)

const (
	// MajorVersion is the current major version for pachyderm.
	MajorVersion = 0
	// MinorVersion is the current minor version for pachyderm.
	MinorVersion = 10
	// MicroVersion is the current micro version for pachyderm.
	MicroVersion = 0
	// AdditionalVersion will be "dev" is this is a development branch, "" otherwise.
	AdditionalVersion = "RC1"
)

var (
	// Version is the current version for pachyderm.
	Version = &protoversion.Version{
		Major:      MajorVersion,
		Minor:      MinorVersion,
		Micro:      MicroVersion,
		Additional: AdditionalVersion,
	}
)

type PfsAPIClient pfs.APIClient
type PpsAPIClient pps.APIClient

type APIClient struct {
	PfsAPIClient
	PpsAPIClient
}

func NewAPIClientFromAddress(pachAddr string) (*APIClient, error) {
	clientConn, err := grpc.Dial(fmt.Sprintf("%s:650", pachAddr), grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return &APIClient{
		pfs.NewAPIClient(clientConn),
		pps.NewAPIClient(clientConn),
	}, nil
}

func NewAPIClient() (*APIClient, error) {
	pachAddr := os.Getenv("PACHD_PORT_650_TCP_ADDR")
	if pachAddr == "" {
		return nil, fmt.Errorf("PACHD_PORT_650_TCP_ADDR not set")
	}

	return NewAPIClientFromAddress(pachAddr)
}
