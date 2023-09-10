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
			orig, err := os.Open(f.Path)
			if err != nil {
				return nil, errors.Wrap(err, "open original disk file")
			}
			return &fsFileHandle{
				Reader: orig,
				fs:     f,
			}, nil
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

func (h *fsFileHandle) Close() error {
	if c, ok := h.Reader.(io.Closer); ok {
		return c.Close()
	}
	return nil

}

func (h *fsFileHandle) Stat() (fs.FileInfo, error) {
	return h.fs.info(), nil
}

type fsDirHandle struct {
	name        string
	size        int64
	mode        fs.FileMode
	fs          *FileFS
	returnedDir bool
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
	all := []fs.DirEntry{h.fs.info()}
	switch {
	case !h.returnedDir && n == -1:
		h.returnedDir = true
		return all, nil
	case h.returnedDir && n == -1:
		return nil, nil
	case !h.returnedDir && n > 0:
		h.returnedDir = true
		return all, io.EOF
	case h.returnedDir && n > 0:
		return nil, io.EOF
	default:
		return nil, fs.ErrInvalid
	}
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
