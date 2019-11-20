package chunk

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"path"

	"github.com/chmduquesne/rollinghash/buzhash64"
	"github.com/gogo/protobuf/proto"
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

// WriterFunc is a callback that returns the annotations within a chunk.
type WriterFunc func([]*Annotation) error

// dataSet is a unit of work for the workers.
// A worker will roll the rolling hash function across the data set
// while processing the associated annotations.
type dataSet struct {
	annotations []*Annotation
	// nextDataSet is used for the edge case where no split point is found in the assigned data set.
	// A worker that did not find a chunk in its assigned data set will pass it to the prior
	// worker in the chain with nextDataSet set to the next worker's dataSet channel. This allows shuffling
	// of data sets between workers until a split point is found.
	nextDataSet <-chan *dataSet
}

// chanSet is a group of channels used to setup the daisy chain and shuffle data between the workers.
// How these channels are used by a worker depends on whether they are associated with
// the previous or next worker in the chain.
type chanSet struct {
	dataSet chan *dataSet
	done    chan struct{}
}

// The following chanSet types enforce the directionality of the channels at compile time
// to help prevent potentially tricky bugs now and in the future with the daisy chain.
type prevChanSet struct {
	dataSet chan<- *dataSet
	done    <-chan struct{}
}

func newPrevChanSet(c *chanSet) *prevChanSet {
	if c == nil {
		return nil
	}
	return &prevChanSet{
		dataSet: c.dataSet,
		done:    c.done,
	}
}

type nextChanSet struct {
	dataSet <-chan *dataSet
	done    chan<- struct{}
}

func newNextChanSet(c *chanSet) *nextChanSet {
	return &nextChanSet{
		dataSet: c.dataSet,
		done:    c.done,
	}
}

type worker struct {
	ctx            context.Context
	objC           obj.Client
	hash           *buzhash64.Buzhash64
	splitMask      uint64
	first          bool
	annotations    []*Annotation
	bufAnnotations []*Annotation
	bufSize        int64
	f              WriterFunc
	fs             []func() error
	prev           *prevChanSet
	next           *nextChanSet
	stats          *stats
}

func (w *worker) run(dataSet *dataSet) error {
	// Roll through the assigned data set.
	if err := w.rollDataSet(dataSet); err != nil {
		return err
	}
	// No split point found.
	if w.prev != nil && w.first {
		dataSet.annotations = w.annotations
		dataSet.nextDataSet = w.next.dataSet
		select {
		case w.prev.dataSet <- dataSet:
		case <-w.ctx.Done():
			return w.ctx.Err()
		}
	} else {
		// Wait for the next data set to roll.
		nextDataSet := w.next.dataSet
		for nextDataSet != nil {
			select {
			case dataSet, more := <-nextDataSet:
				// The next data set channel is closed for the last worker,
				// so it uploads the last buffer as a chunk.
				if !more {
					if w.annotations[len(w.annotations)-1].buf.Len() > 0 {
						// Last chunk in byte stream is an edge chunk.
						if err := w.put(true); err != nil {
							return err
						}
					}
					nextDataSet = nil
					break
				} else if err := w.rollDataSet(dataSet); err != nil {
					return err
				}
				nextDataSet = dataSet.nextDataSet
			case <-w.ctx.Done():
				return w.ctx.Err()
			}
		}
	}
	// Execute the writer function for the chunks that were found.
	return w.executeFuncs()
}

func (w *worker) rollDataSet(dataSet *dataSet) error {
	// Roll across the annotations in the data set.
	for _, a := range dataSet.annotations {
		if a.buf.Len() > 0 {
			if err := w.roll(a); err != nil {
				return err
			}
			// Reset hash between annotations.
			w.resetHash()
		} else {
			if err := w.copyDataReaders(a); err != nil {
				return err
			}
		}
	}
	if err := w.flushDataReaders(); err != nil {
		return err
	}
	return nil
}

func (w *worker) roll(a *Annotation) error {
	offset := 0
	for i, b := range a.buf.Bytes() {
		w.hash.Roll(b)
		if w.hash.Sum64()&w.splitMask == 0 {
			// Split annotation.
			// The annotation before the split is handled at this chunk split point.
			var beforeSplitA *Annotation
			beforeSplitA, a = splitAnnotation(a, i+1-offset)
			w.annotations = joinAnnotations(w.annotations, beforeSplitA)
			offset = i + 1
			// Send the annotations up to this point to the prior worker if this
			// is the first split point encountered by this worker.
			if w.prev != nil && w.first {
				// We do not consider chunk split points within WindowSize bytes
				// of the start of the data rolled in a worker.
				if w.numBytesRolled() < WindowSize {
					continue
				}
				dataSet := &dataSet{annotations: w.annotations}
				w.prev.dataSet <- dataSet
			} else if err := w.put(w.first); err != nil {
				return err
			}
			w.annotations = nil
			w.first = false
		}
	}
	w.annotations = joinAnnotations(w.annotations, a)
	return nil
}

func (w *worker) numBytesRolled() int {
	var numBytes int
	for _, a := range w.annotations {
		numBytes += a.buf.Len()
	}
	return numBytes
}

func (w *worker) put(edge bool) error {
	var chunkBytes []byte
	for _, a := range w.annotations {
		chunkBytes = append(chunkBytes, a.buf.Bytes()...)
	}
	chunk := &Chunk{
		Hash:      hash.EncodeHash(hash.Sum(chunkBytes)),
		SizeBytes: int64(len(chunkBytes)),
		Edge:      edge,
	}
	path := path.Join(prefix, chunk.Hash)
	// If the chunk does not exist, upload it.
	if !w.objC.Exists(w.ctx, path) {
		if err := w.upload(path, chunkBytes); err != nil {
			return err
		}
	}
	chunkRef := &DataRef{
		Chunk:     chunk,
		SizeBytes: int64(len(chunkBytes)),
	}
	// Update the annotations for the current chunk.
	w.updateAnnotations(chunkRef)
	annotations := w.annotations
	w.fs = append(w.fs, func() error {
		return w.f(annotations)
	})
	return nil
}

func (w *worker) updateAnnotations(chunkRef *DataRef) {
	var offset int64
	for _, a := range w.annotations {
		// (bryce) probably a better way to communicate whether to compute datarefs for an annotation.
		if a.NextDataRef != nil {
			a.NextDataRef.Chunk = chunkRef.Chunk
			if len(w.annotations) > 1 {
				a.NextDataRef.Hash = hash.EncodeHash(hash.Sum(a.buf.Bytes()))
			}
			a.NextDataRef.OffsetBytes = offset
			a.NextDataRef.SizeBytes = int64(a.buf.Len())
			a.NextDataRef.Tags, a.tags = splitTags(a.tags, a.buf.Len())
		}
		offset += int64(a.buf.Len())
	}
}

func (w *worker) upload(path string, chunk []byte) error {
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
	_, err = io.Copy(gzipW, bytes.NewReader(chunk))
	return err
}

func (w *worker) resetHash() {
	w.hash.Reset()
	w.hash.Write(initialWindow)
}

func (w *worker) copyDataReaders(a *Annotation) error {
	for _, dr := range a.drs {
		// We can only consider a data reader for buffering / cheap copying
		// when it does not reference an edge chunk and we are at a split point.
		if dr.DataRef().Chunk.Edge || !w.atSplit() {
			if err := w.flushDataReaders(); err != nil {
				return err
			}
			if err := w.rollDataReader(copyAnnotation(a), dr); err != nil {
				return err
			}
			continue
		}
		// Flush buffered annotations if the next data reader is from a different chunk.
		if len(w.bufAnnotations) > 0 {
			lastA := w.bufAnnotations[len(w.bufAnnotations)-1]
			if lastA.drs[0].DataRef().Chunk.Hash != dr.DataRef().Chunk.Hash {
				if err := w.flushDataReaders(); err != nil {
					return err
				}
			}
		}
		// Join annotation with current buffered annotations, then buffer the current data reader.
		w.bufAnnotations = joinAnnotations(w.bufAnnotations, copyAnnotation(a))
		lastA := w.bufAnnotations[len(w.bufAnnotations)-1]
		lastA.drs = append(lastA.drs, dr)
		w.bufSize += dr.Len()
		// Cheap copy if full chunk is buffered.
		if w.bufSize == dr.DataRef().Chunk.SizeBytes {
			for _, a := range w.bufAnnotations {
				// (bryce) need to handle tags.
				var size int64
				for _, dr := range a.drs {
					size += dr.DataRef().SizeBytes
				}
				a.NextDataRef = a.drs[0].DataRef()
				a.NextDataRef.SizeBytes = size
			}
			annotations := w.bufAnnotations
			w.fs = append(w.fs, func() error {
				return w.f(annotations)
			})
			w.bufAnnotations = nil
			w.bufSize = 0
		}
	}
	return nil
}

func (w *worker) atSplit() bool {
	return w.prev != nil && (!w.first && w.numBytesRolled() == 0)
}

func (w *worker) rollDataReader(a *Annotation, dr *DataReader) error {
	if err := dr.Iterate(func(tag *Tag, r io.Reader) error {
		_, err := io.Copy(a.buf, r)
		if err != nil {
			return err
		}
		a.tags = append(a.tags, tag)
		return nil
	}); err != nil {
		return err
	}
	return w.roll(a)
}

func (w *worker) flushDataReaders() error {
	for _, a := range w.bufAnnotations {
		for _, dr := range a.drs {
			if err := w.rollDataReader(copyAnnotation(a), dr); err != nil {
				return err
			}
		}
		// Reset hash between annotations.
		w.resetHash()
	}
	w.bufAnnotations = nil
	w.bufSize = 0
	return nil
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
	chunkCount      int64
	annotationCount int64
	taggedBytesSize int64
}

// Writer splits a byte stream into content defined chunks that are hashed and deduplicated/uploaded to object storage.
// Chunk split points are determined by a bit pattern in a rolling hash function (buzhash64 at https://github.com/chmduquesne/rollinghash).
// The byte stream is split into byte sets for parallel processing. Workers roll the rolling hash function and perform the execution
// of the writer function on these byte sets. The workers are daisy chained such that split points across byte sets can be resolved by shuffling
// bytes between workers in the chain and the writer function is executed on the sequential ordering of the chunks in the byte stream.
type Writer struct {
	ctx, cancelCtx context.Context
	annotations    []*Annotation
	bufSize        int
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
		eg:            eg,
		newWorkerFunc: newWorkerFunc,
		f:             f,
		stats:         stats,
	}
	return w
}

// Annotate associates an annotation with the current data.
func (w *Writer) Annotate(a *Annotation) {
	w.finishTag()
	a.buf = &bytes.Buffer{}
	w.annotations = append(w.annotations, a)
	w.stats.annotationCount++
}

// AnnotationCount returns a count of the number of annotations created/referenced by
// the writer.
func (w *Writer) AnnotationCount() int64 {
	return w.stats.annotationCount
}

func (w *Writer) StartTag(id string) {
	w.finishTag()
	a := w.annotations[len(w.annotations)-1]
	a.tags = append(a.tags, &Tag{Id: id})
}

func (w *Writer) FinishTag(t string) {
	w.finishTag()
}

func (w *Writer) finishTag() {
	if len(w.annotations) > 0 {
		a := w.annotations[len(w.annotations)-1]
		if a.tags != nil {
			a.tags[len(a.tags)-1].SizeBytes = w.stats.taggedBytesSize
			w.stats.taggedBytesSize = 0
		}
	}
}

// ChunkCount returns a count of the number of chunks created/referenced by
// the writer.
func (w *Writer) ChunkCount() int64 {
	return w.stats.chunkCount
}

func (w *Writer) Write(data []byte) (int, error) {
	a := w.annotations[len(w.annotations)-1]
	var written int
	for w.bufSize+len(data) >= bufSize {
		i := bufSize - w.bufSize
		a.buf.Write(data[:i])
		w.stats.taggedBytesSize += int64(i)
		w.writeDataSet()
		a = w.annotations[len(w.annotations)-1]
		written += i
		data = data[i:]
	}
	a.buf.Write(data)
	w.bufSize += len(data)
	w.stats.taggedBytesSize += int64(len(data))
	written += len(data)
	return written, nil
}

func (w *Writer) writeDataSet() {
	w.finishTag()
	prev := w.prev
	next := &chanSet{
		dataSet: make(chan *dataSet, 1),
		done:    make(chan struct{}),
	}
	dataSet := &dataSet{annotations: w.annotations}
	w.eg.Go(func() error {
		return w.newWorkerFunc(w.cancelCtx, newPrevChanSet(prev), newNextChanSet(next)).run(dataSet)
	})
	w.prev = next
	lastA := w.annotations[len(w.annotations)-1]
	_, a := splitAnnotation(lastA, lastA.buf.Len())
	w.annotations = []*Annotation{a}
	w.bufSize = 0
}

func (w *Writer) Copy(dr *DataReader) error {
	lastA := w.annotations[len(w.annotations)-1]
	lastA.drs = append(lastA.drs, dr)
	w.bufSize += int(dr.Len())
	if w.bufSize > bufSize {
		w.writeDataSet()
	}
	return nil
}

// Close closes the writer.
func (w *Writer) Close() error {
	// Write out the last data set.
	if w.bufSize > 0 {
		w.writeDataSet()
	}
	// Signal to the last worker that it is last.
	if w.prev != nil {
		close(w.prev.dataSet)
		w.prev = nil
	}
	// Wait for the workers to finish.
	return w.eg.Wait()
}

func joinAnnotations(as []*Annotation, a *Annotation) []*Annotation {
	// If the annotation being added is the same as the
	// last, then they are merged.
	if as != nil {
		lastA := as[len(as)-1]
		if lastA.Meta == a.Meta {
			lastA.buf.Write(a.buf.Bytes())
			if lastA.tags != nil {
				lastA.tags = joinTags(lastA.tags, a.tags)
			}
			return as
		}
	}
	return append(as, a)
}

func joinTags(ts1, ts2 []*Tag) []*Tag {
	lastT := ts1[len(ts1)-1]
	if lastT.Id == ts2[0].Id {
		lastT.SizeBytes += ts2[0].SizeBytes
		ts2 = ts2[1:]
	}
	return append(ts1, ts2...)
}

func splitAnnotation(a *Annotation, size int) (*Annotation, *Annotation) {
	a1 := copyAnnotation(a)
	a2 := copyAnnotation(a)
	if a.buf != nil {
		a1.buf = bytes.NewBuffer(a.buf.Bytes()[:size])
		a2.buf = bytes.NewBuffer(a.buf.Bytes()[size:])
	}
	if a.tags != nil {
		a1.tags, a2.tags = splitTags(a.tags, size)
	}
	return a1, a2
}

func copyAnnotation(a *Annotation) *Annotation {
	copyA := &Annotation{Meta: a.Meta}
	if a.NextDataRef != nil {
		copyA.NextDataRef = &DataRef{}
	}
	if a.buf != nil {
		copyA.buf = &bytes.Buffer{}
	}
	return copyA
}

func splitTags(ts []*Tag, size int) ([]*Tag, []*Tag) {
	var ts1, ts2 []*Tag
	for _, t := range ts {
		ts2 = append(ts2, proto.Clone(t).(*Tag))
	}
	for {
		if int(ts2[0].SizeBytes) >= size {
			t := proto.Clone(ts2[0]).(*Tag)
			t.SizeBytes = int64(size)
			ts1 = append(ts1, t)
			ts2[0].SizeBytes -= int64(size)
			break
		}
		size -= int(ts2[0].SizeBytes)
		ts1 = append(ts1, ts2[0])
		ts2 = ts2[1:]
	}
	return ts1, ts2
}
