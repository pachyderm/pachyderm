package route

import (
	"hash/adler32"
	"path"

	"github.com/pachyderm/pachyderm/src/pfs"
)

type sharder struct {
	fileModulus  uint64
	blockModulus uint64
}

func newSharder(fileModulus uint64, blockModulus uint64) *sharder {
	return &sharder{
		fileModulus:  fileModulus,
		blockModulus: blockModulus,
	}
}

func (s *sharder) FileModulus() uint64 {
	return s.fileModulus
}

func (s *sharder) BlockModulus() uint64 {
	return s.blockModulus
}

func (s *sharder) GetShard(file *pfs.File) uint64 {
	return uint64(adler32.Checksum([]byte(path.Clean(file.Path)))) % s.fileModulus
}

func (s *sharder) GetBlockShard(block *pfs.Block) uint64 {
	return uint64(adler32.Checksum([]byte(block.Hash))) % s.blockModulus
}

func FileInShard(shard *pfs.Shard, file *pfs.File) bool {
	if shard == nil {
		// this lets us default to no filtering
		return true
	}
	sharder := &sharder{fileModulus: shard.FileModulus}
	return sharder.GetShard(file) == shard.FileNumber
}

func BlockInShard(shard *pfs.Shard, block *pfs.Block) bool {
	if shard == nil {
		// this lets us default to no filtering
		return true
	}
	sharder := &sharder{blockModulus: shard.BlockModulus}
	return sharder.GetBlockShard(block) == shard.BlockNumber
}
