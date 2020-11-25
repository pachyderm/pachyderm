package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"sync"

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

	initServer    sync.Once
	storageConfig *sql.Postgres
	issuer        string
	public        bool

	logger     *logrus.Entry
	dexStorage storage.Storage
	dexServer  *server.Server
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

	s := &dexServer{
		storageConfig: storageConfig,
		issuer:        issuer,
		public:        public,
		logger:        logger,
	}

	if public {
		go func() {
			if err := http.ListenAndServe(dexHttpPort, s); err != nil {
				logger.WithError(err).Error("Dex server stopped")
			}
		}()
	}

	return s, nil
}

func (s *dexServer) lazyStart() {
	s.initServer.Do(func() {
		var err error
		s.dexStorage, err = s.storageConfig.Open(s.logger)
		if err != nil {
			s.logger.WithError(err).Error("dex storage failed to start")
			return
		}
		s.DexServer = server.NewAPI(s.dexStorage, nil)

		if s.public {
			serverConfig := server.Config{
				Storage:            s.dexStorage,
				Issuer:             s.issuer,
				SkipApprovalScreen: true,
				Web: server.WebConfig{
					Dir: "/web",
				},
				Logger: s.logger,
			}

			dexServer, err := server.NewServer(context.Background(), serverConfig)
			if err != nil {
				s.logger.WithError(err).Error("dex server failed to start")
				return
			}

			s.dexServer = dexServer

		}
	})
}

func (s *dexServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.lazyStart()
	if s.dexServer == nil {
		http.Error(w, "unable to start Dex server, check logs", http.StatusInternalServerError)
		return
	}

	s.dexServer.ServeHTTP(w, r)
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

func (s *dexServer) createConnector(id, name, connType string, resourceVersion int, jsonConfig []byte) error {
	s.lazyStart()
	if s.dexStorage == nil {
		return fmt.Errorf("unable to start Dex server, check logs")
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
	s.lazyStart()
	if s.dexStorage == nil {
		return storage.Connector{}, fmt.Errorf("unable to start Dex server, check logs")
	}

	return s.dexStorage.GetConnector(id)
}

func (s *dexServer) updateConnector(id, name string, resourceVersion int, jsonConfig []byte) error {
	s.lazyStart()
	if s.dexStorage == nil {
		return fmt.Errorf("unable to start Dex server, check logs")
	}

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
	s.lazyStart()
	if s.dexStorage == nil {
		return fmt.Errorf("unable to start Dex server, check logs")
	}

	return s.dexStorage.DeleteConnector(id)
}

func (s *dexServer) listConnectors() (connectors []storage.Connector, err error) {
	s.lazyStart()
	if s.dexStorage == nil {
		return nil, fmt.Errorf("unable to start Dex server, check logs")
	}

	return s.dexStorage.ListConnectors()
}

func (s *dexServer) listClients() (connectors []storage.Client, err error) {
	s.lazyStart()
	if s.dexStorage == nil {
		return nil, fmt.Errorf("unable to start Dex server, check logs")
	}

	return s.dexStorage.ListClients()
}
