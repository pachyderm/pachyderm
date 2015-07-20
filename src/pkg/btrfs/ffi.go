package btrfs

/*
#include <stdlib.h>
#include <dirent.h>
#include <btrfs/ioctl.h>
*/
import "C"

import (
	"fmt"
	"path/filepath"
	"syscall"
	"unsafe"
)

var (
	callStrings = map[uintptr]string{
		C.BTRFS_IOC_SUBVOL_SETFLAGS: "btrfs property set",
		C.BTRFS_IOC_SUBVOL_GETFLAGS: "btrfs property get",
		C.BTRFS_IOC_SUBVOL_CREATE:   "btrfs subvolume create",
		C.BTRFS_IOC_SNAP_CREATE_V2:  "btrfs subvolume snapshot",
	}
)

type ffiAPI struct{}

func newFFIAPI() *ffiAPI {
	return &ffiAPI{}
}

func (a *ffiAPI) PropertySetReadonly(path string, readonly bool) error {
	return ffiPropertySetReadonly(path, readonly)
}

func (a *ffiAPI) PropertyGetReadonly(path string) (bool, error) {
	return ffiPropertyGetReadonly(path)
}

func (a *ffiAPI) SubvolumeCreate(path string) error {
	return ffiSubvolumeCreate(path)
}

func (a *ffiAPI) SubvolumeSnapshot(src string, dest string, readonly bool) error {
	return ffiSubvolumeSnapshot(src, dest, readonly)
}

func ffiPropertySetReadonly(path string, readOnly bool) error {
	var flags C.__u64
	if readOnly {
		flags |= C.__u64(C.BTRFS_SUBVOL_RDONLY)
	} else {
		flags = flags &^ C.__u64(C.BTRFS_SUBVOL_RDONLY)
	}
	return ffiIoctl(path, C.BTRFS_IOC_SUBVOL_SETFLAGS, uintptr(unsafe.Pointer(&flags)))
}

func ffiPropertyGetReadonly(path string) (bool, error) {
	var flags C.__u64
	if err := ffiIoctl(path, C.BTRFS_IOC_SUBVOL_GETFLAGS, uintptr(unsafe.Pointer(&flags))); err != nil {
		return false, err
	}
	return flags&C.BTRFS_SUBVOL_RDONLY == C.BTRFS_SUBVOL_RDONLY, nil
}

func ffiSubvolumeCreate(path string) error {
	var args C.struct_btrfs_ioctl_vol_args
	for i, c := range []byte(filepath.Base(path)) {
		args.name[i] = C.char(c)
	}
	return ffiIoctl(filepath.Dir(path), C.BTRFS_IOC_SUBVOL_CREATE, uintptr(unsafe.Pointer(&args)))
}

func ffiSubvolumeSnapshot(src string, dest string, readOnly bool) error {
	srcDir, err := ffiOpenDir(src)
	if err != nil {
		return err
	}
	defer ffiCloseDir(srcDir)
	var args C.struct_btrfs_ioctl_vol_args_v2
	args.fd = C.__s64(ffiGetDirFd(srcDir))
	if readOnly {
		args.flags |= C.__u64(C.BTRFS_SUBVOL_RDONLY)
	}
	for i, c := range []byte(filepath.Base(dest)) {
		args.name[i] = C.char(c)
	}
	return ffiIoctl(filepath.Dir(dest), C.BTRFS_IOC_SNAP_CREATE_V2, uintptr(unsafe.Pointer(&args)))
}

func ffiIoctl(path string, call uintptr, args uintptr) error {
	dir, err := ffiOpenDir(path)
	if err != nil {
		return err
	}
	defer ffiCloseDir(dir)
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, ffiGetDirFd(dir), call, args)
	if errno != 0 {
		return fmt.Errorf("%s failed for %s: %v", callStrings[call], path, errno.Error())
	}
	return nil
}

func ffiOpenDir(path string) (*C.DIR, error) {
	Cpath := C.CString(path)
	defer C.free(unsafe.Pointer(Cpath))
	dir := C.opendir(Cpath)
	if dir == nil {
		return nil, fmt.Errorf("cannot open dir: %s", path)
	}
	return dir, nil
}

func ffiCloseDir(dir *C.DIR) {
	if dir != nil {
		C.closedir(dir)
	}
}

func ffiGetDirFd(dir *C.DIR) uintptr {
	return uintptr(C.dirfd(dir))
}
