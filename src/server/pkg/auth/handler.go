package auth

import (
	"context"

	"github.com/pachyderm/pachyderm/src/client/auth"

	"google.golang.org/grpc"
)

type authHandlerFn func(*Interceptor) authHandler

// authHandler methods return an error if the request is not permitted
type authHandler interface {
	unary(ctx context.Context, info *grpc.UnaryServerInfo, req interface{}) error
	stream(ctx context.Context, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error
}

// unauthenticatedHandler allows all RPCs to succeed, even if there is no auth metadata
type unauthenticatedHandler struct{}

func unauthenticated(_ *Interceptor) authHandler {
	return unauthenticatedHandler{}
}

func (unauthenticatedHandler) unary(ctx context.Context, info *grpc.UnaryServerInfo, req interface{}) error {
	return nil
}

func (unauthenticatedHandler) stream(ctx context.Context, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	return nil
}

// authenticatedHandler allows all RPCs to succeed as long as the user has a valid auth token
type authenticatedHandler struct {
	i *Interceptor
}

func authenticated(i *Interceptor) authHandler {
	return authenticatedHandler{i}
}

func (h authenticatedHandler) isPermitted(ctx context.Context) error {
	pachClient := h.i.env.GetPachClient(ctx)
	_, err := pachClient.WhoAmI(pachClient.Ctx(), &auth.WhoAmIRequest{})
	if err != nil && !auth.IsErrNotActivated(err) {
		return err
	}
	return nil
}

func (h authenticatedHandler) unary(ctx context.Context, info *grpc.UnaryServerInfo, req interface{}) error {
	return h.isPermitted(ctx)
}

func (h authenticatedHandler) stream(ctx context.Context, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	return h.isPermitted(ctx)
}

func adminOnly(i *Interceptor) authHandler {
	return adminOnlyHandler{i}
}

// adminOnlyHandler allows an RPC to succeed only if the user has cluster admin status
type adminOnlyHandler struct {
	i *Interceptor
}

func (h adminOnlyHandler) isPermitted(ctx context.Context, fullMethod string) error {
	pachClient := h.i.env.GetPachClient(ctx)
	me, err := pachClient.WhoAmI(pachClient.Ctx(), &auth.WhoAmIRequest{})
	if err != nil {
		if auth.IsErrNotActivated(err) {
			return nil
		}
		return err
	}

	if me.ClusterRoles != nil {
		for _, s := range me.ClusterRoles.Roles {
			if s == auth.ClusterRole_SUPER {
				return nil
			}
		}
	}

	return &auth.ErrNotAuthorized{
		Subject: me.Username,
		AdminOp: fullMethod,
	}
}

func (h adminOnlyHandler) unary(ctx context.Context, info *grpc.UnaryServerInfo, req interface{}) error {
	return h.isPermitted(ctx, info.FullMethod)
}

func (h adminOnlyHandler) stream(ctx context.Context, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	return h.isPermitted(ctx, info.FullMethod)
}
