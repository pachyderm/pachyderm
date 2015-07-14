package shard

import "github.com/pachyderm/pachyderm/src/pfs"

type Sharder interface {
	NumShards() int
	GetShard(path *pfs.Path) (int, error)
}

func NewSharder(numShards int) Sharder {
	return nil
}
