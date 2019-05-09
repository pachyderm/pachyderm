package fileset

import (
	"archive/tar"
	"context"
	"io"

	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/index"
)

// Reader reads the serialized format of a fileset.
type Reader struct {
	ctx    context.Context
	chunks *chunk.Storage
	ir     *index.Reader
	cr     *chunk.Reader
	tr     *tar.Reader
}

// NewReader creates a new Reader.
func NewReader(ctx context.Context, chunks *chunk.Storage, r io.Reader, prefix string) *Reader {
	cr := chunks.NewReader(ctx)
	return &Reader{
		ctx:    ctx,
		chunks: chunks,
		ir:     index.NewReader(ctx, chunks, r, prefix),
		cr:     cr,
	}
}

// Next returns the next header, and prepares the file's content for reading.
func (r *Reader) Next() (*index.Header, error) {
	hdr, err := r.ir.Next()
	if err != nil {
		return nil, err
	}
	r.cr.NextRange(hdr.Idx.DataOp.DataRefs)
	r.tr = tar.NewReader(r.cr)
	// Remove tar header from content stream.
	if _, err := r.tr.Next(); err != nil {
		return nil, err
	}
	return hdr, nil
}

// Read reads from the current file in the tar stream.
func (r *Reader) Read(data []byte) (int, error) {
	return r.tr.Read(data)
}
