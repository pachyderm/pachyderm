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
		if address == r.localAddress {
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
		if _, ok := addresses[r.localAddress]; ok {
			m[shard] = true
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
	return r.dialer.Dial(address)
}

func (r *router) GetMasterOrReplicaClientConn(shard int) (*grpc.ClientConn, error) {
	addresses, err := r.addresser.GetReplicaAddresses(shard)
	if err != nil {
		return nil, err
	}
	for address := range addresses {
		return r.dialer.Dial(address)
	}
	address, ok, err := r.addresser.GetMasterAddress(shard)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("no master or replica found for %d", shard)
	}
	return r.dialer.Dial(address)
}

func (r *router) GetReplicaClientConns(shard int) ([]*grpc.ClientConn, error) {
	addresses, err := r.addresser.GetReplicaAddresses(shard)
	if err != nil {
		return nil, err
	}
	var result []*grpc.ClientConn
	for address := range addresses {
		conn, err := r.dialer.Dial(address)
		if err != nil {
			return nil, err
		}
		result = append(result, conn)
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
		if address != r.localAddress {
			clientConn, err := r.dialer.Dial(address)
			if err != nil {
				return nil, err
			}
			clientConns = append(clientConns, clientConn)
		}
	}
	return clientConns, nil
}

func (r *router) getAllAddresses() (map[string]bool, error) {
	m := make(map[string]bool, 0)
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
		for address := range addresses {
			m[address] = true
		}
	}
	return m, nil
}
