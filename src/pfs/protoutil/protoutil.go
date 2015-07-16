package protoutil

import (
	"bufio"
	"io"

	"github.com/peter-edge/go-google-protobuf"
)

type StreamingBytesServer interface {
	Send(bytesValue *google_protobuf.BytesValue) error
}

func NewStreamingBytesWriter(streamingBytesServer StreamingBytesServer) io.Writer {
	return newStreamingBytesWriter(streamingBytesServer)
}

func WriteToStreamingBytesServer(reader io.Reader, streamingBytesServer StreamingBytesServer) error {
	_, err := bufio.NewReader(reader).WriteTo(NewStreamingBytesWriter(streamingBytesServer))
	return err
}

type StreamingBytesClient interface {
	Recv() (*google_protobuf.BytesValue, error)
}

type StreamingBytesClientHandler interface {
	Handle(streamingBytesClient StreamingBytesClient) error
}

func NewStreamingBytesClientHandler(handleFunc func(*google_protobuf.BytesValue) error) StreamingBytesClientHandler {
	return newStreamingBytesClientHandler(handleFunc)
}

func WriteFromStreamingBytesClient(streamingBytesClient StreamingBytesClient, writer io.Writer) error {
	return NewStreamingBytesClientHandler(
		func(bytesValue *google_protobuf.BytesValue) error {
			_, err := writer.Write(bytesValue.Value)
			return err
		},
	).Handle(streamingBytesClient)
}

func RelayFromStreamingBytesClient(streamingBytesClient StreamingBytesClient, streamingBytesServer StreamingBytesServer) error {
	return NewStreamingBytesClientHandler(
		func(bytesValue *google_protobuf.BytesValue) error {
			return streamingBytesServer.Send(bytesValue)
		},
	).Handle(streamingBytesClient)
}
