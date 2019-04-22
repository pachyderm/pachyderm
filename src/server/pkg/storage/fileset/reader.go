package fileset

import (
	"archive/tar"
	"context"
	"io"

	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
)

type Reader struct {
	ctx     context.Context
	chunks  *chunk.Storage
	idx     []*tar.Reader
	hdrs    []*Header
	content *tar.Reader
	prefix  string
}

func NewReader(ctx context.Context, chunks *chunk.Storage, r io.Reader, prefix string) *Reader {
	return &Reader{
		ctx:    ctx,
		chunks: chunks,
		trs:    []*tar.Reader{tar.NewReader(r)},
		prefix: prefix,
	}
}

func (r *Reader) Next() (*Header, error) {
	lowestLevel := len(r.trs) - 1
	hdr, err := r.trs[lowestLevel].Next()
	if err != nil {
		if err == io.EOF {

		}
		return nil, err
	}
	info, _ := hdr.PAXRecords[MetaKey]
	for info.IdxRefs != nil {
		r.trs = append(r.trs, tar.NewReader(r))
	}
	r.chunks.NewReader(r.ctx, info.IdxRefs.
}

func (r *Reader) Read(data []byte) (int, error) {
	return r.trs[len(r.trs)-1].Read(data)
}
