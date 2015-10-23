package testing // import "go.pachyderm.com/pachyderm/src/pps/persist/server/testing"

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"go.pachyderm.com/pachyderm/src/pkg/require"
	"go.pachyderm.com/pachyderm/src/pps/persist/server"

	"github.com/satori/go.uuid"
)

func RunTestWithRethinkAPIServer(t *testing.T, testFunc func(t *testing.T, persistAPIServer server.APIServer)) {
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
	databaseName := strings.Replace(uuid.NewV4().String(), "-", "", -1)
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
