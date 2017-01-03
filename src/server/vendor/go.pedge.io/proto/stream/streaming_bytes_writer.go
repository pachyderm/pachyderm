package protostream

import "github.com/gogo/protobuf/types"

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
	if len(p) == 0 {
		return 0, nil
	}
	if err := s.streamingBytesServer.Send(
		&types.BytesValue{
			Value: p,
		},
	); err != nil {
		return 0, err
	}
	return len(p), nil
}
