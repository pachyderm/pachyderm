package grpcutil

import (
	"bufio"
	"bytes"
	"context"
	"io"

	units "github.com/docker/go-units"
	"github.com/gogo/protobuf/types"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

var (
	// MaxMsgSize is used to define the GRPC frame size.
	MaxMsgSize = 20 * units.MiB
	// MaxMsgPayloadSize is the max message payload size.
	// This is slightly less than MaxMsgSize to account
	// for the GRPC message wrapping the payload.
	MaxMsgPayloadSize = MaxMsgSize - units.MiB
)

// Chunk splits a piece of data up, this is useful for splitting up data that's
// bigger than MaxMsgPayloadSize.
func Chunk(data []byte) [][]byte {
	chunkSize := MaxMsgPayloadSize
	var result [][]byte
	for i := 0; i < len(data); i += chunkSize {
		end := i + chunkSize
		if end > len(data) {
			end = len(data)
		}
		result = append(result, data[i:end])
	}
	return result
}

// ChunkReader splits a reader into reasonably sized chunks for the purpose
// of transmitting the chunks over gRPC. For each chunk, it calls the given
// function.
func ChunkReader(r io.Reader, f func([]byte) error) (int, error) {
	var total int
	buf := GetBuffer()
	defer PutBuffer(buf)
	for {
		n, err := r.Read(buf)
		if n == 0 && err != nil {
			if errors.Is(err, io.EOF) {
				return total, nil
			}
			return total, errors.EnsureStack(err)
		}
		if err := f(buf[:n]); err != nil {
			return total, err
		}
		total += n
	}
}

// ChunkWriteCloser is a utility for buffering writes into buffers obtained from a buffer pool.
// The ChunkWriteCloser will buffer up to the capacity of a buffer obtained from a buffer pool,
// then execute a callback that will receive the buffered data. The ChunkWriteCloser will get
// a new buffer from the pool for subsequent writes, so it is expected that the callback will
// return the buffer to the pool.
type ChunkWriteCloser struct {
	bufPool *BufPool
	buf     []byte
	f       func([]byte) error
}

// NewChunkWriteCloser creates a new ChunkWriteCloser.
func NewChunkWriteCloser(bufPool *BufPool, f func(chunk []byte) error) *ChunkWriteCloser {
	return &ChunkWriteCloser{
		bufPool: bufPool,
		buf:     bufPool.GetBuffer()[:0],
		f:       f,
	}
}

// Write performs a write.
func (w *ChunkWriteCloser) Write(data []byte) (int, error) {
	var written int
	for len(w.buf)+len(data) > cap(w.buf) {
		// Write the bytes that fit into w.buf, then
		// remove those bytes from data.
		i := cap(w.buf) - len(w.buf)
		w.buf = append(w.buf, data[:i]...)
		if err := w.f(w.buf); err != nil {
			return 0, err
		}
		w.buf = bufPool.GetBuffer()[:0]
		written += i
		data = data[i:]
	}
	w.buf = append(w.buf, data...)
	written += len(data)
	return written, nil
}

// Close closes the writer.
func (w *ChunkWriteCloser) Close() error {
	if len(w.buf) == 0 {
		return nil
	}
	return w.f(w.buf)
}

// StreamingBytesServer represents a server for an rpc method of the form:
//   rpc Foo(Bar) returns (stream google.protobuf.BytesValue) {}
type StreamingBytesServer interface {
	Send(bytesValue *types.BytesValue) error
}

// StreamingBytesClient represents a client for an rpc method of the form:
//   rpc Foo(Bar) returns (stream google.protobuf.BytesValue) {}
type StreamingBytesClient interface {
	Recv() (*types.BytesValue, error)
}

// NewStreamingBytesReader returns an io.Reader for a StreamingBytesClient.
func NewStreamingBytesReader(streamingBytesClient StreamingBytesClient, cancel context.CancelFunc) io.ReadCloser {
	return &streamingBytesReader{streamingBytesClient: streamingBytesClient, cancel: cancel}
}

type streamingBytesReader struct {
	streamingBytesClient StreamingBytesClient
	buffer               bytes.Buffer
	cancel               context.CancelFunc
}

func (s *streamingBytesReader) Read(p []byte) (int, error) {
	// TODO this is doing an unneeded copy (unless go is smarter than I think it is)
	if s.buffer.Len() == 0 {
		value, err := s.streamingBytesClient.Recv()
		if err != nil {
			return 0, errors.EnsureStack(err)
		}
		s.buffer.Reset()
		if _, err := s.buffer.Write(value.Value); err != nil {
			return 0, errors.EnsureStack(err)
		}
	}
	res, err := s.buffer.Read(p)
	return res, errors.EnsureStack(err)
}

func (s *streamingBytesReader) Close() error {
	if s.cancel != nil {
		s.cancel()
	}
	return nil
}

// WithStreamingBytesWriter sets up a scoped streaming bytes writer that buffers and chunks writes.
// TODO: This should probably use a buffer pool eventually.
func WithStreamingBytesWriter(streamingBytesServer StreamingBytesServer, cb func(io.Writer) error) (retErr error) {
	w := &streamingBytesWriter{streamingBytesServer: streamingBytesServer}
	bufW := bufio.NewWriterSize(w, MaxMsgPayloadSize)
	defer func() {
		if err := bufW.Flush(); retErr == nil {
			retErr = err
		}
	}()
	return cb(bufW)
}

type streamingBytesWriter struct {
	streamingBytesServer StreamingBytesServer
}

func (s *streamingBytesWriter) Write(data []byte) (int, error) {
	var bytesWritten int
	for _, val := range Chunk(data) {
		if err := s.streamingBytesServer.Send(&types.BytesValue{Value: val}); err != nil {
			return bytesWritten, errors.EnsureStack(err)
		}
		bytesWritten += len(val)
	}
	return bytesWritten, nil
}

// ReaderWrapper wraps a reader for the following reason: Go's io.CopyBuffer
// has an annoying optimization wherein if the reader has the WriteTo function
// defined, it doesn't actually use the given buffer.  As a result, we might
// write a large chunk to the gRPC streaming server even though we intend to
// use a small buffer.  Therefore we wrap readers in this wrapper so that only
// Read is defined.
type ReaderWrapper struct {
	Reader io.Reader
}

func (r ReaderWrapper) Read(p []byte) (int, error) {
	res, err := r.Reader.Read(p)
	return res, errors.EnsureStack(err)
}

// TODO: Unused. Remove?
// WriteToStreamingBytesServer writes the data from the io.Reader to the StreamingBytesServer.
func WriteToStreamingBytesServer(reader io.Reader, server StreamingBytesServer) error {
	return WithStreamingBytesWriter(server, func(w io.Writer) error {
		buf := GetBuffer()
		defer PutBuffer(buf)
		_, err := io.CopyBuffer(w, ReaderWrapper{reader}, buf)
		return errors.EnsureStack(err)
	})
}

// WriteFromStreamingBytesClient writes from the StreamingBytesClient to the io.Writer.
func WriteFromStreamingBytesClient(streamingBytesClient StreamingBytesClient, writer io.Writer) error {
	for bytesValue, err := streamingBytesClient.Recv(); !errors.Is(err, io.EOF); bytesValue, err = streamingBytesClient.Recv() {
		if err != nil {
			return errors.EnsureStack(err)
		}
		if _, err = writer.Write(bytesValue.Value); err != nil {
			return errors.EnsureStack(err)
		}
	}
	return nil
}
