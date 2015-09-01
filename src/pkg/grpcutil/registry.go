package grpcutil

import (
	"fmt"

	"github.com/pachyderm/pachyderm/src/pkg/discovery"
	"google.golang.org/grpc"
)

type registry struct {
	discoveryRegistry discovery.Registry
	dialer            Dialer
}

func newRegistry(discoveryRegistry discovery.Registry, dialer Dialer) *registry {
	return &registry{discoveryRegistry, dialer}
}

func (r *registry) RegisterAddress(address string) <-chan error {
	return r.discoveryRegistry.Register(address)
}

func (r *registry) GetClientConn() (*grpc.ClientConn, error) {
	addresses, err := r.discoveryRegistry.GetAll()
	if err != nil {
		return nil, err
	}
	var errs []error
	for _, address := range addresses {
		clientConn, err := r.dialer.Dial(address)
		if err == nil {
			return clientConn, nil
		}
		errs = append(errs, err)
	}
	return nil, fmt.Errorf("%v", errs)
}
