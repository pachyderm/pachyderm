package pfs

import (
	"errors"
)

var ErrFileNotFound = errors.New("file not found")

func ByteRangeSize(byteRange *ByteRange) uint64 {
	return byteRange.Upper - byteRange.Lower
}
