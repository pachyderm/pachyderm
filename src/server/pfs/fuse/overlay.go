package fuse

import (
	"fmt"
	"syscall"
)

// unshare is a wrapper around syscall.Unshare calling this allows you to call
// overlay without being root
func unshare() error {
	return syscall.Unshare(syscall.CLONE_NEWNS | syscall.CLONE_NEWUSER)
}

// overlay is a wrapper around syscall.Mount that mounts an overlay filesystem
// `lowerdir` is the base layer which will be read from but not written to,
// `upperdir` is where writes will go, and thus where we'll find writes once the mount is finished
// `workdir` is a scratch space that overlay uses, overlay seems to clean it up
// after it's unmounted leaving only empty directories, the caller should
// cleanup those directories after unmount.
// target is where the filesystem will be mounted
// all of these directories should exist before calling this
func overlay(lowerdir, upperdir, workdir, target string) error {
	// This is the signature of syscall.Mount:
	// func Mount(source string, target string, fstype string, flags uintptr, data string)
	return syscall.Mount(
		// This is the "source" for the call to Mount, for normal mounts this
		// would be a block device, overlay doesn't have a source though so we
		// can actually put whatever we want here, we put the dummy string
		// "overlay" in since that seems to be standard.
		"pfs",
		// Target is where we want to mount to.
		target,
		// This is the filesystem type, this is where we actually specify that
		// we want to use overlay.
		"overlay",
		// This is where we pass flags, we don't have any flags we want
		// to pass, but MS_MGC_VAL was required for kernels prior to 2.4, and
		// even though it's unlikely that anyone will run this on such an old
		// kernel we might as well be more compatible rather than less
		// compatible.
		syscall.MS_MGC_VAL,
		// These are the args to overlay (as opposed to the mount syscall) this
		// is where we specify how what directories to overlay, "lowerdir" is
		// the base layer which will be read from but not written to,
		// "upperdir" is where writes will go, and thus where we'll find writes
		// once the mount is finished, "workdir" is a scratch space that
		// overlay uses, overlay seems to clean it up after it's unmounted, but
		// it makes sense to delete it after the mount is finished.
		fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s", lowerdir, upperdir, workdir),
	)

	// mount("overlay", "/home/jdoliner/Repos/pachyderm/pfs", "overlay", MS_MGC_VAL, "lowerdir=./lower,upperdir=./uppe"...)
}
