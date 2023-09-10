//go:build unix

package fsutil

import (
	"io/fs"
	"syscall"
)

type HasUID interface{ GetUID() (uint32, bool) }
type HasGID interface{ GetGID() (uint32, bool) }

func FileUID(i fs.FileInfo) (uint32, bool) {
	if r, ok := i.Sys().(HasUID); ok {
		return r.GetUID()
	}
	if r, ok := i.Sys().(*syscall.Stat_t); ok && r != nil {
		return r.Uid, true
	}
	return 0, false
}

func FileGID(i fs.FileInfo) (uint32, bool) {
	if r, ok := i.Sys().(HasGID); ok {
		return r.GetGID()
	}
	if r, ok := i.Sys().(*syscall.Stat_t); ok && r != nil {
		return r.Gid, true
	}
	return 0, false
}
