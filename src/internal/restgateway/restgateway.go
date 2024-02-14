package restgateway

import (
	"context"
	"net/http"

	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"

	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"

	"github.com/pachyderm/pachyderm/v2/src/admin"
	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/debug"
	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/identity"
	"github.com/pachyderm/pachyderm/v2/src/license"
	"github.com/pachyderm/pachyderm/v2/src/logs"
	"github.com/pachyderm/pachyderm/v2/src/proxy"
	"github.com/pachyderm/pachyderm/v2/src/transaction"
	"github.com/pachyderm/pachyderm/v2/src/version/versionpb"
	"github.com/pachyderm/pachyderm/v2/src/worker"
)

func NewMux(ctx context.Context, grpcConn *grpc.ClientConn) (http.Handler, error) {

	ctx = pctx.Child(ctx, "restgateway")

	mux := runtime.NewServeMux(runtime.WithIncomingHeaderMatcher(func(s string) (string, bool) {
		if s != "Content-Length" {
			return s, true
		}
		return s, false
	}))

	var errs error
	if err := pps.RegisterAPIHandler(ctx, mux, grpcConn); err != nil {
		errors.JoinInto(&errs, errors.Wrap(err, "register PPS"))
	}
	if err := pfs.RegisterAPIHandler(ctx, mux, grpcConn); err != nil {
		errors.JoinInto(&errs, errors.Wrap(err, "register PFS"))
	}
	if err := worker.RegisterWorkerHandler(ctx, mux, grpcConn); err != nil {
		errors.JoinInto(&errs, errors.Wrap(err, "register worker"))
	}
	if err := proxy.RegisterAPIHandler(ctx, mux, grpcConn); err != nil {
		errors.JoinInto(&errs, errors.Wrap(err, "register proxy"))
	}
	if err := logs.RegisterAPIHandler(ctx, mux, grpcConn); err != nil {
		errors.JoinInto(&errs, errors.Wrap(err, "register logs"))
	}
	if err := admin.RegisterAPIHandler(ctx, mux, grpcConn); err != nil {
		errors.JoinInto(&errs, errors.Wrap(err, "register admin"))
	}
	if err := auth.RegisterAPIHandler(ctx, mux, grpcConn); err != nil {
		errors.JoinInto(&errs, errors.Wrap(err, "register auth"))
	}
	if err := license.RegisterAPIHandler(ctx, mux, grpcConn); err != nil {
		errors.JoinInto(&errs, errors.Wrap(err, "register license"))
	}
	if err := identity.RegisterAPIHandler(ctx, mux, grpcConn); err != nil {
		errors.JoinInto(&errs, errors.Wrap(err, "register identity"))
	}
	if err := debug.RegisterDebugHandler(ctx, mux, grpcConn); err != nil {
		errors.JoinInto(&errs, errors.Wrap(err, "register debug"))
	}
	if err := enterprise.RegisterAPIHandler(ctx, mux, grpcConn); err != nil {
		errors.JoinInto(&errs, errors.Wrap(err, "register enterprise"))
	}
	if err := transaction.RegisterAPIHandler(ctx, mux, grpcConn); err != nil {
		errors.JoinInto(&errs, errors.Wrap(err, "register transaction"))
	}
	if err := versionpb.RegisterAPIHandler(ctx, mux, grpcConn); err != nil {
		errors.JoinInto(&errs, errors.Wrap(err, "register version"))
	}
	return mux, errs
}
