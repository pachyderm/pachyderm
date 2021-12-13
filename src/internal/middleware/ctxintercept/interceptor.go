package ctxintercept

import (
	"context"
	"time"

	"github.com/pachyderm/pachyderm/v2/src/internal/middleware/util"
	"google.golang.org/grpc"
)

const (
	defaultMethodTimeout time.Duration = 20 * time.Second
	unlimited            time.Duration = -1
)

var customTimeoutMethods = map[string]time.Duration{
	//
	// Debug API
	//
	"/debug_v2.Debug/Profile": 10 * time.Second,
	"/debug_v2.Debug/Binary":  10 * time.Second,
	"/debug_v2.Debug/Dump":    10 * time.Second,

	//
	// PFS API
	//
	"/pfs_v2.API/ActivateAuth":       unlimited,
	"/pfs_v2.API/ListRepo":           10 * time.Second,
	"/pfs_v2.API/DeleteRepo":         10 * time.Second,
	"/pfs_v2.API/FinishCommit":       20 * time.Second,
	"/pfs_v2.API/InspectCommit":      unlimited,
	"/pfs_v2.API/InspectCommitSet":   unlimited,
	"/pfs_v2.API/ListCommit":         20 * time.Second,
	"/pfs_v2.API/SubscribeCommit":    unlimited,
	"/pfs_v2.API/ListCommitSet":      20 * time.Second,
	"/pfs_v2.API/SquashCommitSet":    unlimited,
	"/pfs_v2.API/DropCommitSet":      unlimited,
	"/pfs_v2.API/ListBranch":         20 * time.Second,
	"/pfs_v2.API/ModifyFile":         20 * time.Second,
	"/pfs_v2.API/GetFile":            20 * time.Second,
	"/pfs_v2.API/GetFileTAR":         unlimited,
	"/pfs_v2.API/ListFile":           20 * time.Second,
	"/pfs_v2.API/WalkFile":           20 * time.Second,
	"/pfs_v2.API/GlobFile":           20 * time.Second,
	"/pfs_v2.API/DiffFile":           20 * time.Second,
	"/pfs_v2.API/DeleteAll":          unlimited,
	"/pfs_v2.API/Fsck":               unlimited,
	"/pfs_v2.API/RenewFileSet":       unlimited,
	"/pfs_v2.API/RunLoadTest":        unlimited,
	"/pfs_v2.API/RunLoadTestDefault": unlimited,

	//
	// PPS API
	//
	"/pps_v2.API/InspectJob":      unlimited,
	"/pps_v2.API/InspectJobSet":   unlimited,
	"/pps_v2.API/ListJob":         20 * time.Second,
	"/pps_v2.API/ListJobStream":   20 * time.Second,
	"/pps_v2.API/SubscribeJob":    unlimited,
	"/pps_v2.API/StopJob":         unlimited,
	"/pps_v2.API/ListJobSet":      20 * time.Second,
	"/pps_v2.API/ListDatum":       unlimited,
	"/pps_v2.API/ListDatumStream": 20 * time.Second,
	"/pps_v2.API/RestartDatum":    unlimited,
	"/pps_v2.API/StopPipeline":    unlimited,
	"/pps_v2.API/GetLogs":         unlimited,
	"/pps_v2.API/GarbageCollect":  unlimited,
	"/pps_v2.API/UpdateJobState":  unlimited,
	"/pps_v2.API/ListPipeline":    20 * time.Second,
	"/pps_v2.API/ActivateAuth":    unlimited,
	"/pps_v2.API/DeleteAll":       unlimited,

	"/pps_v2.API/ListSecret":         10 * time.Second,
	"/pps_v2.API/RunLoadTest":        unlimited,
	"/pps_v2.API/RunLoadTestDefault": unlimited,
}

type ContextInterceptor struct {
	enabled bool
}

func NewContextInterceptor(enabled bool) *ContextInterceptor {
	return &ContextInterceptor{enabled: enabled}
}

func (ci *ContextInterceptor) setTimeout(fullMethod string, ctx context.Context) context.Context {
	if !ci.enabled {
		return ctx
	}
	if timeout, ok := customTimeoutMethods[fullMethod]; ok {
		if timeout == unlimited {
			return ctx
		}
		newCtx, cf := context.WithTimeout(ctx, timeout)
		defer cf()
		return newCtx
	}
	newCtx, cf := context.WithTimeout(ctx, defaultMethodTimeout)
	defer cf()
	return newCtx
}

func (ci *ContextInterceptor) UnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	ctx = ci.setTimeout(info.FullMethod, ctx)
	return handler(ctx, req)
}

func (ci *ContextInterceptor) StreamServerInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	ctx := ci.setTimeout(info.FullMethod, stream.Context())
	streamProxy := util.ServerStreamWrapper{Stream: stream, Ctx: ctx}
	return handler(srv, streamProxy)
}
