package pfs

import (
	"errors"
	"github.com/pachyderm/pachyderm/src/client/pfs"
)

var ErrFileNotFound = errors.New("file not found")

func ByteRangeSize(byteRange *pfs.ByteRange) uint64 {
	return byteRange.Upper - byteRange.Lower
}
