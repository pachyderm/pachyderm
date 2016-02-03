package pfs

import (
	"hash/adler32"
	"path"
)

type Sharder struct {
	FileModulus  uint64
	BlockModulus uint64
}

func NewSharder(fileModulus uint64, blockModulus uint64) *Sharder {
	return &Sharder{
		FileModulus:  fileModulus,
		BlockModulus: blockModulus,
	}
}

func (s *Sharder) GetShard(file *File) uint64 {
	return uint64(adler32.Checksum([]byte(path.Clean(file.Path)))) % s.FileModulus
}

func (s *Sharder) GetBlockShard(block *Block) uint64 {
	return uint64(adler32.Checksum([]byte(block.Hash))) % s.BlockModulus
}

func FileInShard(shard *Shard, file *File) bool {
	if shard == nil {
		// this lets us default to no filtering
		return true
	}
	sharder := &Sharder{FileModulus: shard.FileModulus}
	return sharder.GetShard(file) == shard.FileNumber
}

func BlockInShard(shard *Shard, block *Block) bool {
	if shard == nil {
		// this lets us default to no filtering
		return true
	}
	sharder := &Sharder{BlockModulus: shard.BlockModulus}
	return sharder.GetBlockShard(block) == shard.BlockNumber
}
