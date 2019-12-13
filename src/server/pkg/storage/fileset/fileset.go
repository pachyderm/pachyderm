package fileset

import (
	"bytes"
	"context"
	"io"
	"path"
	"sort"
	"strconv"

	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/index"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/tar"
)

type memFile struct {
	hdr  *tar.Header
	op   index.Op
	data *bytes.Buffer
}

func (mf *memFile) Write(data []byte) (int, error) {
	mf.hdr.Size += int64(len(data))
	return mf.data.Write(data)
}

// FileSet is a set of files.
// This may be a full filesystem or a subfilesystem (e.g. datum / datum set / shard).
type FileSet struct {
	ctx                        context.Context
	storage                    *Storage
	root                       string
	memAvailable, memThreshold int64
	name                       string
	tag                        string
	fs                         map[string]*memFile
	subFileSet                 int
}

func newFileSet(ctx context.Context, storage *Storage, name string, memThreshold int64, tag string, opts ...Option) *FileSet {
	f := &FileSet{
		ctx:          ctx,
		storage:      storage,
		memAvailable: memThreshold,
		memThreshold: memThreshold,
		name:         name,
		tag:          tag,
		fs:           make(map[string]*memFile),
	}
	for _, opt := range opts {
		opt(f)
	}
	return f
}

func (f *FileSet) Put(r io.Reader) error {
	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		mFile := f.createFile(hdr)
		for {
			n, err := io.CopyN(mFile, tr, f.memAvailable)
			f.memAvailable -= n
			if err != nil {
				if err == io.EOF {
					break
				}
				return err
			}
			if f.memAvailable == 0 {
				if err := f.serialize(); err != nil {
					return err
				}
				mFile = f.createFile(hdr)
			}
		}
	}
	return nil
}

func (f *FileSet) createFile(hdr *tar.Header) *memFile {
	hdr.Name = path.Join(f.root, hdr.Name)
	// Create entry for path if it does not exist.
	if _, ok := f.fs[hdr.Name]; !ok {
		f.createParent(hdr.Name)
		hdr.Size = 0
		f.fs[hdr.Name] = &memFile{
			hdr:  hdr,
			data: &bytes.Buffer{},
		}
	}
	return f.fs[hdr.Name]
}

func (f *FileSet) createParent(name string) {
	name, _ = path.Split(name)
	if _, ok := f.fs[name]; ok {
		return
	}
	f.fs[name] = &memFile{
		hdr: &tar.Header{
			Typeflag: tar.TypeDir,
			Name:     name,
		},
	}
	f.createParent(name)
}

// Delete deletes a file from the file set.
// (bryce) might need to delete ancestor directories in certain cases.
func (f *FileSet) Delete(name string) {
	name = path.Join(f.root, name)
	if _, ok := f.fs[name]; !ok {
		f.fs[name] = &memFile{
			hdr: &tar.Header{
				Name: name,
			},
		}
	}
	f.fs[name].hdr.Size = 0
	f.fs[name].op = index.Op_DELETE
	f.fs[name].data = nil
}

// serialize will be called whenever the in-memory file set is past the memory threshold.
// A new in-memory file set will be created for the following operations.
func (f *FileSet) serialize() error {
	// Sort names.
	names := make([]string, len(f.fs))
	i := 0
	for name := range f.fs {
		names[i] = name
		i++
	}
	sort.Strings(names)
	// Serialize file set.
	w := f.storage.newWriter(f.ctx, path.Join(f.name, strconv.Itoa(f.subFileSet)))
	for _, name := range names {
		n := f.fs[name]
		// (bryce) skipping serialization of deletion operations for the time being.
		if n.op == index.Op_DELETE {
			continue
		}
		if err := w.WriteHeader(n.hdr); err != nil {
			return err
		}
		if n.hdr.Typeflag != tar.TypeDir {
			w.Tag(f.tag)
			if _, err := w.Write(n.data.Bytes()); err != nil {
				return err
			}
		}
	}
	if err := w.Close(); err != nil {
		return err
	}
	// Reset in-memory file set.
	f.fs = make(map[string]*memFile)
	f.memAvailable = f.memThreshold
	f.subFileSet++
	return nil
}

// Close closes the file set.
func (f *FileSet) Close() error {
	return f.serialize()
}
