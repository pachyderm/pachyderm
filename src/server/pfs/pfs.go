package pfs

import (
	"errors"
	. "github.com/pachyderm/pachyderm/src/client/pfs"
)

var ErrFileNotFound = errors.New("file not found")

func ByteRangeSize(byteRange *ByteRange) uint64 {
	return byteRange.Upper - byteRange.Lower
}
