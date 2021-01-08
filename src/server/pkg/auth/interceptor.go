package auth

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
)

// RPCs to allow without checking authentication
var rpcWhitelist = map[string]bool{
	"/admin.API/InspectCluster": true,
	"/admin.API/InspectCluster": true,
}

type AuthInterceptor struct{}

func (ai *AuthInterceptor) InterceptUnary(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	fmt.Printf("Intercepted unary call %v\n", info.FullMethod)
	return handler(ctx, req)
}
