package client

import (
	"fmt"
	"os"
	"strconv"

	"go.pedge.io/proto/version"
	"google.golang.org/grpc"

	"github.com/pachyderm/pachyderm/src/client/pfs"
	"github.com/pachyderm/pachyderm/src/client/pps"
)

const (
	// MajorVersion is the current major version for pachyderm.
	MajorVersion = 1
	// MinorVersion is the current minor version for pachyderm.
	MinorVersion = 0
	// AdditionalVersion will be "dev" is this is a development branch, "" otherwise.
	AdditionalVersion = ""
)

var (
	// Version is the current version for pachyderm.
	Version = &protoversion.Version{
		Major:      MajorVersion,
		Minor:      MinorVersion,
		Micro:      getMicroVersion(),
		Additional: AdditionalVersion,
	}
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

	return NewFromAddress(fmt.Sprintf("%v:650", pachAddr))
}

func getMicroVersion() (v uint32) {
	value := os.Getenv("BUILD_NUMBER")
	if value == "" {
		v = 0
	} else {
		number, err := strconv.Atoi(value)
		if err != nil {
			panic(fmt.Sprintf("Invalid build number provided via BUILD_NUMBER env variable: (%v)\n", value))
		}
		v = uint32(number)
	}
	return v
}
