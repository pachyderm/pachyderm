package index

import (
	"archive/tar"
	"context"
	"encoding/base64"
	"io"
	"math"
	"strings"

	"github.com/gogo/protobuf/proto"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/chunk"
)

// (bryce) this might make more sense as a general abstraction outside of this package (tbd).
type levelReader struct {
	cr  io.ReadCloser
	tr  *tar.Reader
	hdr *tar.Header
}

// Peek peeks ahead.
func (r *levelReader) Peek() (*tar.Header, error) {
	if r.hdr != nil {
		return r.hdr, nil
	}
	var err error
	r.hdr, err = r.next()
	return r.hdr, err
}

// Next gets the next header.
func (r *levelReader) Next() (*tar.Header, error) {
	if r.hdr != nil {
		tmp := r.hdr
		r.hdr = nil
		return tmp, nil
	}
	return r.next()
}

func (r *levelReader) next() (*tar.Header, error) {
	hdr, err := r.tr.Next()
	if err != nil {
		return nil, err
	}
	return hdr, err
}

// Close closes the reader.
func (r *levelReader) Close() error {
	return r.cr.Close()
}

// Reader is used for reading a multi-level index.
type Reader struct {
	ctx    context.Context
	chunks *chunk.Storage
	prefix string
	levels []*levelReader
}

// NewReader create a new IndexReader.
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
		level := r.levels[len(r.levels)-1]
		hdr, err := level.Next()
		if err != nil {
			if err == io.EOF && len(r.levels) > 1 {
				if err := level.cr.Close(); err != nil {
					return nil, err
				}
				r.levels = r.levels[:len(r.levels)-1]
				continue
			}
			return nil, err
		}
		// Handle the base level index.
		if _, ok := hdr.PAXRecords[base]; ok {
			// Return the header if it has the right prefix.
			if strings.HasPrefix(hdr.Name, r.prefix) {
				return deserialize(hdr)
			}
			continue
		}
		// Handle the indexes above the base level.
		// Skip to the starting header (last header that is before
		// or equal to the prefix).
		if hdr, err = r.skipAhead(hdr); err != nil {
			return nil, err
		}
		// If a header with the prefix cannot show up after the current header,
		// then we are done.
		cmpSize := int64(math.Min(float64(len(hdr.Name)), float64(len(r.prefix))))
		if strings.Compare(hdr.Name[:cmpSize], r.prefix[:cmpSize]) > 0 {
			for i := 1; i < len(r.levels); i++ {
				if err := level.Close(); err != nil {
					return nil, err
				}
			}
			return nil, io.EOF
		}
		// Setup next level.
		fullHdr, err := deserialize(hdr)
		if err != nil {
			return nil, err
		}
		cr := r.chunks.NewReader(r.ctx, opsToRefs(fullHdr.info.DataOps))
		r.levels = append(r.levels, &levelReader{
			cr: cr,
			tr: tar.NewReader(cr),
		})
	}
}

func (r *Reader) skipAhead(hdr *tar.Header) (*tar.Header, error) {
	level := r.levels[len(r.levels)-1]
	for {
		nextHdr, err := level.Peek()
		if err != nil {
			if err == io.EOF {
				return hdr, nil
			}
			return nil, err
		}
		if strings.Compare(nextHdr.Name, r.prefix) > 0 {
			return hdr, nil
		}
		hdr, err = level.Next()
		if err != nil {
			return nil, err
		}
	}
}

func deserialize(hdr *tar.Header) (*Header, error) {
	bytes, err := base64.StdEncoding.DecodeString(hdr.PAXRecords[fileInfo])
	if err != nil {
		return nil, err
	}
	fileInfo := &FileInfo{}
	if err := proto.Unmarshal(bytes, fileInfo); err != nil {
		return nil, err
	}
	return &Header{
		hdr:  hdr,
		info: fileInfo,
	}, nil
}
