package server

import (
	"sync"

	dex_storage "github.com/dexidp/dex/storage"
	dex_sql "github.com/dexidp/dex/storage/sql"

	logrus "github.com/sirupsen/logrus"
)

type StorageProvider interface {
	GetStorage(logger *logrus.Entry) (dex_storage.Storage, error)
}

// lazyPostgresStorage instantiates a postgres connection the first time
// one is requested.
type LazyPostgresStorage struct {
	sync.Once

	storageConfig *dex_sql.Postgres
	storage       dex_storage.Storage
	err           error
}

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

func (s *LazyPostgresStorage) GetStorage(logger *logrus.Entry) (dex_storage.Storage, error) {
	s.Do(func() {
		s.storage, s.err = s.storageConfig.Open(logger)
		if s.err != nil {
			logger.WithError(s.err).Error("dex storage failed to start")
		}
	})

	if s.err != nil {
		return nil, s.err
	}

	return s.storage, nil
}
