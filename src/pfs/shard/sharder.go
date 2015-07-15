package shard

import (
	"hash/adler32"
	"path/filepath"

	"github.com/pachyderm/pachyderm/src/pfs"
)

type sharder struct {
	numShards int
}

func newSharder(numShards int) *sharder {
	return &sharder{numShards}
}

func (s *sharder) NumShards() int {
	return s.numShards
}

func (s *sharder) GetShard(path *pfs.Path) (int, error) {
	return int(
		adler32.Checksum(
			[]byte(
				filepath.Join(
					path.Commit.Repository.Name,
					path.Commit.Id,
					path.Path,
				),
			),
		),
	), nil
}
