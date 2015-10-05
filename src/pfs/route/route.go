package route //import "go.pachyderm.com/pachyderm/src/pfs/route"

import (
	"go.pachyderm.com/pachyderm/src/pfs"
	"go.pachyderm.com/pachyderm/src/pkg/discovery"
	"go.pachyderm.com/pachyderm/src/pkg/grpcutil"
	"google.golang.org/grpc"
)

type Sharder interface {
	NumShards() uint64
	NumReplicas() uint64
	GetBlock(value []byte) *pfs.Block
	GetShard(file *pfs.File) uint64
	GetBlockShard(block *pfs.Block) uint64
}

func NewSharder(numShards uint64, numReplicas uint64) Sharder {
	return newSharder(numShards, numReplicas)
}

type Addresser interface {
	GetMasterAddress(shard uint64, version int64) (string, bool, error)
	GetReplicaAddresses(shard uint64, version int64) (map[string]bool, error)
	GetShardToMasterAddress(version int64) (map[uint64]string, error)
	GetShardToReplicaAddresses(version int64) (map[uint64]map[string]bool, error)

	InspectServer(server *pfs.Server) (*pfs.ServerInfo, error)
	ListServer() ([]*pfs.ServerInfo, error)

	Register(cancel chan bool, id string, address string, server Server) error
	AssignRoles(chan bool) error
	Version() (int64, error)
}

type TestAddresser interface {
	Addresser
	WaitForAvailability(ids []string) error
}

func NewDiscoveryAddresser(discoveryClient discovery.Client, sharder Sharder, namespace string) Addresser {
	return newDiscoveryAddresser(discoveryClient, sharder, namespace)
}

func NewDiscoveryTestAddresser(discoveryClient discovery.Client, sharder Sharder, namespace string) TestAddresser {
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
	Version() (int64, error)
	InspectServer(server *pfs.Server) (*pfs.ServerInfo, error)
	ListServer() ([]*pfs.ServerInfo, error)
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
