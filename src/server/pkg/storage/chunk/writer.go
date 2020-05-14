package chunk

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"path"

	"github.com/chmduquesne/rollinghash/buzhash64"
	units "github.com/docker/go-units"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/gc"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/hash"
	"golang.org/x/sync/errgroup"
)

const (
	// WindowSize is the size of the rolling hash window.
	WindowSize = 64
)

// initialWindow is the set of bytes used to initialize the window
// of the rolling hash function.
var initialWindow = make([]byte, WindowSize)

// Annotation is used to associate information with data
// written into the chunk storage layer.
type Annotation struct {
	RefDataRefs []*DataRef
	NextDataRef *DataRef
	tags        []*Tag
	Data        interface{}
}

// WriterFunc is a callback that returns the annotations within a chunk.
type WriterFunc func([]*Annotation) error

type stats struct {
	chunkCount      int64
	annotationCount int64
}

// (bryce) true max is avg + max, might want to reword or apply the max as max - avg.
const (
	defaultAverageBits  = 23
	defaultSeed         = 1
	defaultMinChunkSize = 1 * units.MB
	defaultMaxChunkSize = 20 * units.MB
)

type chunkSize struct {
	avg, min, max int
}

// Writer splits a byte stream into content defined chunks that are hashed and deduplicated/uploaded to object storage.
// Chunk split points are determined by a bit pattern in a rolling hash function (buzhash64 at https://github.com/chmduquesne/rollinghash).
type Writer struct {
	ctx                     context.Context
	objC                    obj.Client
	gcC                     gc.Client
	chunkSize               *chunkSize
	annotations             []*Annotation
	numChunkBytesAnnotation int
	splitMask               uint64
	hash                    *buzhash64.Buzhash64
	buf                     *bytes.Buffer
	drs                     []*DataReader
	prevChan                chan struct{}
	eg                      *errgroup.Group
	tmpID                   string
	f                       WriterFunc
	noUpload                bool
	stats                   *stats
}

func newWriter(ctx context.Context, objC obj.Client, gcC gc.Client, tmpID string, f WriterFunc, opts ...WriterOption) *Writer {
	eg, errCtx := errgroup.WithContext(ctx)
	w := &Writer{
		ctx:  errCtx,
		objC: objC,
		gcC:  gcC,
		chunkSize: &chunkSize{
			min: defaultMinChunkSize,
			max: defaultMaxChunkSize,
		},
		buf:   &bytes.Buffer{},
		eg:    eg,
		tmpID: tmpID,
		f:     f,
		stats: &stats{},
	}
	WithRollingHashConfig(defaultAverageBits, defaultSeed)(w)
	for _, opt := range opts {
		opt(w)
	}
	w.resetHash()
	return w
}

func (w *Writer) resetHash() {
	w.hash.Reset()
	w.hash.Write(initialWindow)
}

// AnnotationCount returns a count of the number of annotations created/referenced by
// the writer.
func (w *Writer) AnnotationCount() int64 {
	return w.stats.annotationCount
}

// ChunkCount returns a count of the number of chunks created/referenced by
// the writer.
func (w *Writer) ChunkCount() int64 {
	return w.stats.chunkCount
}

// Annotate associates an annotation with the current data.
func (w *Writer) Annotate(a *Annotation) {
	// Create chunks at annotation boundaries if past the average chunk size.
	if w.buf.Len() >= w.chunkSize.avg {
		w.createChunk()
	}
	if w.buf.Len() == 0 && w.drs == nil {
		w.annotations = nil
	}
	w.annotations = append(w.annotations, a)
	w.numChunkBytesAnnotation = 0
	w.resetHash()
	w.stats.annotationCount++
}

// Tag starts a tag in the current annotation with the passed in id.
// (bryce) maybe add some validation for calling annotate / tag function in incorrect order.
func (w *Writer) Tag(id string) {
	lastA := w.annotations[len(w.annotations)-1]
	lastA.tags = joinTags(lastA.tags, []*Tag{&Tag{Id: id}})
}

func (w *Writer) Write(data []byte) (int, error) {
	if w.done() {
		return 0, w.eg.Wait()
	}
	if err := w.flushDataReaders(); err != nil {
		return 0, err
	}
	w.roll(data)
	return len(data), nil
}

func (w *Writer) done() bool {
	select {
	case <-w.ctx.Done():
		return true
	default:
		return false
	}
}

func (w *Writer) roll(data []byte) {
	offset := 0
	for i, b := range data {
		w.hash.Roll(b)
		if w.hash.Sum64()&w.splitMask == 0 {
			if w.numChunkBytesAnnotation+len(data[offset:i+1]) < w.chunkSize.min {
				continue
			}
			w.writeData(data[offset : i+1])
			w.createChunk()
			offset = i + 1
		}
	}
	for w.numChunkBytesAnnotation+len(data[offset:]) >= w.chunkSize.max {
		bytesLeft := w.chunkSize.max - w.numChunkBytesAnnotation
		w.writeData(data[offset : offset+bytesLeft])
		w.createChunk()
		offset += bytesLeft
	}
	w.writeData(data[offset:])
}

func (w *Writer) writeData(data []byte) {
	lastA := w.annotations[len(w.annotations)-1]
	lastTag := lastA.tags[len(lastA.tags)-1]
	lastTag.SizeBytes += int64(len(data))
	w.numChunkBytesAnnotation += len(data)
	w.buf.Write(data)
}

func (w *Writer) createChunk() {
	chunk := w.buf.Bytes()
	w.appendToChain(func(annotations []*Annotation, prevChan, nextChan chan struct{}) error {
		return w.processChunk(chunk, annotations, prevChan, nextChan)
	})
	w.numChunkBytesAnnotation = 0
	w.resetHash()
	w.buf = &bytes.Buffer{}
}

func (w *Writer) appendToChain(f func([]*Annotation, chan struct{}, chan struct{}) error, last ...bool) {
	// (bryce) add limiter
	annotations := w.annotations
	w.annotations = []*Annotation{copyAnnotation(w.annotations[len(w.annotations)-1])}
	prevChan := w.prevChan
	var nextChan chan struct{}
	if len(last) == 0 || !last[0] {
		nextChan = make(chan struct{})
	}
	w.prevChan = nextChan
	w.eg.Go(func() error {
		return f(annotations, prevChan, nextChan)
	})
	w.stats.chunkCount++
}

func copyAnnotation(a *Annotation) *Annotation {
	copyA := &Annotation{Data: a.Data}
	if a.RefDataRefs != nil {
		copyA.RefDataRefs = a.RefDataRefs
	}
	if a.NextDataRef != nil {
		copyA.NextDataRef = &DataRef{}
	}
	if a.tags != nil {
		lastTag := a.tags[len(a.tags)-1]
		copyA.tags = []*Tag{&Tag{Id: lastTag.Id}}
	}
	return copyA
}

func (w *Writer) processChunk(chunkBytes []byte, annotations []*Annotation, prevChan, nextChan chan struct{}) error {
	chunk := &Chunk{Hash: hash.EncodeHash(hash.Sum(chunkBytes))}
	if err := w.maybeUpload(chunk, chunkBytes); err != nil {
		return err
	}
	chunkRef := &DataRef{
		ChunkInfo: &ChunkInfo{
			Chunk:     chunk,
			SizeBytes: int64(len(chunkBytes)),
			Edge:      prevChan == nil || nextChan == nil,
		},
		SizeBytes: int64(len(chunkBytes)),
	}
	// Process the annotations for the current chunk.
	if err := w.processAnnotations(chunkRef, chunkBytes, annotations); err != nil {
		return err
	}
	return w.executeFunc(annotations, prevChan, nextChan)
}

func (w *Writer) maybeUpload(chunk *Chunk, chunkBytes []byte) error {
	// Skip the upload if no upload is configured.
	if w.noUpload {
		return nil
	}
	path := path.Join(prefix, chunk.Hash)
	if err := w.gcC.ReserveChunk(w.ctx, path, w.tmpID); err != nil {
		return err
	}
	// Skip the upload if the chunk already exists.
	if w.objC.Exists(w.ctx, path) {
		return nil
	}
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
	_, err = io.Copy(gzipW, bytes.NewReader(chunkBytes))
	return err
}

func (w *Writer) processAnnotations(chunkRef *DataRef, chunkBytes []byte, annotations []*Annotation) error {
	var offset int64
	var prevRefChunk string
	for _, a := range annotations {
		// Update the annotation fields.
		// (bryce) probably a better way to communicate whether to compute datarefs for an annotation.
		size := sizeOfTags(a.tags)
		if a.NextDataRef != nil {
			a.NextDataRef.ChunkInfo = chunkRef.ChunkInfo
			if len(annotations) > 1 {
				a.NextDataRef.Hash = hash.EncodeHash(hash.Sum(chunkBytes[offset : offset+size]))
			}
			a.NextDataRef.OffsetBytes = offset
			a.NextDataRef.SizeBytes = size
			a.NextDataRef.Tags = a.tags
		}
		offset += size
		// Skip creating the cross chunk references if no upload is configured.
		if w.noUpload {
			continue
		}
		// Create the cross chunk references.
		// We keep track of the previous referenced chunk to prevent duplicate create
		// reference calls.
		for _, dataRef := range a.RefDataRefs {
			refChunk := dataRef.ChunkInfo.Chunk.Hash
			if prevRefChunk == refChunk {
				continue
			}
			if err := w.gcC.CreateReference(w.ctx, &gc.Reference{
				Sourcetype: "chunk",
				Source:     path.Join(prefix, chunkRef.ChunkInfo.Chunk.Hash),
				Chunk:      path.Join(prefix, refChunk),
			}); err != nil {
				return err
			}
			prevRefChunk = refChunk
		}
	}
	return nil
}

func sizeOfTags(tags []*Tag) int64 {
	var size int64
	for _, tag := range tags {
		size += tag.SizeBytes
	}
	return size
}

func (w *Writer) executeFunc(annotations []*Annotation, prevChan, nextChan chan struct{}) error {
	// Wait for the previous chunk to be processed before executing callback.
	if prevChan != nil {
		select {
		case <-prevChan:
		case <-w.ctx.Done():
			return w.ctx.Err()
		}
	}
	if err := w.f(annotations); err != nil {
		return err
	}
	if nextChan != nil {
		close(nextChan)
	}
	return nil
}

// Copy copies data from a data reader to the writer.
// The copy will either be by reading the referenced data, or just
// copying the data reference (cheap copy).
func (w *Writer) Copy(dr *DataReader) error {
	if w.done() {
		return w.eg.Wait()
	}
	if err := w.maybeBufferDataReader(dr); err != nil {
		return err
	}
	w.maybeCheapCopy()
	return nil
}

func (w *Writer) maybeBufferDataReader(dr *DataReader) error {
	if len(w.drs) > 0 {
		// Flush the buffered data readers if the next data reader is for a different chunk
		// or it is not the next data reader for the chunk.
		lastBufDataRef := w.drs[len(w.drs)-1].DataRef()
		if lastBufDataRef.ChunkInfo.Chunk.Hash != dr.DataRef().ChunkInfo.Chunk.Hash ||
			lastBufDataRef.OffsetBytes+lastBufDataRef.SizeBytes != dr.DataRef().OffsetBytes {
			if err := w.flushDataReaders(); err != nil {
				return err
			}
		}
	}
	// We can only consider a data reader for buffering / cheap copying when:
	// - We are at a chunk split point.
	// - It does not reference an edge chunk.
	// - If no other data readers are buffered, it is the first data reader in the chunk.
	// - The full data reference is readable (bounded data readers don't return the
	//   full data reference).
	if w.buf.Len() > 0 || dr.DataRef().ChunkInfo.Edge || (len(w.drs) == 0 && dr.DataRef().OffsetBytes > 0) || dr.Len() != dr.DataRef().SizeBytes {
		if err := w.flushDataReaders(); err != nil {
			return err
		}
		return w.flushDataReader(dr)
	}
	w.drs = append(w.drs, dr)
	return nil
}

func (w *Writer) flushDataReaders() error {
	if w.drs != nil {
		// (bryce) this could probably be refactored, we are basically replaying the
		// annotations and writing the buffered readers.
		annotations := w.annotations
		w.annotations = nil
		for i, dr := range w.drs {
			w.Annotate(copyAnnotation(annotations[i]))
			// (bryce) avoid double counting.
			w.stats.annotationCount--
			if err := w.flushDataReader(dr); err != nil {
				return err
			}
		}
		if len(annotations) > len(w.drs) {
			w.Annotate(copyAnnotation(annotations[len(annotations)-1]))
			// (bryce) avoid double counting.
			w.stats.annotationCount--
		}
		w.drs = nil
	}
	return nil
}

func (w *Writer) flushDataReader(dr *DataReader) error {
	return dr.Iterate(func(tag *Tag, r io.Reader) error {
		w.Tag(tag.Id)
		buf := &bytes.Buffer{}
		if _, err := io.Copy(buf, r); err != nil {
			return err
		}
		w.roll(buf.Bytes())
		return nil
	})
}

func (w *Writer) maybeCheapCopy() {
	if len(w.drs) == 0 {
		return
	}
	// Cheap copy if full chunk is buffered.
	lastBufDataRef := w.drs[len(w.drs)-1].DataRef()
	if lastBufDataRef.OffsetBytes+lastBufDataRef.SizeBytes == lastBufDataRef.ChunkInfo.SizeBytes {
		for i, a := range w.annotations {
			dataRef := w.drs[i].DataRef()
			a.NextDataRef = dataRef
			a.tags = dataRef.Tags
		}
		w.appendToChain(func(annotations []*Annotation, prevChan, nextChan chan struct{}) error {
			return w.executeFunc(annotations, prevChan, nextChan)
		})
		w.drs = nil
	}
}

// Close closes the writer.
func (w *Writer) Close() error {
	if w.done() {
		return w.eg.Wait()
	}
	if err := w.flushDataReaders(); err != nil {
		return err
	}
	if w.buf.Len() > 0 {
		chunk := w.buf.Bytes()
		w.appendToChain(func(annotations []*Annotation, prevChan, nextChan chan struct{}) error {
			return w.processChunk(chunk, annotations, prevChan, nextChan)
		}, true)
	}
	return w.eg.Wait()
}
