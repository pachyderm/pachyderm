//nolint:wrapcheck
package logging

import (
	"context"
	"fmt"
	"io"
	"reflect"
	"time"

	"github.com/gogo/protobuf/types"

	"github.com/gogo/protobuf/proto"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/identity"
	"github.com/pachyderm/pachyderm/v2/src/license"
)

type logConfig struct {
	// interfaces here are the request and response protobufs, respectively
	level             func(error) logrus.Level
	transformRequest  func(interface{}) interface{}
	transformResponse func(interface{}) interface{}
}

func defaultLevel(err error) logrus.Level {
	switch status.Code(err) {
	case codes.OK:
		// status.Code(nil) returns OK
		return logrus.InfoLevel
	case codes.InvalidArgument, codes.OutOfRange:
		return logrus.InfoLevel
	case codes.NotFound, codes.AlreadyExists, codes.Unauthenticated:
		return logrus.WarnLevel
	default:
		return logrus.ErrorLevel
	}
}

var defaultConfig = logConfig{level: defaultLevel}

// The config used for auth endpoints suppresses 'not activated' errors to the
// debug level
var authConfig = logConfig{
	level: func(err error) logrus.Level {
		if err == nil {
			return logrus.InfoLevel
		} else if auth.IsErrNotActivated(err) {
			return logrus.DebugLevel
		}
		return logrus.ErrorLevel
	},
}

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

	"/auth_v2.API/WhoAmI": {
		level: func(err error) logrus.Level {
			if err == nil {
				return logrus.DebugLevel
			}
			return authConfig.level(err)
		},
	},

	"/auth_v2.API/Activate": {
		level: authConfig.level,
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
		level: authConfig.level,
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
		level: authConfig.level,
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
		level: authConfig.level,
		transformResponse: func(r interface{}) interface{} {
			copyResp := proto.Clone(r.(*auth.GetOIDCLoginResponse)).(*auth.GetOIDCLoginResponse)
			copyResp.LoginURL = ""
			copyResp.State = ""
			return copyResp
		},
	},

	"/auth_v2.API/SetConfiguration": {
		level: authConfig.level,
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
		level: authConfig.level,
		transformResponse: func(r interface{}) interface{} {
			copyResp := proto.Clone(r.(*auth.GetRobotTokenResponse)).(*auth.GetRobotTokenResponse)
			copyResp.Token = ""
			return copyResp
		},
	},

	"/auth_v2.API/RevokeAuthToken": {
		level: authConfig.level,
		transformRequest: func(r interface{}) interface{} {
			copyReq := proto.Clone(r.(*auth.RevokeAuthTokenRequest)).(*auth.RevokeAuthTokenRequest)
			copyReq.Token = ""
			return copyReq
		},
	},

	"/auth_v2.API/ExtractAuthTokens": {
		level: authConfig.level,
		transformResponse: func(r interface{}) interface{} {
			copyResp := proto.Clone(r.(*auth.ExtractAuthTokensResponse)).(*auth.ExtractAuthTokensResponse)
			copyResp.Tokens = nil
			return copyResp
		},
	},

	"/auth_v2.API/RestoreAuthToken": {
		level: authConfig.level,
		transformRequest: func(r interface{}) interface{} {
			copyReq := proto.Clone(r.(*auth.RestoreAuthTokenRequest)).(*auth.RestoreAuthTokenRequest)
			copyReq.Token = nil
			return copyReq
		},
	},

	"/auth_v2.API/GetConfiguration": {
		level: authConfig.level,
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

func (li *LoggingInterceptor) UnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, retErr error) {
	ctx = withMethodName(ctx, info.FullMethod)
	start := time.Now()
	config := getConfig(info.FullMethod)
	level := defaultLevel(nil)
	if config.level != nil {
		level = config.level(nil)
	}

	logReq := req
	if config.transformRequest != nil && !isNilInterface(req) {
		logReq = config.transformRequest(req)
	}

	li.logUnaryBefore(ctx, level, logReq, info.FullMethod, start)
	defer func() {
		level := defaultLevel(retErr)
		if config.level != nil {
			level = config.level(retErr)
		}

		logResp := resp
		if config.transformResponse != nil && !isNilInterface(resp) {
			logResp = config.transformResponse(resp)
		}
		li.logUnaryAfter(ctx, level, logReq, info.FullMethod, start, logResp, retErr)
	}()

	return handler(ctx, req)
}

func (li *LoggingInterceptor) StreamServerInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (retErr error) {
	start := time.Now()
	config := getConfig(info.FullMethod)
	wrapper := &streamWrapper{stream: stream,
		ctx: withMethodName(stream.Context(), info.FullMethod),
	}

	// Log the first received message as the request
	var req interface{}
	var logReq interface{}
	wrapper.onFirstRecv = func(m interface{}) {
		req = m
		level := defaultLevel(nil)
		if config.level != nil {
			level = config.level(nil)
		}

		logReq = req
		if config.transformRequest != nil && !isNilInterface(req) {
			logReq = config.transformRequest(req)
		}
		li.logUnaryBefore(stream.Context(), level, logReq, info.FullMethod, start)
	}

	var resp interface{}
	if !info.IsServerStream {
		// Not a streaming response, save the first sent message for the defer
		wrapper.onFirstSend = func(m interface{}) {
			resp = m
		}
	}

	defer func() {
		level := defaultLevel(retErr)
		if config.level != nil {
			level = config.level(retErr)
		}

		logResp := resp
		if config.transformResponse != nil && !isNilInterface(resp) {
			logResp = config.transformResponse(resp)
		} else if info.IsServerStream {
			logResp = fmt.Sprintf("stream containing %d objects", wrapper.sent)
		}
		li.logUnaryAfter(stream.Context(), level, logReq, info.FullMethod, start, logResp, retErr)
	}()

	return handler(srv, wrapper)
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
