//go:build !unix

package fsutil

import (
	"io/fs"
)

type HasUID interface{ GetUID() (uint32, bool) }
type HasGID interface{ GetGID() (uint32, bool) }

func FileUID(i fs.FileInfo) (uint32, bool) {
	if r, ok := i.Sys().(HasUID); ok {
		return r.GetUID()
	}
	return 0, false
}

func FileGID(i fs.FileInfo) (uint32, bool) {
	if r, ok := i.Sys().(HasGID); ok {
		return r.GetGID()
	}
	return 0, false
}
