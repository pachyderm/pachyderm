package fileset

import (
	"bytes"
	"context"
	"io"

	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/index"
	"github.com/pachyderm/pachyderm/src/server/pkg/tar"
)

// Reader reads the serialized format of a fileset.
type Reader struct {
	ctx context.Context
	ir  *index.Reader
	cr  *chunk.Reader
}

func newReader(ctx context.Context, store Store, chunks *chunk.Storage, p string, opts ...index.Option) (*Reader, error) {
	topIdx, err := store.GetIndex(ctx, p)
	if err != nil {
		return nil, err
	}
	cr := chunks.NewReader(ctx)
	return &Reader{
		ctx: ctx,
		ir:  index.NewReader(ctx, topIdx, chunks, opts...),
		cr:  cr,
	}, nil
}

// Peek returns the next file index without progressing the reader.
func (r *Reader) Peek() (*index.Index, error) {
	return r.ir.Peek()
}

// Next returns the next file reader and progresses the reader.
func (r *Reader) Next() (*FileReader, error) {
	idx, err := r.ir.Next()
	if err != nil {
		return nil, err
	}
	r.cr.NextDataRefs(idx.DataOp.DataRefs)
	return newFileReader(idx, r.cr), nil
}

// Iterate iterates over the file readers in the fileset.
// pathBound is an optional parameter for specifiying the upper bound (exclusive) of the iteration.
func (r *Reader) Iterate(ctx context.Context, f func(File) error, pathBound ...string) error {
	return r.iterate(func(fr *FileReader) error {
		return f(fr)
	}, pathBound...)
}

func (r *Reader) iterate(f func(*FileReader) error, pathBound ...string) error {
	return r.ir.Iterate(func(idx *index.Index) error {
		r.cr.NextDataRefs(idx.DataOp.DataRefs)
		return f(newFileReader(idx, r.cr))
	}, pathBound...)
}

// Get writes the fileset.
func (r *Reader) Get(w io.Writer) error {
	return r.iterate(func(fr *FileReader) error {
		return fr.Get(w)
	})
}

// FileReader is an abstraction for reading a file.
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

// Index returns the index for the file.
func (fr *FileReader) Index() *index.Index {
	return fr.idx
}

// Header returns the tar header for the file.
func (fr *FileReader) Header() (*tar.Header, error) {
	if fr.hdr == nil {
		buf := &bytes.Buffer{}
		if err := fr.cr.NextTagReader().Get(buf); err != nil {
			return nil, err
		}
		hdr, err := tar.NewReader(buf).Next()
		if err != nil {
			return nil, err
		}
		if !IsCleanTarPath(hdr.Name, hdr.FileInfo().IsDir()) {
			return nil, errors.Errorf("uncleaned tar header name: %s", hdr.Name)
		}
		fr.hdr = hdr
	}
	return fr.hdr, nil
}

// PeekTag returns the next tag in the file without progressing the reader.
func (fr *FileReader) PeekTag() (*chunk.Tag, error) {
	return fr.cr.PeekTag()
}

// NextTagReader returns a tag reader for the next tagged data in the file.
func (fr *FileReader) NextTagReader() *chunk.TagReader {
	return fr.cr.NextTagReader()
}

// Iterate iterates over the data readers for the data in the file.
// tagUpperBound is an optional parameter for specifiying the upper bound (exclusive) of the iteration.
func (fr *FileReader) Iterate(f func(*chunk.DataReader) error, tagUpperBound ...string) error {
	return fr.cr.Iterate(f, tagUpperBound...)
}

// Get writes the file.
func (fr *FileReader) Get(w io.Writer) error {
	if _, err := fr.Header(); err != nil {
		return err
	}
	return fr.cr.Get(w)
}

// Content writes the contents of the file excluding the header to w.
func (fr *FileReader) Content(w io.Writer) error {
	return errors.New("FileReader.GetContents not implemented")
}
