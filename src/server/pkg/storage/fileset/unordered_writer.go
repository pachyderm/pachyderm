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
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/index"
	"github.com/pachyderm/pachyderm/src/server/pkg/tar"
)

type memFileOp struct {
	Op         index.Op
	memDataOps map[string]*memDataOp
}
type memDataOp struct {
	op      index.Op
	memFile *memFile
	tag     string
}

type memFile struct {
	hdr *tar.Header
	buf *bytes.Buffer
}

func (mf *memFile) Write(data []byte) (int, error) {
	mf.hdr.Size += int64(len(data))
	return mf.buf.Write(data)
}

// UnorderedWriter allows writing Files, unordered by path, into multiple ordered filesets.
// This may be a full filesystem or a subfilesystem (e.g. datum / datum set / shard).
type UnorderedWriter struct {
	ctx                        context.Context
	storage                    *Storage
	memAvailable, memThreshold int64
	name                       string
	defaultTag                 string
	fs                         map[string]*memFileOp
	subFileSet                 int64
	ttl                        time.Duration
	renewer                    *Renewer
}

func newUnorderedWriter(ctx context.Context, storage *Storage, name string, memThreshold int64, defaultTag string, opts ...UnorderedWriterOption) (*UnorderedWriter, error) {
	if err := storage.filesetSem.Acquire(ctx, 1); err != nil {
		return nil, err
	}
	uw := &UnorderedWriter{
		ctx:          ctx,
		storage:      storage,
		memAvailable: memThreshold,
		memThreshold: memThreshold,
		name:         name,
		defaultTag:   defaultTag,
		fs:           make(map[string]*memFileOp),
	}
	for _, opt := range opts {
		opt(uw)
	}
	return uw, nil
}

// Put reads files from a tar stream and adds them to the fileset.
// TODO: Make overwrite work with tags.
func (uw *UnorderedWriter) Put(r io.Reader, overwrite bool, customTag ...string) error {
	tag := uw.defaultTag
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
			uw.createParent(hdr.Name, overwrite, tag)
			continue
		}
		mf := uw.createMemFile(hdr, overwrite, tag)
		for {
			n, err := io.CopyN(mf, tr, uw.memAvailable)
			uw.memAvailable -= n
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return err
			}
			if uw.memAvailable == 0 {
				if err := uw.serialize(); err != nil {
					return err
				}
				mf = uw.createMemFile(hdr, overwrite, tag)
			}
		}
	}
}
func (uw *UnorderedWriter) createMemFile(hdr *tar.Header, overwrite bool, tag string) *memFile {
	uw.createParent(hdr.Name, overwrite, tag)
	hdr.Size = 0
	mf := &memFile{
		hdr: hdr,
		buf: &bytes.Buffer{},
	}
	var op index.Op
	if overwrite {
		op = index.Op_OVERWRITE
	}
	uw.createMemDataOp(hdr.Name, op, tag, mf)
	return mf
}

func (uw *UnorderedWriter) createMemDataOp(name string, op index.Op, tag string, memFile *memFile) {
	mfo := uw.getMemFileOp(name, op)
	mdo := &memDataOp{
		memFile: memFile,
		tag:     tag,
	}
	if _, ok := mfo.memDataOps[tag]; !ok {
		mfo.memDataOps[tag] = mdo
	}
	mfo.memDataOps[tag] = mergeMemDataOp(mfo.memDataOps[tag], mdo)
}

func (uw *UnorderedWriter) getMemFileOp(name string, op index.Op) *memFileOp {
	if _, ok := uw.fs[name]; !ok {
		uw.fs[name] = &memFileOp{
			memDataOps: make(map[string]*memDataOp),
		}
	}
	mfo := uw.fs[name]
	if op == index.Op_DELETE || op == index.Op_OVERWRITE {
		mfo.Op = op
		mfo.memDataOps = make(map[string]*memDataOp)
	}
	return mfo
}

func (uw *UnorderedWriter) createParent(name string, overwrite bool, tag string) {
	if name == "" {
		return
	}
	name, _ = path.Split(name)
	if _, ok := uw.fs[name]; ok {
		return
	}
	mf := &memFile{
		hdr: &tar.Header{
			Typeflag: tar.TypeDir,
			Name:     CleanTarPath(name, true),
		},
		buf: &bytes.Buffer{},
	}
	var op index.Op
	if overwrite {
		op = index.Op_OVERWRITE
	}
	uw.createMemDataOp(name, op, tag, mf)
	name = strings.TrimRight(name, "/")
	uw.createParent(name, overwrite, tag)
}

// Delete deletes a file from the file set.
// TODO: Directory deletion needs more invariant checks.
// Right now you have to specify the trailing slash explicitly.
func (uw *UnorderedWriter) Delete(name string, customTag ...string) {
	name = CleanTarPath(name, IsDir(name))
	var tag string
	if len(customTag) > 0 {
		tag = customTag[0]
	}
	if tag == headerTag {
		uw.getMemFileOp(name, index.Op_DELETE)
		return
	}
	uw.createMemDataOp(name, index.Op_DELETE, tag, nil)
}

// serialize will be called whenever the in-memory file set is past the memory threshold.
// A new in-memory file set will be created for the following operations.
func (uw *UnorderedWriter) serialize() error {
	// Sort names.
	names := make([]string, len(uw.fs))
	i := 0
	for name := range uw.fs {
		names[i] = name
		i++
	}
	sort.Strings(names)
	// Serialize file set.
	var writerOpts []WriterOption
	if uw.ttl > 0 {
		writerOpts = append(writerOpts, WithTTL(uw.ttl))
	}
	p := path.Join(uw.name, SubFileSetStr(uw.subFileSet))
	w := uw.storage.newWriter(uw.ctx, p, writerOpts...)
	for _, name := range names {
		fileOp := uw.fs[name]
		if err := writeMemFileOp(w, name, fileOp); err != nil {
			return err
		}
	}
	if err := w.Close(); err != nil {
		return err
	}
	if uw.renewer != nil {
		uw.renewer.Add(p)
	}
	// Reset in-memory file set.
	uw.fs = make(map[string]*memFileOp)
	uw.memAvailable = uw.memThreshold
	uw.subFileSet++
	return nil
}

// Close closes the writer.
func (uw *UnorderedWriter) Close() error {
	defer uw.storage.filesetSem.Release(1)
	return uw.serialize()
}

func mergeMemDataOp(mdo1, mdo2 *memDataOp) *memDataOp {
	switch mdo1.op {
	// Handle merge into append / overwrite operation.
	case index.Op_APPEND:
	case index.Op_OVERWRITE:
		switch mdo2.op {
		case index.Op_APPEND:
			hdr := mdo2.memFile.hdr
			hdr.Size += mdo1.memFile.hdr.Size
			mdo1.memFile.hdr = hdr
			mdo1.memFile.buf.Write(mdo2.memFile.buf.Bytes())
		case index.Op_OVERWRITE:
			mdo1.op = index.Op_OVERWRITE
			mdo1.memFile = mdo2.memFile
		case index.Op_DELETE:
			mdo1.op = index.Op_DELETE
			mdo1.memFile = nil
		}
	// Handle merge into delete operation.
	case index.Op_DELETE:
		if mdo2.op == index.Op_DELETE {
			return mdo1
		}
		mdo1.op = index.Op_OVERWRITE
		mdo1.memFile = mdo2.memFile
	}
	return mdo1
}

func writeMemFileOp(w *Writer, name string, memFileOp *memFileOp) error {
	var mdos []*memDataOp
	for _, mdo := range memFileOp.memDataOps {
		mdos = append(mdos, mdo)
	}
	sort.SliceStable(mdos, func(i, j int) bool {
		return mdos[i].tag < mdos[j].tag
	})
	switch memFileOp.Op {
	case index.Op_APPEND:
		return w.Append(name, func(fw *FileWriter) error {
			return writeMemDataOps(fw, mdos)
		})
	case index.Op_OVERWRITE:
		return w.Overwrite(name, func(fw *FileWriter) error {
			return writeMemDataOps(fw, mdos)
		})
	case index.Op_DELETE:
		w.Delete(name)
	}
	return nil
}

func writeMemDataOps(fw *FileWriter, memDataOps []*memDataOp) error {
	for _, mdo := range memDataOps {
		switch mdo.op {
		case index.Op_APPEND:
			if err := WithTarFileWriter(fw, mdo.memFile.hdr, func(tfw *TarFileWriter) error {
				fw.Append(mdo.tag)
				_, err := tfw.Write(mdo.memFile.buf.Bytes())
				return err
			}); err != nil {
				return err
			}
		case index.Op_OVERWRITE:
			if err := WithTarFileWriter(fw, mdo.memFile.hdr, func(tfw *TarFileWriter) error {
				tfw.Overwrite(mdo.tag)
				_, err := tfw.Write(mdo.memFile.buf.Bytes())
				return err
			}); err != nil {
				return err
			}
		case index.Op_DELETE:
			fw.Delete(mdo.tag)
		}
	}
	return nil
}
