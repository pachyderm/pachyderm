package server

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"

	dex_api "github.com/dexidp/dex/api/v2"
	dex_server "github.com/dexidp/dex/server"
	dex_storage "github.com/dexidp/dex/storage"
	"github.com/pkg/errors"
	logrus "github.com/sirupsen/logrus"
)

// dexApi wraps an api.DexServer and extends it with CRUD operations
// that are currently missing from the dex api. It also handles lazily
// instantiating the api and storage on the first request.
type dexApi struct {
	init sync.Once

	server          dex_api.DexServer
	logger          *logrus.Entry
	storageProvider StorageProvider
	err             error
}

func newDexApi(sp StorageProvider, logger *logrus.Entry) *dexApi {
	return &dexApi{
		storageProvider: sp,
		logger:          logger,
	}
}

func (a *dexApi) api() (dex_api.DexServer, error) {
	a.init.Do(func() {
		storage, err := a.storageProvider.GetStorage(a.logger)
		if err != nil {
			a.err = err
			return
		}

		a.server = dex_server.NewAPI(storage, nil)
	})

	if a.err != nil {
		return nil, a.err
	}

	return a.server, nil
}

func (a *dexApi) createClient(ctx context.Context, in *dex_api.CreateClientReq) (*dex_api.CreateClientResp, error) {
	api, err := a.api()
	if err != nil {
		return nil, err
	}
	return api.CreateClient(ctx, in)
}

func (a *dexApi) updateClient(ctx context.Context, in *dex_api.UpdateClientReq) (*dex_api.UpdateClientResp, error) {
	api, err := a.api()
	if err != nil {
		return nil, err
	}
	return api.UpdateClient(ctx, in)
}

func (a *dexApi) deleteClient(ctx context.Context, in *dex_api.DeleteClientReq) (*dex_api.DeleteClientResp, error) {
	api, err := a.api()
	if err != nil {
		return nil, err
	}
	return api.DeleteClient(ctx, in)
}

func (a *dexApi) createConnector(id, name, connType string, resourceVersion int, jsonConfig []byte) error {
	storage, err := a.storageProvider.GetStorage(a.logger)
	if err != nil {
		return err
	}

	if id == "" {
		return errors.New("no id specified")
	}

	if connType == "" {
		return errors.New("no connType specified")
	}

	if name == "" {
		return errors.New("no name specified")
	}

	if err := a.validateConfig(id, connType, jsonConfig); err != nil {
		return err
	}

	conn := dex_storage.Connector{
		ID:              id,
		Type:            connType,
		Name:            name,
		ResourceVersion: strconv.Itoa(resourceVersion),
		Config:          jsonConfig,
	}

	if err := storage.CreateConnector(conn); err != nil {
		return err
	}

	return nil
}

func (a *dexApi) getConnector(id string) (dex_storage.Connector, error) {
	storage, err := a.storageProvider.GetStorage(a.logger)
	if err != nil {
		return dex_storage.Connector{}, err
	}

	return storage.GetConnector(id)
}

func (a *dexApi) updateConnector(id, name string, resourceVersion int, jsonConfig []byte) error {
	storage, err := a.storageProvider.GetStorage(a.logger)
	if err != nil {
		return err
	}

	return storage.UpdateConnector(id, func(c dex_storage.Connector) (dex_storage.Connector, error) {
		oldVersion, _ := strconv.Atoi(c.ResourceVersion)
		if oldVersion+1 != resourceVersion {
			return dex_storage.Connector{}, fmt.Errorf("new config version is %v, expected %v", resourceVersion, oldVersion+1)
		}

		if name == "" {
			name = c.Name
		}

		if string(jsonConfig) == "" {
			jsonConfig = c.Config
		}

		if err := a.validateConfig(id, c.Type, jsonConfig); err != nil {
			return dex_storage.Connector{}, err
		}

		return dex_storage.Connector{
			ID:              id,
			Type:            c.Type,
			Name:            name,
			ResourceVersion: strconv.Itoa(resourceVersion),
			Config:          jsonConfig,
		}, nil
	})
}

func (a *dexApi) deleteConnector(id string) error {
	storage, err := a.storageProvider.GetStorage(a.logger)
	if err != nil {
		return err
	}

	return storage.DeleteConnector(id)
}

func (a *dexApi) listConnectors() ([]dex_storage.Connector, error) {
	storage, err := a.storageProvider.GetStorage(a.logger)
	if err != nil {
		return nil, err
	}

	return storage.ListConnectors()
}

func (a *dexApi) listClients() ([]*dex_api.Client, error) {
	storage, err := a.storageProvider.GetStorage(a.logger)
	if err != nil {
		return nil, err
	}

	storageClients, err := storage.ListClients()
	if err != nil {
		return nil, err
	}

	clients := make([]*dex_api.Client, len(storageClients))
	for i, c := range storageClients {
		clients[i] = storageClientToAPIClient(c)
	}
	return clients, nil
}

func (a *dexApi) getClient(id string) (*dex_api.Client, error) {
	storage, err := a.storageProvider.GetStorage(a.logger)
	if err != nil {
		return nil, err
	}

	client, err := storage.GetClient(id)
	if err != nil {
		return nil, err
	}
	return storageClientToAPIClient(client), nil
}

func (a *dexApi) validateConfig(id, connType string, jsonConfig []byte) error {
	typeConf, ok := dex_server.ConnectorsConfig[connType]
	if !ok {
		return fmt.Errorf("unknown connector type %q", connType)
	}

	conf := typeConf()
	if err := json.Unmarshal(jsonConfig, conf); err != nil {
		return fmt.Errorf("unable to deserialize JSON: %w", err)
	}

	if _, err := conf.Open(id, a.logger); err != nil {
		return fmt.Errorf("unable to open connector: %w", err)
	}

	return nil
}

func storageClientToAPIClient(c dex_storage.Client) *dex_api.Client {
	return &dex_api.Client{
		Id:           c.ID,
		Secret:       c.Secret,
		RedirectUris: c.RedirectURIs,
		TrustedPeers: c.TrustedPeers,
		Name:         c.Name,
	}
}
