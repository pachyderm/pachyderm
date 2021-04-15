package fileset

import (
	"context"
	"io"
	"strings"

	globlib "github.com/pachyderm/ohmyglob"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset/index"
)

var _ FileSet = &globFilter{}

type globFilter struct {
	fileSet FileSet
	glob    string
	full    bool
}

func NewGlobFilter(fs FileSet, glob string, full ...bool) FileSet {
	gf := &globFilter{
		fileSet: fs,
		glob:    glob,
	}
	if len(full) > 0 {
		gf.full = full[0]
	}
	return gf
}

func (gf *globFilter) Iterate(ctx context.Context, cb func(File) error, _ ...bool) error {
	g, err := globlib.Compile(gf.glob, '/')
	if err != nil {
		return err
	}
	mf := func(path string) bool {
		// TODO: This does not seem like a good approach for this edge case.
		if path == "/" && gf.glob == "/" {
			return true
		}
		path = strings.TrimRight(path, "/")
		return g.Match(path)
	}
	var dir string
	return NewIndexFilter(gf.fileSet, func(idx *index.Index) bool {
		if gf.full {
			if dir != "" && strings.HasPrefix(idx.Path, dir) {
				return true
			}
			match := mf(idx.Path)
			if match && IsDir(idx.Path) {
				dir = idx.Path
			}
			return match
		}
		return mf(idx.Path)
	}).Iterate(ctx, cb)
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
