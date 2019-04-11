package chunk

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"io/ioutil"
	"path"

	"github.com/pachyderm/pachyderm/src/client/pkg/grpcutil"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
)

// Storage is the abstraction that manages chunk storage.
type Storage struct {
	objC   obj.Client
	prefix string
}

// NewStorage creates a new Storage.
func NewStorage(objC obj.Client, prefix string) *Storage {
	return &Storage{
		objC:   objC,
		prefix: prefix,
	}
}

// NewReader creates an io.ReadCloser for a chunk.
// (bryce) The whole chunk is in-memory right now. Could be a problem with
// concurrency, particularly the merge process.
// May want to handle concurrency here (pass in multiple data refs)
func (s *Storage) NewReader(ctx context.Context, dataRef *DataRef) (io.ReadCloser, error) {
	objR, err := s.objC.Reader(ctx, path.Join(s.prefix, dataRef.Hash), 0, 0)
	if err != nil {
		return nil, err
	}
	defer objR.Close()
	gzipR, err := gzip.NewReader(objR)
	if err != nil {
		return nil, err
	}
	defer gzipR.Close()
	chunk := &bytes.Buffer{}
	buf := grpcutil.GetBuffer()
	defer grpcutil.PutBuffer(buf)
	if _, err := io.CopyBuffer(chunk, gzipR, buf); err != nil {
		return nil, err
	}
	data := chunk.Bytes()[dataRef.Offset : dataRef.Offset+dataRef.Size]
	return ioutil.NopCloser(bytes.NewReader(data)), nil
}

// NewWriter creates an io.WriteCloser for a stream of bytes to be chunked.
// Chunks are created based on the content, then hashed and deduplicated/uploaded to
// object storage.
// The callback arguments are the chunk hash and content.
func (s *Storage) NewWriter(ctx context.Context, f func([]*DataRef) error) *Writer {
	return newWriter(ctx, s.objC, s.prefix, f)
}

// Clear deletes all of the chunks in object storage.
func (s *Storage) Clear(ctx context.Context) error {
	return s.objC.Walk(ctx, s.prefix, func(hash string) error {
		return s.objC.Delete(ctx, hash)
	})
}
