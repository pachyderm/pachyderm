package route

import (
	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/address"
	"github.com/pachyderm/pachyderm/src/pfs/dial"
	"github.com/pachyderm/pachyderm/src/pfs/shard"
)

type Router interface {
	IsLocalShard(shard int) (bool, error)
	GetAPIClient(shard int) (pfs.ApiClient, error)
}

func NewRouter(
	sharder shard.Sharder,
	addresser address.Addresser,
	dialer dial.Dialer,
	localAddress string,
) Router {
	return nil
}
