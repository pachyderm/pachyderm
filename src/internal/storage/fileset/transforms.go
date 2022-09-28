package fileset

import (
	"context"
	"io"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/chunk"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset/index"
)

var _ FileSet = &indexFilter{}

type indexFilter struct {
	fs        FileSet
	predicate func(*index.Index) bool
	full      bool
}

// NewIndexFilter filters fs using predicate.
func NewIndexFilter(fs FileSet, predicate func(idx *index.Index) bool, full ...bool) FileSet {
	idxf := &indexFilter{
		fs:        fs,
		predicate: predicate,
	}
	if len(full) > 0 {
		idxf.full = full[0]
	}
	return idxf
}

func (idxf *indexFilter) Iterate(ctx context.Context, cb func(File) error, opts ...index.Option) error {
	var dir string
	err := idxf.fs.Iterate(ctx, func(f File) error {
		idx := f.Index()
		if idxf.full {
			if dir != "" && strings.HasPrefix(idx.Path, dir) {
				return cb(f)
			}
			match := idxf.predicate(idx)
			if match && IsDir(idx.Path) {
				dir = idx.Path
			}
		}
		if idxf.predicate(idx) {
			return cb(f)
		}
		return nil
	}, opts...)
	return errors.EnsureStack(err)
}

func (idxf *indexFilter) IterateDeletes(_ context.Context, _ func(File) error, opts ...index.Option) error {
	return errors.Errorf("iterating deletes in an index filter file set")
}

func (idxf *indexFilter) Shards(_ context.Context) ([]*index.PathRange, error) {
	return nil, errors.Errorf("sharding an index filter file set")
}

var _ FileSet = &indexMapper{}

type indexMapper struct {
	fn func(idx *index.Index) *index.Index
	x  FileSet
}

// NewIndexMapper performs a map operation on the index entries of the files in the file set.
func NewIndexMapper(x FileSet, fn func(*index.Index) *index.Index) FileSet {
	return &indexMapper{x: x, fn: fn}
}

func (im *indexMapper) Iterate(ctx context.Context, cb func(File) error, opts ...index.Option) error {
	err := im.x.Iterate(ctx, func(fr File) error {
		y := im.fn(fr.Index())
		return cb(&indexMap{
			idx:   y,
			inner: fr,
		})
	}, opts...)
	return errors.EnsureStack(err)
}

func (im *indexMapper) IterateDeletes(_ context.Context, _ func(File) error, opts ...index.Option) error {
	return errors.Errorf("iterating deletes in an index mapper file set")
}

func (im *indexMapper) Shards(_ context.Context) ([]*index.PathRange, error) {
	return nil, errors.Errorf("sharding an index mapper file set")
}

var _ File = &indexMap{}

type indexMap struct {
	idx   *index.Index
	inner File
}

func (im *indexMap) Index() *index.Index {
	return im.idx
}

func (im *indexMap) Content(ctx context.Context, w io.Writer, opts ...chunk.ReaderOption) error {
	return im.inner.Content(ctx, w)
}

func (im *indexMap) Hash(ctx context.Context) ([]byte, error) {
	return im.inner.Hash(ctx)
}
