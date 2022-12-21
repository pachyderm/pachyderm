//nolint:wrapcheck
package logging

import (
	"context"
	"io"
	"reflect"
	"time"

	"github.com/gogo/protobuf/types"
	"go.uber.org/zap"

	"github.com/gogo/protobuf/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/identity"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	mauth "github.com/pachyderm/pachyderm/v2/src/internal/middleware/auth"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/license"
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
	"/license_v2.API/Activate": {
		transformRequest: func(r interface{}) interface{} {
			copyReq := proto.Clone(r.(*license.ActivateRequest)).(*license.ActivateRequest)
			copyReq.ActivationCode = ""
			return copyReq
		},
	},

	"/license_v2.API/GetActivationCode": {
		transformResponse: func(r interface{}) interface{} {
			copyResp := proto.Clone(r.(*license.GetActivationCodeResponse)).(*license.GetActivationCodeResponse)
			copyResp.ActivationCode = ""
			return copyResp
		},
	},

	"/license_v2.API/AddCluster": {
		transformRequest: func(r interface{}) interface{} {
			copyReq := proto.Clone(r.(*license.AddClusterRequest)).(*license.AddClusterRequest)
			copyReq.Secret = ""
			return copyReq
		},
		transformResponse: func(r interface{}) interface{} {
			copyResp := proto.Clone(r.(*license.AddClusterResponse)).(*license.AddClusterResponse)
			copyResp.Secret = ""
			return copyResp
		},
	},

	"/license_v2.API/Heartbeat": {
		transformRequest: func(r interface{}) interface{} {
			copyReq := proto.Clone(r.(*license.HeartbeatRequest)).(*license.HeartbeatRequest)
			copyReq.Secret = ""
			return copyReq
		},
		transformResponse: func(r interface{}) interface{} {
			copyResp := proto.Clone(r.(*license.HeartbeatResponse)).(*license.HeartbeatResponse)
			copyResp.License.ActivationCode = ""
			return copyResp
		},
	},

	"/enterprise_v2.API/GetActivationCode": {
		transformResponse: func(r interface{}) interface{} {
			copyResp := proto.Clone(r.(*enterprise.GetActivationCodeResponse)).(*enterprise.GetActivationCodeResponse)
			copyResp.ActivationCode = ""
			return copyResp
		},
	},

	"/enterprise_v2.API/GetState": {
		transformResponse: func(r interface{}) interface{} {
			copyResp := proto.Clone(r.(*enterprise.GetStateResponse)).(*enterprise.GetStateResponse)
			copyResp.ActivationCode = ""
			return copyResp
		},
	},

	"/enterprise_v2.API/Activate": {
		transformRequest: func(r interface{}) interface{} {
			copyReq := proto.Clone(r.(*enterprise.ActivateRequest)).(*enterprise.ActivateRequest)
			copyReq.Secret = ""
			return copyReq
		},
	},

	"/identity_v2.API/CreateIDPConnector": {
		transformRequest: func(r interface{}) interface{} {
			copyReq := proto.Clone(r.(*identity.CreateIDPConnectorRequest)).(*identity.CreateIDPConnectorRequest)
			if copyReq.Connector != nil {
				copyReq.Connector.Config = &types.Struct{}
				copyReq.Connector.JsonConfig = ""
			}
			return copyReq
		},
	},

	"/identity_v2.API/GetIDPConnector": {
		transformResponse: func(r interface{}) interface{} {
			copyReq := proto.Clone(r.(*identity.GetIDPConnectorResponse)).(*identity.GetIDPConnectorResponse)
			copyReq.Connector.Config = &types.Struct{}
			copyReq.Connector.JsonConfig = ""
			return copyReq
		},
	},

	"/identity_v2.API/UpdateIDPConnector": {
		transformRequest: func(r interface{}) interface{} {
			copyReq := proto.Clone(r.(*identity.UpdateIDPConnectorRequest)).(*identity.UpdateIDPConnectorRequest)
			if copyReq.Connector != nil {
				copyReq.Connector.Config = &types.Struct{}
				copyReq.Connector.JsonConfig = ""
			}
			return copyReq
		},
	},

	"/identity_v2.API/ListIDPConnectors": {
		transformResponse: func(r interface{}) interface{} {
			copyResp := proto.Clone(r.(*identity.ListIDPConnectorsResponse)).(*identity.ListIDPConnectorsResponse)
			for _, c := range copyResp.Connectors {
				c.Config = &types.Struct{}
				c.JsonConfig = ""
			}
			return copyResp
		},
	},

	"/identity_v2.API/CreateOIDCClient": {
		transformRequest: func(r interface{}) interface{} {
			copyReq := proto.Clone(r.(*identity.CreateOIDCClientRequest)).(*identity.CreateOIDCClientRequest)
			if copyReq.Client != nil {
				copyReq.Client.Secret = ""
			}
			return copyReq
		},
	},

	"/identity_v2.API/UpdateOIDCClient": {
		transformRequest: func(r interface{}) interface{} {
			copyReq := proto.Clone(r.(*identity.UpdateOIDCClientRequest)).(*identity.UpdateOIDCClientRequest)
			if copyReq.Client != nil {
				copyReq.Client.Secret = ""
			}
			return copyReq
		},
	},

	"/identity_v2.API/GetOIDCClient": {
		transformResponse: func(r interface{}) interface{} {
			copyResp := proto.Clone(r.(*identity.GetOIDCClientResponse)).(*identity.GetOIDCClientResponse)
			copyResp.Client.Secret = ""
			return copyResp
		},
	},

	"/identity_v2.API/ListOIDCClients": {
		transformResponse: func(r interface{}) interface{} {
			copyResp := proto.Clone(r.(*identity.ListOIDCClientsResponse)).(*identity.ListOIDCClientsResponse)
			for _, c := range copyResp.Clients {
				c.Secret = ""
			}
			return copyResp
		},
	},

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

	"/auth_v2.API/Activate": {
		transformRequest: func(r interface{}) interface{} {
			copyReq := proto.Clone(r.(*auth.ActivateRequest)).(*auth.ActivateRequest)
			copyReq.RootToken = ""
			return copyReq
		},
		transformResponse: func(r interface{}) interface{} {
			copyResp := proto.Clone(r.(*auth.ActivateResponse)).(*auth.ActivateResponse)
			copyResp.PachToken = ""
			return copyResp
		},
	},

	"/auth_v2.API/Authenticate": {
		transformRequest: func(r interface{}) interface{} {
			copyReq := proto.Clone(r.(*auth.AuthenticateRequest)).(*auth.AuthenticateRequest)
			copyReq.OIDCState = ""
			copyReq.IdToken = ""
			return copyReq
		},
		transformResponse: func(r interface{}) interface{} {
			copyResp := proto.Clone(r.(*auth.AuthenticateResponse)).(*auth.AuthenticateResponse)
			copyResp.PachToken = ""
			return copyResp
		},
	},

	"/auth_v2.API/RotateRootToken": {
		transformRequest: func(r interface{}) interface{} {
			copyReq := proto.Clone(r.(*auth.RotateRootTokenRequest)).(*auth.RotateRootTokenRequest)
			copyReq.RootToken = ""
			return copyReq
		},
		transformResponse: func(r interface{}) interface{} {
			copyResp := proto.Clone(r.(*auth.RotateRootTokenResponse)).(*auth.RotateRootTokenResponse)
			copyResp.RootToken = ""
			return copyResp
		},
	},

	"/auth_v2.API/GetOIDCLogin": {
		transformResponse: func(r interface{}) interface{} {
			copyResp := proto.Clone(r.(*auth.GetOIDCLoginResponse)).(*auth.GetOIDCLoginResponse)
			copyResp.LoginURL = ""
			copyResp.State = ""
			return copyResp
		},
	},

	"/auth_v2.API/SetConfiguration": {
		transformRequest: func(r interface{}) interface{} {
			req := r.(*auth.SetConfigurationRequest)
			if req.Configuration == nil {
				return r
			}
			copyReq := proto.Clone(req).(*auth.SetConfigurationRequest)
			copyReq.Configuration.ClientSecret = ""
			return copyReq
		},
	},

	"/auth_v2.API/GetRobotToken": {
		transformResponse: func(r interface{}) interface{} {
			copyResp := proto.Clone(r.(*auth.GetRobotTokenResponse)).(*auth.GetRobotTokenResponse)
			copyResp.Token = ""
			return copyResp
		},
	},

	"/auth_v2.API/RevokeAuthToken": {
		transformRequest: func(r interface{}) interface{} {
			copyReq := proto.Clone(r.(*auth.RevokeAuthTokenRequest)).(*auth.RevokeAuthTokenRequest)
			copyReq.Token = ""
			return copyReq
		},
	},

	"/auth_v2.API/ExtractAuthTokens": {
		transformResponse: func(r interface{}) interface{} {
			copyResp := proto.Clone(r.(*auth.ExtractAuthTokensResponse)).(*auth.ExtractAuthTokensResponse)
			copyResp.Tokens = nil
			return copyResp
		},
	},

	"/auth_v2.API/RestoreAuthToken": {
		transformRequest: func(r interface{}) interface{} {
			copyReq := proto.Clone(r.(*auth.RestoreAuthTokenRequest)).(*auth.RestoreAuthTokenRequest)
			copyReq.Token = nil
			return copyReq
		},
	},

	"/auth_v2.API/GetConfiguration": {
		transformResponse: func(r interface{}) interface{} {
			resp := r.(*auth.GetConfigurationResponse)
			if resp.Configuration == nil {
				return resp
			}
			copyResp := proto.Clone(resp).(*auth.GetConfigurationResponse)
			copyResp.Configuration.ClientSecret = ""
			return copyResp
		},
	},

	"/pfs_v2.API/CreateFileSet": {
		transformRequest: func(r interface{}) interface{} {
			return nil
		},
	},

	"/pfs_v2.API/ModifyFile": { // TODO(jonathan): Provisional.  This is the only RPC that has big requests via wrapped BytesValues.
		transformRequest: func(r interface{}) interface{} {
			return "<modify file>"
		},
	},

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

	if user := mauth.GetWhoAmI(ctx); user != "" {
		f = append(f, zap.String("user", user))
	}
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
	// call methods on a nil status, though.  It also returns stats=Unknown, ok=false if the
	// error doesn't have a gRPC code in it.  So we want to copy status information into the log
	// even when ok is false.
	s, _ := status.FromError(err)
	f = append(f, zap.Uint32("grpc.code", uint32(s.Code()))) // always want code, even if it's 0 (= "OK")
	if msg := s.Message(); msg != "" {
		f = append(f, zap.String("grpc.message", msg))
	}
	return pctx.Child(ctx, "", pctx.WithFields(f...))
}

func (li *LoggingInterceptor) UnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, retErr error) {
	start := time.Now()
	config := getConfig(info.FullMethod)
	lvl := li.Level

	logReq := req
	if config.transformRequest != nil && !isNilInterface(req) {
		logReq = config.transformRequest(req)
	}

	service, method := parseMethod(info.FullMethod)
	ctx = getCommonLogger(ctx, service, method)

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

func (li *LoggingInterceptor) StreamServerInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (retErr error) {
	start := time.Now()
	config := getConfig(info.FullMethod)
	service, method := parseMethod(info.FullMethod)
	lvl := li.Level

	ctx := getCommonLogger(stream.Context(), service, method)

	reqCtx := getRequestLogger(ctx, nil)
	// If client streaming, then our "main" log message for this RPC is 'first message received'
	// (which is one of many messages).  If not client streaming the main message is 'request'.
	// This debug call is mostly to catch cases where the client spends a long time not sending
	// a first message.
	dolog(reqCtx, lvl, "stream started for "+service+"/"+method)

	wrapper := &streamWrapper{
		stream: stream,
		ctx:    ctx,
	}

	// Log the first received message as the request
	wrapper.onFirstRecv = func(m any) {
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
	}

	var resp any
	wrapper.onFirstSend = func(m any) {
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
			dolog(resCtx, lvl, "first mesasge sent for "+service+"/"+method)
			resp = logResp
		}
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

type streamWrapper struct {
	stream      grpc.ServerStream
	ctx         context.Context
	received    int
	sent        int
	onFirstSend func(interface{})
	onFirstRecv func(interface{})
}

func (sw *streamWrapper) SetHeader(m metadata.MD) error {
	return sw.stream.SetHeader(m)
}

func (sw *streamWrapper) SendHeader(m metadata.MD) error {
	return sw.stream.SendHeader(m)
}

func (sw *streamWrapper) SetTrailer(m metadata.MD) {
	sw.stream.SetTrailer(m)
}

func (sw *streamWrapper) Context() context.Context {
	return sw.ctx
}

func (sw *streamWrapper) SendMsg(m interface{}) error {
	err := sw.stream.SendMsg(m)
	if err == nil {
		if sw.sent == 0 && sw.onFirstSend != nil {
			sw.onFirstSend(m)
		}
		sw.sent++
	}
	return err
}

func (sw *streamWrapper) RecvMsg(m interface{}) error {
	err := sw.stream.RecvMsg(m)
	if err != io.EOF {
		if sw.received == 0 && sw.onFirstRecv != nil {
			sw.onFirstRecv(m)
		}
		sw.received++
	}
	return err
}
