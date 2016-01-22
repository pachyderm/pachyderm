package protostream

import (
	"io"

	"go.pedge.io/pb/go/google/protobuf"
)

type streamingBytesClientHandler struct {
	handleFunc func(*google_protobuf.BytesValue) error
}

func newStreamingBytesClientHandler(handleFunc func(*google_protobuf.BytesValue) error) *streamingBytesClientHandler {
	return &streamingBytesClientHandler{handleFunc}
}

func (s *streamingBytesClientHandler) Handle(streamingBytesClient StreamingBytesClient) error {
	for bytesValue, err := streamingBytesClient.Recv(); err != io.EOF; bytesValue, err = streamingBytesClient.Recv() {
		if err != nil {
			return err
		}
		if handleErr := s.handleFunc(bytesValue); handleErr != nil {
			return handleErr
		}
	}
	return nil
}
