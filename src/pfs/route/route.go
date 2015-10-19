package route //import "go.pachyderm.com/pachyderm/src/pfs/route"

import (
	"go.pachyderm.com/pachyderm/src/pfs"
	"go.pachyderm.com/pachyderm/src/pkg/grpcutil"
	"go.pachyderm.com/pachyderm/src/pkg/shard"
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
	sharder shard.Sharder,
	dialer grpcutil.Dialer,
	localAddress string,
) Router {
	return newRouter(
		sharder,
		dialer,
		localAddress,
	)
}
