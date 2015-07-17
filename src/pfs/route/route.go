package route

import (
	"github.com/pachyderm/pachyderm/src/pfs/address"
	"github.com/pachyderm/pachyderm/src/pfs/dial"
	"google.golang.org/grpc"
)

type Router interface {
	IsLocalMasterShard(shard int) (bool, error)
	IsLocalSlaveShard(shard int) (bool, error)
	GetMasterClientConn(shard int) (*grpc.ClientConn, error)
	GetMasterOrSlaveClientConn(shard int) (*grpc.ClientConn, error)
	GetAllClientConns() ([]*grpc.ClientConn, error)
}

func NewRouter(
	addresser address.Addresser,
	dialer dial.Dialer,
	localAddress string,
) Router {
	return newRouter(
		addresser,
		dialer,
		localAddress,
	)
}
