package protostream

import "go.pedge.io/pb/go/google/protobuf"

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
		&google_protobuf.BytesValue{
			Value: p,
		},
	); err != nil {
		return 0, err
	}
	return len(p), nil
}
