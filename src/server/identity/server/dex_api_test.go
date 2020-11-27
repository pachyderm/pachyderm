package server

import (
	"errors"
	"testing"

	"github.com/pachyderm/pachyderm/src/client/identity"
	"github.com/pachyderm/pachyderm/src/client/pkg/require"

	dex_memory "github.com/dexidp/dex/storage/memory"
	logrus "github.com/sirupsen/logrus"
)

// TestLazyStartAPI tests that the API server tries to connect to the database on each request
func TestLazyStartAPI(t *testing.T) {
	logger := logrus.NewEntry(logrus.New())
	sp := &InMemoryStorageProvider{err: errors.New("unable to connect to database")}

	// api fails to connect to the database initially
	api := newDexAPI(sp, logger)
	req := &identity.CreateIDPConnectorRequest{
		Config: &identity.IDPConnector{
			Id:            "id",
			Name:          "name",
			Type:          "github",
			ConfigVersion: 0,
			JsonConfig:    "{}",
		},
	}
	err := api.createConnector(req)
	require.YesError(t, err, "unable to connect to database")

	// once the database is available the API requests succeed
	sp.provider = dex_memory.New(logger)
	sp.err = nil
	err = api.createConnector(req)
	require.NoError(t, err, "unable to connect to database")
}
