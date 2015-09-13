package route

import (
	"hash/adler32"
	"path"

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

func (s *sharder) GetShard(file *pfs.File) (int, error) {
	return int(adler32.Checksum([]byte(path.Clean(file.Path)))) % s.numShards, nil
}
