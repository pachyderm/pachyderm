package jobs

import (
	"fmt"
	"io/fs"
	"os"
)

// WithFS is something representend by a filesystem.
type WithFS interface {
	FS() fs.FS
}

// File represents something on disk (including directories).
type File struct {
	Name      string
	Path      string
	Directory bool
	Digest    Digest
}

func (f *File) String() string {
	if f == nil {
		return "<nil file>"
	}
	return fmt.Sprintf("file:%v@%v", f.Path, f.Digest.String())
}

func (f *File) FS() fs.FS {
	if f.Directory {
		return os.DirFS(f.Path)
	} else {
		return &FileFS{Name: f.Name, Path: f.Path, Mode: 0o666}
	}
}
