package protoutil

import (
	"bufio"
	"io"
	"time"

	"go.pedge.io/google-protobuf"
)

func TimeToTimestamp(t time.Time) *google_protobuf.Timestamp {
	return &google_protobuf.Timestamp{
		Seconds: t.UnixNano() / int64(time.Second),
		Nanos:   int32(t.UnixNano() % int64(time.Second)),
	}
}

func TimestampToTime(timestamp *google_protobuf.Timestamp) time.Time {
	return time.Unix(
		timestamp.Seconds,
		int64(timestamp.Nanos),
	).UTC()
}

func TimestampLess(i *google_protobuf.Timestamp, j *google_protobuf.Timestamp) bool {
	if i == nil {
		return true
	}
	if j == nil {
		return false
	}
	if i.Seconds < j.Seconds {
		return true
	}
	if i.Seconds > j.Seconds {
		return false
	}
	return i.Nanos < j.Nanos
}

type StreamingBytesServer interface {
	Send(bytesValue *google_protobuf.BytesValue) error
}

func NewStreamingBytesReader(streamingBytesClient StreamingBytesClient) io.Reader {
	return newStreamingBytesReader(streamingBytesClient)
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
