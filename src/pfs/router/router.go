package router

import (
	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/addresser"
	"github.com/pachyderm/pachyderm/src/pfs/dialer"
	"github.com/pachyderm/pachyderm/src/pfs/sharder"
)

type Router interface {
	LocalShards() map[int]bool
	GetAPIClient(shard int) (pfs.ApiClient, error)
}

func NewRouter(
	sharder sharder.Sharder,
	addresser addresser.Addresser,
	dialer dialer.Dialer,
	localShards map[int]bool,
) Router {
	return nil
}
