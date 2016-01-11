package route

import (
	"hash/adler32"
	"path"

	"github.com/pachyderm/pachyderm/src/pfs"
	"github.com/pachyderm/pachyderm/src/pfs/drive"
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

func (s *sharder) GetBlockShard(block *drive.Block) uint64 {
	return uint64(adler32.Checksum([]byte(block.Hash))) % s.blockModulus
}
