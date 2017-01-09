package grpcutil

import (
	"strings"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	log "github.com/Sirupsen/logrus"
)

type fullMethod struct {
	Package string
	Service string
	Method  string
}

// parseFullMethod parses grpc.UnaryServerInfo.FullMethod "/package.service/method"
// Note that the package itself can contain '.'
func parseFullMethod(s string) fullMethod {
	dot := strings.LastIndexByte(s, '.')
	slash := strings.LastIndexByte(s, '/')
	return fullMethod{
		Package: string(s[1:dot]),
		Service: string(s[dot+1 : slash]),
		Method:  string(s[slash+1:]),
	}
}

// UnaryServerInterceptor implements a hook to intercept the execution of a unary RPC on the server.
func UnaryServerInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	m := parseFullMethod(info.FullMethod)
	start := time.Now()
	resp, err := handler(ctx, req)
	l := log.WithFields(log.Fields{
		"service":  m.Service,
		"method":   m.Method,
		"duration": time.Since(start).Seconds(),
	})
	if err == nil {
		l.Info("rpclog")
		return resp, nil
	}
	l.WithField("req", req).WithError(err).Error("rpclog")
	return nil, err
}
