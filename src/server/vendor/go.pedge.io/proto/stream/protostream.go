package protostream // import "go.pedge.io/proto/stream"

import (
	"bufio"
	"errors"
	"io"

	"golang.org/x/net/context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

	"go.pedge.io/pb/go/google/protobuf"
)

var (
	// ErrAlreadyClosed is the error returned if CloseSend() is called twice on a StreamingBytesRelayer.
	ErrAlreadyClosed = errors.New("protostream: already closed")
)

// StreamingBytesServer represents a server for an rpc method of the form:
//   rpc Foo(Bar) returns (stream google.protobuf.BytesValue) {}
type StreamingBytesServer interface {
	Send(bytesValue *google_protobuf.BytesValue) error
}

// StreamingBytesServeCloser is a StreamingBytesServer with close.
type StreamingBytesServeCloser interface {
	StreamingBytesServer
	CloseSend() error
}

// StreamingBytesClient represents a client for an rpc method of the form:
//   rpc Foo(Bar) returns (stream google.protobuf.BytesValue) {}
type StreamingBytesClient interface {
	Recv() (*google_protobuf.BytesValue, error)
}

// StreamingBytesDuplexer is both a StreamingBytesClient and StreamingBytesServer.
type StreamingBytesDuplexer interface {
	StreamingBytesClient
	StreamingBytesServer
}

// StreamingBytesDuplexCloser is a StreamingBytesDuplexer with close.
type StreamingBytesDuplexCloser interface {
	StreamingBytesClient
	StreamingBytesServeCloser
}

// StreamingBytesClientHandler handles a StreamingBytesClient.
type StreamingBytesClientHandler interface {
	Handle(streamingBytesClient StreamingBytesClient) error
}

// NewStreamingBytesReader returns an io.Reader for a StreamingBytesClient.
func NewStreamingBytesReader(streamingBytesClient StreamingBytesClient) io.Reader {
	return newStreamingBytesReader(streamingBytesClient)
}

// NewStreamingBytesWriter returns an io.Writer for a StreamingBytesServer.
func NewStreamingBytesWriter(streamingBytesServer StreamingBytesServer) io.Writer {
	return newStreamingBytesWriter(streamingBytesServer)
}

// WriteToStreamingBytesServer writes the data from the io.Reader to the StreamingBytesServer.
func WriteToStreamingBytesServer(reader io.Reader, streamingBytesServer StreamingBytesServer) error {
	_, err := bufio.NewReader(reader).WriteTo(NewStreamingBytesWriter(streamingBytesServer))
	return err
}

// NewStreamingBytesClientHandler returns a StreamingBytesClientHandler for the given handleFunc.
func NewStreamingBytesClientHandler(handleFunc func(*google_protobuf.BytesValue) error) StreamingBytesClientHandler {
	return newStreamingBytesClientHandler(handleFunc)
}

// WriteFromStreamingBytesClient writes from the StreamingBytesClient to the io.Writer.
func WriteFromStreamingBytesClient(streamingBytesClient StreamingBytesClient, writer io.Writer) error {
	return NewStreamingBytesClientHandler(
		func(bytesValue *google_protobuf.BytesValue) error {
			_, err := writer.Write(bytesValue.Value)
			return err
		},
	).Handle(streamingBytesClient)
}

// RelayFromStreamingBytesClient relays *google_protobuf.BytesValues from the StreamingBytesClient to the StreamingBytesServer.
func RelayFromStreamingBytesClient(streamingBytesClient StreamingBytesClient, streamingBytesServer StreamingBytesServer) error {
	return NewStreamingBytesClientHandler(
		func(bytesValue *google_protobuf.BytesValue) error {
			return streamingBytesServer.Send(bytesValue)
		},
	).Handle(streamingBytesClient)
}

// StreamingBytesRelayer represents both generated Clients and servers for streams of *google_protobuf.BytesValue.
type StreamingBytesRelayer interface {
	StreamingBytesDuplexer
	Header() (metadata.MD, error)
	Trailer() metadata.MD
	CloseSend() error
	SendHeader(metadata.MD) error
	SetTrailer(metadata.MD)
	grpc.Stream
}

// NewStreamingBytesRelayer returns a new StreamingBytesRelayer for the context.Context.
func NewStreamingBytesRelayer(ctx context.Context) StreamingBytesRelayer {
	return newStreamingBytesRelayer(ctx)
}
