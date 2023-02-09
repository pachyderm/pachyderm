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
	switch x := req.GetLevel().(type) {
	case nil:
		return result, status.Error(codes.InvalidArgument, "no level provided")
	case *debug.SetLogLevelRequest_Grpc:
		switch x.Grpc {
		case debug.SetLogLevelRequest_DEBUG:
			log.SetGRPCLogLevel(zap.DebugLevel)
			log.Info(ctx, "set grpc log level to debug")
		case debug.SetLogLevelRequest_INFO:
			log.SetGRPCLogLevel(zap.InfoLevel)
			log.Info(ctx, "set grpc log level to info")
		case debug.SetLogLevelRequest_ERROR:
			log.SetGRPCLogLevel(zap.ErrorLevel)
			log.Info(ctx, "set grpc log level to error")
		case debug.SetLogLevelRequest_OFF:
			log.SetGRPCLogLevel(zap.FatalLevel)
			log.Info(ctx, "set grpc log level to fatal")
		default:
			return result, status.Errorf(codes.InvalidArgument, "cannot set grpc log level to %v", x.Grpc.String())
		}
	case *debug.SetLogLevelRequest_Pachyderm:
		switch x.Pachyderm {
		case debug.SetLogLevelRequest_DEBUG:
			log.SetLevel(log.DebugLevel)
			log.Debug(ctx, "set log level to debug")
		case debug.SetLogLevelRequest_INFO:
			log.SetLevel(log.InfoLevel)
			log.Info(ctx, "set log level to info")
		case debug.SetLogLevelRequest_ERROR:
			log.SetLevel(log.ErrorLevel)
			log.Error(ctx, "set log level to error")
		default:
			return result, status.Errorf(codes.InvalidArgument, "cannot set log level to %v", x.Pachyderm.String())
		}
	}
	return result, nil
}
