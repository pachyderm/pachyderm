package fileset

import (
	"context"
	"fmt"
	"strings"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/internal/storage/fileset/index"
)

var _ FileSet = &pathValidator{}

type pathValidator struct {
	fs FileSet
}

// NewPathValidator checks if there is collision between file and dir names.
func NewPathValidator(fs FileSet) FileSet {
	pf := &pathValidator{
		fs: fs,
	}
	return pf
}

func (pf *pathValidator) Iterate(ctx context.Context, cb func(File) error, opts ...index.Option) error {
	ctx = pctx.Child(ctx, "indexFilter")
	var prev *index.Index
	var retErr error = nil
	if err := pf.fs.Iterate(ctx, func(f File) error {
		idx := f.Index()
		if retErr == nil {
			retErr = CheckIndex(prev, idx)
		}
		prev = idx
		return cb(f)
	}, opts...); err != nil {
		return errors.EnsureStack(err)
	}
	if retErr != nil {
		return errors.EnsureStack(retErr)
	}
	return nil
}

func (pf *pathValidator) IterateDeletes(_ context.Context, _ func(File) error, opts ...index.Option) error {
	return errors.Errorf("iterating deletes in an index filter file set")
}

func (pf *pathValidator) Shards(_ context.Context, _ ...index.Option) ([]*index.PathRange, error) {
	return nil, errors.Errorf("sharding an index filter file set")
}

func CheckIndex(prev, curr *index.Index) error {
	if prev == nil || curr == nil {
		return nil
	}
	if curr.Path == prev.Path {
		return errors.New(fmt.Sprintf("duplicate path output by different datums (%v from %v and %v from %v)", prev.Path, prev.File.Datum, curr.Path, curr.File.Datum))
	}
	if strings.HasPrefix(curr.Path, prev.Path+"/") {
		return errors.New(fmt.Sprintf("file / directory path collision (%v)", curr.Path))
	}
	return nil
}
