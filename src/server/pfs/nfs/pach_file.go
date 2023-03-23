package main

import "github.com/go-git/go-billy/v5"

type pachFile struct {
	fs   pachFS
	name string
	idx  int
}

// Name returns the name of the file as presented to Open.
func (p pachFile) Name() string {
	return p.name
}

// io.Writer
func (f pachFile) Write(p []byte) (n int, err error) {
	return 0, billy.ErrReadOnly
}

// io.Reader
func (f pachFile) Read(p []byte) (n int, err error) {
	// TODO
	return 0, nil
}

// io.ReaderAt
func (f pachFile) ReadAt(p []byte, off int64) (n int, err error) {
	// TODO
	return 0, nil
}

// io.Seeker
func Seek(offset int64, whence int) (int64, error) {
	// TODO
	return 0, nil
}

// io.Closer
func (f pachFile) Close() error {
	// TODO
	return 0, nil
}

// Lock locks the file like e.g. flock. It protects against access from
// other processes.
func (f pachFile) Lock() error {
}

// Unlock unlocks the file.
func (f pachFile) Unlock() error {
}

// Truncate the file.
func (f pachFile) Truncate(size int64) error {
}
