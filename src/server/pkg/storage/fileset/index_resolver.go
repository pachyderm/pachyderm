package fileset

import (
	"context"

	"github.com/pachyderm/pachyderm/src/client/pkg/errors"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/index"
)

// NewIndexResolver ensures the indexes in the FileSource are correct
// based on the content
func NewIndexResolver(x FileSource) FileSource {
	switch x := x.(type) {
	case *mergeSource:
		return &mergeResolver{
			getReader: x.getReader,
			s:         x.s,
		}
	default:
		panic("cannot resolve indexes")
	}
}

type mergeResolver struct {
	s         *Storage
	getReader func() (*MergeReader, error)
}

func (mr *mergeResolver) Iterate(ctx context.Context, cb func(File) error, stopBefore ...string) error {
	mr1, err := mr.getReader()
	if err != nil {
		return err
	}
	mr2, err := mr.getReader()
	if err != nil {
		return err
	}
	w := mr.s.newWriter(ctx, "", WithNoUpload(), WithIndexCallback(func(idx *index.Index) error {
		// Index entries that do not reference any data are for propagating deletes.
		if len(idx.DataOp.DataRefs) == 0 {
			return nil
		}
		fmr, err := mr2.Next()
		if err != nil {
			return err
		}
		if fmr.Index().Path != idx.Path {
			return errors.Errorf("merge resolver has been given 2 different merge readers")
		}
		fmr.fullIdx = idx
		return cb(fmr)
	}))
	if err := mr1.WriteTo(w); err != nil {
		return err
	}
	return w.Close()
}
