package chunk

import (
	"bytes"
	"context"
	"io"
	"path"
	"time"

	"github.com/chmduquesne/rollinghash/buzhash64"
	units "github.com/docker/go-units"
	"github.com/pachyderm/pachyderm/src/server/pkg/obj"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/gc"
	"github.com/pachyderm/pachyderm/src/server/pkg/storage/hash"
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
	Data        interface{}
	size        int64
}

// WriterCallback is a callback that returns the updated annotations within a chunk.
type WriterCallback func([]*Annotation) error

type stats struct {
	chunkCount      int64
	annotationCount int64
}

// TODO True max is avg + max, might want to reword or apply the max as max - avg.
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
	objC                    obj.Client
	gcC                     gc.Client
	chunkSize               *chunkSize
	cb                      WriterCallback
	ctx                     context.Context
	cancel                  context.CancelFunc
	err                     error
	chain                   *TaskChain
	annotations             []*Annotation
	numChunkBytesAnnotation int
	hash                    *buzhash64.Buzhash64
	splitMask               uint64
	buf                     *bytes.Buffer
	tmpID                   string
	noUpload                bool
	stats                   *stats
	chunkTTL                time.Duration
	buffering               bool
	first, last             bool
}

func newWriter(ctx context.Context, objC obj.Client, gcC gc.Client, tmpID string, cb WriterCallback, opts ...WriterOption) *Writer {
	cancelCtx, cancel := context.WithCancel(ctx)
	w := &Writer{
		objC: objC,
		gcC:  gcC,
		chunkSize: &chunkSize{
			min: defaultMinChunkSize,
			max: defaultMaxChunkSize,
		},
		cb:     cb,
		ctx:    cancelCtx,
		cancel: cancel,
		chain:  NewTaskChain(cancelCtx),
		buf:    &bytes.Buffer{},
		tmpID:  tmpID,
		stats:  &stats{},
		first:  true,
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
func (w *Writer) Annotate(a *Annotation) error {
	// Create chunks at annotation boundaries if past the average chunk size.
	if w.buf.Len() >= w.chunkSize.avg {
		if err := w.createChunk(); err != nil {
			return err
		}
	}
	w.annotations = append(w.annotations, a)
	w.numChunkBytesAnnotation = 0
	w.stats.annotationCount++
	w.resetHash()
	return nil
}

func (w *Writer) Write(data []byte) (int, error) {
	if err := w.maybeDone(func() error {
		if err := w.flushBuffer(); err != nil {
			return err
		}
		w.roll(data)
		return nil
	}); err != nil {
		return 0, err
	}
	return len(data), nil
}

func (w *Writer) maybeDone(cb func() error) (retErr error) {
	if w.err != nil {
		return w.err
	}
	defer func() {
		if retErr != nil {
			w.err = retErr
			w.cancel()
		}
	}()
	select {
	case <-w.ctx.Done():
		return w.ctx.Err()
	default:
	}
	return cb()
}

func (w *Writer) roll(data []byte) error {
	offset := 0
	for i, b := range data {
		w.hash.Roll(b)
		if w.hash.Sum64()&w.splitMask == 0 {
			if w.numChunkBytesAnnotation+len(data[offset:i+1]) < w.chunkSize.min {
				continue
			}
			w.writeData(data[offset : i+1])
			if err := w.createChunk(); err != nil {
				return err
			}
			offset = i + 1
		}
	}
	for w.numChunkBytesAnnotation+len(data[offset:]) >= w.chunkSize.max {
		bytesLeft := w.chunkSize.max - w.numChunkBytesAnnotation
		w.writeData(data[offset : offset+bytesLeft])
		if err := w.createChunk(); err != nil {
			return err
		}
		offset += bytesLeft
	}
	w.writeData(data[offset:])
	return nil
}

func (w *Writer) writeData(data []byte) {
	lastA := w.annotations[len(w.annotations)-1]
	lastA.size += int64(len(data))
	w.numChunkBytesAnnotation += len(data)
	w.buf.Write(data)
}

func (w *Writer) createChunk() error {
	chunk := w.buf.Bytes()
	edge := w.first || w.last
	annotations := w.splitAnnotations()
	if err := w.chain.CreateTask(func(ctx context.Context, serial func(func() error) error) error {
		return w.processChunk(ctx, chunk, edge, annotations, serial)
	}); err != nil {
		return err
	}
	w.first = false
	w.numChunkBytesAnnotation = 0
	w.buf = &bytes.Buffer{}
	w.stats.chunkCount++
	w.resetHash()
	return nil
}

func (w *Writer) splitAnnotations() []*Annotation {
	annotations := w.annotations
	lastA := w.annotations[len(w.annotations)-1]
	w.annotations = []*Annotation{copyAnnotation(lastA)}
	return annotations
}

func copyAnnotation(a *Annotation) *Annotation {
	copyA := &Annotation{}
	if a.RefDataRefs != nil {
		copyA.RefDataRefs = a.RefDataRefs
	}
	if a.Data != nil {
		copyA.Data = a.Data
	}
	return copyA
}

func (w *Writer) processChunk(ctx context.Context, chunkBytes []byte, edge bool, annotations []*Annotation, serial func(func() error) error) error {
	chunk := &Chunk{Hash: hash.EncodeHash(hash.Sum(chunkBytes))}
	if err := w.maybeUpload(ctx, chunk, chunkBytes); err != nil {
		return err
	}
	chunkRef := &DataRef{
		ChunkInfo: &ChunkInfo{
			Chunk:     chunk,
			SizeBytes: int64(len(chunkBytes)),
			Edge:      edge,
		},
		SizeBytes: int64(len(chunkBytes)),
	}
	if err := w.processAnnotations(ctx, chunkRef, chunkBytes, annotations); err != nil {
		return err
	}
	return serial(func() error {
		return w.cb(annotations)
	})
}

func (w *Writer) maybeUpload(ctx context.Context, chunk *Chunk, chunkBytes []byte) error {
	// Skip the upload if no upload is configured.
	if w.noUpload {
		return nil
	}
	path := path.Join(prefix, chunk.Hash)
	if err := w.gcC.ReserveChunk(ctx, path, w.tmpID, w.getExpiresAt()); err != nil {
		return err
	}
	// Skip the upload if the chunk already exists.
	if w.objC.Exists(ctx, path) {
		return nil
	}
	objW, err := w.objC.Writer(ctx, path)
	if err != nil {
		return err
	}
	defer objW.Close()
	_, err = io.Copy(objW, bytes.NewReader(chunkBytes))
	return err
}

func (w *Writer) processAnnotations(ctx context.Context, chunkRef *DataRef, chunkBytes []byte, annotations []*Annotation) error {
	var offset int64
	var prevRefChunk string
	for _, a := range annotations {
		// TODO: Empty data reference for size zero annotation?
		if a.size == 0 {
			continue
		}
		a.NextDataRef = newDataRef(chunkRef, chunkBytes, offset, a.size)
		offset += a.size
		// Skip references if no upload is configured.
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
			if err := w.gcC.CreateReference(ctx, &gc.Reference{
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

func newDataRef(chunkRef *DataRef, chunkBytes []byte, offset, size int64) *DataRef {
	dataRef := &DataRef{}
	dataRef.ChunkInfo = chunkRef.ChunkInfo
	if chunkRef.SizeBytes == size {
		dataRef.Hash = chunkRef.Hash
	} else {
		dataRef.Hash = hash.EncodeHash(hash.Sum(chunkBytes[offset : offset+size]))
	}
	dataRef.OffsetBytes = offset
	dataRef.SizeBytes = size
	return dataRef
}

// Copy copies a data reference to the writer.
func (w *Writer) Copy(dataRef *DataRef) error {
	return w.maybeDone(func() error {
		if err := w.maybeBufferDataRef(dataRef); err != nil {
			return err
		}
		return w.maybeCheapCopy()
	})
}

func (w *Writer) maybeBufferDataRef(dataRef *DataRef) error {
	lastA := w.annotations[len(w.annotations)-1]
	if lastA.NextDataRef != nil && lastA.NextDataRef.OffsetBytes != 0 {
		if err := w.flushBuffer(); err != nil {
			return err
		}
	}
	if !w.buffering {
		// We can only begin buffering data refs when:
		// - We are at a chunk split point.
		// - The data ref does not reference an edge chunk.
		// - It is the first data reference for the chunk.
		if w.buf.Len() != 0 || dataRef.ChunkInfo.Edge || dataRef.OffsetBytes != 0 {
			return w.flushDataRef(dataRef)
		}
	} else {
		// We can only continue buffering data refs if each subsequent data ref is the next in the chunk.
		prevDataRef := w.getPrevDataRef()
		if prevDataRef.ChunkInfo.Chunk.Hash != dataRef.ChunkInfo.Chunk.Hash || prevDataRef.OffsetBytes+prevDataRef.SizeBytes != dataRef.OffsetBytes {
			if err := w.flushBuffer(); err != nil {
				return err
			}
			return w.flushDataRef(dataRef)
		}
	}
	lastA.NextDataRef = mergeDataRef(lastA.NextDataRef, dataRef)
	w.buffering = true
	return nil
}

func mergeDataRef(dr1, dr2 *DataRef) *DataRef {
	if dr1 == nil {
		return dr2
	}
	dr1.SizeBytes += dr2.SizeBytes
	if dr1.SizeBytes == dr1.ChunkInfo.SizeBytes {
		dr1.Hash = dr1.ChunkInfo.Chunk.Hash
	}
	return dr1
}

func (w *Writer) getPrevDataRef() *DataRef {
	for i := len(w.annotations) - 1; i >= 0; i-- {
		if w.annotations[i].NextDataRef != nil {
			return w.annotations[i].NextDataRef
		}
	}
	// TODO: Reaching here would be a bug, maybe panic?
	return nil
}

func (w *Writer) flushBuffer() error {
	if w.buffering {
		annotations := w.annotations
		w.annotations = nil
		for _, annotation := range annotations {
			if err := w.Annotate(copyAnnotation(annotation)); err != nil {
				return err
			}
			w.stats.annotationCount--
			if annotation.NextDataRef != nil {
				if err := w.flushDataRef(annotation.NextDataRef); err != nil {
					return err
				}
			}
		}
		w.buffering = false
	}
	return nil
}

func (w *Writer) flushDataRef(dataRef *DataRef) error {
	buf := &bytes.Buffer{}
	r := newDataReader(w.ctx, w.objC, dataRef, nil)
	if err := r.Get(buf); err != nil {
		return err
	}
	return w.roll(buf.Bytes())
}

func (w *Writer) maybeCheapCopy() error {
	if w.buffering {
		// Cheap copy if a full chunk is buffered.
		lastDataRef := w.annotations[len(w.annotations)-1].NextDataRef
		if lastDataRef.OffsetBytes+lastDataRef.SizeBytes == lastDataRef.ChunkInfo.SizeBytes {
			annotations := w.splitAnnotations()
			if err := w.chain.CreateTask(func(_ context.Context, serial func(func() error) error) error {
				return serial(func() error {
					return w.cb(annotations)
				})
			}); err != nil {
				return err
			}
			w.buffering = false
		}
	}
	return nil
}

// Close closes the writer.
func (w *Writer) Close() error {
	defer w.cancel()
	return w.maybeDone(func() error {
		if err := w.flushBuffer(); err != nil {
			return err
		}
		if len(w.annotations) > 0 {
			if err := w.createChunk(); err != nil {
				return err
			}
		}
		return w.chain.Wait()
	})
}

func (w *Writer) getExpiresAt() time.Time {
	return time.Now().UTC().Add(w.chunkTTL)
}
