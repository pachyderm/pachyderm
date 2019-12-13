package fileset

import (
	"bytes"
	"context"
	"io"

	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/index"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/tar"
)

// Reader reads the serialized format of a fileset.
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
	return newFileReader(idx, r.cr), nil
}

func (r *Reader) Iterate(f func(*FileReader) error, pathBound ...string) error {
	return r.ir.Iterate(func(idx *index.Index) error {
		r.cr.NextDataRefs(idx.DataOp.DataRefs)
		return f(newFileReader(idx, r.cr))
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
	hdr *tar.Header
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

func (fr *FileReader) Header() (*tar.Header, error) {
	if fr.hdr == nil {
		buf := &bytes.Buffer{}
		if err := fr.cr.NextTagReader().Get(buf); err != nil {
			return nil, err
		}
		var err error
		fr.hdr, err = tar.NewReader(buf).Next()
		return fr.hdr, err
	}
	return fr.hdr, nil
}

func (fr *FileReader) PeekTag() (*chunk.Tag, error) {
	return fr.cr.PeekTag()
}

func (fr *FileReader) NextTagReader() *chunk.TagReader {
	return fr.cr.NextTagReader()
}

func (fr *FileReader) Iterate(f func(*chunk.DataReader) error, tagUpperBound ...string) error {
	return fr.cr.Iterate(f, tagUpperBound...)
}

func (fr *FileReader) Get(w io.Writer) error {
	return fr.cr.Get(w)
}
