package fileset

import (
	"context"
	"io"

	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/index"
)

var _ FileSet = &indexFilter{}

type indexFilter struct {
	pred func(idx *index.Index) bool
	x    FileSet
}

// NewIndexFilter filters x using pred
func NewIndexFilter(x FileSet, pred func(idx *index.Index) bool) FileSet {
	return &indexFilter{x: x, pred: pred}
}

func (fil *indexFilter) Iterate(ctx context.Context, cb func(File) error, _ ...bool) error {
	return fil.x.Iterate(ctx, func(fr File) error {
		idx := fr.Index()
		if fil.pred(idx) {
			return cb(fr)
		}
		return nil
	})
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

func (im *indexMapper) Iterate(ctx context.Context, cb func(File) error, _ ...bool) error {
	return im.x.Iterate(ctx, func(fr File) error {
		y := im.fn(fr.Index())
		return cb(&indexMap{
			idx:   y,
			inner: fr,
		})
	})
}

var _ File = &indexMap{}

type indexMap struct {
	idx   *index.Index
	inner File
}

func (im *indexMap) Index() *index.Index {
	return im.idx
}

func (im *indexMap) Content(w io.Writer) error {
	return im.inner.Content(w)
}
