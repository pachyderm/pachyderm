package route

import (
	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pkg/discovery"
	"github.com/pachyderm/pachyderm/src/pkg/grpcutil"
	"google.golang.org/grpc"
)

type Sharder interface {
	NumShards() int
	GetShard(file *pfs.File) (int, error)
}

func NewSharder(numShards int) Sharder {
	return newSharder(numShards)
}

// namespace/pfs/shard/num/master -> address
// namespace/pfs/shard/num/replica/address -> true

type Address struct {
	Address     string
	Backfilling bool
}

type Addresser interface {
	// TODO consider splitting Addresser's interface into read an write methods.
	// Each user of Addresser seems to use only one of these interfaces.
	GetMasterAddress(shard int, version string) (Address, bool, error)
	GetReplicaAddresses(shard int, version string) (map[Address]bool, error)
	GetShardToMasterAddress(version string) (map[int]Address, error)
	WatchShardToAddress(chan bool, func(map[int]Address, map[int]map[int]Address) (uint64, error)) error
	GetShardToReplicaAddresses() (map[int]map[int]Address, error)
	SetMasterAddress(shard int, address Address) (uint64, error)
	ClaimMasterAddress(shard int, address Address, prevAddress Address) (uint64, error)
	HoldMasterAddress(shard int, address Address, cancel chan bool) error
	SetReplicaAddress(shard int, index int, address Address) (uint64, error)
	ClaimReplicaAddress(shard int, index int, address Address, prevAddress Address) (uint64, error)
	HoldReplicaAddress(shard int, index int, address Address, cancel chan bool) error
	DeleteMasterAddress(shard int) (uint64, error)
	DeleteReplicaAddress(shard int, index int, address Address) (uint64, error)

	Announce(cancel chan bool, address string, server Server) error
	WatchServers(chan bool, func(map[string]ServerInfo) error) error
	Version() (string, error)
	Set(layout Layout) error
}

func NewDiscoveryAddresser(discoveryClient discovery.Client, namespace string) Addresser {
	return newDiscoveryAddresser(discoveryClient, namespace)
}

type Server interface {
	// Master tells the server that it is now the master for shard.
	// After this returns the Peer is expected to service Master requests for shard.
	Master(shard int) error
	// Replica tells the server that it is now a replica for shard.
	// After this returns the Server is expected to service Replica requests for shard.
	Replica(shard int) error
	// Clear tells the server that it is no longer filling any role for shard.
	Clear(shard int) error
}

// Announcer announces a server to the outside world.
type Announcer interface {
	Announce(address string, server Server) error
}

type Router interface {
	GetMasterShards() (map[int]bool, error)
	GetReplicaShards() (map[int]bool, error)
	GetAllShards() (map[int]bool, error)
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
