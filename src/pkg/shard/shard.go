package shard

import (
	"go.pachyderm.com/pachyderm/src/pkg/discovery"
)

// Sharder distributes shards between a set of servers.
type Sharder interface {
	GetMasterAddress(shard uint64, version int64) (string, bool, error)
	GetReplicaAddresses(shard uint64, version int64) (map[string]bool, error)
	GetShardToMasterAddress(version int64) (map[uint64]string, error)
	GetShardToReplicaAddresses(version int64) (map[uint64]map[string]bool, error)

	Register(cancel chan bool, id string, address string, server Server) error
	RegisterFrontend(cancel chan bool, address string, frontend Frontend) error
	AssignRoles(chan bool) error
}

type TestSharder interface {
	Sharder
	WaitForAvailability(frontendIds []string, serverIds []string) error
}

func NewSharder(discoveryClient discovery.Client, numShards uint64, numReplicas uint64, namespace string) Sharder {
	return newSharder(discoveryClient, numShards, numReplicas, namespace)
}

func NewTestSharder(discoveryClient discovery.Client, numShards uint64, numReplicas uint64, namespace string) TestSharder {
	return newSharder(discoveryClient, numShards, numReplicas, namespace)
}

type Server interface {
	// AddShard tells the server it now has a role for a shard.
	AddShard(shard uint64, version int64) error
	// RemoveShard tells the server it no longer has a role for a shard.
	RemoveShard(shard uint64, version int64) error
	// LocalRoles asks the server which shards it has on disk and how many commits each shard has.
	LocalShards() (map[uint64]bool, error)
}

type Frontend interface {
	// Version tells the Frontend a new version exists.
	// Version should block until the Frontend is done using the previous version.
	Version(version int64) error
}
