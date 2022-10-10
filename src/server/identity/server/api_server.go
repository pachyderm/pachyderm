package server

import (
	"bytes"
	"context"
	"net/http"

	"github.com/ghodss/yaml"
	"github.com/gogo/protobuf/jsonpb"

	"github.com/pachyderm/pachyderm/v2/src/identity"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/server/identityutil"

	logrus "github.com/sirupsen/logrus"
)

const (
	dexHTTPPort = ":1658"

	configKey = 1
)

type apiServer struct {
	env Env
	api *dexAPI
}

type IdentityServerOption func(web *dexWeb)

func UnitTestOption(addr, assetsDir string) IdentityServerOption {
	return func(web *dexWeb) {
		web.addr = addr
		web.assetsDir = assetsDir
	}
}

// NewIdentityServer returns an implementation of identity.APIServer.
func NewIdentityServer(env Env, public bool, options ...IdentityServerOption) *apiServer {
	server := &apiServer{
		env: env,
		api: newDexAPI(env.DexStorage),
	}

	if public {
		web := newDexWeb(env, server, options...)
		go func() {
			if err := http.ListenAndServe(web.addr, web); err != nil {
				logrus.WithError(err).Fatalf("error setting up and/or running the identity server")
			}
		}()
	}

	return server
}

func (a *apiServer) EnvBootstrap(ctx context.Context) error {
	if a.env.Config.IdentityConfig != "" {
		a.env.Logger.Info("Started to configure identity server config via environment")
		var config identity.IdentityServerConfig
		if err := yaml.Unmarshal([]byte(a.env.Config.IdentityConfig), &config); err != nil {
			return errors.Wrapf(err, "unmarshal IdentityConfig with value: %q", a.env.Config.IdentityConfig)
		}
		if _, err := a.setIdentityServerConfig(ctx, &identity.SetIdentityServerConfigRequest{Config: &config}); err != nil {
			return errors.Wrapf(err, "failed to set identity server config")
		}
		a.env.Logger.Info("Successfully configured identity server config via environment")
	}
	if a.env.Config.IdentityConnectors != "" {
		a.env.Logger.Info("Started to configure identity connectors via environment")
		var connectors identityutil.IDPConnectors
		// this is a no-op if the config.IdentityConnectors is already json.
		jsonBytes, err := yaml.YAMLToJSON([]byte(a.env.Config.IdentityConnectors))
		if err != nil {
			return errors.Wrapf(err, "convert IdentityConnectors from YAML to JSON: %q", a.env.Config.IdentityConfig)
		}
		err = jsonpb.Unmarshal(bytes.NewReader(jsonBytes), &connectors)
		if err != nil {
			return errors.Wrapf(err, "unmarshal IdentityConnectors: %q", a.env.Config.IdentityConfig)
		}
		existing, err := a.ListIDPConnectors(ctx, &identity.ListIDPConnectorsRequest{})
		if err != nil {
			return errors.Wrap(err, "failed to retrieve list of IDP connectors.")
		}
		oldCons := make(map[string]*identity.IDPConnector)
		for _, e := range existing.Connectors {
			oldCons[e.Id] = e
		}
		for _, c := range connectors {
			if oc, ok := oldCons[c.Id]; ok {
				c.ConfigVersion = oc.ConfigVersion + 1
				if _, err := a.updateIDPConnector(ctx, &identity.UpdateIDPConnectorRequest{Connector: &c}); err != nil {
					return errors.Wrapf(err, "update connector with ID: %q", c.Id)
				}
				delete(oldCons, c.Id)
			} else {
				if _, err := a.createIDPConnector(ctx, &identity.CreateIDPConnectorRequest{Connector: &c}); err != nil {
					return errors.Wrapf(err, "create connector with ID: %q", c.Id)
				}
			}
		}
		for e := range oldCons {
			if _, err = a.deleteIDPConnector(ctx, &identity.DeleteIDPConnectorRequest{Id: e}); err != nil {
				a.env.Logger.Errorf(errors.Wrapf(err, "delete connector %q", e).Error())
			}
		}
		a.env.Logger.Info("Successfully configured identity connectors via environment")
	}
	return nil
}

func (a *apiServer) SetIdentityServerConfig(ctx context.Context, req *identity.SetIdentityServerConfigRequest) (resp *identity.SetIdentityServerConfigResponse, retErr error) {
	if a.env.Config.IdentityConfig != "" {
		return nil, errors.New("identity.SetIdentityServerConfig() is disable when the Identity Config is set via environment.")
	}
	return a.setIdentityServerConfig(ctx, req)
}

func (a *apiServer) setIdentityServerConfig(ctx context.Context, req *identity.SetIdentityServerConfigRequest) (resp *identity.SetIdentityServerConfigResponse, retErr error) {
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

func (a *apiServer) CreateIDPConnector(ctx context.Context, req *identity.CreateIDPConnectorRequest) (*identity.CreateIDPConnectorResponse, error) {
	if a.env.Config.IdentityConnectors != "" {
		return nil, errors.New("identity.CreateIDPConnector() is disabled when the Identity Connectors are set via environment.")
	}
	return a.createIDPConnector(ctx, req)
}

func (a *apiServer) createIDPConnector(ctx context.Context, req *identity.CreateIDPConnectorRequest) (resp *identity.CreateIDPConnectorResponse, retErr error) {
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

func (a *apiServer) UpdateIDPConnector(ctx context.Context, req *identity.UpdateIDPConnectorRequest) (*identity.UpdateIDPConnectorResponse, error) {
	if a.env.Config.IdentityConnectors != "" {
		return nil, errors.New("identity.UpdateIDPConnector() is disabled when the Identity Connectors are set via environment.")
	}
	return a.updateIDPConnector(ctx, req)
}

func (a *apiServer) updateIDPConnector(ctx context.Context, req *identity.UpdateIDPConnectorRequest) (resp *identity.UpdateIDPConnectorResponse, retErr error) {
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

func (a *apiServer) DeleteIDPConnector(ctx context.Context, req *identity.DeleteIDPConnectorRequest) (*identity.DeleteIDPConnectorResponse, error) {
	if a.env.Config.IdentityConnectors != "" {
		return nil, errors.New("identity.DeleteIDPConnector() is disabled when the Identity Connectors are set via environment.")
	}
	return a.deleteIDPConnector(ctx, req)
}

func (a *apiServer) deleteIDPConnector(ctx context.Context, req *identity.DeleteIDPConnectorRequest) (resp *identity.DeleteIDPConnectorResponse, retErr error) {
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
