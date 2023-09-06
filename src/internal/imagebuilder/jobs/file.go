package jobs

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"testing/fstest"
	"time"
)

// WithFS is something representend by a filesystem.
type WithFS interface {
	FS() fs.FS
}

// File represents something on disk (including directories).
type File struct {
	Name   string
	Path   string
	Digest Digest
}

func (f *File) String() string {
	if f == nil {
		return "<nil file>"
	}
	return fmt.Sprintf("file:%v@%v", f.Path, f.Digest.String())
}

func (f *File) FS() fs.FS {
	return os.DirFS(f.Path)
}

type MemFile struct {
	Name
	fs fs.FS
}

func (f *MemFile) FS() fs.FS {
	return f.fs
}

func JSONFile(name Name, filename string, x any) (*MemFile, error) {
	js, err := json.Marshal(x)
	if err != nil {
		return nil, err
	}
	return &MemFile{
		Name: name,
		fs: fstest.MapFS{
			filename: &fstest.MapFile{
				Data:    js,
				Mode:    0o600,
				ModTime: time.Now(),
			},
		},
	}, nil
}
