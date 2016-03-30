package testing

import (
	"errors"
	"fmt"
	"os"
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
	address, err := getTestRethinkAddress()
	if err != nil {
		return nil, err
	}
	databaseName := uuid.NewWithoutDashes()
	if err := server.InitDBs(address, databaseName); err != nil {
		return nil, err
	}
	return server.NewRethinkAPIServer(address, databaseName)
}

func getTestRethinkAddress() (string, error) {
	rethinkAddr := os.Getenv("RETHINK_PORT_28015_TCP_ADDR")
	if rethinkAddr == "" {
		return "", errors.New("RETHINK_PORT_28015_TCP_ADDR not set")
	}
	return fmt.Sprintf("%s:28015", rethinkAddr), nil
}
