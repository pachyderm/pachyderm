package restgateway

import (
	"context"
	"net/http"

	"github.com/pachyderm/pachyderm/v2/src/internal/client"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	pfsgw "github.com/pachyderm/pachyderm/v2/src/pfs"
	ppsgw "github.com/pachyderm/pachyderm/v2/src/pps"

	admingw "github.com/pachyderm/pachyderm/v2/src/admin"
	authgw "github.com/pachyderm/pachyderm/v2/src/auth"
	debuggw "github.com/pachyderm/pachyderm/v2/src/debug"
	enterprisegw "github.com/pachyderm/pachyderm/v2/src/enterprise"
	identitygw "github.com/pachyderm/pachyderm/v2/src/identity"
	licensegw "github.com/pachyderm/pachyderm/v2/src/license"
	proxygw "github.com/pachyderm/pachyderm/v2/src/proxy"
	transactiongw "github.com/pachyderm/pachyderm/v2/src/transaction"
	versiongw "github.com/pachyderm/pachyderm/v2/src/version/versionpb"
	workergw "github.com/pachyderm/pachyderm/v2/src/worker"
)

func NewRestGatewayMux(pachClientFactory func(context.Context) *client.APIClient) http.Handler {
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

	err := ppsgw.RegisterAPIHandler(ctx, mux, client.ClientConn())
	if err != nil {
		log.Error(ctx, "restgateway: failed to register pps api with grpc-gateway")
		return nil
	}

	err = pfsgw.RegisterAPIHandler(ctx, mux, client.ClientConn())
	if err != nil {
		log.Error(ctx, "restgateway: failed to register pfs api with grpc-gateway")
		return nil
	}

	err = workergw.RegisterWorkerHandler(ctx, mux, client.ClientConn())
	if err != nil {
		log.Error(ctx, "restgateway: failed to register wroker api with grpc-gateway")
		return nil
	}

	err = proxygw.RegisterAPIHandler(ctx, mux, client.ClientConn())
	if err != nil {
		log.Error(ctx, "restgateway: failed to register proxy api with grpc-gateway")
		return nil
	}

	err = admingw.RegisterAPIHandler(ctx, mux, client.ClientConn())
	if err != nil {
		log.Error(ctx, "restgateway: failed to register admin api with grpc-gateway")
		return nil
	}

	err = authgw.RegisterAPIHandler(ctx, mux, client.ClientConn())
	if err != nil {
		log.Error(ctx, "restgateway: failed to register auth api with grpc-gateway")
		return nil
	}

	err = licensegw.RegisterAPIHandler(ctx, mux, client.ClientConn())
	if err != nil {
		log.Error(ctx, "restgateway: failed to register license api with grpc-gateway")
		return nil
	}

	err = identitygw.RegisterAPIHandler(ctx, mux, client.ClientConn())
	if err != nil {
		log.Error(ctx, "restgateway: failed to register identity api with grpc-gateway")
		return nil
	}

	err = debuggw.RegisterDebugHandler(ctx, mux, client.ClientConn())
	if err != nil {
		log.Error(ctx, "restgateway: failed to register debug api with grpc-gateway")
		return nil
	}

	err = enterprisegw.RegisterAPIHandler(ctx, mux, client.ClientConn())
	if err != nil {
		log.Error(ctx, "restgateway: failed to register enterprise api with grpc-gateway")
		return nil
	}

	err = transactiongw.RegisterAPIHandler(ctx, mux, client.ClientConn())
	if err != nil {
		log.Error(ctx, "restgateway: failed to register transaction api with grpc-gateway")
		return nil
	}

	err = versiongw.RegisterAPIHandler(ctx, mux, client.ClientConn())
	if err != nil {
		log.Error(ctx, "restgateway: failed to register version api with grpc-gateway")
		return nil
	}

	return mux
}
