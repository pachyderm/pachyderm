// This file implements the billy.Filesystem interface for the pachnfs type

package main

import (
	"os"
	"path"

	"github.com/go-git/go-billy/v5"
)

type pachFS struct {
	nfs pachNFS
}

// Create creates the named file with mode 0666 (before umask), truncating
// it if it already exists. If successful, methods on the returned File can
// be used for I/O; the associated file descriptor has mode O_RDWR.
func (p pachnfs) Create(filename string) (billy.File, error) {
	return nil, billy.ErrReadOnly
}

// Open opens the named file for reading. If successful, methods on the
// returned file can be used for reading; the associated file descriptor has
// mode O_RDONLY.
func (p pachnfs) Open(filename string) (billy.File, error) {
	return pachfile{}
}

// OpenFile is the generalized open call; most users will use Open or Create
// instead. It opens the named file with specified flag (O_RDONLY etc.) and
// perm, (0666 etc.) if applicable. If successful, methods on the returned
// File can be used for I/O.
func (p pachnfs) OpenFile(filename string, flag int, perm os.FileMode) (billy.File, error) {
	return pachfile{}

}

// Stat returns a FileInfo describing the named file.
func (p pachnfs) Stat(filename string) (os.FileInfo, error) {
	return nil, nil
}

// Rename renames (moves) oldpath to newpath. If newpath already exists and
// is not a directory, Rename replaces it. OS-specific restrictions may
// apply when oldpath and newpath are in different directories.
func (p pachnfs) Rename(oldpath, newpath string) error {
	return billy.ErrReadOnly
}

// Remove removes the named file or directory.
func (p pachnfs) Remove(filename string) error {
	return billy.ErrReadOnly
}

// Join joins any number of path elements into a single path, adding a
// Separator if necessary. Join calls filepath.Clean on the result; in
// particular, all empty strings are ignored. On Windows, the result is a
// UNC path if and only if the first path element is a UNC path.
func (p pachnfs) Join(elem ...string) string {
	return path.Join(elem...)
}
