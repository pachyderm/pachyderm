package fileset

import (
	"bytes"
	"context"
	"io"
	"path"
	"sort"

	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/index"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/tar"
)

type memFile struct {
	hdr  *tar.Header
	tag  string
	data *bytes.Buffer
	op   index.Op
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
	defaultTag                 string
	fs                         map[string][]*memFile
	subFileSet                 int64
}

func newFileSet(ctx context.Context, storage *Storage, name string, memThreshold int64, defaultTag string, opts ...Option) *FileSet {
	f := &FileSet{
		ctx:          ctx,
		storage:      storage,
		memAvailable: memThreshold,
		memThreshold: memThreshold,
		name:         name,
		defaultTag:   defaultTag,
		fs:           make(map[string][]*memFile),
	}
	for _, opt := range opts {
		opt(f)
	}
	return f
}

// Put reads files from a tar stream and adds them to the fileset.
// (bryce) probably should prevent / clean files that end with "/", since that will indicate a directory.
func (f *FileSet) Put(r io.Reader, customTag ...string) error {
	tag := f.defaultTag
	if len(customTag) > 0 && customTag[0] != "" {
		tag = customTag[0]
	}
	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		hdr.Name = path.Join(f.root, hdr.Name)
		mf := f.createFile(hdr, tag)
		for {
			n, err := io.CopyN(mf, tr, f.memAvailable)
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
				mf = f.createFile(hdr, tag)
			}
		}
	}
	return nil
}

func (f *FileSet) createFile(hdr *tar.Header, tag string) *memFile {
	f.createParent(hdr.Name)
	hdr.Size = 0
	mf := &memFile{
		hdr:  hdr,
		tag:  tag,
		data: &bytes.Buffer{},
	}
	f.fs[hdr.Name] = append(f.fs[hdr.Name], mf)
	return mf
}

func (f *FileSet) createParent(name string) {
	name, _ = path.Split(name)
	if _, ok := f.fs[name]; ok {
		return
	}
	f.fs[name] = append(f.fs[name], &memFile{
		hdr: &tar.Header{
			Typeflag: tar.TypeDir,
			Name:     name,
		},
	})
	f.createParent(name)
}

// Delete deletes a file from the file set.
// (bryce) might need to delete ancestor directories in certain cases.
func (f *FileSet) Delete(name string) {
	name = path.Join(f.root, name)
	f.fs[name] = []*memFile{&memFile{op: index.Op_DELETE}}
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
	w := f.storage.newWriter(f.ctx, path.Join(f.name, SubFileSetStr(f.subFileSet)), nil)
	for _, name := range names {
		mfs := f.fs[name]
		sort.Slice(mfs, func(i, j int) bool {
			return mfs[i].tag < mfs[j].tag
		})
		// (bryce) skipping serialization of deletion operations for the time being.
		hdr := mfs[len(mfs)-1].hdr
		if err := w.WriteHeader(hdr); err != nil {
			return err
		}
		if hdr.Typeflag != tar.TypeDir {
			for _, mf := range mfs {
				w.Tag(mf.tag)
				if _, err := w.Write(mf.data.Bytes()); err != nil {
					return err
				}
			}
		}
	}
	if err := w.Close(); err != nil {
		return err
	}
	// Reset in-memory file set.
	f.fs = make(map[string][]*memFile)
	f.memAvailable = f.memThreshold
	f.subFileSet++
	return nil
}

// Close closes the file set.
func (f *FileSet) Close() error {
	return f.serialize()
}
