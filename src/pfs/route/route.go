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
// namespace/pfs/shard/num/replica/address -> true

type Addresser interface {
	// TODO consider splitting Addresser's interface into read an write methods.
	// Each user of Addresser seems to use only one of these interfaces.
	GetMasterAddress(shard int) (string, bool, error)
	GetReplicaAddresses(shard int) (map[string]bool, error)
	GetShardToMasterAddress() (map[int]string, error)
	WatchShardToMasterAddress(chan bool, func(map[int]string) error) error
	GetShardToReplicaAddresses() (map[int]map[string]bool, error)
	SetMasterAddress(shard int, address string, ttl uint64) error
	HoldMasterAddress(shard int, address string, prevAddress string, cancel chan bool) error
	SetReplicaAddress(shard int, address string, ttl uint64) error
	HoldReplicaAddress(shard int, address string, prevAddress string, cancel chan bool) error
	DeleteMasterAddress(shard int) error
	DeleteReplicaAddress(shard int, address string) error
}

func NewDiscoveryAddresser(discoveryClient discovery.Client, namespace string) Addresser {
	return newDiscoveryAddresser(discoveryClient, namespace)
}

type Router interface {
	GetMasterShards() (map[int]bool, error)
	GetReplicaShards() (map[int]bool, error)
	GetMasterClientConn(shard int) (*grpc.ClientConn, error)
	GetMasterOrReplicaClientConn(shard int) (*grpc.ClientConn, error)
	GetReplicaClientConns(shard int) ([]*grpc.ClientConn, error)
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
