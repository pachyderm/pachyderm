package pfs

import (
	"github.com/pachyderm/pachyderm/src/client/pfs"
	"hash/adler32"
	"path"
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

func (s *Hasher) HashFile(file *pfs.File) uint64 {
	return uint64(adler32.Checksum([]byte(path.Clean(file.Path)))) % s.FileModulus
}

func (s *Hasher) HashBlock(block *pfs.Block) uint64 {
	return uint64(adler32.Checksum([]byte(block.Hash))) % s.BlockModulus
}

func FileInShard(shard *pfs.Shard, file *pfs.File) bool {
	if shard == nil {
		// this lets us default to no filtering
		return true
	}
	sharder := &Hasher{FileModulus: shard.FileModulus}
	return sharder.HashFile(file) == shard.FileNumber
}

func BlockInShard(shard *pfs.Shard, block *pfs.Block) bool {
	if shard == nil {
		// this lets us default to no filtering
		return true
	}
	sharder := &Hasher{BlockModulus: shard.BlockModulus}
	return sharder.HashBlock(block) == shard.BlockNumber
}
