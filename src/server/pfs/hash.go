package pfs

import (
	"hash/adler32"
	"path"
	"strings"

	"github.com/pachyderm/pachyderm/src/client/pfs"
)

type Hasher struct {
	FileModulus  uint64
	BlockModulus uint64
}

func NewHasher(fileModulus uint64, blockModulus uint64) *Hasher {
	return &Hasher{
		FileModulus:  fileModulus,
		BlockModulus: blockModulus,
	}
}

// HashFile hashes the first part of the file path.  That is, for a path like
// foo/bar/buzz, HashFile only considers foo
func (s *Hasher) HashFile(file *pfs.File) uint64 {
	parts := strings.Split(path.Clean(file.Path), "/")
	var topLevelPath string
	if len(parts) > 0 {
		topLevelPath = parts[0]
	}
	return uint64(adler32.Checksum([]byte(topLevelPath))) % s.FileModulus
}

func (s *Hasher) HashBlock(block *pfs.Block) uint64 {
	return uint64(adler32.Checksum([]byte(block.Hash))) % s.BlockModulus
}

func FileInShard(shard *pfs.Shard, file *pfs.File) bool {
	if shard == nil || shard.FileModulus == 0 {
		// this lets us default to no filtering
		return true
	}
	sharder := &Hasher{FileModulus: shard.FileModulus}
	return sharder.HashFile(file) == shard.FileNumber
}

func BlockInShard(shard *pfs.Shard, block *pfs.Block) bool {
	if shard == nil || shard.BlockModulus == 0 {
		// this lets us default to no filtering
		return true
	}
	sharder := &Hasher{BlockModulus: shard.BlockModulus}
	return sharder.HashBlock(block) == shard.BlockNumber
}
