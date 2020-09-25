package fileset

import (
	"bytes"
	"context"
	"io"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/server/pkg/tar"
)

// TODO Might want to rework this a bit later or add some additional validation.
type dataOp struct {
	deleteTags map[string]struct{}
	memFiles   map[string]*memFile
}

type memFile struct {
	hdr  *tar.Header
	tag  string
	data *bytes.Buffer
}

func (mf *memFile) Write(data []byte) (int, error) {
	mf.hdr.Size += int64(len(data))
	return mf.data.Write(data)
}

// UnorderedWriter allows writing Files, unordered by path, into multiple ordered filesets.
// This may be a full filesystem or a subfilesystem (e.g. datum / datum set / shard).
type UnorderedWriter struct {
	ctx                        context.Context
	storage                    *Storage
	memAvailable, memThreshold int64
	name                       string
	defaultTag                 string
	fs                         map[string]*dataOp
	subFileSet                 int64
	writerOpts                 []WriterOption
	expiresAt                  *time.Time
}

func newUnorderedWriter(ctx context.Context, storage *Storage, name string, memThreshold int64, defaultTag string, opts ...UnorderedWriterOption) (*UnorderedWriter, error) {
	if err := storage.filesetSem.Acquire(ctx, 1); err != nil {
		return nil, err
	}
	f := &UnorderedWriter{
		ctx:          ctx,
		storage:      storage,
		memAvailable: memThreshold,
		memThreshold: memThreshold,
		name:         name,
		defaultTag:   defaultTag,
		fs:           make(map[string]*dataOp),
	}
	for _, opt := range opts {
		opt(f)
	}
	return f, nil
}

// Put reads files from a tar stream and adds them to the fileset.
func (f *UnorderedWriter) Put(r io.Reader, customTag ...string) error {
	tag := f.defaultTag
	if len(customTag) > 0 && customTag[0] != "" {
		tag = customTag[0]
	}
	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		hdr.Name = CleanTarPath(hdr.Name, hdr.FileInfo().IsDir())
		if hdr.Typeflag == tar.TypeDir {
			f.createParent(hdr.Name, tag)
			continue
		}
		mf := f.createFile(hdr, tag)
		for {
			n, err := io.CopyN(mf, tr, f.memAvailable)
			f.memAvailable -= n
			if err != nil {
				if errors.Is(err, io.EOF) {
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
}

func (f *UnorderedWriter) createFile(hdr *tar.Header, tag string) *memFile {
	f.createParent(hdr.Name, tag)
	hdr.Size = 0
	mf := &memFile{
		hdr:  hdr,
		tag:  tag,
		data: &bytes.Buffer{},
	}
	dataOp := f.getDataOp(hdr.Name)
	dataOp.memFiles[tag] = mf
	return mf
}

func (f *UnorderedWriter) getDataOp(name string) *dataOp {
	if _, ok := f.fs[name]; !ok {
		f.fs[name] = &dataOp{
			deleteTags: make(map[string]struct{}),
			memFiles:   make(map[string]*memFile),
		}
	}
	return f.fs[name]
}

func (f *UnorderedWriter) createParent(name string, tag string) {
	if name == "" {
		return
	}
	name, _ = path.Split(name)
	if _, ok := f.fs[name]; ok {
		return
	}
	mf := &memFile{
		hdr: &tar.Header{
			Typeflag: tar.TypeDir,
			Name:     CleanTarPath(name, true),
		},
		tag:  tag,
		data: &bytes.Buffer{},
	}
	dataOp := f.getDataOp(name)
	dataOp.memFiles[tag] = mf
	name = strings.TrimRight(name, "/")
	f.createParent(name, tag)
}

// Delete deletes a file from the file set.
func (f *UnorderedWriter) Delete(name string, customTag ...string) {
	var tag string
	if len(customTag) > 0 {
		tag = customTag[0]
	}
	if tag == headerTag {
		deleteTags := make(map[string]struct{})
		deleteTags[headerTag] = struct{}{}
		f.fs[name] = &dataOp{deleteTags: deleteTags}
		return
	}
	dataOp := f.getDataOp(name)
	if _, ok := dataOp.deleteTags[headerTag]; ok {
		return
	}
	dataOp.deleteTags[tag] = struct{}{}
	dataOp.memFiles[tag] = nil
}

// PathsWritten returns the full paths (not prefixes) written by this UnorderedWriter
func (f *UnorderedWriter) PathsWritten() (ret []string) {
	name := removePrefix(f.name)
	for i := int64(0); i < f.subFileSet; i++ {
		p := path.Join(name, SubFileSetStr(i))
		ret = append(ret, p)
	}
	return ret
}

// serialize will be called whenever the in-memory file set is past the memory threshold.
// A new in-memory file set will be created for the following operations.
func (f *UnorderedWriter) serialize() error {
	// Sort names.
	names := make([]string, len(f.fs))
	i := 0
	for name := range f.fs {
		names[i] = name
		i++
	}
	sort.Strings(names)
	// Serialize file set.
	w := f.storage.newWriter(f.ctx, path.Join(f.name, SubFileSetStr(f.subFileSet)), f.writerOpts...)
	for _, name := range names {
		dataOp := f.fs[name]
		deleteTags := getSortedTags(dataOp.deleteTags)
		mfs := getSortedMemFiles(dataOp.memFiles)
		if len(mfs) == 0 {
			if err := w.DeleteFile(name, deleteTags...); err != nil {
				return err
			}
			continue
		}
		// TODO Tar header validation?
		hdr := mfs[len(mfs)-1].hdr
		if err := w.WriteHeader(hdr); err != nil {
			return err
		}
		for _, tag := range deleteTags {
			w.DeleteTag(tag)
		}
		for _, mf := range mfs {
			w.Tag(mf.tag)
			if _, err := w.Write(mf.data.Bytes()); err != nil {
				return err
			}
		}
	}
	if err := w.Close(); err != nil {
		return err
	}
	if expiresAt := w.ExpiresAt(); expiresAt != nil {
		if f.expiresAt == nil || expiresAt.Before(*f.expiresAt) {
			f.expiresAt = expiresAt
		}
	}

	// Reset in-memory file set.
	f.fs = make(map[string]*dataOp)
	f.memAvailable = f.memThreshold
	f.subFileSet++
	return nil
}

func getSortedMemFiles(memFiles map[string]*memFile) []*memFile {
	var mfs []*memFile
	for _, mf := range memFiles {
		mfs = append(mfs, mf)
	}
	sort.SliceStable(mfs, func(i, j int) bool {
		return mfs[i].tag < mfs[j].tag
	})
	return mfs
}

func getSortedTags(tags map[string]struct{}) []string {
	var tgs []string
	for tag := range tags {
		tgs = append(tgs, tag)
	}
	sort.Strings(tgs)
	return tgs
}

// Close closes the writer.
func (f *UnorderedWriter) Close() error {
	defer f.storage.filesetSem.Release(1)
	return f.serialize()
}

// ExpiresAt returns the time the fileset expires
// - returns nil if the fileset is permanent
// - should not be called until the UnorderedWriter is closed
func (uw *UnorderedWriter) ExpiresAt() *time.Time {
	return uw.expiresAt
}
