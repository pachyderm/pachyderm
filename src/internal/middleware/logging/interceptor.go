//nolint:wrapcheck
package logging

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"runtime/trace"
	"time"

	"go.uber.org/zap"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

type logConfig struct {
	// interfaces here are the request and response protobufs, respectively
	transformRequest  func(interface{}) interface{}
	transformResponse func(interface{}) interface{}
	leveler           func(defaultLevel log.Level, response any, err error) log.Level
}

var defaultConfig = logConfig{}

// The config used for auth endpoints suppresses 'not activated' errors to the
// debug level
var authConfig = logConfig{}

// Special handling for some endpoints, usually regarding redaction or
// suppressing to a different log level.
// TODO: would be nice if we could annotate the protobuf fields for redaction,
// then auto-generate this code (or even just handle it dynamically).
var endpoints = map[string]logConfig{
	"/auth_v2.API/Deactivate":                 authConfig,
	"/auth_v2.API/Authorize":                  authConfig,
	"/auth_v2.API/GetPermissionsForPrincipal": authConfig,
	"/auth_v2.API/GetPermissions":             authConfig,
	"/auth_v2.API/ModifyRoleBinding":          authConfig,
	"/auth_v2.API/GetRoleBinding":             authConfig,
	"/auth_v2.API/SetGroupsForUser":           authConfig,
	"/auth_v2.API/ModifyMembers":              authConfig,
	"/auth_v2.API/GetGroups":                  authConfig,
	"/auth_v2.API/GetGroupsForPrincipal":      authConfig,
	"/auth_v2.API/GetUsers":                   authConfig,
	"/auth_v2.API/GetRolesForPermission":      authConfig,
	"/auth_v2.API/DeleteExpiredAuthTokens":    authConfig,
	"/auth_v2.API/RevokeAuthTokensForUser":    authConfig,

	"/auth_v2.API/WhoAmI": defaultConfig,

	"/grpc.health.v1.Health/Check": {
		leveler: func(lvl log.Level, a any, err error) log.Level {
			if err != nil {
				return log.ErrorLevel
			}
			if res, ok := a.(*healthpb.HealthCheckResponse); ok {
				if res.GetStatus() == healthpb.HealthCheckResponse_SERVING {
					return lvl
				}
			}
			return log.ErrorLevel
		},
	},
}

func getConfig(fullMethod string) logConfig {
	if config, ok := endpoints[fullMethod]; ok {
		return config
	}
	return defaultConfig
}

func isNilInterface(x interface{}) bool {
	val := reflect.ValueOf(x)
	return x == nil || (val.Kind() == reflect.Ptr && val.IsNil())
}

func getCommonLogger(ctx context.Context, service, method string) context.Context {
	var f []log.Field
	f = append(f, zap.String("service", service), zap.String("method", method))

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if rids := md.Get("x-request-id"); rids != nil {
			// There shouldn't be multiple copies of the x-request-id header, but if
			// there are, log all of them.
			f = append(f, zap.Strings("x-request-id", rids))
		}
		if command := md.Get("command"); command != nil {
			f = append(f, zap.Strings("command", command))
		}
	}
	if peer, ok := peer.FromContext(ctx); ok {
		f = append(f, zap.Stringer("peer", peer.Addr))
	}

	if service == "grpc.health.v1.Health" {
		// The health check logger is rate-limited to one unique message per hour.
		ctx = log.HealthCheckLogger(ctx)
	}
	return pctx.Child(ctx, service+"/"+method, pctx.WithFields(f...))
}

func getRequestLogger(ctx context.Context, req any) context.Context {
	var f []log.Field
	switch x := req.(type) {
	case nil:
	case proto.Message:
		f = append(f, log.Proto("request", x))
	default:
		f = append(f, zap.Any("request", x))
	}
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		f = append(f, log.Metadata("metadata", md))
	}
	if deadline, ok := ctx.Deadline(); ok {
		f = append(f, zap.Duration("deadline", time.Until(deadline)))
	}
	return pctx.Child(ctx, "", pctx.WithFields(f...))
}

func getResponseLogger(ctx context.Context, res any, sent, rcvd int, err error) context.Context {
	var f []log.Field
	switch x := res.(type) {
	case nil:
	case proto.Message:
		f = append(f, log.Proto("response", x))
	default:
		f = append(f, zap.Any("response", x))
	}
	if sent > 0 {
		f = append(f, zap.Int("messagesSent", sent))
	}
	if rcvd > 0 {
		f = append(f, zap.Int("messagesReceived", rcvd))
	}
	if err != nil {
		f = append(f, zap.Error(err))
	}
	// FromError is pretty weird.  It returns (status=nil, ok=true) for nil errors.  It's OK to
	// call methods on a nil status, though.  It also returns status=Unknown, ok=false if the
	// error doesn't have a gRPC code in it.  So we want to copy status information into the log
	// even when ok is false.
	s, _ := status.FromError(err)
	f = append(f, zap.Uint32("grpc.code", uint32(s.Code()))) // always want code, even if it's 0 (= "OK")
	if msg := s.Message(); msg != "" {
		f = append(f, zap.String("grpc.message", msg))
	}
	if dd := s.Details(); len(dd) > 0 {
		for i, d := range dd {
			switch d := d.(type) {
			case error:
				f = append(f, zap.NamedError(fmt.Sprintf("grpc.details[%d]", i), d))
			case proto.Message:
				f = append(f, log.Proto(fmt.Sprintf("grpc.details[%d]", i), d))
			default:
				f = append(f, zap.Any(fmt.Sprintf("grpc.details[%d]", i), d))
			}
		}
	}
	return pctx.Child(ctx, "", pctx.WithFields(f...))
}

// UnarySetup is a grpc unary server interceptor that sets up the grpc request for logging.  It
// should be near the top of the interceptor chain.
func (li *LoggingInterceptor) UnarySetup(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, retErr error) {
	ctx, task := trace.NewTask(ctx, "grpc server "+info.FullMethod)
	defer task.End()
	service, method := parseMethod(info.FullMethod)
	ctx = getCommonLogger(ctx, service, method)
	return handler(ctx, req)
}

// UnaryAnnounce is a grpc unary server interceptor that announces the beginning and end of the
// request.  It should be after the auth middleware in the interceptor chain.
func (li *LoggingInterceptor) UnaryAnnounce(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, retErr error) {
	start := time.Now()
	config := getConfig(info.FullMethod)
	lvl := li.Level
	logReq := req
	if config.transformRequest != nil && !isNilInterface(req) {
		logReq = config.transformRequest(req)
	}
	service, method := parseMethod(info.FullMethod)

	// NOTE(jonathan): We use service/method in the log messages so that rate limiting applies
	// per-RPC instead of for all RPCs.  (Rate limiting algorithm looks at the message and the
	// severity.)
	dolog(getRequestLogger(ctx, logReq), lvl, "request for "+service+"/"+method)
	defer func() {
		logResp := resp
		if config.transformResponse != nil && !isNilInterface(resp) {
			logResp = config.transformResponse(resp)
		}
		if config.leveler != nil && !isNilInterface(resp) {
			lvl = config.leveler(lvl, resp, retErr)
		}
		li.logUnaryAfter(getResponseLogger(ctx, logResp, 1, 1, retErr), lvl, service, method, start, retErr)
	}()

	return handler(ctx, req)
}

// StreamSetup is a grpc stream server interceptor that sets up the grpc request for logging.  It
// should be near the top of the interceptor chain.
func (li *LoggingInterceptor) StreamSetup(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (retErr error) {
	ctx, task := trace.NewTask(stream.Context(), "grpc server stream "+info.FullMethod)
	defer task.End()
	service, method := parseMethod(info.FullMethod)
	ctx = getCommonLogger(ctx, service, method)

	wrapper := &streamSetupWrapper{
		ServerStream: stream,
		ctx:          ctx,
	}
	return handler(srv, wrapper)
}

// StreamAnnounce is a grpc stream server interceptor that announces the beginning and end of the
// request.  It should be after the auth middleware in the interceptor chain.
func (li *LoggingInterceptor) StreamAnnounce(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (retErr error) {
	start := time.Now()
	config := getConfig(info.FullMethod)
	service, method := parseMethod(info.FullMethod)
	lvl := li.Level

	ctx := stream.Context()
	reqCtx := getRequestLogger(ctx, nil)
	// If client streaming, then our "main" log message for this RPC is 'first message received'
	// (which is one of many messages).  If not client streaming the main message is 'request'.
	// This debug call is mostly to catch cases where the client spends a long time not sending
	// a first message.
	dolog(reqCtx, lvl, "stream started for "+service+"/"+method)

	var resp any
	wrapper := &streamAnnounceWrapper{
		ServerStream: stream,
		onFirstRecv: func(m any) {
			// Log the first received message as the request
			logReq := m
			if config.transformRequest != nil && !isNilInterface(m) {
				logReq = config.transformRequest(m)
			}
			reqCtx := getRequestLogger(ctx, logReq)
			if info.IsClientStream {
				dolog(reqCtx, lvl, "first message received for "+service+"/"+method)
			} else {
				dolog(reqCtx, lvl, "request for "+service+"/"+method)
			}
		},
		onFirstSend: func(m any) {
			logResp := m
			if config.transformResponse != nil && !isNilInterface(m) {
				logResp = config.transformResponse(m)
			}

			// If server stream, log the first message.  If not, the call should be ending soon,
			// and we'll log it as the response.  The !IsServerStream case is for catching
			// slowness between sending the first and only message, and actually returning.
			if info.IsServerStream {
				resCtx := getResponseLogger(ctx, logResp, 0, 0, nil)
				dolog(resCtx, lvl, "first message sent for "+service+"/"+method)
				resp = nil
			} else {
				resCtx := getResponseLogger(ctx, nil, 0, 0, nil)
				dolog(resCtx, lvl, "first message sent for "+service+"/"+method)
				resp = logResp
			}
		},
	}

	defer func() {
		resCtx := getResponseLogger(ctx, resp, wrapper.sent, wrapper.received, retErr)
		if config.leveler != nil && !isNilInterface(resp) {
			lvl = config.leveler(lvl, resp, retErr)
		}
		li.logUnaryAfter(resCtx, lvl, service, method, start, retErr)
	}()
	return handler(srv, wrapper)
}

func dolog(ctx context.Context, lvl log.Level, msg string, f ...log.Field) {
	switch lvl { //exhaustive:enforce
	case log.DebugLevel:
		log.Debug(ctx, msg, f...)
	case log.InfoLevel:
		log.Info(ctx, msg, f...)
	case log.ErrorLevel:
		log.Error(ctx, msg, f...)
	}
}

type streamSetupWrapper struct {
	grpc.ServerStream
	ctx context.Context
}

func (sw *streamSetupWrapper) Context() context.Context {
	return sw.ctx
}

type streamAnnounceWrapper struct {
	grpc.ServerStream
	received    int
	sent        int
	onFirstSend func(interface{})
	onFirstRecv func(interface{})
}

func (sw *streamAnnounceWrapper) SendMsg(m interface{}) error {
	trace.Logf(sw.Context(), "grpc server", "grpc server send %T", m)
	err := sw.ServerStream.SendMsg(m)
	if err == nil {
		if sw.sent == 0 && sw.onFirstSend != nil {
			sw.onFirstSend(m)
		}
		sw.sent++
	}
	return err
}

func (sw *streamAnnounceWrapper) RecvMsg(m interface{}) error {
	trace.Logf(sw.Context(), "grpc server", "grpc server recv %T", m)
	err := sw.ServerStream.RecvMsg(m)
	if err != io.EOF {
		if sw.received == 0 && sw.onFirstRecv != nil {
			sw.onFirstRecv(m)
		}
		sw.received++
	}
	return err
}
