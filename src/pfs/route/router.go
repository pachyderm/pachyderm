package route

import (
	"fmt"

	"github.com/pachyderm/pachyderm/src/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/pkg/shard"
	"google.golang.org/grpc"
)

type router struct {
	sharder      shard.Sharder
	dialer       grpcutil.Dialer
	localAddress string
}

func newRouter(
	sharder shard.Sharder,
	dialer grpcutil.Dialer,
	localAddress string,
) *router {
	return &router{
		sharder,
		dialer,
		localAddress,
	}
}

func (r *router) GetMasterShards(version int64) (map[uint64]bool, error) {
	shardToMasterAddress, err := r.sharder.GetShardToMasterAddress(version)
	if err != nil {
		return nil, err
	}
	result := make(map[uint64]bool)
	for shard, address := range shardToMasterAddress {
		if address == r.localAddress {
			result[shard] = true
		}
	}
	return result, nil
}

func (r *router) GetReplicaShards(version int64) (map[uint64]bool, error) {
	shardToReplicaAddresses, err := r.sharder.GetShardToReplicaAddresses(version)
	if err != nil {
		return nil, err
	}
	result := make(map[uint64]bool)
	for shard, addresses := range shardToReplicaAddresses {
		for address := range addresses {
			if address == r.localAddress {
				result[shard] = true
			}
		}
	}
	return result, nil
}

func (r *router) GetAllShards(version int64) (map[uint64]bool, error) {
	shardToMasterAddress, err := r.sharder.GetShardToMasterAddress(version)
	if err != nil {
		return nil, err
	}
	shardToReplicaAddresses, err := r.sharder.GetShardToReplicaAddresses(version)
	if err != nil {
		return nil, err
	}
	result := make(map[uint64]bool)
	for shard, address := range shardToMasterAddress {
		if address == r.localAddress {
			result[shard] = true
		}
	}
	for shard, addresses := range shardToReplicaAddresses {
		for address := range addresses {
			if address == r.localAddress {
				result[shard] = true
			}
		}
	}
	return result, nil
}

func (r *router) GetMasterClientConn(shard uint64, version int64) (*grpc.ClientConn, error) {
	address, ok, err := r.sharder.GetMasterAddress(shard, version)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("no master found for %d", shard)
	}
	return r.dialer.Dial(address)
}

func (r *router) GetMasterOrReplicaClientConn(shard uint64, version int64) (*grpc.ClientConn, error) {
	addresses, err := r.sharder.GetReplicaAddresses(shard, version)
	if err != nil {
		return nil, err
	}
	for address := range addresses {
		return r.dialer.Dial(address)
	}
	return r.GetMasterClientConn(shard, version)
}

func (r *router) GetReplicaClientConns(shard uint64, version int64) ([]*grpc.ClientConn, error) {
	addresses, err := r.sharder.GetReplicaAddresses(shard, version)
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

func (r *router) GetAllClientConns(version int64) ([]*grpc.ClientConn, error) {
	addresses, err := r.getAllAddresses(version)
	if err != nil {
		return nil, err
	}
	var clientConns []*grpc.ClientConn
	for address := range addresses {
		// TODO: huge race, this whole thing is bad
		clientConn, err := r.dialer.Dial(address)
		if err != nil {
			return nil, err
		}
		clientConns = append(clientConns, clientConn)
	}
	return clientConns, nil
}

func (r *router) getAllAddresses(version int64) (map[string]bool, error) {
	result := make(map[string]bool)
	shardToMasterAddress, err := r.sharder.GetShardToMasterAddress(version)
	if err != nil {
		return nil, err
	}
	for _, address := range shardToMasterAddress {
		result[address] = true
	}
	shardToReplicaAddresses, err := r.sharder.GetShardToReplicaAddresses(version)
	if err != nil {
		return nil, err
	}
	for _, addresses := range shardToReplicaAddresses {
		for address := range addresses {
			result[address] = true
		}
	}
	return result, nil
}
