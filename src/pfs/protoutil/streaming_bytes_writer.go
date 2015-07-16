package protoutil

import (
	"bufio"
	"io"

	"github.com/peter-edge/go-google-protobuf"
)

type streamingBytesWriter struct {
	streamingBytesServer StreamingBytesServer
}

func newStreamingBytesWriter(
	streamingBytesServer StreamingBytesServer,
) *streamingBytesWriter {
	return &streamingBytesWriter{
		streamingBytesServer,
	}
}

func (s *streamingBytesWriter) Write(p []byte) (int, error) {
	if err := s.streamingBytesServer.Send(
		&google_protobuf.BytesValue{
			Value: p,
		},
	); err != nil {
		return 0, err
	}
	return len(p), nil
}

func (s *streamingBytesWriter) ReadFrom(reader io.Reader) error {
	_, err := bufio.NewReader(reader).WriteTo(s)
	return err
}
