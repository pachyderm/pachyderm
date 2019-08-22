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
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

const (
	// MB is Megabytes.
	MB = 1024 * 1024
	// WindowSize is the size of the rolling hash window.
	WindowSize = 64
	// (bryce) this should be configurable.
	bufSize = 50 * MB
)

var initialWindow = make([]byte, WindowSize)

// WriterFunc is a callback that returns a data reference to the next chunk and the annotations within the chunk.
type WriterFunc func(*DataRef, []*Annotation) error

type byteSet struct {
	data        []byte
	annotations []*Annotation
}

type chanSet struct {
	bytes chan *byteSet
	done  chan struct{}
}

type worker struct {
	ctx         context.Context
	objC        obj.Client
	hash        *buzhash64.Buzhash64
	splitMask   uint64
	first       bool
	buf         *bytes.Buffer
	annotations []*Annotation
	f           WriterFunc
	callbacks   []func() error
	prev, next  *chanSet
	stats       *stats
}

// (bryce) edge case where no split point was found, or found before window size bytes.
func (w *worker) run(byteSet *byteSet) error {
	// Roll through the assigned byte set.
	if err := w.rollByteSet(byteSet); err != nil {
		return err
	}
	select {
	case byteSet, more := <-w.next.bytes:
		if !more {
			if err := w.put(); err != nil {
				return err
			}
		} else if err := w.rollByteSet(byteSet); err != nil {
			return err
		}
	case <-w.ctx.Done():
		return w.ctx.Err()
	}
	if w.prev != nil {
		select {
		case <-w.prev.done:
		case <-w.ctx.Done():
			return w.ctx.Err()
		}
	}
	// Execute the writer function for each chunk.
	for _, c := range w.callbacks {
		w.stats.chunkCount++
		if err := c(); err != nil {
			return err
		}
	}
	// Signal to the next worker that this worker is done.
	w.next.done <- struct{}{}
	return nil
}

func (w *worker) rollByteSet(byteSet *byteSet) error {
	// Roll across the byte set.
	for i, a := range byteSet.annotations {
		w.annotations = joinAnnotations(w.annotations, a)
		var data []byte
		if i == len(byteSet.annotations)-1 {
			data = byteSet.data[a.Offset:len(byteSet.data)]
		} else {
			data = byteSet.data[a.Offset:byteSet.annotations[i+1].Offset]
		}
		// Convert from byte set offset to chunk offset.
		a.Offset = int64(w.buf.Len())
		if err := w.roll(data); err != nil {
			return err
		}
		// Reset hash between annotations.
		w.resetHash()
	}
	return nil
}

func joinAnnotations(as []*Annotation, a *Annotation) []*Annotation {
	if as != nil && as[len(as)-1].Meta == a.Meta {
		return as
	}
	return append(as, a)
}

func (w *worker) roll(data []byte) error {
	offset := 0
	for i, b := range data {
		w.hash.Roll(b)
		if w.hash.Sum64()&w.splitMask == 0 {
			w.buf.Write(data[offset : i+1])
			offset = i + 1
			if w.prev != nil && w.first {
				// We do not consider chunk split points within WindowSize bytes
				// of the start of the byte set.
				if w.first && w.buf.Len() < WindowSize {
					continue
				}
				w.first = false
				byteSet := &byteSet{
					data:        w.buf.Bytes(),
					annotations: w.annotations,
				}
				w.prev.bytes <- byteSet
			} else if err := w.put(); err != nil {
				return err
			}
			w.buf = &bytes.Buffer{}
			w.annotations = splitAnnotations(w.annotations)
		}
	}
	w.buf.Write(data[offset:])
	return nil
}

func (w *worker) put() error {
	chunk := &Chunk{Hash: hash.EncodeHash(hash.Sum(w.buf.Bytes()))}
	path := path.Join(prefix, chunk.Hash)
	// If the chunk does not exist, upload it.
	if !w.objC.Exists(w.ctx, path) {
		if err := w.upload(path); err != nil {
			return err
		}
	}
	chunkRef := &DataRef{
		Chunk:     chunk,
		SizeBytes: int64(w.buf.Len()),
	}
	// Update the annotations for the current chunk.
	w.updateAnnotations(chunkRef)
	annotations := w.annotations
	w.callbacks = append(w.callbacks, func() error {
		return w.f(chunkRef, annotations)
	})
	return nil
}

func (w *worker) updateAnnotations(chunkRef *DataRef) {
	// Fast path for next data reference being full chunk.
	if len(w.annotations) == 1 && w.annotations[0].NextDataRef != nil {
		w.annotations[0].NextDataRef = chunkRef
	}
	for i, a := range w.annotations {
		// (bryce) probably a better way to communicate whether to compute datarefs for an annotation.
		if a.NextDataRef != nil {
			a.NextDataRef.Chunk = chunkRef.Chunk
			var data []byte
			if i == len(w.annotations)-1 {
				data = w.buf.Bytes()[a.Offset:w.buf.Len()]
			} else {
				data = w.buf.Bytes()[a.Offset:w.annotations[i+1].Offset]
			}
			a.NextDataRef.Hash = hash.EncodeHash(hash.Sum(data))
			a.NextDataRef.OffsetBytes = a.Offset
			a.NextDataRef.SizeBytes = int64(len(data))
		}
	}
}

func (w *worker) resetHash() {
	w.hash.Reset()
	w.hash.Write(initialWindow)
}

func (w *worker) upload(path string) error {
	objW, err := w.objC.Writer(w.ctx, path)
	if err != nil {
		return err
	}
	defer objW.Close()
	gzipW, err := gzip.NewWriterLevel(objW, gzip.BestSpeed)
	if err != nil {
		return err
	}
	defer gzipW.Close()
	// (bryce) Encrypt?
	_, err = io.Copy(gzipW, bytes.NewReader(w.buf.Bytes()))
	return err
}

type stats struct {
	chunkCount     int64
	annotationSize int64
}

// Writer splits a byte stream into content defined chunks that are hashed and deduplicated/uploaded to object storage.
// Chunk split points are determined by a bit pattern in a rolling hash function (buzhash64 at https://github.com/chmduquesne/rollinghash).
type Writer struct {
	ctx           context.Context
	buf           *bytes.Buffer
	annotations   []*Annotation
	memoryLimiter *semaphore.Weighted
	eg            *errgroup.Group
	newWorkerFunc func(*chanSet, *chanSet) *worker
	prev          *chanSet
	stats         *stats
}

func newWriter(ctx context.Context, objC obj.Client, memoryLimiter *semaphore.Weighted, averageBits int, f WriterFunc, seed int64) *Writer {
	eg, cancelCtx := errgroup.WithContext(ctx)
	stats := &stats{}
	newWorkerFunc := func(prev, next *chanSet) *worker {
		w := &worker{
			ctx:       cancelCtx,
			objC:      objC,
			hash:      buzhash64.NewFromUint64Array(buzhash64.GenerateHashes(seed)),
			splitMask: (1 << uint64(averageBits)) - 1,
			first:     true,
			buf:       &bytes.Buffer{},
			prev:      prev,
			next:      next,
			f:         f,
			stats:     stats,
		}
		w.resetHash()
		return w
	}
	w := &Writer{
		ctx:           cancelCtx,
		buf:           &bytes.Buffer{},
		memoryLimiter: memoryLimiter,
		eg:            eg,
		newWorkerFunc: newWorkerFunc,
		stats:         stats,
	}
	return w
}

// Annotate associates an annotation with the current byte set.
func (w *Writer) Annotate(a *Annotation) {
	a.Offset = int64(w.buf.Len())
	if a.Offset == 0 {
		w.annotations = nil
	}
	w.annotations = append(w.annotations, a)
	w.stats.annotationSize = 0
}

// AnnotationSize returns the size of the current annotation.
func (w *Writer) AnnotationSize() int64 {
	return w.stats.annotationSize
}

func (w *Writer) Flush() error {
	// (bryce) should wait for outstanding work.
	// might need to treat buffering differently here.
	return nil
}

// Reset resets the buffer and annotations.
func (w *Writer) Reset() {
	// (bryce) should cancel all workers.
	w.buf = &bytes.Buffer{}
	w.annotations = nil
	w.stats.annotationSize = 0
}

// ChunkCount returns a count of the number of chunks created/referenced by
// the writer.
func (w *Writer) ChunkCount() int64 {
	return w.stats.chunkCount
}

// Write rolls through the data written, calling c.f when a chunk is found.
// Note: If making changes to this function, be wary of the performance
// implications (check before and after performance with chunker benchmarks).
func (w *Writer) Write(data []byte) (int, error) {
	var written int
	for w.buf.Len()+len(data) >= bufSize {
		i := bufSize - w.buf.Len()
		w.buf.Write(data[:i])
		w.writeByteSet()
		written += i
		data = data[i:]
	}
	w.buf.Write(data)
	written += len(data)
	w.stats.annotationSize += int64(written)
	return written, nil
}

func (w *Writer) writeByteSet() {
	w.memoryLimiter.Acquire(w.ctx, bufSize)
	prev := w.prev
	next := &chanSet{
		bytes: make(chan *byteSet, 1),
		done:  make(chan struct{}, 1),
	}
	byteSet := &byteSet{
		data:        w.buf.Bytes(),
		annotations: w.annotations,
	}
	w.eg.Go(func() error {
		defer w.memoryLimiter.Release(bufSize)
		return w.newWorkerFunc(prev, next).run(byteSet)
	})
	w.prev = next
	w.buf = &bytes.Buffer{}
	w.annotations = splitAnnotations(w.annotations)
}

func splitAnnotations(as []*Annotation) []*Annotation {
	if len(as) == 0 {
		return nil
	}
	lastA := as[len(as)-1]
	copyA := &Annotation{Meta: lastA.Meta}
	if lastA.NextDataRef != nil {
		copyA.NextDataRef = &DataRef{}
	}
	return []*Annotation{copyA}
}

// Copy does a cheap copy from a reader to a writer.
func (w *Writer) Copy(r *Reader, n ...int64) error {
	c, err := r.ReadCopy(n...)
	if err != nil {
		return err
	}
	return w.WriteCopy(c)
}

// WriteCopy writes copy data to the writer.
func (w *Writer) WriteCopy(c *Copy) error {
	if _, err := io.Copy(w, c.before); err != nil {
		return err
	}
	for _, chunkRef := range c.chunkRefs {
		w.stats.chunkCount++
		// (bryce) might want to double check if this is correct.
		w.stats.annotationSize += chunkRef.SizeBytes
		//if err := w.executeFunc(chunkRef); err != nil {
		//	return err
		//}
	}
	_, err := io.Copy(w, c.after)
	return err
}

func (w *Writer) Close() error {
	// Write out the last buffer.
	if w.buf.Len() > 0 {
		w.writeByteSet()
	}
	// Signal to the last worker that it is last.
	if w.prev != nil {
		close(w.prev.bytes)
	}
	// Wait for the workers to finish.
	return w.eg.Wait()
}
