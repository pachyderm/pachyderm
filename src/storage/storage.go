package storage

import (
	"net/http"

	"github.com/pachyderm/pachyderm/src/btrfs"
	"github.com/pachyderm/pachyderm/src/etcache"
)

// json structures that shard will return in response to requests.

type BranchMsg struct {
	Name   string `json:"name"`
	TStamp string `json:"tstamp"`
}

type CommitMsg struct {
	Name   string `json:"name"`
	TStamp string `json:"tstamp"`
}

type Shard interface {
	EnsureRepos() error
	Peers() ([]string, error)
	SyncFromPeers() error
	SyncToPeers() error
	FillRole(cancel chan bool) error
}

func NewShard(
	url string,
	dataRepo string,
	compRepo string,
	pipelinePrefix string,
	shardNum uint64,
	modulos uint64,
	cache etcache.Cache,
) Shard {
	return newShard(
		url,
		dataRepo,
		compRepo,
		pipelinePrefix,
		shardNum,
		modulos,
		cache,
	)
}

func NewShardReplica(url string) btrfs.Replica {
	return newShardReplica(url)
}

func NewShardHTTPHandler(shard Shard) http.Handler {
	return newShardHTTPHandler(shard)
}
