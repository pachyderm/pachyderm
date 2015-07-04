package storage

import (
	"net/http"

	"github.com/pachyderm/pachyderm/src/etcache"
)

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

func NewShardHTTPHandler(shard Shard) http.Handler {
	return newShardHTTPHandler(shard)
}
