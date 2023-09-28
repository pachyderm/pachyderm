package restgateway

import (
	"context"
	"net/http"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	"github.com/pachyderm/pachyderm/v2/src/pfs"
	"github.com/pachyderm/pachyderm/v2/src/pps"

	"github.com/pachyderm/pachyderm/v2/src/admin"
	"github.com/pachyderm/pachyderm/v2/src/auth"
	"github.com/pachyderm/pachyderm/v2/src/debug"
	"github.com/pachyderm/pachyderm/v2/src/enterprise"
	"github.com/pachyderm/pachyderm/v2/src/identity"
	"github.com/pachyderm/pachyderm/v2/src/license"
	"github.com/pachyderm/pachyderm/v2/src/proxy"
	"github.com/pachyderm/pachyderm/v2/src/transaction"
	"github.com/pachyderm/pachyderm/v2/src/version/versionpb"
	"github.com/pachyderm/pachyderm/v2/src/worker"
)

func NewMux(pachClientFactory func(context.Context) *client.APIClient) http.Handler {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	client := pachClientFactory(ctx)

	mux := runtime.NewServeMux(runtime.WithIncomingHeaderMatcher(func(s string) (string, bool) {
		if s != "Content-Length" {
			return s, true
		}
		return s, false
	}))

	err := pps.RegisterAPIHandler(ctx, mux, client.ClientConn())
	if err != nil {
		log.Error(ctx, "restgateway: failed to register pps api with grpc-gateway")
		return nil
	}

	err = pfs.RegisterAPIHandler(ctx, mux, client.ClientConn())
	if err != nil {
		log.Error(ctx, "restgateway: failed to register pfs api with grpc-gateway")
		return nil
	}

	err = worker.RegisterWorkerHandler(ctx, mux, client.ClientConn())
	if err != nil {
		log.Error(ctx, "restgateway: failed to register wroker api with grpc-gateway")
		return nil
	}

	err = proxy.RegisterAPIHandler(ctx, mux, client.ClientConn())
	if err != nil {
		log.Error(ctx, "restgateway: failed to register proxy api with grpc-gateway")
		return nil
	}

	err = admin.RegisterAPIHandler(ctx, mux, client.ClientConn())
	if err != nil {
		log.Error(ctx, "restgateway: failed to register admin api with grpc-gateway")
		return nil
	}

	err = auth.RegisterAPIHandler(ctx, mux, client.ClientConn())
	if err != nil {
		log.Error(ctx, "restgateway: failed to register auth api with grpc-gateway")
		return nil
	}

	err = license.RegisterAPIHandler(ctx, mux, client.ClientConn())
	if err != nil {
		log.Error(ctx, "restgateway: failed to register license api with grpc-gateway")
		return nil
	}

	err = identity.RegisterAPIHandler(ctx, mux, client.ClientConn())
	if err != nil {
		log.Error(ctx, "restgateway: failed to register identity api with grpc-gateway")
		return nil
	}

	err = debug.RegisterDebugHandler(ctx, mux, client.ClientConn())
	if err != nil {
		log.Error(ctx, "restgateway: failed to register debug api with grpc-gateway")
		return nil
	}

	err = enterprise.RegisterAPIHandler(ctx, mux, client.ClientConn())
	if err != nil {
		log.Error(ctx, "restgateway: failed to register enterprise api with grpc-gateway")
		return nil
	}

	err = transaction.RegisterAPIHandler(ctx, mux, client.ClientConn())
	if err != nil {
		log.Error(ctx, "restgateway: failed to register transaction api with grpc-gateway")
		return nil
	}

	err = versionpb.RegisterAPIHandler(ctx, mux, client.ClientConn())
	if err != nil {
		log.Error(ctx, "restgateway: failed to register version api with grpc-gateway")
		return nil
	}

	return mux
}
