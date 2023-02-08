package server

import (
	"context"

	"github.com/pachyderm/pachyderm/v2/src/debug"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *debugServer) SetLogLevel(ctx context.Context, req *debug.SetLogLevelRequest) (*debug.SetLogLevelResponse, error) {
	result := new(debug.SetLogLevelResponse)
	switch req.GetLevel() {
	case debug.SetLogLevelRequest_DEBUG:
		log.SetLevel(log.DebugLevel)
		log.SetGRPCLogLevel(zap.DebugLevel)
	case debug.SetLogLevelRequest_INFO:
		log.SetLevel(log.InfoLevel)
		log.SetGRPCLogLevel(zap.InfoLevel)
	case debug.SetLogLevelRequest_ERROR:
		log.SetLevel(log.ErrorLevel)
		log.SetGRPCLogLevel(zap.ErrorLevel)
	default:
		return result, status.Error(codes.InvalidArgument, "invalid log level")
	}
	log.Error(ctx, "log level set by user request", zap.Stringer("level", req.GetLevel()))
	return result, nil
}
