package fileset

import (
	"bytes"
	"context"
	"io"

	"github.com/gogo/protobuf/proto"
	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/index"
	"github.com/pachyderm/pachyderm/src/server/pkg/tar"
)

// Reader is an abstraction for reading a fileset.
type Reader struct {
	chunks *chunk.Storage
	ir     *index.Reader
}

func newReader(objC obj.Client, chunks *chunk.Storage, path string, opts ...index.Option) *Reader {
	return &Reader{
		chunks: chunks,
		ir:     index.NewReader(objC, chunks, path, opts...),
	}
}

// Iterate iterates over the files in the fileset.
func (r *Reader) Iterate(ctx context.Context, cb func(File) error) error {
	return r.ir.Iterate(ctx, func(idx *index.Index) error {
		return cb(newFileReader(ctx, r.chunks, idx))
	})
}

// FileReader is an abstraction for reading a file.
type FileReader struct {
	ctx    context.Context
	chunks *chunk.Storage
	idx    *index.Index
	hdr    *tar.Header
}

func newFileReader(ctx context.Context, chunks *chunk.Storage, idx *index.Index) *FileReader {
	return &FileReader{
		ctx:    ctx,
		chunks: chunks,
		idx:    proto.Clone(idx).(*index.Index),
	}
}

// Index returns the index for the file.
func (fr *FileReader) Index() *index.Index {
	return proto.Clone(fr.idx).(*index.Index)
}

// Header returns the tar header for the file.
func (fr *FileReader) Header() (*tar.Header, error) {
	if fr.hdr == nil {
		buf := &bytes.Buffer{}
		dataRefs := getHeaderDataOp(fr.idx.FileOp.DataOps).DataRefs
		r := fr.chunks.NewReader(fr.ctx, dataRefs)
		if err := r.Get(buf); err != nil {
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

// Content writes the content of the file.
func (fr *FileReader) Content(w io.Writer) error {
	dataRefs := getDataRefs(getContentDataOps(fr.idx.FileOp.DataOps))
	r := fr.chunks.NewReader(fr.ctx, dataRefs)
	return r.Get(w)
}
