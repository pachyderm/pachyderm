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
	requireAuthEnabled()
}

func requireAuthEnabled(f authHandlerFn) authHandlerFn {
	return func(i *Interceptor) authHandler {
		h := f(i)
		h.requireAuthEnabled()
		return h
	}
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

func (unauthenticatedHandler) requireAuthEnabled() {}

// authenticatedHandler allows all RPCs to succeed as long as the user has a valid auth token
type authenticatedHandler struct {
	interceptor *Interceptor

	// if checkAuthEnabled is set, fail if auth is not in the active state
	// this blocks calls when auth is disabled or paritally active
	checkAuthEnabled bool
}

func authenticated(i *Interceptor) authHandler {
	return &authenticatedHandler{
		interceptor: i,
	}
}

func (h authenticatedHandler) isPermitted(ctx context.Context) error {
	pachClient := h.interceptor.env.GetPachClient(ctx)
	resp, err := pachClient.WhoAmI(pachClient.Ctx(), &auth.WhoAmIRequest{})
	if err != nil {
		if auth.IsErrNotActivated(err) && !h.checkAuthEnabled {
			return nil
		}
		return err
	}

	if !resp.FullyActivated && h.checkAuthEnabled {
		return auth.ErrPartiallyActivated
	}
	return nil
}

func (h *authenticatedHandler) requireAuthEnabled() {
	h.checkAuthEnabled = true
}

func (h authenticatedHandler) unary(ctx context.Context, info *grpc.UnaryServerInfo, req interface{}) error {
	return h.isPermitted(ctx)
}

func (h authenticatedHandler) stream(ctx context.Context, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	return h.isPermitted(ctx)
}

func adminOnly(i *Interceptor) authHandler {
	return &adminOnlyHandler{
		interceptor: i,
	}
}

// adminOnlyHandler allows an RPC to succeed only if the user has cluster admin status
type adminOnlyHandler struct {
	interceptor *Interceptor

	// if checkAuthEnabled is set, fail if auth is not in the active state
	// this blocks calls when auth is disabled or paritally active
	checkAuthEnabled bool
}

func (h *adminOnlyHandler) requireAuthEnabled() {
	h.checkAuthEnabled = true
}

func (h adminOnlyHandler) isPermitted(ctx context.Context, fullMethod string) error {
	pachClient := h.interceptor.env.GetPachClient(ctx)
	me, err := pachClient.WhoAmI(pachClient.Ctx(), &auth.WhoAmIRequest{})
	if err != nil {
		if auth.IsErrNotActivated(err) && !h.checkAuthEnabled {
			return nil
		}
		return err
	}

	if !me.FullyActivated && h.checkAuthEnabled {
		return auth.ErrPartiallyActivated
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
