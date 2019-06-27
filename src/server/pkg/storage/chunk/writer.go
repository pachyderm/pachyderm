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
	// WindowSize is the size of the rolling hash window.
	WindowSize = 64
)

var initialWindow = make([]byte, WindowSize)

type WriterFunc func(*DataRef, []*Annotation) error

// Writer splits a byte stream into content defined chunks that are hashed and deduplicated/uploaded to object storage.
// Chunk split points are determined by a bit pattern in a rolling hash function (buzhash64 at https://github.com/chmduquesne/rollinghash).
type Writer struct {
	ctx            context.Context
	objC           obj.Client
	buf            *bytes.Buffer
	hash           *buzhash64.Buzhash64
	splitMask      uint64
	f              WriterFunc
	annotations    []*Annotation
	chunkCount     int64
	annotationSize int64
}

// newWriter creates a new Writer.
func newWriter(ctx context.Context, objC obj.Client, averageBits int, f WriterFunc) *Writer {
	// Initialize buzhash64 with WindowSize window.
	hash := buzhash64.New()
	w := &Writer{
		ctx:       ctx,
		objC:      objC,
		buf:       &bytes.Buffer{},
		hash:      hash,
		splitMask: (1 << uint64(averageBits)) - 1,
		f:         f,
	}
	w.resetHash()
	return w
}

func (w *Writer) resetHash() {
	w.hash.Reset()
	w.hash.Write(initialWindow)
}

func (w *Writer) Annotate(a *Annotation) {
	a.Offset = int64(w.buf.Len())
	// Handle the edge case when the last annotation from the prior chunk
	// should not carry over into the next.
	if a.Offset == 0 {
		w.annotations = nil
	}
	w.annotations = append(w.annotations, a)
	w.annotationSize = 0
	// Reset hash between annotations.
	w.resetHash()
}

// AnnotationSize returns the size of the current annotation.
func (w *Writer) AnnotationSize() int64 {
	return w.annotationSize
}

// ChunkCount returns a count of the number of chunks created/referenced by
// the writer.
func (w *Writer) ChunkCount() int64 {
	return w.chunkCount
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
			// Reset hash between chunks.
			w.resetHash()
			offset = i + 1
			size = 0
		}
	}
	w.buf.Write(data[offset:])
	w.annotationSize += int64(len(data))
	return len(data), nil
}

func (w *Writer) put() error {
	chunk := &Chunk{Hash: hash.EncodeHash(hash.Sum(w.buf.Bytes()))}
	path := path.Join(prefix, chunk.Hash)
	// If the chunk does not exist, upload it.
	if !w.objC.Exists(w.ctx, path) {
		if err := w.upload(path); err != nil {
			return err
		}
	}
	w.chunkCount++
	chunkRef := &DataRef{
		Chunk:     chunk,
		SizeBytes: int64(len(w.buf.Bytes())),
	}
	return w.executeFunc(chunkRef)
}

func (w *Writer) upload(path string) error {
	objW, err := w.objC.Writer(w.ctx, path)
	if err != nil {
		return err
	}
	defer objW.Close()
	gzipW := gzip.NewWriter(objW)
	defer gzipW.Close()
	// (bryce) Encrypt?
	_, err = io.Copy(gzipW, bytes.NewReader(w.buf.Bytes()))
	return err
}

func (w *Writer) executeFunc(chunkRef *DataRef) error {
	// Update the annotations for the current chunk.
	w.updateAnnotations(chunkRef)
	// Execute the chunk callback.
	if err := w.f(chunkRef, w.annotations); err != nil {
		return err
	}
	// Keep the last annotation because it may span into the next chunk.
	lastA := w.annotations[len(w.annotations)-1]
	a := &Annotation{Meta: lastA.Meta}
	if lastA.NextDataRef != nil {
		a.NextDataRef = &DataRef{}
	}
	w.annotations = []*Annotation{a}
	return nil
}

func (w *Writer) updateAnnotations(chunkRef *DataRef) {
	// Fast path for next data reference being full chunk.
	if len(w.annotations) == 1 && w.annotations[0].NextDataRef != nil {
		w.annotations[0].NextDataRef = chunkRef
	}
	for i, a := range w.annotations {
		// (bryce) probably a better way to communicate whether to compute datarefs for an annotation.
		if a.NextDataRef != nil {
			a.NextDataRef.Chunk = chunkRef.Chunk
			data := w.buf.Bytes()
			if i == len(w.annotations)-1 {
				data = data[a.Offset:w.buf.Len()]
			} else {
				data = data[a.Offset:w.annotations[i+1].Offset]
			}
			a.NextDataRef.Hash = hash.EncodeHash(hash.Sum(data))
			a.NextDataRef.OffsetBytes = a.Offset
			a.NextDataRef.SizeBytes = int64(len(data))
		}
	}
}

func (w *Writer) atSplit() bool {
	return w.buf.Len() == 0
}

func (w *Writer) writeChunk(chunkRef *DataRef) error {
	w.chunkCount++
	w.annotationSize += chunkRef.SizeBytes
	return w.executeFunc(chunkRef)
}

// Close closes the writer and flushes the remaining bytes to a chunk and finishes
// the final range.
func (w *Writer) Close() error {
	return w.put()
}
