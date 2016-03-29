package protostream

import (
	"bytes"
)

type streamingBytesReader struct {
	streamingBytesClient StreamingBytesClient
	buffer               bytes.Buffer
}

func newStreamingBytesReader(
	streamingBytesClient StreamingBytesClient,
) *streamingBytesReader {
	return &streamingBytesReader{
		streamingBytesClient: streamingBytesClient,
	}
}

func (s *streamingBytesReader) Read(p []byte) (int, error) {
	// TODO this is doing an unneeded copy (unless go is smarter than I think it is)
	if s.buffer.Len() == 0 {
		value, err := s.streamingBytesClient.Recv()
		if err != nil {
			return 0, err
		}
		if _, err := s.buffer.Write(value.Value); err != nil {
			return 0, err
		}
	}
	return s.buffer.Read(p)
}
