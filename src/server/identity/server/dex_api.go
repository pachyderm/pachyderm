package server

import (
	"context"
	"encoding/json"
	"strconv"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/pachyderm/pachyderm/v2/src/identity"
	"github.com/pachyderm/pachyderm/v2/src/internal/errors"
	"github.com/pachyderm/pachyderm/v2/src/internal/log"
	"github.com/pachyderm/pachyderm/v2/src/internal/pctx"
	"github.com/pachyderm/pachyderm/v2/src/server/identityutil"

	dex_api "github.com/dexidp/dex/api/v2"
	dex_log "github.com/dexidp/dex/pkg/log"
	dex_server "github.com/dexidp/dex/server"
	dex_storage "github.com/dexidp/dex/storage"
)

// dexAPI wraps an api.DexServer and extends it with CRUD operations
// that are currently missing from the dex api. It also handles lazily
// instantiating the api and storage on the first request.
type dexAPI struct {
	api     dex_api.DexServer
	storage dex_storage.Storage
	logger  dex_log.Logger
}

func newDexAPI(sp dex_storage.Storage) *dexAPI {
	ctx := pctx.Background("dexAPI")
	logger := log.NewLogrus(ctx)
	return &dexAPI{
		api:     dex_server.NewAPI(sp, logger, ""),
		storage: sp,
		logger:  logger,
	}
}

func (a *dexAPI) createClient(ctx context.Context, in *identity.CreateOIDCClientRequest) (*identity.OIDCClient, error) {
	if in.Client.Name == "" {
		return nil, errors.New("no client name specified")
	}

	if in.Client.Id == "" {
		return nil, errors.New("no client id specified")
	}

	req := &dex_api.CreateClientReq{
		Client: &dex_api.Client{
			Id:           in.Client.Id,
			Secret:       in.Client.Secret,
			Name:         in.Client.Name,
			RedirectUris: in.Client.RedirectUris,
			TrustedPeers: in.Client.TrustedPeers,
		},
	}

	resp, err := a.api.CreateClient(ctx, req)
	if err != nil {
		return nil, errors.EnsureStack(err)
	}

	if resp.AlreadyExists {
		return nil, identity.ErrAlreadyExists
	}

	return dexClientToPach(resp.Client), nil
}

func (a *dexAPI) updateClient(ctx context.Context, in *identity.UpdateOIDCClientRequest) error {
	req := &dex_api.UpdateClientReq{
		Id:           in.Client.Id,
		Name:         in.Client.Name,
		RedirectUris: in.Client.RedirectUris,
		TrustedPeers: in.Client.TrustedPeers,
	}

	resp, err := a.api.UpdateClient(ctx, req)
	if err != nil {
		return errors.EnsureStack(err)
	}

	if resp.NotFound {
		return errors.Errorf("unable to find OIDC client with id %q", req.Id)
	}

	return nil
}

func (a *dexAPI) deleteClient(ctx context.Context, id string) error {
	resp, err := a.api.DeleteClient(ctx, &dex_api.DeleteClientReq{Id: id})
	if err != nil {
		return errors.EnsureStack(err)
	}

	if resp.NotFound {
		return errors.Errorf("unable to find OIDC client with id %q", id)
	}

	return nil
}

func (a *dexAPI) createConnector(req *identity.CreateIDPConnectorRequest) error {
	if req.Connector.Id == "" {
		return errors.New("no id specified")
	}

	if req.Connector.Type == "" {
		return errors.New("no type specified")
	}

	if req.Connector.Name == "" {
		return errors.New("no name specified")
	}

	// dexidp doesn't support unmarshalling from yaml, so we have to convert to json.
	// If config is already json, then this is a no-op under the hood.
	config, err := identityutil.PickConfig(req.Connector.Config, req.Connector.JsonConfig)
	if err != nil {
		return errors.EnsureStack(err)
	}

	if err := a.validateConnector(req.Connector.Id, req.Connector.Type, config); err != nil {
		return err
	}

	conn := dex_storage.Connector{
		ID:              req.Connector.Id,
		Type:            req.Connector.Type,
		Name:            req.Connector.Name,
		ResourceVersion: strconv.Itoa(int(req.Connector.ConfigVersion)),
		Config:          config,
	}

	if err := a.storage.CreateConnector(conn); err != nil {
		if errors.Is(err, dex_storage.ErrAlreadyExists) {
			return identity.ErrAlreadyExists
		}
		return errors.EnsureStack(err)
	}

	return nil
}

func (a *dexAPI) getConnector(id string) (*identity.IDPConnector, error) {
	conn, err := a.storage.GetConnector(id)
	if err != nil {
		if errors.Is(err, dex_storage.ErrNotFound) {
			return nil, identity.ErrInvalidID
		}
		return nil, errors.EnsureStack(err)
	}

	return dexConnectorToPach(conn)
}

func (a *dexAPI) updateConnector(in *identity.UpdateIDPConnectorRequest) error {
	err := a.storage.UpdateConnector(in.Connector.Id, func(c dex_storage.Connector) (dex_storage.Connector, error) {
		oldVersion, _ := strconv.Atoi(c.ResourceVersion)
		if oldVersion+1 != int(in.Connector.ConfigVersion) {
			return dex_storage.Connector{}, errors.Errorf("new config version is %v, expected %v", in.Connector.ConfigVersion, oldVersion+1)
		}

		c.ResourceVersion = strconv.Itoa(int(in.Connector.ConfigVersion))

		if in.Connector.Name != "" {
			c.Name = in.Connector.Name
		}

		if in.Connector.Type != "" {
			c.Type = in.Connector.Type
		}

		config, err := identityutil.PickConfig(in.Connector.Config, in.Connector.JsonConfig)
		if err != nil && err.Error() != identityutil.NoConfigErr {
			return dex_storage.Connector{}, errors.EnsureStack(err)
		}
		// override the current config only if a config is defined.
		if err == nil {
			c.Config = config
		}
		if err := a.validateConnector(c.ID, c.Type, c.Config); err != nil {
			return dex_storage.Connector{}, err
		}
		return c, nil
	})
	return errors.EnsureStack(err)
}

func (a *dexAPI) deleteConnector(id string) error {
	return errors.EnsureStack(a.storage.DeleteConnector(id))
}

func (a *dexAPI) listConnectors() ([]*identity.IDPConnector, error) {
	dexConnectors, err := a.storage.ListConnectors()
	if err != nil {
		return nil, errors.EnsureStack(err)
	}

	connectors := make([]*identity.IDPConnector, len(dexConnectors))
	for i, c := range dexConnectors {
		conn, err := dexConnectorToPach(c)
		if err != nil {
			return nil, errors.EnsureStack(err)
		}
		connectors[i] = conn
	}
	return connectors, nil
}

func (a *dexAPI) listClients() ([]*identity.OIDCClient, error) {
	storageClients, err := a.storage.ListClients()
	if err != nil {
		return nil, errors.EnsureStack(err)
	}

	clients := make([]*identity.OIDCClient, len(storageClients))
	for i, c := range storageClients {
		clients[i] = storageClientToPach(c)
	}
	return clients, nil
}

func (a *dexAPI) getClient(id string) (*identity.OIDCClient, error) {
	client, err := a.storage.GetClient(id)
	if err != nil {
		if errors.Is(err, dex_storage.ErrNotFound) {
			return nil, identity.ErrInvalidID
		}
		return nil, errors.EnsureStack(err)
	}
	return storageClientToPach(client), nil
}

func (a *dexAPI) validateConnector(id, connType string, jsonConfig []byte) error {
	typeConf, ok := dex_server.ConnectorsConfig[connType]
	if !ok {
		return errors.Errorf("unknown connector type %q", connType)
	}
	conf := typeConf()
	if err := json.Unmarshal(jsonConfig, conf); err != nil {
		return errors.Errorf("unable to deserialize JSON: %v", err)
	}

	if _, err := conf.Open(id, a.logger); err != nil {
		return errors.Errorf("unable to open connector: %v", err)
	}

	return nil
}

func storageClientToPach(c dex_storage.Client) *identity.OIDCClient {
	return &identity.OIDCClient{
		Id:           c.ID,
		Secret:       c.Secret,
		RedirectUris: c.RedirectURIs,
		TrustedPeers: c.TrustedPeers,
		Name:         c.Name,
	}
}

func dexConnectorToPach(c dex_storage.Connector) (*identity.IDPConnector, error) {
	config := &structpb.Struct{}
	if c.Config != nil && len(c.Config) != 0 {
		if err := protojson.Unmarshal(c.Config, config); err != nil {
			return nil, errors.Wrap(err, "unmarshal dex connector into a structpb.Struct")
		}
	}
	// If the version isn't an int, set it to zero
	version, _ := strconv.Atoi(c.ResourceVersion)
	return &identity.IDPConnector{
		Id:            c.ID,
		Name:          c.Name,
		Type:          c.Type,
		ConfigVersion: int64(version),
		Config:        config,
	}, nil
}

func dexClientToPach(c *dex_api.Client) *identity.OIDCClient {
	return &identity.OIDCClient{
		Id:           c.Id,
		Secret:       c.Secret,
		RedirectUris: c.RedirectUris,
		TrustedPeers: c.TrustedPeers,
		Name:         c.Name,
	}
}
