package fileset

import (
	"bytes"
	"context"
	"path"
	"sort"
	"strconv"

	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/index"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/tar"
)

type node struct {
	hdr  *tar.Header
	op   index.Op
	data *bytes.Buffer
	tag  string
}

// FileSet is a set of files.
// This may be a full filesystem or a subfilesystem (e.g. datum / datum set / shard).
type FileSet struct {
	ctx                        context.Context
	storage                    *Storage
	root                       string
	memAvailable, memThreshold int64
	name                       string
	fs                         map[string]*node
	curr                       *node
	tag                        string
	part                       int
}

func newFileSet(ctx context.Context, storage *Storage, name string, memThreshold int64, opts ...Option) *FileSet {
	f := &FileSet{
		ctx:          ctx,
		storage:      storage,
		memAvailable: memThreshold,
		memThreshold: memThreshold,
		name:         name,
		fs:           make(map[string]*node),
	}
	for _, opt := range opts {
		opt(f)
	}
	return f
}

// StartTag starts a tag for the next set of files.
func (f *FileSet) StartTag(tag string) {
	f.tag = tag
}

// WriteHeader writes a tar header and prepares to accept the file's contents.
// (bryce) should we prevent directories from being written here?
func (f *FileSet) WriteHeader(hdr *tar.Header) error {
	hdr.Name = path.Join(f.root, hdr.Name)
	// Create entry for path if it does not exist.
	if _, ok := f.fs[hdr.Name]; !ok {
		f.createParent(hdr.Name)
		hdr.Size = 0
		f.fs[hdr.Name] = &node{
			hdr:  hdr,
			data: &bytes.Buffer{},
		}
	}
	// (bryce) should make a note about the implication of this
	// a file in a file set being written can only have one tag.
	// multiple come into play when merging.
	f.fs[hdr.Name].tag = f.tag
	f.curr = f.fs[hdr.Name]
	return nil
}

func (f *FileSet) createParent(name string) {
	name, _ = path.Split(name)
	if _, ok := f.fs[name]; ok {
		return
	}
	f.fs[name] = &node{
		hdr: &tar.Header{
			Typeflag: tar.TypeDir,
			Name:     name,
		},
	}
	f.createParent(name)
}

// Write writes to the current file in the tar stream.
func (f *FileSet) Write(data []byte) (int, error) {
	for int64(len(data)) > f.memAvailable {
		n, _ := f.curr.data.Write(data[:int(f.memAvailable)])
		f.curr.hdr.Size += int64(n)
		data = data[n:]
		if err := f.serialize(); err != nil {
			return 0, err
		}
	}
	n, _ := f.curr.data.Write(data)
	f.curr.hdr.Size += int64(n)
	f.memAvailable -= int64(n)
	return n, nil
}

// Delete deletes a file from the file set.
// (bryce) might need to delete ancestor directories in certain cases.
func (f *FileSet) Delete(name string) {
	name = path.Join(f.root, name)
	if _, ok := f.fs[name]; !ok {
		f.fs[name] = &node{
			hdr: &tar.Header{
				Name: name,
			},
		}
	}
	f.fs[name].hdr.Size = 0
	f.fs[name].op = index.Op_DELETE
	f.fs[name].data = nil
	f.curr = nil
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
	w := f.storage.newWriter(f.ctx, path.Join(f.name, strconv.Itoa(f.part)))
	for _, name := range names {
		n := f.fs[name]
		// (bryce) skipping serialization of deletion operations for the time being.
		// only testing against basic interface without multiple parts.
		if n.op == index.Op_DELETE {
			continue
		}
		if err := w.WriteHeader(&index.Header{Hdr: n.hdr}); err != nil {
			return err
		}
		if n.hdr.Typeflag != tar.TypeDir {
			w.StartTag(n.tag)
			if _, err := w.Write(n.data.Bytes()); err != nil {
				return err
			}
		}
	}
	if err := w.Close(); err != nil {
		return err
	}
	// Reset in-memory file set.
	f.fs = make(map[string]*node)
	f.memAvailable = f.memThreshold
	f.part++
	return nil
}

// Close closes the file set.
func (f *FileSet) Close() error {
	return f.serialize()
}
