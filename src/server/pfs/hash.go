package pfs

import (
	"hash/adler32"
	"path"
	"strings"

	"github.com/pachyderm/pachyderm/src/client/pfs"
)

// A Hasher represents a file/block hasher.
type Hasher struct {
	FileModulus  uint64
	BlockModulus uint64
}

// NewHasher creates a Hasher.
func NewHasher(fileModulus uint64, blockModulus uint64) *Hasher {
	return &Hasher{
		FileModulus:  fileModulus,
		BlockModulus: blockModulus,
	}
}

// HashFile computes and returns a hash of a file.
func (s *Hasher) HashFile(file *pfs.File) uint64 {
	return uint64(adler32.Checksum([]byte(path.Clean(file.Path)))) % s.FileModulus
}

// HashBlock computes and returns a hash of a block.
func (s *Hasher) HashBlock(file *pfs.File, block *pfs.Block) uint64 {
	var str string
	if file != nil {
		str += file.Path
	}
	if block != nil {
		str += block.Hash
	}
	return uint64(adler32.Checksum([]byte(str))) % s.BlockModulus
}

// FileInShard checks if a given file belongs in a given shard, using only the
// file's top-level path.  That is, for a path like foo/bar/buzz, FileInShard only
// considers foo
func FileInShard(shard *pfs.Shard, file *pfs.File) bool {
	if shard == nil || shard.FileModulus == 0 {
		// this lets us default to no filtering
		return true
	}
	path := path.Clean(file.Path)
	if len(path) > 0 && path[0] == '/' {
		path = path[1:]
	}
	parts := strings.Split(path, "/")
	var topLevelPath string
	if len(parts) > 0 {
		topLevelPath = parts[0]
	}
	sharder := &Hasher{FileModulus: shard.FileModulus}
	f := &pfs.File{Path: topLevelPath}
	return sharder.HashFile(f) == shard.FileNumber
}

// BlockInShard returns true if the block is in the given shard.
func BlockInShard(shard *pfs.Shard, file *pfs.File, block *pfs.Block) bool {
	if shard == nil || shard.BlockModulus == 0 {
		// this lets us default to no filtering
		return true
	}
	sharder := &Hasher{BlockModulus: shard.BlockModulus}
	return sharder.HashBlock(file, block) == shard.BlockNumber
}
