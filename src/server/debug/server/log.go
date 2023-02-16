package server

import (
	"context"

	"github.com/gogo/protobuf/types"
	"github.com/pachyderm/pachyderm/v2/src/debug"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *debugServer) SetLogLevel(ctx context.Context, req *debug.SetLogLevelRequest) (*debug.SetLogLevelResponse, error) {
	result := new(debug.SetLogLevelResponse)
	d, err := types.DurationFromProto(req.GetDuration())
	if err != nil {
		return result, status.Errorf(codes.InvalidArgument, "invalid duration: %v", err)
	}
	switch x := req.GetLevel().(type) {
	case nil:
		return result, status.Error(codes.InvalidArgument, "no level provided")
	case *debug.SetLogLevelRequest_Grpc:
		switch x.Grpc {
		case debug.SetLogLevelRequest_DEBUG:
			log.SetGRPCLogLevelFor(zap.DebugLevel, d)
			log.Info(ctx, "set grpc log level to debug", zap.Duration("revert_after", d))
		case debug.SetLogLevelRequest_INFO:
			log.SetGRPCLogLevelFor(zap.InfoLevel, d)
			log.Info(ctx, "set grpc log level to info", zap.Duration("revert_after", d))
		case debug.SetLogLevelRequest_ERROR:
			log.SetGRPCLogLevelFor(zap.ErrorLevel, d)
			log.Info(ctx, "set grpc log level to error", zap.Duration("revert_after", d))
		case debug.SetLogLevelRequest_OFF:
			log.SetGRPCLogLevelFor(zap.FatalLevel, d)
			log.Info(ctx, "set grpc log level to fatal", zap.Duration("revert_after", d))
		default:
			return result, status.Errorf(codes.InvalidArgument, "cannot set grpc log level to %v", x.Grpc.String())
		}
	case *debug.SetLogLevelRequest_Pachyderm:
		switch x.Pachyderm {
		case debug.SetLogLevelRequest_DEBUG:
			log.SetLevelFor(log.DebugLevel, d)
			log.Debug(ctx, "set log level to debug", zap.Duration("revert_after", d))
		case debug.SetLogLevelRequest_INFO:
			log.SetLevelFor(log.InfoLevel, d)
			log.Info(ctx, "set log level to info", zap.Duration("revert_after", d))
		case debug.SetLogLevelRequest_ERROR:
			log.SetLevelFor(log.ErrorLevel, d)
			log.Error(ctx, "set log level to error", zap.Duration("revert_after", d))
		default:
			return result, status.Errorf(codes.InvalidArgument, "cannot set log level to %v", x.Pachyderm.String())
		}
	}
	return result, nil
}
