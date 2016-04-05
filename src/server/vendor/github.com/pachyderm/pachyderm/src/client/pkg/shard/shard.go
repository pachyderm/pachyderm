package shard

import (
	"github.com/pachyderm/pachyderm/src/client/pkg/discovery"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"google.golang.org/grpc"
)

// Sharder distributes shards between a set of servers.
type Sharder interface {
	GetAddress(shard uint64, version int64) (string, bool, error)
	GetShardToAddress(version int64) (map[uint64]string, error)

	Register(cancel chan bool, address string, servers []Server) error
	RegisterFrontends(cancel chan bool, address string, frontends []Frontend) error
	AssignRoles(address string, cancel chan bool) error
}

type TestSharder interface {
	Sharder
	WaitForAvailability(frontendIds []string, serverIds []string) error
}

func NewSharder(discoveryClient discovery.Client, numShards uint64, namespace string) Sharder {
	return newSharder(discoveryClient, numShards, namespace)
}

func NewTestSharder(discoveryClient discovery.Client, numShards uint64, namespace string) TestSharder {
	return newSharder(discoveryClient, numShards, namespace)
}

func NewLocalSharder(addresses []string, numShards uint64) Sharder {
	return newLocalSharder(addresses, numShards)
}

type Server interface {
	// AddShard tells the server it now has a role for a shard.
	AddShard(shard uint64) error
	// RemoveShard tells the server it no longer has a role for a shard.
	DeleteShard(shard uint64) error
}

type Frontend interface {
	// Version tells the Frontend a new version exists.
	// Version should block until the Frontend is done using the previous version.
	Version(version int64) error
}

type Router interface {
	GetShards(version int64) (map[uint64]bool, error)
	GetClientConn(shard uint64, version int64) (*grpc.ClientConn, error)
	GetAllClientConns(version int64) ([]*grpc.ClientConn, error)
}

func NewRouter(
	sharder Sharder,
	dialer grpcutil.Dialer,
	localAddress string,
) Router {
	return newRouter(
		sharder,
		dialer,
		localAddress,
	)
}
