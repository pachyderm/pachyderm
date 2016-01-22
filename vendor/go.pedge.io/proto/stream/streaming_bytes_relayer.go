package protostream

import (
	"errors"
	"io"
	"sync"

	"golang.org/x/net/context"

	"google.golang.org/grpc/metadata"

	"go.pedge.io/pb/go/google/protobuf"
	"go.pedge.io/pkg/sync"
)

type streamingBytesRelayer struct {
	ctx     context.Context
	header  metadata.MD
	trailer metadata.MD
	values  []*google_protobuf.BytesValue
	cv      *sync.Cond
	closed  pkgsync.VolatileBool
}

func newStreamingBytesRelayer(ctx context.Context) *streamingBytesRelayer {
	return &streamingBytesRelayer{
		ctx,
		nil,
		nil,
		make([]*google_protobuf.BytesValue, 0),
		sync.NewCond(&sync.Mutex{}),
		pkgsync.NewVolatileBool(false),
	}
}

func (s *streamingBytesRelayer) Send(bytesValue *google_protobuf.BytesValue) error {
	if bytesValue == nil {
		return nil
	}
	s.cv.L.Lock()
	defer s.cv.L.Unlock()
	if s.closed.Value() {
		return ErrAlreadyClosed
	}
	value := make([]byte, len(bytesValue.Value))
	copy(value, bytesValue.Value)
	s.values = append(s.values, &google_protobuf.BytesValue{Value: value})
	s.cv.Signal()
	return nil
}

func (s *streamingBytesRelayer) Recv() (*google_protobuf.BytesValue, error) {
	s.cv.L.Lock()
	for len(s.values) == 0 {
		if s.closed.Value() {
			return nil, io.EOF
		}
		s.cv.Wait()
	}
	defer s.cv.L.Unlock()
	value := s.values[0]
	s.values = s.values[1:]
	return value, nil
}

func (s *streamingBytesRelayer) Header() (metadata.MD, error) {
	return s.header, nil
}

func (s *streamingBytesRelayer) Trailer() metadata.MD {
	return s.trailer
}

func (s *streamingBytesRelayer) CloseSend() error {
	s.cv.L.Lock()
	defer s.cv.L.Unlock()
	if !s.closed.CompareAndSwap(false, true) {
		return ErrAlreadyClosed
	}
	s.cv.Broadcast()
	return nil
}

func (s *streamingBytesRelayer) SendHeader(md metadata.MD) error {
	s.header = md
	return nil
}

func (s *streamingBytesRelayer) SetTrailer(md metadata.MD) {
	s.trailer = md
}

func (s *streamingBytesRelayer) Context() context.Context {
	return s.ctx
}

func (s *streamingBytesRelayer) SendMsg(m interface{}) error {
	return errors.New("protostream: not implemented")
}

func (s *streamingBytesRelayer) RecvMsg(m interface{}) error {
	return errors.New("protostream: not implemented")
}
