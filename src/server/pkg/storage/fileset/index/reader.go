package index

import (
	"bytes"
	"context"
	"io"
	"math"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/fileset/tar"
)

type levelReader struct {
	cr *chunk.Reader
	tr *tar.Reader
}

// Reader is used for reading a multi-level index.
type Reader struct {
	ctx    context.Context
	objC   obj.Client
	chunks *chunk.Storage
	path   string
	prefix string
	levels []*levelReader
	done   bool
}

// NewReader create a new Reader.
func NewReader(ctx context.Context, objC obj.Client, chunks *chunk.Storage, path, prefix string) *Reader {
	return &Reader{
		ctx:    ctx,
		objC:   objC,
		chunks: chunks,
		path:   path,
		prefix: prefix,
	}
}

// Next gets the next header in the index.
func (r *Reader) Next() (*Header, error) {
	if r.done {
		return nil, io.EOF
	}
	if r.levels == nil {
		return r.setupLevels()
	}
	return r.next(len(r.levels) - 1)
}

func (r *Reader) setupLevels() (*Header, error) {
	// Setup top level.
	objR, err := r.objC.Reader(r.ctx, r.path, 0, 0)
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	if _, err := io.Copy(buf, objR); err != nil {
		return nil, err
	}
	if err := objR.Close(); err != nil {
		return nil, err
	}
	r.levels = []*levelReader{&levelReader{tr: tar.NewReader(buf)}}
	// Traverse until we reach the first entry in the lowest level.
	for {
		hdr, err := r.next(len(r.levels) - 1)
		if err != nil {
			return nil, err
		}
		// Return when we are at the lowest level.
		if hdr.Hdr.Typeflag == indexType {
			return hdr, nil
		}
		// Setup next level.
		// (bryce) this whole process of reading at the offset is janky, needs to be re-thought.
		cr := r.chunks.NewReader(r.ctx, r.callback(len(r.levels)))
		dataRef := hdr.Idx.DataOp.DataRefs[0]
		dataRef.OffsetBytes = hdr.Idx.Range.Offset
		dataRef.SizeBytes -= hdr.Idx.Range.Offset
		cr.NextRange([]*chunk.DataRef{dataRef})
		r.levels = append(r.levels, &levelReader{
			cr: cr,
			tr: tar.NewReader(cr),
		})
	}
}

func (r *Reader) next(level int) (*Header, error) {
	l := r.levels[level]
	for {
		hdr, err := l.tr.Next()
		if err != nil {
			return nil, err
		}
		// Handle lowest level index.
		if hdr.Typeflag == indexType {
			cmpSize := int64(math.Min(float64(len(hdr.Name)), float64(len(r.prefix))))
			cmp := strings.Compare(hdr.Name[:cmpSize], r.prefix[:cmpSize])
			// If a header with the prefix cannot show up after the current header,
			// then we are done.
			if cmp > 0 {
				r.done = true
				return nil, io.EOF
			} else if cmp != 0 {
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
		return fullHdr, nil
	}
}

func (r *Reader) callback(level int) chunk.ReaderFunc {
	return func() ([]*chunk.DataRef, error) {
		hdr, err := r.next(level - 1)
		if err != nil {
			return nil, err
		}
		return hdr.Idx.DataOp.DataRefs, nil
	}
}

// Close closes the reader.
func (r *Reader) Close() error {
	for i := 1; i < len(r.levels); i++ {
		if err := r.levels[i].cr.Close(); err != nil {
			return err
		}
	}
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
