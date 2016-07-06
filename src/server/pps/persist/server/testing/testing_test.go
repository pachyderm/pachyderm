package testing

import (
	"testing"

	"github.com/pachyderm/pachyderm/src/client/pkg/require"
	"github.com/pachyderm/pachyderm/src/client/pkg/uuid"
	"github.com/pachyderm/pachyderm/src/server/pps/persist"
	"github.com/pachyderm/pachyderm/src/server/pps/persist/server"
)

func RunTestWithRethinkAPIServer(t *testing.T, testFunc func(t *testing.T, persistAPIServer persist.APIServer)) {
	if testing.Short() {
		t.Skip("Skipping test because of short mode.")
	}

	apiServer, err := NewTestRethinkAPIServer()
	require.NoError(t, err)
	defer func() {
		require.NoError(t, apiServer.Close())
	}()
	testFunc(t, apiServer)
}

func NewTestRethinkAPIServer() (server.APIServer, error) {
	address := "0.0.0.0:28015"
	databaseName := uuid.NewWithoutDashes()
	if err := server.InitDBs(address, databaseName); err != nil {
		return nil, err
	}
	return server.NewRethinkAPIServer(address, databaseName)
}
