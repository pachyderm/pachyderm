// Package client contains GRPC client interceptors for logging.
package client

import (
	"context"
	"io"
	"strings"
	"sync/atomic"

	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/proto"
)

type loggingStream struct {
	grpc.ClientStream
	done       func(...log.Field)
	desc       *grpc.StreamDesc
	closedSend atomic.Bool
}

var _ grpc.ClientStream = new(loggingStream)

func (s *loggingStream) RecvMsg(m any) error {
	if err := s.ClientStream.RecvMsg(m); err != nil {
		if err == io.EOF {
			s.done(log.Metadata("trailer", s.Trailer()))
			return err //nolint:wrapcheck
		}
		log.Debug(s.Context(), "stream ended unexpectedly", zap.Error(err))
		s.done(zap.Error(err), log.Metadata("trailer", s.Trailer()))
		return err //nolint:wrapcheck
	}
	var field log.Field
	if p, ok := m.(proto.Message); ok {
		field = log.Proto("reply", p)
	} else {
		field = zap.Any("reply", m)
	}
	log.Debug(s.Context(), "received message", field)
	if !s.desc.ServerStreams && s.closedSend.Load() {
		// When the server sends a unary response, the generated client typically only calls
		// CloseAndRecv, which means they never call RecvMsg in a way that returns io.EOF to
		// indiciate the end of the stream.  To handle this case, we consider the stream
		// done if we CloseSend'd and this not a streaming server call.
		s.done(log.Metadata("trailer", s.Trailer()))
	}
	return nil
}

func (s *loggingStream) SendMsg(m any) (retErr error) {
	var field log.Field
	if p, ok := m.(proto.Message); ok {
		field = log.Proto("request", p)
	} else {
		field = zap.Any("request", m)
	}
	if err := s.ClientStream.SendMsg(m); err != nil {
		log.Debug(s.Context(), "error sending message", zap.Error(err), field)
		return err //nolint:wrapcheck
	}
	log.Debug(s.Context(), "sent message", field)
	return nil
}

func (s *loggingStream) CloseSend() error {
	err := s.ClientStream.CloseSend()
	log.Debug(s.Context(), "send side of stream closed", zap.Error(err))
	s.closedSend.Store(true)
	return err //nolint:wrapcheck
}

func LogStream(rctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	ctx, done := log.SpanContext(pctx.Child(rctx, "grpcClient.stream", pctx.WithOptions(zap.WithCaller(false))), strings.TrimPrefix(method, "/"))
	underlying, err := streamer(ctx, desc, cc, method, opts...)
	if err != nil {
		return underlying, err
	}
	stream := &loggingStream{
		ClientStream: underlying,
		desc:         desc,
		done:         done,
	}
	log.Debug(ctx, "stream started", log.OutgoingMetadata(ctx))
	go func() {
		if md, err := stream.Header(); err != nil {
			log.Debug(ctx, "problem getting server headers", zap.Error(err))
		} else {
			log.Debug(ctx, "server headers", log.Metadata("headers", md))
		}
	}()
	return stream, nil
}

var _ grpc.StreamClientInterceptor = LogStream

func LogUnary(rctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	headers, trailer := make(metadata.MD), make(metadata.MD)
	opts = append(opts, grpc.Header(&headers), grpc.Trailer(&trailer))
	ctx, done := log.SpanContext(pctx.Child(rctx, "grpcClient", pctx.WithOptions(zap.WithCaller(false))), strings.TrimPrefix(method, "/"))
	var reqF, replyF log.Field
	if p, ok := req.(proto.Message); ok {
		reqF = log.Proto("request", p)
	} else {
		reqF = zap.Any("request", req)
	}
	log.Debug(ctx, "request", reqF, log.OutgoingMetadata(ctx))
	err := invoker(ctx, method, req, reply, cc, opts...)
	if p, ok := reply.(proto.Message); ok {
		replyF = log.Proto("reply", p)
	} else {
		replyF = zap.Any("reply", reply)
	}
	done(zap.Error(err), log.Metadata("headers", headers), log.Metadata("trailer", trailer), replyF)
	return err
}

var _ grpc.UnaryClientInterceptor = LogUnary
