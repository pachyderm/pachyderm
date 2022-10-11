package fileset

import (
	"context"
	"io"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset/index"
)

// Reader is an abstraction for reading a file set.
type Reader struct {
	store    MetadataStore
	chunks   *chunk.Storage
	idxCache *index.Cache
	id       ID
}

func newReader(store MetadataStore, chunks *chunk.Storage, idxCache *index.Cache, id ID) *Reader {
	return &Reader{
		store:    store,
		chunks:   chunks,
		idxCache: idxCache,
		id:       id,
	}
}

func (r *Reader) Iterate(ctx context.Context, cb func(File) error, opts ...index.Option) error {
	prim, err := r.getPrimitive(ctx)
	if err != nil {
		return err
	}
	ir := index.NewReader(r.chunks, r.idxCache, prim.Additive, opts...)
	return ir.Iterate(ctx, func(idx *index.Index) error {
		return cb(newFileReader(r.chunks, idx))
	})
}

func (r *Reader) getPrimitive(ctx context.Context) (*Primitive, error) {
	md, err := r.store.Get(ctx, r.id)
	if err != nil {
		return nil, err
	}
	prim := md.GetPrimitive()
	if prim == nil {
		return nil, errors.Errorf("file set %v is not primitive", r.id)
	}
	return prim, nil
}

func (r *Reader) IterateDeletes(ctx context.Context, cb func(File) error, opts ...index.Option) error {
	prim, err := r.getPrimitive(ctx)
	if err != nil {
		return err
	}
	ir := index.NewReader(r.chunks, r.idxCache, prim.Deletive, opts...)
	return ir.Iterate(ctx, func(idx *index.Index) error {
		return cb(newFileReader(r.chunks, idx))
	})
}

func (r *Reader) Shards(ctx context.Context) ([]*index.PathRange, error) {
	prim, err := r.getPrimitive(ctx)
	if err != nil {
		return nil, err
	}
	ir := index.NewReader(r.chunks, nil, prim.Additive)
	return ir.Shards(ctx)
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
