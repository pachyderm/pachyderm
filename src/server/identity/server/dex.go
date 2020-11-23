package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/dexidp/dex/api/v2"
	"github.com/dexidp/dex/server"
	"github.com/dexidp/dex/storage"
	"github.com/dexidp/dex/storage/sql"
	"github.com/pkg/errors"
	logrus "github.com/sirupsen/logrus"
)

const (
	dexHttpPort = ":658"
)

type dexServer struct {
	api.DexServer

	logger     *logrus.Entry
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
		dexStorage: storage,
		logger:     logger,
		DexServer:  server.NewAPI(storage, nil),
	}, nil
}

func (s *dexServer) validateConfig(id, connType string, jsonConfig []byte) error {
	typeConf, ok := server.ConnectorsConfig[connType]
	if !ok {
		return fmt.Errorf("unknown connector type %q", connType)
	}

	conf := typeConf()
	if err := json.Unmarshal(jsonConfig, conf); err != nil {
		return fmt.Errorf("unable to deserialize JSON: %w", err)
	}

	if _, err := conf.Open(id, s.logger); err != nil {
		return fmt.Errorf("unable to open connector: %w", err)
	}

	return nil
}

func (s *dexServer) createConnector(id, connType, name string, resourceVersion int, jsonConfig []byte) error {
	if id == "" {
		return errors.New("no id specified")
	}

	if connType == "" {
		return errors.New("no connType specified")
	}

	if name == "" {
		return errors.New("no name specified")
	}

	if err := s.validateConfig(id, connType, jsonConfig); err != nil {
		return err
	}

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

func (s *dexServer) getConnector(id string) (storage.Connector, error) {
	return s.dexStorage.GetConnector(id)
}

func (s *dexServer) updateConnector(id, name string, resourceVersion int, jsonConfig []byte) error {
	return s.dexStorage.UpdateConnector(id, func(c storage.Connector) (storage.Connector, error) {
		oldVersion, _ := strconv.Atoi(c.ResourceVersion)
		if oldVersion != resourceVersion+1 {
			return storage.Connector{}, fmt.Errorf("new config version is %v, expected %v", resourceVersion, oldVersion+1)
		}

		if name == "" {
			name = c.Name
		}

		if string(jsonConfig) == "" {
			jsonConfig = c.Config
		}

		if err := s.validateConfig(id, c.Type, jsonConfig); err != nil {
			return storage.Connector{}, err
		}

		return storage.Connector{
			ID:              id,
			Type:            c.Type,
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
