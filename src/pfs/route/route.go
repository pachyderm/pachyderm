package route

import (
	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pkg/discovery"
	"github.com/pachyderm/pachyderm/src/pkg/grpcutil"
	"google.golang.org/grpc"
)

type Sharder interface {
	NumShards() int
	GetShard(path *pfs.Path) (int, error)
}

func NewSharder(numShards int) Sharder {
	return newSharder(numShards)
}

// namespace/pfs/shard/num/master -> address
// namespace/pfs/shard/num/slave/address -> true

type Addresser interface {
	GetMasterAddress(shard int) (string, bool, error)
	GetSlaveAddresses(shard int) (map[string]bool, error)
	GetShardToMasterAddress() (map[int]string, error)
	GetShardToSlaveAddresses() (map[int]map[string]bool, error)
	SetMasterAddress(shard int, address string, ttl uint64) error
	SetSlaveAddress(shard int, address string, ttl uint64) error
	DeleteMasterAddress(shard int) error
	DeleteSlaveAddress(shard int, address string) error
}

func NewDiscoveryAddresser(discoveryClient discovery.Client, namespace string) Addresser {
	return newDiscoveryAddresser(discoveryClient, namespace)
}

type Router interface {
	GetMasterShards() (map[int]bool, error)
	GetSlaveShards() (map[int]bool, error)
	GetMasterClientConn(shard int) (*grpc.ClientConn, error)
	GetMasterOrSlaveClientConn(shard int) (*grpc.ClientConn, error)
	GetSlaveClientConns(shard int) ([]*grpc.ClientConn, error)
	GetAllClientConns() ([]*grpc.ClientConn, error)
}

func NewRouter(
	addresser Addresser,
	dialer grpcutil.Dialer,
	localAddress string,
) Router {
	return newRouter(
		addresser,
		dialer,
		localAddress,
	)
}
