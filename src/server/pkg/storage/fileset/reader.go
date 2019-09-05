package fileset

import (
	"context"
	"io"

	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/index"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/tar"
)

// Reader reads the serialized format of a fileset.
type Reader struct {
	ctx     context.Context
	chunks  *chunk.Storage
	ir      *index.Reader
	cr      *chunk.Reader
	tr      *tar.Reader
	hdr     *index.Header
	peekHdr *index.Header
	tags    []*index.Tag
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
	var hdr *index.Header
	if r.peekHdr != nil {
		hdr, r.peekHdr = r.peekHdr, nil
	} else {
		var err error
		hdr, err = r.next()
		if err != nil {
			return nil, err
		}
	}
	// Store header and reset reader for if a read is attempted.
	r.hdr = hdr
	r.tr = nil
	r.tags = hdr.Idx.DataOp.Tags[1:]
	return hdr, nil
}

func (r *Reader) next() (*index.Header, error) {
	hdr, err := r.ir.Next()
	if err != nil {
		return nil, err
	}
	// Convert index tar header to corresponding content tar header.
	indexToContentHeader(hdr)
	return hdr, nil
}

// Peek returns the next header without progressing the reader.
func (r *Reader) Peek() (*index.Header, error) {
	if r.peekHdr != nil {
		return r.peekHdr, nil
	}
	hdr, err := r.next()
	if err != nil {
		return nil, err
	}
	r.peekHdr = hdr
	return hdr, err
}

// PeekTag returns the next tag without progressing the reader.
func (r *Reader) PeekTag() (*index.Tag, error) {
	if len(r.tags) == 0 {
		return nil, io.EOF
	}
	return r.tags[0], nil
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
	if err := r.setupReader(); err != nil {
		return 0, err
	}
	return r.tr.Read(data)
}

func (r *Reader) setupReader() error {
	if r.tr == nil {
		r.cr.NextRange(r.hdr.Idx.DataOp.DataRefs)
		r.tr = tar.NewReader(r.cr)
		// Remove tar header from content stream.
		if _, err := r.tr.Next(); err != nil {
			return err
		}
	}
	return nil
}

// WriteToFiles does an efficient copy of files from a file set reader to writer.
// The optional pathBound specifies the upper bound (exclusive) for files to copy.
func (r *Reader) WriteToFiles(w *Writer, pathBound ...string) error {
	for {
		hdr, err := r.Peek()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if len(pathBound) > 0 && hdr.Hdr.Name >= pathBound[0] {
			return nil
		}
		if _, err := r.Next(); err != nil {
			return err
		}
		if err := w.WriteHeader(hdr); err != nil {
			return err
		}
		if err := r.WriteToTags(w); err != nil {
			return err
		}
	}
}

// WriteToTags does an efficient copy of tagged content in the current file from a file set reader to writer.
// The optional tagBound specifies the upper bound (inclusive, but duplicate tags should not exist) for tags to copy.
func (r *Reader) WriteToTags(w *Writer, tagBound ...string) error {
	// Lazily setup reader for underlying file.
	if err := r.setupReader(); err != nil {
		return err
	}
	var idx int
	var numBytes int64
	for i, tag := range r.tags {
		if len(tagBound) > 0 && tag.Id > tagBound[0] {
			break
		}
		idx = i + 1
		numBytes += tag.SizeBytes

	}
	err := r.cr.WriteToN(w.cw, numBytes)
	if err != nil {
		return err
	}
	if err := w.writeTags(r.tags[:idx]); err != nil {
		return err
	}
	r.tags = r.tags[idx:]
	return r.tr.Skip(numBytes)
}

// Close closes the reader.
func (r *Reader) Close() error {
	if err := r.cr.Close(); err != nil {
		return err
	}
	return r.ir.Close()
}
