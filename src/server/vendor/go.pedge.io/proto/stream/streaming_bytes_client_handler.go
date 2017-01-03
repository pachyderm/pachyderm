package protostream

import (
	"io"

	"github.com/gogo/protobuf/types"
)

type streamingBytesClientHandler struct {
	handleFunc func(*types.BytesValue) error
}

func newStreamingBytesClientHandler(handleFunc func(*types.BytesValue) error) *streamingBytesClientHandler {
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
