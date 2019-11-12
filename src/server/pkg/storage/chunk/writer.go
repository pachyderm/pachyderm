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
	drs         []*DataReader
	size        int64
}

func (w *worker) run(dataSet *dataSet) error {
	// Roll through the assigned data set.
	if err := w.rollDataSet(dataSet); err != nil {
		return err
	}
	if err := w.flushDataReaders(); err != nil {
		return err
	}
	// No split point found.
	if w.prev != nil && w.first {
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
					if err := w.put(); err != nil {
						return err
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
		a.Offset = int64(w.buf.Len())
		w.annotations = joinAnnotations(w.annotations, a)
		a = w.annotations[len(w.annotations)-1]
		if err := w.copyDataReaders(a); err != nil {
			return err
		}
		if err := w.roll(a); err != nil {
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
		as[len(as)-1].tags = joinTags(as[len(as)-1].tags, a.tags)
		return as
	}
	return append(as, a)
}

func joinTags(ts1, ts2 []*Tag) []*Tag {
	if ts1[len(ts1)-1].Id == ts2[0].Id {
		ts1[len(ts1)-1].SizeBytes += ts2[0].SizeBytes
		ts2 = ts2[1:]
	}
	return append(ts1, ts2...)
}

func (w *worker) copyDataReaders(a *Annotation) error {
	for _, dr := range a.drs {
		if w.first {
			if err := w.rollDataReader(dr); err != nil {
				return err
			}
			continue
		}
		// Flush buffered data readers if they will not be cheap copied.
		if len(w.drs) > 0 && w.drs[0].DataRef().Chunk.Hash != dr.DataRef().Chunk.Hash {
			if err := w.flushDataReaders(); err != nil {
				return err
			}
			w.drs = nil
			w.size = 0
		}
		w.drs = append(w.drs, dr)
		w.size += dr.Len()
		// Cheap copy if full chunk is buffered.
		if w.size == dr.DataRef().Chunk.SizeBytes {
			// (bryce) I think passing the chunk ref into the callback is no longer necessary
			// in the indexing and can be removed.
			chunkRef := &(*dr.DataRef())
			chunkRef.Hash = ""
			chunkRef.OffsetBytes = 0
			chunkRef.SizeBytes = dr.DataRef().Chunk.SizeBytes
			for i, a := range w.annotations {
				a.NextDataRef = w.drs[i].DataRef()
			}
			if err := w.f(chunkRef, w.annotations); err != nil {
				return err
			}
			w.drs = nil
			w.size = 0
		}
	}
	return nil
}

func (w *worker) rollDataReader(dr *DataReader) error {
	a := w.annotations[len(w.annotations)-1]
	if err := dr.Iterate(func(tag *Tag, r io.Reader) error {
		a.tags = joinTags(a.tags, []*Tag{tag})
		_, err := io.Copy(a.buf, r)
		return err
	}); err != nil {
		return err
	}
	defer a.buf.Reset()
	return w.roll(a)
}

func (w *worker) flushDataReaders() error {
	for _, dr := range w.drs {
		if err := w.rollDataReader(dr); err != nil {
			return err
		}
	}
	return nil
}

func (w *worker) roll(a *Annotation) error {
	data := a.buf.Bytes()
	offset := 0
	for i, b := range data {
		w.hash.Roll(b)
		if w.hash.Sum64()&w.splitMask == 0 {
			w.buf.Write(data[offset : i+1])
			offset = i + 1
			if w.prev != nil && w.first {
				// We do not consider chunk split points within WindowSize bytes
				// of the start of the data set.
				if w.buf.Len() < WindowSize {
					continue
				}
				w.annotations[len(w.annotations)-1].buf = w.buf
				dataSet := &dataSet{annotations: w.annotations}
				w.prev.dataSet <- dataSet
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
			a.NextDataRef.Tags, a.tags = splitTags(a.tags, a.NextDataRef.SizeBytes)
		}
	}
}

func splitAnnotations(as []*Annotation) []*Annotation {
	if len(as) == 0 {
		return nil
	}
	// Copy the last annotation.
	lastA := as[len(as)-1]
	copyA := &Annotation{
		Meta: lastA.Meta,
		buf:  &bytes.Buffer{},
	}
	if lastA.NextDataRef != nil {
		copyA.NextDataRef = &DataRef{}
	}
	return []*Annotation{copyA}
}

func splitTags(tags []*Tag, size int64) ([]*Tag, []*Tag) {
	var beforeSplit []*Tag
	afterSplit := tags
	for {
		if size == 0 {
			break
		}
		if afterSplit[0].SizeBytes > size {
			beforeTag := &(*afterSplit[0])
			beforeTag.SizeBytes = size
			beforeSplit = append(beforeSplit, beforeTag)
			afterSplit[0].SizeBytes = afterSplit[0].SizeBytes - size
			break
		}
		size -= afterSplit[0].SizeBytes
		beforeSplit = append(beforeSplit, afterSplit[0])
		afterSplit = afterSplit[1:]
	}
	return beforeSplit, afterSplit
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
	a.tags = []*Tag{}
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
		if len(a.tags) > 0 {
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
		w.writeDataSet()
		written += i
		data = data[i:]
	}
	a.buf.Write(data)
	w.bufSize += len(data)
	written += len(data)
	w.stats.taggedBytesSize += int64(written)
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
	w.annotations = splitAnnotations(w.annotations)
	w.bufSize = 0
}

func (w *Writer) Copy(dr *DataReader, tagBound ...string) error {
	var err error
	dr, err = dr.LimitReader(tagBound...)
	if err == nil {
		return err
	}
	a := w.annotations[len(w.annotations)-1]
	a.drs = append(a.drs, dr)
	w.bufSize += int(dr.Len())
	if w.bufSize > bufSize {
		w.writeDataSet()
	}
	return nil
}

// Close closes the writer.
func (w *Writer) Close() error {
	// Write out the last data set.
	w.writeDataSet()
	// Signal to the last worker that it is last.
	if w.prev != nil {
		close(w.prev.dataSet)
		w.prev = nil
	}
	// Wait for the workers to finish.
	return w.eg.Wait()
}
