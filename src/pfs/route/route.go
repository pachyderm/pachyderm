package route

import (
	"github.com/pachyderm/pachyderm/src/pfs"
)

type Sharder interface {
	FileModulus() uint64
	BlockModulus() uint64
	GetShard(file *pfs.File) uint64
	GetBlockShard(block *pfs.Block) uint64
}

func NewSharder(fileModulus uint64, blockModulus uint64) Sharder {
	return newSharder(fileModulus, blockModulus)
}
