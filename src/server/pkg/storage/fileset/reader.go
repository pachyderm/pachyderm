package fileset

import (
	"context"
	"io"

	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/index"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/tar"
)

// Reader reads the serialized format of a file set.
type Reader struct {
	ctx context.Context
	ir  *index.Reader
	cr  *chunk.Reader
}

func newReader(ctx context.Context, objC obj.Client, chunks *chunk.Storage, path string, opts ...index.Option) *Reader {
	cr := chunks.NewReader(ctx)
	return &Reader{
		ctx: ctx,
		ir:  index.NewReader(ctx, objC, chunks, path, opts...),
		cr:  cr,
	}
}

func (r *Reader) Peek() (*index.Index, error) {
	return r.ir.Peek()
}

func (r *Reader) Next() (*FileReader, error) {
	idx, err := r.ir.Next()
	if err != nil {
		return nil, err
	}
	r.cr.NextDataRefs(idx.DataOp.DataRefs)
	return &FileReader{
		idx: idx,
		cr:  r.cr,
	}
}

func (r *Reader) Iterate(f func(*FileReader) error, pathBound ...string) error {
	return r.ir.Iterate(func(idx *index.Index) error {
		r.cr.NextDataRefs(idx.DataOp.DataRefs)
		return f(&FileReader{
			idx: idx,
			cr:  r.cr,
		})
	}, pathBound...)
}

func (r *Reader) Get(w io.Writer) error {
	return r.Iterate(func(fr *FileReader) error {
		return fr.Get(w)
	})
}

type FileReader struct {
	idx *index.Index
	cr  *chunk.Reader
}

func newFileReader(idx *index.Index, cr *chunk.Reader) *FileReader {
	return &FileReader{
		idx: idx,
		cr:  cr,
	}
}

func (fr *FileReader) Index() *index.Index {
	return fr.idx
}

func (fr *FileReader) Get(w io.Writer) error {
	return fr.cr.Get(w)
}

func (fr *FileReader) TagReader() *TagReader {
	return newTagReader(fr.idx.DataOp.Tags, fr.cr)
}

type TagReader struct {
	tags []*index.Tag
	cr   *chunk.Reader
	hdr  *tar.Header
}

func newTagReader(tags []*index.Tag, cr *chunk.Reader) *TagReader {
	return &TagReader{
		tags: tags,
		cr:   cr,
	}
}

func (tr *TagReader) Header() (*tar.Header, error) {
	if tr.hdr == nil {
		var err error
		tr.hdr, err = tar.NewReader(tr.cr).Next()
		if err != nil {
			return nil, err
		}
		tr.tags = tr.tags[1:]
	}
	return tr.hdr, nil
}

func (tr *TagReader) Peek() (*index.Tag, error) {
	if _, err := tr.Header(); err != nil {
		return nil, err
	}
	if len(tr.tags) == 0 {
		return nil, io.EOF
	}
	return tr.tags[0], nil
}

func (tr *TagReader) Iterate(f func(*index.Tag, io.Reader) error, tagBound ...string) error {
	for {
		tag, err := tr.Peek()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if !index.BeforeBound(tag.Id, tagBound...) {
			return nil
		}
		if err := f(tag, io.LimitReader(tr.cr, tag.SizeBytes)); err != nil {
			return err
		}
		tr.tags = tr.tags[1:]
	}
}
