package route

import (
	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/address"
	"github.com/pachyderm/pachyderm/src/pfs/dial"
)

type Router interface {
	IsLocalMasterShard(shard int) (bool, error)
	IsLocalSlaveShard(shard int) (bool, error)
	GetAPIClient(shard int) (pfs.ApiClient, error)
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
