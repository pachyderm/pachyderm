package chunk

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"path"

	"github.com/chmduquesne/rollinghash/buzhash64"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/hash"
)

const (
	// MB is Megabytes.
	MB = 1024 * 1024
	// MinSize is the minimum chunk size.
	MinSize = 5 * MB
	// AverageBits determines the average chunk size (2^AverageBits).
	AverageBits = 23
	// MaxSize is the maximum chunk size.
	MaxSize = 15 * MB
	// WindowSize is the size of the rolling hash window.
	WindowSize = 64
)

// Writer splits written data into content defined chunks that are deduplicated/uploaded to object storage.
// Chunk split points are determined by a bit pattern in a rolling hash function (buzhash64 at https://github.com/chmduquesne/rollinghash).
// A min/max chunk size is also enforced to avoid creating chunks that are too small/big.
// (bryce) The chunking/hashing/uploading could be made concurrent by reading ahead a certain amount and splitting the data among chunking/hashing/uploading workers
// in a circular array where the first identified chunk (or whole chunk if there is no chunk split point) in a worker is appended to the prior workers data. This would
// handle chunk splits that show up when the rolling hash window is across the data splits. The callback would still be executed sequentially so that the order would be
// correct for the file index.
// - An improvement to this would be to just append WindowSize bytes to the prior worker's data, then stitch together the correct chunks.
//   It doesn't make sense to roll the window over the same data twice.
// - This does not work with the Max criteria, need to revisit.
type Writer struct {
	ctx       context.Context
	objC      obj.Client
	prefix    string
	f         func(string, io.Reader) error
	buf       *bytes.Buffer
	hash      *buzhash64.Buzhash64
	splitMask uint64
}

// newWriter creates a new Writer.
func newWriter(ctx context.Context, objC obj.Client, prefix string, f func(string, io.Reader) error) io.WriteCloser {
	// Initialize buzhash64 with WindowSize window.
	hash := buzhash64.New()
	hash.Write(make([]byte, WindowSize))
	return &Writer{
		ctx:       ctx,
		objC:      objC,
		prefix:    prefix,
		f:         f,
		buf:       &bytes.Buffer{},
		hash:      hash,
		splitMask: (1 << uint64(AverageBits)) - 1,
	}
}

// Write rolls through the data written, calling c.f when a chunk is found.
// Note: If making changes to this function, be wary of the performance
// implications (check before and after performance with chunker benchmarks).
func (w *Writer) Write(data []byte) (int, error) {
	offset := 0
	size := w.buf.Len()
	for i, b := range data {
		size++
		w.hash.Roll(b)
		// If the AverageBits lowest bits are zero and the chunk is >= MinSize, then a split point is created.
		// If the chunk is >= MaxSize, then a split point is created regardless.
		if w.hash.Sum64()&w.splitMask == 0 && size >= MinSize || size >= MaxSize {
			w.buf.Write(data[offset : i-offset+1])
			hash, err := w.put()
			if err != nil {
				return 0, err
			}
			if err := w.f(hash, w.buf); err != nil {
				return 0, err
			}
			w.buf.Reset()
			offset = i + 1
			size = 0
		}
	}
	w.buf.Write(data[offset:])
	return len(data), nil
}

func (w *Writer) put() (string, error) {
	hash := hash.EncodeHash(hash.Sum(w.buf.Bytes()))
	path := path.Join(w.prefix, hash)
	// If it does not exist, compress and write it.
	if !w.objC.Exists(w.ctx, path) {
		objW, err := w.objC.Writer(w.ctx, path)
		if err != nil {
			return "", err
		}
		defer objW.Close()
		gzipW := gzip.NewWriter(objW)
		defer gzipW.Close()
		// (bryce) Encrypt?
		if _, err := io.Copy(gzipW, bytes.NewReader(w.buf.Bytes())); err != nil {
			return "", err
		}
	}
	return hash, nil
}

func (w *Writer) Close() error {
	hash, err := w.put()
	if err != nil {
		return err
	}
	return w.f(hash, w.buf)
}
