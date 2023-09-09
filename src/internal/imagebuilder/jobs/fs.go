package jobs

import (
	"bytes"
	"io"
	"io/fs"
	"os"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

// FileFS is an fs.FS representing a single file.
type FileFS struct {
	Name string // The name that this file will have in the FS.
	Path string // If an on-disk file, the location.
	Data []byte // If an in-memory file, the data.
	Mode fs.FileMode
}

var _ fs.FS = (*FileFS)(nil)

func (f *FileFS) Open(x string) (fs.File, error) {
	if f.Name == "" {
		return nil, errors.New("invalid FileFS: empty name")
	}
	switch x {
	case ".":
		return &fsDirHandle{fs: f}, nil
	case f.Name:
		if f.Path != "" {
			return os.Open(f.Path)
		}
		return &fsFileHandle{
			Reader: bytes.NewReader(f.Data),
			fs:     f,
		}, nil
	}
	return nil, fs.ErrNotExist
}

func (f *FileFS) info() *fsFileInfo {
	return &fsFileInfo{
		name: f.Name,
		size: int64(len(f.Data)),
		mode: f.Mode,
	}
}

type fsFileHandle struct {
	io.Reader
	fs *FileFS
}

var _ fs.File = (*fsFileHandle)(nil)

func (*fsFileHandle) Close() error { return nil }
func (h *fsFileHandle) Stat() (fs.FileInfo, error) {
	return h.fs.info(), nil
}

type fsDirHandle struct {
	name string
	size int64
	mode fs.FileMode
	fs   *FileFS
}

var _ fs.File = (*fsDirHandle)(nil)
var _ fs.ReadDirFile = (*fsDirHandle)(nil)

func (*fsDirHandle) Read(p []byte) (int, error) { return 0, fs.ErrInvalid }
func (*fsDirHandle) Close() error               { return nil }
func (h *fsDirHandle) Stat() (fs.FileInfo, error) {
	return &fsFileInfo{
		name: ".",
		size: 4096,
		mode: 0o777 | fs.ModeDir,
	}, nil
}
func (h *fsDirHandle) ReadDir(n int) ([]fs.DirEntry, error) {
	if n == 0 {
		return nil, fs.ErrInvalid
	}
	return []fs.DirEntry{h.fs.info()}, nil
}

type fsFileInfo struct {
	name string
	size int64
	mode fs.FileMode
}

var _ fs.FileInfo = (*fsFileInfo)(nil)
var _ fs.DirEntry = (*fsFileInfo)(nil)

func (i *fsFileInfo) Name() string               { return i.name }
func (i *fsFileInfo) Size() int64                { return i.size }
func (i *fsFileInfo) Mode() fs.FileMode          { return i.mode }
func (i *fsFileInfo) Type() fs.FileMode          { return i.mode.Type() }
func (i *fsFileInfo) ModTime() time.Time         { return time.Time{} }
func (i *fsFileInfo) IsDir() bool                { return i.mode&fs.ModeDir != 0 }
func (i *fsFileInfo) Sys() any                   { return i }
func (i *fsFileInfo) Info() (fs.FileInfo, error) { return i, nil }
func (i *fsFileInfo) String() string             { return fs.FormatFileInfo(i) }
