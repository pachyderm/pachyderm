package server

import (
	"context"
	"net/http"
	"time"

	"github.com/gogo/protobuf/proto"
	logrus "github.com/sirupsen/logrus"

	"github.com/pachyderm/pachyderm/v2/src/identity"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/serviceenv"
)

const (
	dexHTTPPort = ":658"

	configKey = 1
)

type apiServer struct {
	pachLogger log.Logger
	env        serviceenv.ServiceEnv

	api *dexAPI
}

// LogReq is like log.Logger.Log(), but it assumes that it's being called from
// the top level of a GRPC method implementation, and correspondingly extracts
// the method name from the parent stack frame
func (a *apiServer) LogReq(request interface{}) {
	a.pachLogger.Log(request, nil, nil, 0)
}

// LogResp is like log.Logger.Log(). However,
// 1) It assumes that it's being called from a defer() statement in a GRPC
//    method , and correspondingly extracts the method name from the grandparent
//    stack frame
func (a *apiServer) LogResp(request interface{}, response interface{}, err error, duration time.Duration) {
	if err == nil {
		a.pachLogger.LogAtLevelFromDepth(request, response, err, duration, logrus.InfoLevel, 4)
	} else {
		a.pachLogger.LogAtLevelFromDepth(request, response, err, duration, logrus.ErrorLevel, 4)
	}
}

// NewIdentityServer returns an implementation of identity.APIServer.
func NewIdentityServer(env serviceenv.ServiceEnv, public bool) identity.APIServer {
	server := &apiServer{
		env:        env,
		pachLogger: log.NewLogger("identity.API", env.Logger()),
		api:        newDexAPI(env.GetDexDB()),
	}

	if public {
		web := newDexWeb(env, server)
		go func() {
			if err := http.ListenAndServe(dexHTTPPort, web); err != nil {
				logrus.WithError(err).Fatalf("error setting up and/or running the identity server")
			}
		}()
	}

	return server
}

func (a *apiServer) SetIdentityServerConfig(ctx context.Context, req *identity.SetIdentityServerConfigRequest) (resp *identity.SetIdentityServerConfigResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	if _, err := a.env.GetDBClient().ExecContext(ctx, `INSERT INTO identity.config (id, issuer) VALUES ($1, $2) ON CONFLICT (id) DO UPDATE SET issuer=$2 `, configKey, req.Config.Issuer); err != nil {
		return nil, err
	}

	return &identity.SetIdentityServerConfigResponse{}, nil
}

func (a *apiServer) GetIdentityServerConfig(ctx context.Context, req *identity.GetIdentityServerConfigRequest) (resp *identity.GetIdentityServerConfigResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	var config []*identity.IdentityServerConfig
	err := a.env.GetDBClient().SelectContext(ctx, &config, "SELECT issuer FROM identity.config WHERE id=$1;", configKey)
	if err != nil {
		return nil, err
	}
	if len(config) == 0 {
		return &identity.GetIdentityServerConfigResponse{Config: &identity.IdentityServerConfig{}}, nil
	}

	return &identity.GetIdentityServerConfigResponse{Config: config[0]}, nil
}

func (a *apiServer) CreateIDPConnector(ctx context.Context, req *identity.CreateIDPConnectorRequest) (resp *identity.CreateIDPConnectorResponse, retErr error) {
	removeJSONConfig := func(r *identity.CreateIDPConnectorRequest) *identity.CreateIDPConnectorRequest {
		copyReq := proto.Clone(r).(*identity.CreateIDPConnectorRequest)
		copyReq.Connector.JsonConfig = ""
		return copyReq
	}

	a.LogReq(removeJSONConfig(req))
	defer func(start time.Time) { a.LogResp(removeJSONConfig(req), resp, retErr, time.Since(start)) }(time.Now())

	if err := a.api.createConnector(req); err != nil {
		return nil, err
	}
	return &identity.CreateIDPConnectorResponse{}, nil
}

func (a *apiServer) GetIDPConnector(ctx context.Context, req *identity.GetIDPConnectorRequest) (resp *identity.GetIDPConnectorResponse, retErr error) {
	removeJSONConfig := func(r *identity.GetIDPConnectorResponse) *identity.GetIDPConnectorResponse {
		copyResp := proto.Clone(r).(*identity.GetIDPConnectorResponse)
		copyResp.Connector.JsonConfig = ""
		return copyResp
	}

	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, removeJSONConfig(resp), retErr, time.Since(start)) }(time.Now())

	c, err := a.api.getConnector(req.Id)
	if err != nil {
		return nil, err
	}

	return &identity.GetIDPConnectorResponse{
		Connector: c,
	}, nil
}

func (a *apiServer) UpdateIDPConnector(ctx context.Context, req *identity.UpdateIDPConnectorRequest) (resp *identity.UpdateIDPConnectorResponse, retErr error) {
	removeJSONConfig := func(r *identity.UpdateIDPConnectorRequest) *identity.UpdateIDPConnectorRequest {
		copyReq := proto.Clone(r).(*identity.UpdateIDPConnectorRequest)
		copyReq.Connector.JsonConfig = ""
		return copyReq
	}

	a.LogReq(removeJSONConfig(req))
	defer func(start time.Time) { a.LogResp(removeJSONConfig(req), resp, retErr, time.Since(start)) }(time.Now())

	if err := a.api.updateConnector(req); err != nil {
		return nil, err
	}

	return &identity.UpdateIDPConnectorResponse{}, nil
}

func (a *apiServer) ListIDPConnectors(ctx context.Context, req *identity.ListIDPConnectorsRequest) (resp *identity.ListIDPConnectorsResponse, retErr error) {
	removeJSONConfig := func(r *identity.ListIDPConnectorsResponse) *identity.ListIDPConnectorsResponse {
		copyResp := proto.Clone(r).(*identity.ListIDPConnectorsResponse)
		for _, c := range copyResp.Connectors {
			c.JsonConfig = ""
		}
		return copyResp
	}

	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, removeJSONConfig(resp), retErr, time.Since(start)) }(time.Now())

	connectors, err := a.api.listConnectors()
	if err != nil {
		return nil, err
	}

	return &identity.ListIDPConnectorsResponse{Connectors: connectors}, nil
}

func (a *apiServer) DeleteIDPConnector(ctx context.Context, req *identity.DeleteIDPConnectorRequest) (resp *identity.DeleteIDPConnectorResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	if err := a.api.deleteConnector(req.Id); err != nil {
		return nil, err
	}

	return &identity.DeleteIDPConnectorResponse{}, nil
}

func (a *apiServer) CreateOIDCClient(ctx context.Context, req *identity.CreateOIDCClientRequest) (resp *identity.CreateOIDCClientResponse, retErr error) {
	removeSecret := func(r *identity.CreateOIDCClientRequest) *identity.CreateOIDCClientRequest {
		copyReq := proto.Clone(r).(*identity.CreateOIDCClientRequest)
		copyReq.Client.Secret = ""
		return copyReq
	}

	a.LogReq(removeSecret(req))
	defer func(start time.Time) { a.LogResp(removeSecret(req), nil, retErr, time.Since(start)) }(time.Now())

	client, err := a.api.createClient(ctx, req)
	if err != nil {
		return nil, err
	}

	return &identity.CreateOIDCClientResponse{
		Client: client,
	}, nil
}

func (a *apiServer) UpdateOIDCClient(ctx context.Context, req *identity.UpdateOIDCClientRequest) (resp *identity.UpdateOIDCClientResponse, retErr error) {
	removeSecret := func(r *identity.UpdateOIDCClientRequest) *identity.UpdateOIDCClientRequest {
		copyReq := proto.Clone(r).(*identity.UpdateOIDCClientRequest)
		copyReq.Client.Secret = ""
		return copyReq
	}
	a.LogReq(removeSecret(req))
	defer func(start time.Time) { a.LogResp(removeSecret(req), nil, retErr, time.Since(start)) }(time.Now())

	if err := a.api.updateClient(ctx, req); err != nil {
		return nil, err
	}

	return &identity.UpdateOIDCClientResponse{}, nil
}

func (a *apiServer) GetOIDCClient(ctx context.Context, req *identity.GetOIDCClientRequest) (resp *identity.GetOIDCClientResponse, retErr error) {
	removeSecret := func(r *identity.GetOIDCClientResponse) *identity.GetOIDCClientResponse {
		copyResp := proto.Clone(r).(*identity.GetOIDCClientResponse)
		copyResp.Client.Secret = ""
		return copyResp
	}
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, removeSecret(resp), retErr, time.Since(start)) }(time.Now())

	client, err := a.api.getClient(req.Id)
	if err != nil {
		return nil, err
	}

	return &identity.GetOIDCClientResponse{Client: client}, nil
}

func (a *apiServer) ListOIDCClients(ctx context.Context, req *identity.ListOIDCClientsRequest) (resp *identity.ListOIDCClientsResponse, retErr error) {
	removeSecret := func(r *identity.ListOIDCClientsResponse) *identity.ListOIDCClientsResponse {
		copyResp := proto.Clone(r).(*identity.ListOIDCClientsResponse)
		for _, c := range copyResp.Clients {
			c.Secret = ""
		}
		return copyResp
	}
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, removeSecret(resp), retErr, time.Since(start)) }(time.Now())

	clients, err := a.api.listClients()
	if err != nil {
		return nil, err
	}

	return &identity.ListOIDCClientsResponse{Clients: clients}, nil
}

func (a *apiServer) DeleteOIDCClient(ctx context.Context, req *identity.DeleteOIDCClientRequest) (resp *identity.DeleteOIDCClientResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	if err := a.api.deleteClient(ctx, req.Id); err != nil {
		return nil, err
	}

	return &identity.DeleteOIDCClientResponse{}, nil
}

func (a *apiServer) DeleteAll(ctx context.Context, req *identity.DeleteAllRequest) (resp *identity.DeleteAllResponse, retErr error) {
	a.LogReq(req)
	defer func(start time.Time) { a.LogResp(req, resp, retErr, time.Since(start)) }(time.Now())

	clients, err := a.api.listClients()
	if err != nil {
		return nil, err
	}

	connectors, err := a.api.listConnectors()
	if err != nil {
		return nil, err
	}

	for _, client := range clients {
		if err := a.api.deleteClient(ctx, client.Id); err != nil {
			return nil, err
		}
	}

	for _, conn := range connectors {
		if err := a.api.deleteConnector(conn.Id); err != nil {
			return nil, err
		}
	}

	if _, err := a.env.GetDBClient().ExecContext(ctx, `DELETE FROM identity.config`); err != nil {
		return nil, err
	}

	return &identity.DeleteAllResponse{}, nil
}
