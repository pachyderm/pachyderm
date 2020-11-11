package fileset

import (
	"context"
	"io"

	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/index"
	"github.com/pachyderm/pachyderm/src/server/pkg/tar"
)

var _ FileSet = &headerFilter{}

type headerFilter struct {
	pred func(th *tar.Header) bool
	x    FileSet
}

// NewHeaderFilter filters x using pred
func NewHeaderFilter(x FileSet, pred func(th *tar.Header) bool) FileSet {
	return &headerFilter{x: x, pred: pred}
}

func (hf *headerFilter) Iterate(ctx context.Context, cb func(File) error) error {
	return hf.x.Iterate(ctx, func(fr File) error {
		th, err := fr.Header()
		if err != nil {
			return err
		}
		if hf.pred(th) {
			return cb(fr)
		}
		return nil
	})
}

var _ FileSet = &indexFilter{}

type indexFilter struct {
	pred func(idx *index.Index) bool
	x    FileSet
}

// NewIndexFilter filters x using pred
func NewIndexFilter(x FileSet, pred func(idx *index.Index) bool) FileSet {
	return &indexFilter{x: x, pred: pred}
}

func (fil *indexFilter) Iterate(ctx context.Context, cb func(File) error) error {
	return fil.x.Iterate(ctx, func(fr File) error {
		idx := fr.Index()
		if fil.pred(idx) {
			return cb(fr)
		}
		return nil
	})
}

// NewDeleteFilter creates an index filter that filters out deleted files.
func NewDeleteFilter(x FileSet) FileSet {
	return NewIndexFilter(x, func(idx *index.Index) bool {
		if idx.FileOp.Op == index.Op_DELETE {
			return false
		}
		if len(idx.FileOp.DataRefs) == 0 && len(getDataRefs(idx.FileOp.DataOps)) == 0 {
			return false
		}
		return true
	})
}

var _ FileSet = &headerMapper{}

type headerMapper struct {
	fn func(th *tar.Header) *tar.Header
	x  FileSet
}

// NewHeaderMapper filters x using pred
func NewHeaderMapper(x FileSet, fn func(*tar.Header) *tar.Header) FileSet {
	return &headerMapper{x: x, fn: fn}
}

func (hm *headerMapper) Iterate(ctx context.Context, cb func(File) error) error {
	return hm.x.Iterate(ctx, func(fr File) error {
		x, err := fr.Header()
		if err != nil {
			return err
		}
		y := hm.fn(x)
		return cb(headerMap{
			header: y,
			inner:  fr,
		})
	})
}

var _ File = headerMap{}

type headerMap struct {
	header *tar.Header
	inner  File
}

func (hm headerMap) Index() *index.Index {
	return nil
}

func (hm headerMap) Header() (*tar.Header, error) {
	return hm.header, nil
}

func (hm headerMap) Content(w io.Writer) error {
	return hm.inner.Content(w)
}
