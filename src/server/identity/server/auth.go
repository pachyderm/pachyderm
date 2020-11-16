package server

import (
	"context"
	"fmt"
	"path"
	"strconv"

	"github.com/dexidp/dex/api/v2"
	"github.com/dexidp/dex/server"
	"github.com/dexidp/dex/storage"
	"github.com/dexidp/dex/storage/etcd"
)

const (
	etcdNamespace = "/enterprise"
)

type dexServer struct {
	dexStorage storage.Storage
	dexAPI     api.DexServer
}

func newDexServer(etcdEndpoints, etcdPrefix, issuer string, public bool) (*dexServer, error) {
	storageConfig := &etcd.Etcd{
		Endpoints: []string{etcdEndpoints},
		Namespace: path.Join(etcdPrefix, etcdNamespace),
	}

	storage, err := storageConfig.Open(nil)
	if err != nil {
		return nil, err
	}

	if public {
		serverConfig := server.Config{
			Storage:            storage,
			Issuer:             issuer,
			SkipApprovalScreen: true,
		}

		_, err := server.NewServer(context.Background(), serverConfig)
		if err != nil {
			return nil, err
		}
	}

	return &dexServer{
		dexStorage: &proxyStorage{storage},
		dexAPI:     server.NewAPI(storage, nil),
	}, nil
}

func (s *dexServer) addConnector(id, connType, name string, resourceVersion int, jsonConfig []byte) error {
	conn := storage.Connector{
		ID:              id,
		Type:            connType,
		Name:            name,
		ResourceVersion: strconv.Itoa(resourceVersion),
		Config:          jsonConfig,
	}

	if err := s.dexStorage.CreateConnector(conn); err != nil {
		return err
	}

	return nil
}

func (s *dexServer) updateConnector(id, name string, resourceVersion int, jsonConfig []byte) error {
	return s.dexStorage.UpdateConnector(id, func(s storage.Connector) (storage.Connector, error) {
		oldVersion, _ := strconv.Atoi(s.ResourceVersion)
		if oldVersion > resourceVersion {
			return storage.Connector{}, fmt.Errorf("existing resource version is %v, new version is %v", oldVersion, resourceVersion)
		}

		return storage.Connector{
			ID:              id,
			Name:            name,
			ResourceVersion: strconv.Itoa(resourceVersion),
			Config:          jsonConfig,
		}, nil
	})
}

func (s *dexServer) deleteConnector(id string) error {
	return s.dexStorage.DeleteConnector(id)
}

func (s *dexServer) listConnectors() (connectors []storage.Connector, err error) {
	return s.dexStorage.ListConnectors()
}
