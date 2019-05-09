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
	cr *chunk.Reader
	tr *tar.Reader
}

// Reader is used for reading a multi-level index.
type Reader struct {
	ctx       context.Context
	chunks    *chunk.Storage
	prefix    string
	levels    []*levelReader
	currLevel int
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
		l := r.levels[r.currLevel]
		hdr, err := l.tr.Next()
		if err != nil {
			if err == io.EOF {
				// Potentially not done, return to the index level above.
				if r.currLevel > 0 {
					r.currLevel--
					continue
				}
				// We are done, so we should close our reader.
				if err := r.Close(); err != nil {
					return nil, err
				}
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
		if strings.Compare(fullHdr.Idx.Range.LastPath, r.prefix) < 0 {
			continue
		}
		// If a header with the prefix cannot show up after the current header,
		// then we are done.
		cmpSize := int64(math.Min(float64(len(hdr.Name)), float64(len(r.prefix))))
		if strings.Compare(hdr.Name[:cmpSize], r.prefix[:cmpSize]) > 0 {
			if err := r.Close(); err != nil {
				return nil, err
			}
			return nil, io.EOF
		}
		// Traverse into indexed tar stream.
		r.currLevel++
		// Create next level reader if it does not exist.
		if r.currLevel == len(r.levels) {
			r.levels = append(r.levels, &levelReader{
				cr: r.chunks.NewReader(r.ctx),
			})
		}
		// Set the next range.
		r.levels[r.currLevel].cr.NextRange(fullHdr.Idx.DataOp.DataRefs)
		r.levels[r.currLevel].tr = tar.NewReader(r.levels[r.currLevel].cr)
	}
}

// Close closes the reader.
func (r *Reader) Close() error {
	for i := 1; i < len(r.levels); i++ {
		if err := r.levels[i].cr.Close(); err != nil {
			return err
		}
	}
	r.levels = r.levels[:1]
	r.currLevel = 0
	return nil
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
		Hdr: hdr,
		Idx: idx,
	}, nil
}
