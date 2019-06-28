package fileset

import (
	"context"

	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/index"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/tar"
)

// Reader reads the serialized format of a fileset.
type Reader struct {
	ctx    context.Context
	chunks *chunk.Storage
	ir     *index.Reader
	cr     *chunk.Reader
	tr     *tar.Reader
	hdr    *index.Header
}

func newReader(ctx context.Context, objC obj.Client, chunks *chunk.Storage, path, prefix string) *Reader {
	cr := chunks.NewReader(ctx)
	return &Reader{
		ctx:    ctx,
		chunks: chunks,
		ir:     index.NewReader(ctx, objC, chunks, path, prefix),
		cr:     cr,
	}
}

// Next returns the next header, and prepares the file's content for reading.
func (r *Reader) Next() (*index.Header, error) {
	hdr, err := r.ir.Next()
	if err != nil {
		return nil, err
	}
	// Convert index tar header to corresponding content tar header.
	indexToContentHeader(hdr)
	// Store header and reset reader for if a read is attempted.
	r.hdr = hdr
	r.tr = nil
	return hdr, nil
}

func indexToContentHeader(idx *index.Header) {
	idx.Hdr = &tar.Header{
		Name: idx.Hdr.Name,
		Size: idx.Idx.SizeBytes,
	}
}

// Read reads from the current file in the tar stream.
func (r *Reader) Read(data []byte) (int, error) {
	// Lazily setup reader for underlying file.
	if r.tr == nil {
		if err := r.setupReader(); err != nil {
			return 0, err
		}
	}
	return r.tr.Read(data)
}

func (r *Reader) setupReader() error {
	r.cr.NextRange(r.hdr.Idx.DataOp.DataRefs)
	r.tr = tar.NewReader(r.cr)
	// Remove tar header from content stream.
	if _, err := r.tr.Next(); err != nil {
		return err
	}
	return nil
}
