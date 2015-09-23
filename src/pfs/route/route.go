package route

import (
	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pkg/discovery"
	"github.com/pachyderm/pachyderm/src/pkg/grpcutil"
	"google.golang.org/grpc"
)

type Sharder interface {
	NumShards() int
	NumReplicas() int
	GetShard(file *pfs.File) (int, error)
}

func NewSharder(numShards int, numReplicas int) Sharder {
	return newSharder(numShards, numReplicas)
}

type Addresser interface {
	GetMasterAddress(shard uint64, version int64) (string, bool, error)
	GetReplicaAddresses(shard uint64, version int64) (map[string]bool, error)
	GetShardToMasterAddress(version int64) (map[uint64]string, error)
	GetShardToReplicaAddresses(version int64) (map[uint64]map[string]bool, error)

	Register(cancel chan bool, id string, address string, server Server) error
	AssignRoles(chan bool) error
	Version() (int64, error)
	WaitOneVersion() error
}

func NewDiscoveryAddresser(discoveryClient discovery.Client, sharder Sharder, namespace string) Addresser {
	return newDiscoveryAddresser(discoveryClient, sharder, namespace)
}

type Server interface {
	// AddShard tells the server it now has a role for a shard.
	AddShard(shard uint64) error
	// RemoveShard tells the server it no longer has a role for a shard.
	RemoveShard(shard uint64) error
	// LocalRoles asks the server which shards it has on disk and how many commits each shard has.
	LocalShards() (map[uint64]bool, error)
}

// Announcer announces a server to the outside world.
type Announcer interface {
	Announce(address string, server Server) error
}

type Router interface {
	GetMasterShards(version int64) (map[uint64]bool, error)
	GetReplicaShards(version int64) (map[uint64]bool, error)
	GetAllShards(version int64) (map[uint64]bool, error)
	GetMasterClientConn(shard uint64, version int64) (*grpc.ClientConn, error)
	GetMasterOrReplicaClientConn(shard uint64, version int64) (*grpc.ClientConn, error)
	GetReplicaClientConns(shard uint64, version int64) ([]*grpc.ClientConn, error)
	GetAllClientConns(version int64) ([]*grpc.ClientConn, error)
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
