package route

import (
	"fmt"
	"math/rand"

	"google.golang.org/grpc"
)

type router struct {
	addresser    Addresser
	dialer       Dialer
	localAddress string
}

func newRouter(
	addresser Addresser,
	dialer Dialer,
	localAddress string,
) *router {
	return &router{
		addresser,
		dialer,
		localAddress,
	}
}

func (r *router) GetMasterShards() (map[int]bool, error) {
	return r.addresser.GetMasterShards(r.localAddress)
}

func (r *router) GetSlaveShards() (map[int]bool, error) {
	return r.addresser.GetSlaveShards(r.localAddress)
}

func (r *router) GetMasterClientConn(shard int) (*grpc.ClientConn, error) {
	address, err := r.getAddress(shard, r.addresser.GetMasterShards)
	if err != nil {
		return nil, err
	}
	if address == "" {
		return nil, fmt.Errorf("no master found for %d", shard)
	}
	return r.dialer.Dial(address)
}

func (r *router) GetMasterOrSlaveClientConn(shard int) (*grpc.ClientConn, error) {
	address, err := r.getAddress(shard, r.addresser.GetSlaveShards)
	if err != nil {
		return nil, err
	}
	if address == "" {
		address, err = r.getAddress(shard, r.addresser.GetMasterShards)
		if err != nil {
			return nil, err
		}
		if address == "" {
			return nil, fmt.Errorf("no slave or master found for %d", shard)
		}
	}
	return r.dialer.Dial(address)
}

func (r *router) getAddress(shard int, testFunc func(string) (map[int]bool, error)) (string, error) {
	addresses, err := r.addresser.GetAllAddresses()
	if err != nil {
		return "", err
	}
	var foundAddresses []string
	for _, address := range addresses {
		shards, err := testFunc(address)
		if err != nil {
			return "", err
		}
		if _, ok := shards[shard]; ok {
			foundAddresses = append(foundAddresses, address)
		}
	}
	if len(foundAddresses) == 0 {
		return "", nil
	}
	return foundAddresses[int(rand.Uint32())%len(foundAddresses)], nil
}

func (r *router) GetAllClientConns() ([]*grpc.ClientConn, error) {
	addresses, err := r.addresser.GetAllAddresses()
	if err != nil {
		return nil, err
	}
	clientConns := make([]*grpc.ClientConn, len(addresses)-1)
	for i, address := range addresses {
		// TODO(pedge): huge race, this whole thing is bad
		if address != r.localAddress {
			clientConn, err := r.dialer.Dial(address)
			if err != nil {
				return nil, err
			}
			clientConns[i] = clientConn
		}
	}
	return clientConns, nil
}
