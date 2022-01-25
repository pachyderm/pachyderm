package server

import (
	"context"
	"net/http"

	logrus "github.com/sirupsen/logrus"

	"github.com/pachyderm/pachyderm/v2/src/identity"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
)

const (
	dexHTTPPort = ":1658"

	configKey = 1
)

type apiServer struct {
	env Env
	api *dexAPI
}

// NewIdentityServer returns an implementation of identity.APIServer.
func NewIdentityServer(env Env, public bool) identity.APIServer {
	server := &apiServer{
		env: env,
		api: newDexAPI(env.DexStorage),
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
	if _, err := a.env.DB.ExecContext(ctx, `INSERT INTO identity.config (id, issuer, id_token_expiry, rotation_token_expiry) VALUES ($1, $2, $3, $4) ON CONFLICT (id) DO UPDATE SET issuer=$2, id_token_expiry=$3, rotation_token_expiry=$4`, configKey, req.Config.Issuer, req.Config.IdTokenExpiry, req.Config.RotationTokenExpiry); err != nil {
		return nil, errors.EnsureStack(err)
	}

	return &identity.SetIdentityServerConfigResponse{}, nil
}

func (a *apiServer) GetIdentityServerConfig(ctx context.Context, req *identity.GetIdentityServerConfigRequest) (resp *identity.GetIdentityServerConfigResponse, retErr error) {
	var config []*identity.IdentityServerConfig
	err := a.env.DB.SelectContext(ctx, &config, "SELECT issuer, id_token_expiry, rotation_token_expiry FROM identity.config WHERE id=$1;", configKey)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}
	if len(config) == 0 {
		return &identity.GetIdentityServerConfigResponse{Config: &identity.IdentityServerConfig{}}, nil
	}

	return &identity.GetIdentityServerConfigResponse{Config: config[0]}, nil
}

func (a *apiServer) CreateIDPConnector(ctx context.Context, req *identity.CreateIDPConnectorRequest) (resp *identity.CreateIDPConnectorResponse, retErr error) {
	if err := a.api.createConnector(req); err != nil {
		return nil, err
	}
	return &identity.CreateIDPConnectorResponse{}, nil
}

func (a *apiServer) GetIDPConnector(ctx context.Context, req *identity.GetIDPConnectorRequest) (resp *identity.GetIDPConnectorResponse, retErr error) {
	c, err := a.api.getConnector(req.Id)
	if err != nil {
		return nil, err
	}

	return &identity.GetIDPConnectorResponse{
		Connector: c,
	}, nil
}

func (a *apiServer) UpdateIDPConnector(ctx context.Context, req *identity.UpdateIDPConnectorRequest) (resp *identity.UpdateIDPConnectorResponse, retErr error) {
	if err := a.api.updateConnector(req); err != nil {
		return nil, err
	}

	return &identity.UpdateIDPConnectorResponse{}, nil
}

func (a *apiServer) ListIDPConnectors(ctx context.Context, req *identity.ListIDPConnectorsRequest) (resp *identity.ListIDPConnectorsResponse, retErr error) {
	connectors, err := a.api.listConnectors()
	if err != nil {
		return nil, err
	}

	return &identity.ListIDPConnectorsResponse{Connectors: connectors}, nil
}

func (a *apiServer) DeleteIDPConnector(ctx context.Context, req *identity.DeleteIDPConnectorRequest) (resp *identity.DeleteIDPConnectorResponse, retErr error) {
	if err := a.api.deleteConnector(req.Id); err != nil {
		return nil, err
	}

	return &identity.DeleteIDPConnectorResponse{}, nil
}

func (a *apiServer) CreateOIDCClient(ctx context.Context, req *identity.CreateOIDCClientRequest) (resp *identity.CreateOIDCClientResponse, retErr error) {
	client, err := a.api.createClient(ctx, req)
	if err != nil {
		return nil, err
	}

	return &identity.CreateOIDCClientResponse{
		Client: client,
	}, nil
}

func (a *apiServer) UpdateOIDCClient(ctx context.Context, req *identity.UpdateOIDCClientRequest) (resp *identity.UpdateOIDCClientResponse, retErr error) {
	if err := a.api.updateClient(ctx, req); err != nil {
		return nil, err
	}

	return &identity.UpdateOIDCClientResponse{}, nil
}

func (a *apiServer) GetOIDCClient(ctx context.Context, req *identity.GetOIDCClientRequest) (resp *identity.GetOIDCClientResponse, retErr error) {
	client, err := a.api.getClient(req.Id)
	if err != nil {
		return nil, err
	}

	return &identity.GetOIDCClientResponse{Client: client}, nil
}

func (a *apiServer) ListOIDCClients(ctx context.Context, req *identity.ListOIDCClientsRequest) (resp *identity.ListOIDCClientsResponse, retErr error) {
	clients, err := a.api.listClients()
	if err != nil {
		return nil, err
	}

	return &identity.ListOIDCClientsResponse{Clients: clients}, nil
}

func (a *apiServer) DeleteOIDCClient(ctx context.Context, req *identity.DeleteOIDCClientRequest) (resp *identity.DeleteOIDCClientResponse, retErr error) {
	if err := a.api.deleteClient(ctx, req.Id); err != nil {
		return nil, err
	}

	return &identity.DeleteOIDCClientResponse{}, nil
}

func (a *apiServer) DeleteAll(ctx context.Context, req *identity.DeleteAllRequest) (resp *identity.DeleteAllResponse, retErr error) {
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

	if _, err := a.env.DB.ExecContext(ctx, `DELETE FROM identity.config`); err != nil {
		return nil, errors.EnsureStack(err)
	}

	return &identity.DeleteAllResponse{}, nil
}
