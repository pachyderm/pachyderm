package fileset

import (
	"context"
	"io"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset/index"
)

// Reader is an abstraction for reading a fileset.
type Reader struct {
	store     MetadataStore
	chunks    *chunk.Storage
	idxCache  *index.Cache
	id        ID
	indexOpts []index.Option
}

func newReader(store MetadataStore, chunks *chunk.Storage, idxCache *index.Cache, id ID, opts ...index.Option) *Reader {
	r := &Reader{
		store:     store,
		chunks:    chunks,
		idxCache:  idxCache,
		id:        id,
		indexOpts: opts,
	}
	return r
}

// Iterate iterates over the files in the file set.
func (r *Reader) Iterate(ctx context.Context, cb func(File) error, deletive ...bool) error {
	md, err := r.store.Get(ctx, r.id)
	if err != nil {
		return err
	}
	prim := md.GetPrimitive()
	if prim == nil {
		return errors.Errorf("fileset %v is not primitive", r.id)
	}
	if len(deletive) > 0 && deletive[0] {
		ir := index.NewReader(r.chunks, r.idxCache, prim.Deletive, r.indexOpts...)
		return ir.Iterate(ctx, func(idx *index.Index) error {
			return cb(newFileReader(r.chunks, idx))
		})
	}
	ir := index.NewReader(r.chunks, r.idxCache, prim.Additive, r.indexOpts...)
	return ir.Iterate(ctx, func(idx *index.Index) error {
		return cb(newFileReader(r.chunks, idx))
	})
}

// FileReader is an abstraction for reading a file.
type FileReader struct {
	chunks *chunk.Storage
	idx    *index.Index
}

func newFileReader(chunks *chunk.Storage, idx *index.Index) *FileReader {
	return &FileReader{
		chunks: chunks,
		idx:    idx,
	}
}

// Index returns the index for the file.
func (fr *FileReader) Index() *index.Index {
	return fr.idx
}

// Content writes the content of the file.
func (fr *FileReader) Content(ctx context.Context, w io.Writer, opts ...chunk.ReaderOption) error {
	r := fr.chunks.NewReader(ctx, fr.idx.File.DataRefs, opts...)
	return r.Get(w)
}

// Hash returns the hash of the file.
func (fr *FileReader) Hash(_ context.Context) ([]byte, error) {
	var hashes [][]byte
	for _, dataRef := range fr.idx.File.DataRefs {
		hashes = append(hashes, dataRef.Hash)
	}
	return computeFileHash(hashes)
}
