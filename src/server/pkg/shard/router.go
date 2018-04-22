package shard

import (
	"fmt"

	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"google.golang.org/grpc"
)

type router struct {
	sharder      Sharder
	dialer       grpcutil.Dialer
	localAddress string
}

func newRouter(
	sharder Sharder,
	dialer grpcutil.Dialer,
	localAddress string,
) *router {
	return &router{
		sharder,
		dialer,
		localAddress,
	}
}

func (r *router) GetShards(version int64) (map[uint64]bool, error) {
	shardToAddress, err := r.sharder.GetShardToAddress(version)
	if err != nil {
		return nil, err
	}
	result := make(map[uint64]bool)
	for shard, address := range shardToAddress {
		if address == r.localAddress {
			result[shard] = true
		}
	}
	return result, nil
}

func (r *router) GetAllShards(version int64) (map[uint64]bool, error) {
	shardToAddress, err := r.sharder.GetShardToAddress(version)
	if err != nil {
		return nil, err
	}
	result := make(map[uint64]bool)
	for shard, address := range shardToAddress {
		if address == r.localAddress {
			result[shard] = true
		}
	}
	return result, nil
}

func (r *router) GetClientConn(shard uint64, version int64) (*grpc.ClientConn, error) {
	address, ok, err := r.sharder.GetAddress(shard, version)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("no master found for %d", shard)
	}
	return r.dialer.Dial(address)
}

func (r *router) GetAllClientConns(version int64) ([]*grpc.ClientConn, error) {
	addresses, err := r.getAllAddresses(version)
	if err != nil {
		return nil, err
	}
	var clientConns []*grpc.ClientConn
	for address := range addresses {
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
	shardToAddress, err := r.sharder.GetShardToAddress(version)
	if err != nil {
		return nil, err
	}
	for _, address := range shardToAddress {
		result[address] = true
	}
	return result, nil
}

func (r *router) CloseClientConns() error {
	return r.dialer.CloseConns()
}
