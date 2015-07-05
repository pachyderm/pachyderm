package storage

import (
	"io"
	"net/http"
	"os"
	"time"

	"github.com/pachyderm/pachyderm/src/btrfs"
	"github.com/pachyderm/pachyderm/src/etcache"
)

type File struct {
	Name    string
	ModTime time.Time
	File    *os.File
}

type Commit struct {
	Name    string
	ModTime time.Time
}

type Branch struct {
	Name    string
	ModTime time.Time
}

type Shard interface {
	Filesystem
	Replica
	FillRole(cancel chan bool) error
	EnsureRepos() error
}

type Filesystem interface {
	FileGet(name string, commit string) (File, error)
	FileGetAll(name string, commit string) ([]File, error)
	FileCreate(name string, content io.Reader, branch string) error

	CommitGet(name string) (Commit, error)
	CommitList() ([]Commit, error)
	CommitCreate(name string, branch string) (Commit, error)

	BranchGet(name string) (Branch, error)
	BranchList() ([]Branch, error)
	BranchCreate(name string, commit string) (Branch, error)
}

type Replica interface {
	From() (string, error)
	Push(diff io.Reader) error
	Pull(from string, p btrfs.Pusher) error
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
