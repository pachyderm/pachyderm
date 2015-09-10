package protostream

import (
	"bufio"
	"io"

	"go.pedge.io/google-protobuf"
)

// StreamingBytesServer represents a server for an rpc method of the form:
//   rpc Foo(Bar) returns (stream google.protobuf.BytesValue) {}
type StreamingBytesServer interface {
	Send(bytesValue *google_protobuf.BytesValue) error
}

// StreamingBytesClient represents a client for an rpc method of the form:
//   rpc Foo(Bar) returns (stream google.protobuf.BytesValue) {}
type StreamingBytesClient interface {
	Recv() (*google_protobuf.BytesValue, error)
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
