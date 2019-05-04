package index

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"math"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
)

type levelReader struct {
	cr io.ReadCloser
	tr *tar.Reader
}

// Reader is used for reading a multi-level index.
type Reader struct {
	ctx    context.Context
	chunks *chunk.Storage
	prefix string
	levels []*levelReader
}

// NewReader create a new Reader.
func NewReader(ctx context.Context, chunks *chunk.Storage, r io.Reader, prefix string) *Reader {
	return &Reader{
		ctx:    ctx,
		chunks: chunks,
		prefix: prefix,
		levels: []*levelReader{&levelReader{tr: tar.NewReader(r)}},
	}
}

// Next gets the next header in the index.
func (r *Reader) Next() (*Header, error) {
	for {
		l := r.levels[len(r.levels)-1]
		hdr, err := l.tr.Next()
		if err != nil {
			if err == io.EOF && len(r.levels) > 1 {
				if err := l.cr.Close(); err != nil {
					return nil, err
				}
				r.levels = r.levels[:len(r.levels)-1]
				continue
			}
			return nil, err
		}
		// Handle lowest level index.
		if hdr.Typeflag == indexType {
			if !strings.HasPrefix(hdr.Name, r.prefix) {
				continue
			}
			return deserialize(l.tr, hdr)
		}
		// Handle index level above lowest.
		fullHdr, err := deserialize(l.tr, hdr)
		if err != nil {
			return nil, err
		}
		// Skip to the starting header.
		if strings.Compare(fullHdr.idx.Range.LastPath, r.prefix) < 0 {
			continue
		}
		// If a header with the prefix cannot show up after the current header,
		// then we are done.
		cmpSize := int64(math.Min(float64(len(hdr.Name)), float64(len(r.prefix))))
		if strings.Compare(hdr.Name[:cmpSize], r.prefix[:cmpSize]) > 0 {
			for i := 1; i < len(r.levels); i++ {
				if err := l.cr.Close(); err != nil {
					return nil, err
				}
			}
			r.levels = r.levels[:1]
			return nil, io.EOF
		}
		// Setup next level.
		cr := r.chunks.NewReader(r.ctx, fullHdr.idx.DataOp.DataRefs)
		r.levels = append(r.levels, &levelReader{
			cr: cr,
			tr: tar.NewReader(cr),
		})
	}
}

func deserialize(tr *tar.Reader, hdr *tar.Header) (*Header, error) {
	data := &bytes.Buffer{}
	buf := grpcutil.GetBuffer()
	defer grpcutil.PutBuffer(buf)
	if _, err := io.CopyBuffer(data, tr, buf); err != nil {
		return nil, err
	}
	idx := &Index{}
	if err := proto.Unmarshal(data.Bytes(), idx); err != nil {
		return nil, err
	}
	return &Header{
		hdr: hdr,
		idx: idx,
	}, nil
}
