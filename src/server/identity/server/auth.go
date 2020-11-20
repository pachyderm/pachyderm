package server

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/dexidp/dex/api/v2"
	"github.com/dexidp/dex/server"
	"github.com/dexidp/dex/storage"
	"github.com/dexidp/dex/storage/sql"
	logrus "github.com/sirupsen/logrus"
)

const (
	dexHttpPort = ":658"
)

type dexServer struct {
	api.DexServer

	dexStorage storage.Storage
}

func newDexServer(pgHost, pgDatabase, pgUser, pgPwd, pgSSL, issuer string, pgPort int, public bool) (*dexServer, error) {
	storageConfig := &sql.Postgres{
		NetworkDB: sql.NetworkDB{
			Database: pgDatabase,
			User:     pgUser,
			Password: pgPwd,
			Host:     pgHost,
			Port:     uint16(pgPort),
		},
		SSL: sql.SSL{
			Mode: pgSSL,
		},
	}
	logger := logrus.NewEntry(logrus.New()).WithField("source", "dex")
	storage, err := storageConfig.Open(logger)
	if err != nil {
		return nil, err
	}

	if public {
		serverConfig := server.Config{
			Storage:            storage,
			Issuer:             issuer,
			SkipApprovalScreen: true,
			Web: server.WebConfig{
				Dir: "/web",
			},
			Logger: logger,
		}

		serv, err := server.NewServer(context.Background(), serverConfig)
		if err != nil {
			return nil, err
		}

		go func() {
			if err = http.ListenAndServe(dexHttpPort, serv); err != nil {
				logger.WithError(err).Error("Dex server stopped")
			}
		}()
	}

	return &dexServer{
		dexStorage: &proxyStorage{storage},
		DexServer:  server.NewAPI(storage, nil),
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
