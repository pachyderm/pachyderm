package route

import (
	"fmt"

	"github.com/pachyderm/pachyderm/src/pkg/grpcutil"
	"google.golang.org/grpc"
)

type router struct {
	addresser    Addresser
	dialer       grpcutil.Dialer
	localAddress string
}

func newRouter(
	addresser Addresser,
	dialer grpcutil.Dialer,
	localAddress string,
) *router {
	return &router{
		addresser,
		dialer,
		localAddress,
	}
}

func (r *router) GetMasterShards() (map[int]bool, error) {
	shardToMasterAddress, err := r.addresser.GetShardToMasterAddress()
	if err != nil {
		return nil, err
	}
	m := make(map[int]bool, 0)
	for shard, address := range shardToMasterAddress {
		if address.Address == r.localAddress && !address.Backfilling {
			m[shard] = true
		}
	}
	return m, nil
}

func (r *router) GetReplicaShards() (map[int]bool, error) {
	shardToReplicaAddresses, err := r.addresser.GetShardToReplicaAddresses()
	if err != nil {
		return nil, err
	}
	m := make(map[int]bool, 0)
	for shard, addresses := range shardToReplicaAddresses {
		for _, address := range addresses {
			if address.Address == r.localAddress && !address.Backfilling {
				m[shard] = true
			}
		}
	}
	return m, nil
}

func (r *router) GetAllShards() (map[int]bool, error) {
	shardToMasterAddress, err := r.addresser.GetShardToMasterAddress()
	if err != nil {
		return nil, err
	}
	shardToReplicaAddresses, err := r.addresser.GetShardToReplicaAddresses()
	if err != nil {
		return nil, err
	}
	m := make(map[int]bool, 0)
	for shard, address := range shardToMasterAddress {
		if address.Address == r.localAddress && !address.Backfilling {
			m[shard] = true
		}
	}
	for shard, addresses := range shardToReplicaAddresses {
		for _, address := range addresses {
			if address.Address == r.localAddress && !address.Backfilling {
				m[shard] = true
			}
		}
	}
	return m, nil
}

func (r *router) GetMasterClientConn(shard int) (*grpc.ClientConn, error) {
	address, ok, err := r.addresser.GetMasterAddress(shard)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("no master found for %d", shard)
	}
	if address.Backfilling {
		return nil, fmt.Errorf("master %s for shard %d is backfilling", address.Address, shard)
	}
	return r.dialer.Dial(address.Address)
}

func (r *router) GetMasterOrReplicaClientConn(shard int) (*grpc.ClientConn, error) {
	addresses, err := r.addresser.GetReplicaAddresses(shard)
	if err != nil {
		return nil, err
	}
	for address := range addresses {
		if !address.Backfilling {
			return r.dialer.Dial(address.Address)
		}
	}
	return r.GetMasterClientConn(shard)
}

func (r *router) GetReplicaClientConns(shard int) ([]*grpc.ClientConn, error) {
	addresses, err := r.addresser.GetReplicaAddresses(shard)
	if err != nil {
		return nil, err
	}
	var result []*grpc.ClientConn
	for address := range addresses {
		if !address.Backfilling {
			conn, err := r.dialer.Dial(address.Address)
			if err != nil {
				return nil, err
			}
			result = append(result, conn)
		}
	}
	return result, nil
}

func (r *router) GetAllClientConns() ([]*grpc.ClientConn, error) {
	addresses, err := r.getAllAddresses()
	if err != nil {
		return nil, err
	}
	var clientConns []*grpc.ClientConn
	for address := range addresses {
		// TODO(pedge): huge race, this whole thing is bad
		if address.Address != r.localAddress && !address.Backfilling {
			clientConn, err := r.dialer.Dial(address.Address)
			if err != nil {
				return nil, err
			}
			clientConns = append(clientConns, clientConn)
		}
	}
	return clientConns, nil
}

func (r *router) getAllAddresses() (map[Address]bool, error) {
	m := make(map[Address]bool, 0)
	shardToMasterAddress, err := r.addresser.GetShardToMasterAddress()
	if err != nil {
		return nil, err
	}
	for _, address := range shardToMasterAddress {
		m[address] = true
	}
	shardToReplicaAddresses, err := r.addresser.GetShardToReplicaAddresses()
	if err != nil {
		return nil, err
	}
	for _, addresses := range shardToReplicaAddresses {
		for _, address := range addresses {
			m[address] = true
		}
	}
	return m, nil
}
