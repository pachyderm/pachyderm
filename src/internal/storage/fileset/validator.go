package fileset

import (
	"context"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset/index"
)

var _ FileSet = &pathValidator{}

type pathValidator struct {
	fs FileSet
}

// NewPathValidator validates the paths in a file set.
func NewPathValidator(fs FileSet) FileSet {
	return &pathValidator{
		fs: fs,
	}
}

func (pv *pathValidator) Iterate(ctx context.Context, cb func(File) error, opts ...index.Option) error {
	ctx = pctx.Child(ctx, "pathIterator")
	var prev *index.Index
	if err := pv.fs.Iterate(ctx, func(f File) error {
		idx := f.Index()
		if err := CheckIndex(prev, idx); err != nil {
			return err
		}
		prev = idx
		return cb(f)
	}, opts...); err != nil {
		return errors.EnsureStack(err)
	}
	return nil
}

func (pv *pathValidator) IterateDeletes(_ context.Context, _ func(File) error, opts ...index.Option) error {
	return errors.Errorf("iterating deletes in a path validator file set")
}

func (pv *pathValidator) Shards(_ context.Context, _ ...index.Option) ([]*index.PathRange, error) {
	return nil, errors.Errorf("sharding a path validator file set")
}

func CheckIndex(prev, curr *index.Index) error {
	if prev == nil || curr == nil {
		return nil
	}
	if curr.Path == prev.Path {
		return errors.Errorf("duplicate path output by different datums (%v from %v and %v from %v)", prev.Path, prev.File.Datum, curr.Path, curr.File.Datum)
	}
	if strings.HasPrefix(curr.Path, prev.Path+"/") {
		return errors.Errorf("file / directory path collision (%v)", curr.Path)
	}
	return nil
}
