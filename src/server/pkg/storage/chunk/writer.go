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
)

const (
	// MB is Megabytes.
	MB = 1024 * 1024
	// WindowSize is the size of the rolling hash window.
	WindowSize = 64
	// (bryce) this should be configurable.
	bufSize = 50 * MB
)

// initialWindow is the set of bytes used to initialize the window
// of the rolling hash function.
var initialWindow = make([]byte, WindowSize)

// WriterFunc is a callback that returns a data reference to the next chunk and the annotations within the chunk.
type WriterFunc func(*DataRef, []*Annotation) error

// byteSet is a unit of work for the workers.
// A worker will roll the rolling hash function across the bytes
// while processing the associated annotations.
type byteSet struct {
	data        []byte
	annotations []*Annotation
	// nextBytes is used for the edge case where no split point is found in an assigned byte set.
	// A worker that did not find a chunk in its assigned byte set will pass it to the prior
	// worker in the chain with nextBytes set to the next worker's bytes channel. This allows shuffling
	// of byte sets between workers until a split point is found.
	nextBytes <-chan *byteSet
}

// chanSet is a group of channels used to setup the daisy chain and shuffle data between the workers.
// How these channels are used by a worker depends on whether they are associated with
// the previous or next worker in the chain.
type chanSet struct {
	bytes chan *byteSet
	done  chan struct{}
}

// The following chanSet types enforce the directionality of the channels at compile time
// to help prevent potentially tricky bugs now and in the future with the daisy chain.
type prevChanSet struct {
	bytes chan<- *byteSet
	done  <-chan struct{}
}

func newPrevChanSet(c *chanSet) *prevChanSet {
	if c == nil {
		return nil
	}
	return &prevChanSet{
		bytes: c.bytes,
		done:  c.done,
	}
}

type nextChanSet struct {
	bytes <-chan *byteSet
	done  chan<- struct{}
}

func newNextChanSet(c *chanSet) *nextChanSet {
	return &nextChanSet{
		bytes: c.bytes,
		done:  c.done,
	}
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
	fs          []func() error
	prev        *prevChanSet
	next        *nextChanSet
	stats       *stats
}

func (w *worker) run(byteSet *byteSet) error {
	// Roll through the assigned byte set.
	if err := w.rollByteSet(byteSet); err != nil {
		return err
	}
	// No split point found.
	if w.prev != nil && w.first {
		byteSet.nextBytes = w.next.bytes
		select {
		case w.prev.bytes <- byteSet:
		case <-w.ctx.Done():
			return w.ctx.Err()
		}
	} else {
		// Wait for the next byte set to roll.
		nextBytes := w.next.bytes
		for nextBytes != nil {
			select {
			case byteSet, more := <-nextBytes:
				// The next bytes channel is closed for the last worker,
				// so it uploads the last buffer as a chunk.
				if !more {
					if err := w.put(); err != nil {
						return err
					}
					nextBytes = nil
					break
				} else if err := w.rollByteSet(byteSet); err != nil {
					return err
				}
				nextBytes = byteSet.nextBytes
			case <-w.ctx.Done():
				return w.ctx.Err()
			}
		}
	}
	// Execute the writer function for the chunks that were found.
	return w.executeFuncs()
}

func (w *worker) rollByteSet(byteSet *byteSet) error {
	// Roll across the byte set.
	for i, a := range byteSet.annotations {
		var data []byte
		if i == len(byteSet.annotations)-1 {
			data = byteSet.data[a.Offset:len(byteSet.data)]
		} else {
			data = byteSet.data[a.Offset:byteSet.annotations[i+1].Offset]
		}
		// Convert from byte set offset to chunk offset.
		a.Offset = int64(w.buf.Len())
		w.annotations = joinAnnotations(w.annotations, a)
		if err := w.roll(data); err != nil {
			return err
		}
		// Reset hash between annotations.
		w.resetHash()
	}
	return nil
}

func joinAnnotations(as []*Annotation, a *Annotation) []*Annotation {
	// If the annotation being added is the same as the
	// last, then they are merged.
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
				if w.buf.Len() < WindowSize {
					continue
				}
				byteSet := &byteSet{
					data:        w.buf.Bytes(),
					annotations: w.annotations,
				}
				w.prev.bytes <- byteSet
				w.first = false
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
	updateAnnotations(chunkRef, w.buf.Bytes(), w.annotations)
	annotations := w.annotations
	w.fs = append(w.fs, func() error {
		return w.f(chunkRef, annotations)
	})
	return nil
}

func updateAnnotations(chunkRef *DataRef, buf []byte, annotations []*Annotation) {
	// (bryce) add check for no buf and greater than one annotation.
	// Fast path for next data reference being full chunk.
	if len(annotations) == 1 && annotations[0].NextDataRef != nil {
		annotations[0].NextDataRef = chunkRef
	}
	for i, a := range annotations {
		// (bryce) probably a better way to communicate whether to compute datarefs for an annotation.
		if a.NextDataRef != nil {
			a.NextDataRef.Chunk = chunkRef.Chunk
			var data []byte
			if i == len(annotations)-1 {
				data = buf[a.Offset:len(buf)]
			} else {
				data = buf[a.Offset:annotations[i+1].Offset]
			}
			a.NextDataRef.Hash = hash.EncodeHash(hash.Sum(data))
			a.NextDataRef.OffsetBytes = a.Offset
			a.NextDataRef.SizeBytes = int64(len(data))
		}
	}
}

func splitAnnotations(as []*Annotation) []*Annotation {
	if len(as) == 0 {
		return nil
	}
	// Copy the last annotation.
	lastA := as[len(as)-1]
	copyA := &Annotation{Meta: lastA.Meta}
	if lastA.NextDataRef != nil {
		copyA.NextDataRef = &DataRef{}
	}
	return []*Annotation{copyA}
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

func (w *worker) executeFuncs() error {
	// Wait for the prior worker in the chain to signal
	// that it is done.
	if w.prev != nil {
		select {
		case <-w.prev.done:
		case <-w.ctx.Done():
			return w.ctx.Err()
		}
	}
	// Execute the writer function for each chunk.
	for _, f := range w.fs {
		w.stats.chunkCount++
		if err := f(); err != nil {
			return err
		}
	}
	// Signal to the next worker in the chain that this worker is done.
	close(w.next.done)
	return nil
}

type stats struct {
	chunkCount         int64
	annotationCount    int64
	annotatedBytesSize int64
}

// Writer splits a byte stream into content defined chunks that are hashed and deduplicated/uploaded to object storage.
// Chunk split points are determined by a bit pattern in a rolling hash function (buzhash64 at https://github.com/chmduquesne/rollinghash).
// The byte stream is split into byte sets for parallel processing. Workers roll the rolling hash function and perform the execution
// of the writer function on these byte sets. The workers are daisy chained such that split points across byte sets can be resolved by shuffling
// bytes between workers in the chain and the writer function is executed on the sequential ordering of the chunks in the byte stream.
type Writer struct {
	ctx, cancelCtx context.Context
	buf            *bytes.Buffer
	annotations    []*Annotation
	eg             *errgroup.Group
	newWorkerFunc  func(context.Context, *prevChanSet, *nextChanSet) *worker
	prev           *chanSet
	f              WriterFunc
	stats          *stats
}

func newWriter(ctx context.Context, objC obj.Client, averageBits int, f WriterFunc, seed int64) *Writer {
	stats := &stats{}
	newWorkerFunc := func(ctx context.Context, prev *prevChanSet, next *nextChanSet) *worker {
		w := &worker{
			ctx:       ctx,
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
	eg, cancelCtx := errgroup.WithContext(ctx)
	w := &Writer{
		ctx:           ctx,
		cancelCtx:     cancelCtx,
		buf:           &bytes.Buffer{},
		eg:            eg,
		newWorkerFunc: newWorkerFunc,
		f:             f,
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
	w.stats.annotationCount++
	w.stats.annotatedBytesSize = 0
}

// AnnotationCount returns a count of the number of annotations created/referenced by
// the writer.
func (w *Writer) AnnotationCount() int64 {
	return w.stats.annotationCount
}

// AnnotatedBytesSize returns the size of the bytes for the current annotation.
func (w *Writer) AnnotatedBytesSize() int64 {
	return w.stats.annotatedBytesSize
}

// Flush flushes the buffered data.
func (w *Writer) Flush() error {
	// Write out the last buffer.
	if w.buf.Len() > 0 {
		w.writeByteSet()
	}
	// Signal to the last worker that it is last.
	if w.prev != nil {
		close(w.prev.bytes)
		w.prev = nil
	}
	// Wait for the workers to finish.
	if err := w.eg.Wait(); err != nil {
		return err
	}
	w.eg, w.cancelCtx = errgroup.WithContext(w.ctx)
	return nil
}

// Reset resets the buffer and annotations.
func (w *Writer) Reset() {
	// (bryce) should cancel all workers.
	w.buf = &bytes.Buffer{}
	w.annotations = nil
	w.stats.annotatedBytesSize = 0
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
	w.stats.annotatedBytesSize += int64(written)
	return written, nil
}

func (w *Writer) writeByteSet() {
	prev := w.prev
	next := &chanSet{
		bytes: make(chan *byteSet, 1),
		done:  make(chan struct{}),
	}
	byteSet := &byteSet{
		data:        w.buf.Bytes(),
		annotations: w.annotations,
	}
	w.eg.Go(func() error {
		return w.newWorkerFunc(w.cancelCtx, newPrevChanSet(prev), newNextChanSet(next)).run(byteSet)
	})
	w.prev = next
	w.buf = &bytes.Buffer{}
	w.annotations = splitAnnotations(w.annotations)
}

// Close closes the writer.
func (w *Writer) Close() error {
	return w.Flush()
}

//// Copy does a cheap copy from a reader to a writer.
//func (w *Writer) Copy(r *Reader, n ...int64) error {
//	c, err := r.ReadCopy(n...)
//	if err != nil {
//		return err
//	}
//	return w.WriteCopy(c)
//}
//
//// WriteCopy writes copy data to the writer.
//func (w *Writer) WriteCopy(c *Copy) error {
//	if _, err := io.Copy(w, c.before); err != nil {
//		return err
//	}
//	for _, chunkRef := range c.chunkRefs {
//		w.stats.chunkCount++
//		// (bryce) might want to double check if this is correct.
//		w.stats.annotatedBytesSize += chunkRef.SizeBytes
//		updateAnnotations(chunkRef, nil, w.annotations)
//		if err := w.f(chunkRef, w.annotations); err != nil {
//			return err
//		}
//	}
//	_, err := io.Copy(w, c.after)
//	return err
//}
