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

// NewSharder creates a Sharder using a discovery client.
func NewSharder(discoveryClient discovery.Client, numShards uint64, namespace string) Sharder {
	return newSharder(discoveryClient, numShards, namespace)
}

// NewLocalSharder creates a Sharder user a list of addresses.
func NewLocalSharder(addresses []string, numShards uint64) Sharder {
	return newLocalSharder(addresses, numShards)
}

// A Server represents a server that has roles for shards.
type Server interface {
	// AddShard tells the server it now has a role for a shard.
	AddShard(shard uint64) error
	// RemoveShard tells the server it no longer has a role for a shard.
	DeleteShard(shard uint64) error
}

// A Frontend represents a frontend which receives new versions.
type Frontend interface {
	// Version tells the Frontend a new version exists.
	// Version should block until the Frontend is done using the previous version.
	Version(version int64) error
}

// Router represents a router from shard id and version to grpc connections.
type Router interface {
	GetShards(version int64) (map[uint64]bool, error)
	GetClientConn(shard uint64, version int64) (*grpc.ClientConn, error)
	GetAllClientConns(version int64) ([]*grpc.ClientConn, error)
	CloseClientConns() error // close all outstanding client connections
}

// NewRouter creates a Router.
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
