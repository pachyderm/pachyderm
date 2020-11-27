package server

import (
	"sync"

	dex_storage "github.com/dexidp/dex/storage"
	dex_sql "github.com/dexidp/dex/storage/sql"

	logrus "github.com/sirupsen/logrus"
)

// StorageProvider is an interface to let us lazy load dex Storage implementations
type StorageProvider interface {
	GetStorage(logger *logrus.Entry) (dex_storage.Storage, error)
}

// LazyPostgresStorage creates a postgres connection when one is requested
type LazyPostgresStorage struct {
	sync.RWMutex

	storageConfig *dex_sql.Postgres
	storage       dex_storage.Storage
}

// NewLazyPostgresStorage returns a new StorageProvider that lazily connects to Postgres
func NewLazyPostgresStorage(pgHost, pgDatabase, pgUser, pgPwd, pgSSL string, pgPort int) *LazyPostgresStorage {
	return &LazyPostgresStorage{
		storageConfig: &dex_sql.Postgres{
			NetworkDB: dex_sql.NetworkDB{
				Database: pgDatabase,
				User:     pgUser,
				Password: pgPwd,
				Host:     pgHost,
				Port:     uint16(pgPort),
			},
			SSL: dex_sql.SSL{
				Mode: pgSSL,
			},
		}}
}

// GetStorage returns a dex Storage, creating it if one isn't already cached
func (s *LazyPostgresStorage) GetStorage(logger *logrus.Entry) (dex_storage.Storage, error) {
	s.RLock()

	storage := s.storage
	if storage == nil {
		s.RUnlock()
		s.Lock()
		if s.storage != nil {
			storage = s.storage
			s.Unlock()
			return storage, nil
		}
		storage, err := s.storageConfig.Open(logger)
		if err != nil {
			logger.WithError(err).Error("dex storage failed to start")
			s.Unlock()
			return nil, err
		}
		s.storage = storage
		s.Unlock()
		return storage, nil
	}

	s.RUnlock()
	return storage, nil
}
