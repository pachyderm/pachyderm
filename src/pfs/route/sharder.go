package route

import (
	"hash/adler32"

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
	return int(adler32.Checksum([]byte(path.Path))) % s.numShards, nil
}
