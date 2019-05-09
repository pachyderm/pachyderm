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
	// AverageBits determines the average chunk size (2^AverageBits).
	AverageBits = 23
	// WindowSize is the size of the rolling hash window.
	WindowSize = 64
)

var initialWindow = make([]byte, WindowSize)

// Writer splits a byte stream into content defined chunks that are hashed and deduplicated/uploaded to object storage.
// Chunk split points are determined by a bit pattern in a rolling hash function (buzhash64 at https://github.com/chmduquesne/rollinghash).
// (bryce) The chunking/hashing/uploading could be made concurrent by reading ahead a certain amount and splitting the data among chunking/hashing/uploading workers
// in a circular array where the first identified chunk (or whole chunk if there is no chunk split point) in a worker is appended to the prior workers data. This would
// handle chunk splits that show up when the rolling hash window is across the data splits. The callback would still be executed sequentially so that the order would be
// correct for the file index.
// - An improvement to this would be to just append WindowSize bytes to the prior worker's data, then stitch together the correct chunks.
//   It doesn't make sense to roll the window over the same data twice.
type Writer struct {
	ctx        context.Context
	objC       obj.Client
	prefix     string
	cbs        []func([]*DataRef) error
	buf        *bytes.Buffer
	hash       *buzhash64.Buzhash64
	splitMask  uint64
	dataRefs   []*DataRef
	done       [][]*DataRef
	rangeSize  int64
	rangeCount int64
}

// newWriter creates a new Writer.
func newWriter(ctx context.Context, objC obj.Client, prefix string) *Writer {
	// Initialize buzhash64 with WindowSize window.
	hash := buzhash64.New()
	hash.Write(initialWindow)
	return &Writer{
		ctx:       ctx,
		objC:      objC,
		prefix:    prefix,
		cbs:       []func([]*DataRef) error{},
		buf:       &bytes.Buffer{},
		hash:      hash,
		splitMask: (1 << uint64(AverageBits)) - 1,
	}
}

// StartRange specifies the start of a range within the byte stream that is meaningful to the caller.
// When this range has ended (by calling StartRange again or Close) and all of the necessary chunks are written, the
// callback given during initialization will be called with DataRefs that can be used for accessing that range.
func (w *Writer) StartRange(cb func([]*DataRef) error) {
	// Finish prior range.
	if w.dataRefs != nil {
		w.finishRange()
	}
	// Start new range.
	w.cbs = append(w.cbs, cb)
	w.dataRefs = []*DataRef{&DataRef{OffsetBytes: int64(w.buf.Len())}}
	w.rangeSize = 0
	w.rangeCount++
}

func (w *Writer) finishRange() {
	lastDataRef := w.dataRefs[len(w.dataRefs)-1]
	lastDataRef.SizeBytes = int64(w.buf.Len()) - lastDataRef.OffsetBytes
	data := w.buf.Bytes()[lastDataRef.OffsetBytes:w.buf.Len()]
	lastDataRef.Hash = hash.EncodeHash(hash.Sum(data))
	w.done = append(w.done, w.dataRefs)
	// Reset hash between ranges.
	w.hash.Reset()
	w.hash.Write(initialWindow)
}

// RangeSize returns the size of the current range.
func (w *Writer) RangeSize() int64 {
	return w.rangeSize
}

// RangeCount returns a count of the number of ranges associated with
// the writer.
func (w *Writer) RangeCount() int64 {
	return w.rangeCount
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
		if w.hash.Sum64()&w.splitMask == 0 {
			w.buf.Write(data[offset : i+1])
			if err := w.put(); err != nil {
				return 0, err
			}
			w.buf.Reset()
			offset = i + 1
			size = 0
		}
	}
	w.buf.Write(data[offset:])
	w.rangeSize += int64(len(data))
	return len(data), nil
}

func (w *Writer) put() error {
	chunk := &Chunk{Hash: hash.EncodeHash(hash.Sum(w.buf.Bytes()))}
	path := path.Join(w.prefix, chunk.Hash)
	// If it does not exist, compress and write it.
	if !w.objC.Exists(w.ctx, path) {
		objW, err := w.objC.Writer(w.ctx, path)
		if err != nil {
			return err
		}
		defer objW.Close()
		gzipW := gzip.NewWriter(objW)
		defer gzipW.Close()
		// (bryce) Encrypt?
		if _, err := io.Copy(gzipW, bytes.NewReader(w.buf.Bytes())); err != nil {
			return err
		}
	}
	// Update chunk and run callback for ranges within the current chunk.
	for _, dataRefs := range w.done {
		lastDataRef := dataRefs[len(dataRefs)-1]
		// Handle edge case where DataRef is size zero.
		if lastDataRef.SizeBytes == 0 {
			dataRefs = dataRefs[:len(dataRefs)-1]
		} else {
			// Set chunk for last DataRef (from current chunk).
			lastDataRef.Chunk = chunk
		}
		if err := w.cbs[0](dataRefs); err != nil {
			return err
		}
		w.cbs = w.cbs[1:]
	}
	w.done = nil
	// Update chunk and hash (if at first DataRef in range)
	// in last DataRef for current range.
	lastDataRef := w.dataRefs[len(w.dataRefs)-1]
	lastDataRef.Chunk = chunk
	if lastDataRef.OffsetBytes > 0 {
		data := w.buf.Bytes()[lastDataRef.OffsetBytes:w.buf.Len()]
		lastDataRef.Hash = hash.EncodeHash(hash.Sum(data))
	}
	lastDataRef.SizeBytes = int64(w.buf.Len()) - lastDataRef.OffsetBytes
	// Setup DataRef for next chunk.
	w.dataRefs = append(w.dataRefs, &DataRef{})
	return nil
}

// Close closes the writer and flushes the remaining bytes to a chunk and finishes
// the final range.
func (w *Writer) Close() error {
	w.finishRange()
	return w.put()
}
