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

func (s *Hasher) HashFile(file *pfs.File) uint64 {
	return uint64(adler32.Checksum([]byte(path.Clean(file.Path)))) % s.FileModulus
}

func (s *Hasher) HashBlock(block *pfs.Block) uint64 {
	return uint64(adler32.Checksum([]byte(block.Hash))) % s.BlockModulus
}

// FileInShard checks if a given file belongs in a given shard, using only the
// file's top-level path.  That is, for a path like foo/bar/buzz, FileInShard only
// considers foo
func FileInShard(shard *pfs.Shard, file *pfs.File) bool {
	if shard == nil || shard.FileModulus == 0 {
		// this lets us default to no filtering
		return true
	}
	parts := strings.Split(path.Clean(file.Path), "/")
	var topLevelPath string
	if len(parts) > 0 {
		topLevelPath = parts[0]
	}
	sharder := &Hasher{FileModulus: shard.FileModulus}
	f := &pfs.File{Path: topLevelPath}
	return sharder.HashFile(f) == shard.FileNumber
}

func BlockInShard(shard *pfs.Shard, block *pfs.Block) bool {
	if shard == nil || shard.BlockModulus == 0 {
		// this lets us default to no filtering
		return true
	}
	sharder := &Hasher{BlockModulus: shard.BlockModulus}
	return sharder.HashBlock(block) == shard.BlockNumber
}
