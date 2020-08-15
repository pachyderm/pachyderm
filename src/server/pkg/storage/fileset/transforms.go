package fileset

import (
	"context"
	"io"

	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/index"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/tar"
)

var _ FileSource = &headerFilter{}

type headerFilter struct {
	pred func(th *tar.Header) bool
	x    FileSource
}

// NewHeaderFilter filters x using pred
func NewHeaderFilter(x FileSource, pred func(th *tar.Header) bool) FileSource {
	return &headerFilter{x: x, pred: pred}
}

func (hf *headerFilter) Iterate(ctx context.Context, cb func(File) error, stopBefore ...string) error {
	return hf.x.Iterate(ctx, func(fr File) error {
		th, err := fr.Header()
		if err != nil {
			return err
		}
		if hf.pred(th) {
			return cb(fr)
		}
		return nil
	}, stopBefore...)
}

var _ FileSource = &indexFilter{}

type indexFilter struct {
	pred func(idx *index.Index) bool
	x    FileSource
}

// NewIndexFilter filters x using pred
func NewIndexFilter(x FileSource, pred func(idx *index.Index) bool) FileSource {
	return &indexFilter{x: x, pred: pred}
}

func (fil *indexFilter) Iterate(ctx context.Context, cb func(File) error, stopBefore ...string) error {
	return fil.x.Iterate(ctx, func(fr File) error {
		idx := fr.Index()
		if fil.pred(idx) {
			cb(fr)
		}
		return nil
	})
}

var _ FileSource = &headerMapper{}

type headerMapper struct {
	fn func(th *tar.Header) *tar.Header
	x  FileSource
}

// NewHeaderMapper filters x using pred
func NewHeaderMapper(x FileSource, fn func(*tar.Header) *tar.Header) FileSource {
	return &headerMapper{x: x, fn: fn}
}

func (hm *headerMapper) Iterate(ctx context.Context, cb func(File) error, stopBefore ...string) error {
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
