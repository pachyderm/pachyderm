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
	ttl                        time.Duration
	renewer                    *Renewer
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
func (f *UnorderedWriter) Put(r io.Reader, overwrite bool, customTag ...string) error {
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
		mf := f.createFile(hdr, overwrite, tag)
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
				mf = f.createFile(hdr, overwrite, tag)
			}
		}
	}
}

func (f *UnorderedWriter) createFile(hdr *tar.Header, overwrite bool, tag string) *memFile {
	f.createParent(hdr.Name, tag)
	hdr.Size = 0
	mf := &memFile{
		hdr:  hdr,
		tag:  tag,
		data: &bytes.Buffer{},
	}
	dataOp := f.getDataOp(hdr.Name, overwrite)
	dataOp.memFiles[tag] = mf
	return mf
}

func (f *UnorderedWriter) getDataOp(name string, overwrite bool) *dataOp {
	if _, ok := f.fs[name]; !ok {
		f.fs[name] = &dataOp{
			deleteTags: make(map[string]struct{}),
			memFiles:   make(map[string]*memFile),
		}
	}
	dataOp := f.fs[name]
	if overwrite {
		dataOp.deleteTags = make(map[string]struct{})
		dataOp.deleteTags[headerTag] = struct{}{}
		dataOp.memFiles = make(map[string]*memFile)
	}
	return dataOp
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
	dataOp := f.getDataOp(name, false)
	dataOp.memFiles[tag] = mf
	name = strings.TrimRight(name, "/")
	f.createParent(name, tag)
}

// Delete deletes a file from the file set.
// TODO: Directory deletion needs more invariant checks.
// Right now you have to specify the trailing slash explicitly.
func (f *UnorderedWriter) Delete(name string, customTag ...string) {
	name = CleanTarPath(name, IsDir(name))
	var tag string
	if len(customTag) > 0 {
		tag = customTag[0]
	}
	if tag == headerTag {
		f.getDataOp(name, true)
		return
	}
	dataOp := f.getDataOp(name, false)
	if _, ok := dataOp.deleteTags[headerTag]; !ok {
		dataOp.deleteTags[tag] = struct{}{}
	}
	delete(dataOp.memFiles, tag)
}

// PathsWritten returns the full paths (not prefixes) written by this UnorderedWriter
func (f *UnorderedWriter) PathsWritten() (ret []string) {
	for i := int64(0); i < f.subFileSet; i++ {
		p := path.Join(f.name, SubFileSetStr(i))
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
	var writerOpts []WriterOption
	if f.ttl > 0 {
		writerOpts = append(writerOpts, WithTTL(f.ttl))
	}
	p := path.Join(f.name, SubFileSetStr(f.subFileSet))
	w := f.storage.newWriter(f.ctx, p, writerOpts...)
	for _, name := range names {
		dataOp := f.fs[name]
		deleteTags := getSortedKeys(dataOp.deleteTags)
		mfs := getSortedMemFiles(dataOp.memFiles)
		if len(mfs) == 0 {
			if err := w.DeleteFile(name, deleteTags...); err != nil {
				return err
			}
			continue
		}
		hdr := mergeTarHeaders(mfs)
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
	if f.renewer != nil {
		f.renewer.Add(p)
	}

	// Reset in-memory file set.
	f.fs = make(map[string]*dataOp)
	f.memAvailable = f.memThreshold
	f.subFileSet++
	return nil
}

// TODO Tar header validation?
func mergeTarHeaders(memFiles []*memFile) *tar.Header {
	// TODO: Deeper copy?
	hdr := *memFiles[len(memFiles)-1].hdr
	for i := 0; i < len(memFiles)-1; i++ {
		hdr.Size += memFiles[i].hdr.Size
	}
	return &hdr
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

// Close closes the writer.
func (f *UnorderedWriter) Close() error {
	defer f.storage.filesetSem.Release(1)
	return f.serialize()
}
